import { createContext, useContext, useEffect, useMemo, useState } from 'react';
import {
  addProductToWishlist,
  clearGuestWishlist,
  clearWishlist as clearWishlistApi,
  createWishlist,
  deleteWishlist as deleteWishlistApi,
  getGuestWishlist,
  getWishlistDetails,
  getWishlists,
  mergeWishlistItems,
  removeWishlistItem,
  saveGuestWishlist,
  updateWishlist,
} from '../api/wishlistService';
import { useAuth } from './AuthContext';
import { useToast } from './ToastContext';

const WishlistContext = createContext(null);

export function WishlistProvider({ children }) {
  const { isAuthenticated, status } = useAuth();
  const { showToast } = useToast();

  const [wishlists, setWishlists] = useState([]);
  const [activeWishlistId, setActiveWishlistId] = useState(null);
  const [currentWishlist, setCurrentWishlist] = useState({ id: 0, name: 'My Wishlist', items: [] });
  const [guestItems, setGuestItems] = useState([]);
  const [isLoading, setIsLoading] = useState(false);

  function applyWishlistDetails(details, fallbackItems) {
    const updated = details || { ...currentWishlist, items: fallbackItems };
    setCurrentWishlist(updated);
    setWishlists((prev) =>
      prev.map((wishlist) =>
        wishlist.id === updated.id
          ? { ...wishlist, item_count: updated.items?.length ?? wishlist.item_count }
          : wishlist,
      ),
    );
  }

  // Sync wishlists on auth state change
  useEffect(() => {
    let isMounted = true;

    async function syncWishlist() {
      if (status === 'initializing') return;

      setIsLoading(true);
      try {
        if (isAuthenticated) {
          // Merge guest wishlist if any
          const localGuest = getGuestWishlist();
          let lists = await getWishlists();

          if (!lists || lists.length === 0) {
            // Create initial default wishlist
            const created = await createWishlist({ name: 'My Wishlist' });
            if (created) lists = [created];
          }

          const defaultList = lists[0];
          if (localGuest.length > 0 && defaultList) {
            try {
              await mergeWishlistItems({
                wishlist_name: defaultList.name,
                product_ids: localGuest.map((item) => (typeof item === 'object' ? item.id : item)),
              });
              clearGuestWishlist();
            } catch {
              // Ignore merge conflict
            }
          }

          if (isMounted) {
            setWishlists(lists);
            const targetId = defaultList ? defaultList.id : null;
            setActiveWishlistId(targetId);
            if (targetId) {
              const details = await getWishlistDetails(targetId);
              if (isMounted && details) {
                setCurrentWishlist(details);
              }
            }
          }
        } else {
          // Guest mode
          const local = getGuestWishlist();
          if (isMounted) {
            setGuestItems(local);
            setCurrentWishlist({
              id: 0,
              name: 'My Wishlist',
              items: local.map((item, index) =>
                typeof item === 'object'
                  ? { id: index + 1, product_id: item.id, ...item }
                  : { id: index + 1, product_id: item, product_name: 'Saved product' },
              ),
            });
          }
        }
      } catch {
        // Ignore
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    }

    syncWishlist();

    return () => {
      isMounted = false;
    };
  }, [isAuthenticated, status]);

  // Load specific wishlist details
  async function selectWishlist(id) {
    setActiveWishlistId(id);
    setIsLoading(true);
    try {
      const details = await getWishlistDetails(id);
      if (details) setCurrentWishlist(details);
    } finally {
      setIsLoading(false);
    }
  }

  // Check if a product is in the active wishlist
  function isInWishlist(productId) {
    if (!productId) return false;
    if (isAuthenticated) {
      return Boolean(currentWishlist?.items?.some((it) => it.product_id === productId));
    }
    return guestItems.some((it) => (typeof it === 'object' ? it.id === productId : it === productId));
  }

  // Toggle item in wishlist
  async function toggleWishlist(product) {
    if (!product?.id) return;
    const exists = isInWishlist(product.id);

    setIsLoading(true);
    try {
      if (isAuthenticated) {
        const targetList = currentWishlist?.name || 'My Wishlist';
        if (exists) {
          const item = currentWishlist.items.find((it) => it.product_id === product.id);
          if (item && currentWishlist.id) {
            const fallbackItems = currentWishlist.items.filter((it) => it.id !== item.id);
            const updated = await removeWishlistItem(currentWishlist.id, item.id);
            applyWishlistDetails(updated, fallbackItems);
            showToast({
              title: 'Removed from Wishlist',
              message: 'Item removed from your wishlist.',
              type: 'info',
            });
          }
        } else {
          const fallbackItems = [
            ...currentWishlist.items,
            {
              product_id: product.id,
              product_name: product.name,
              product_slug: product.slug,
              image_url: product.primary_image?.image_url,
              starting_price: product.min_price,
            },
          ];
          const updated = await addProductToWishlist({
            wishlist_name: targetList,
            product_id: product.id,
          });
          applyWishlistDetails(updated, fallbackItems);
          showToast({
            title: 'Saved to Wishlist',
            message: 'Item saved to your wishlist.',
            type: 'success',
          });
        }
      } else {
        // Guest mode
        let nextGuest = [...guestItems];
        if (exists) {
          nextGuest = nextGuest.filter((it) => (typeof it === 'object' ? it.id !== product.id : it !== product.id));
          showToast({
            title: 'Removed from Wishlist',
            message: 'Item removed from your guest wishlist.',
            type: 'info',
          });
        } else {
          nextGuest.push({
            id: product.id,
            product_name: product.name,
            product_slug: product.slug,
            image_url: product.primary_image?.image_url,
            starting_price: product.min_price,
          });
          showToast({
            title: 'Saved to Wishlist',
            message: 'Item saved. Sign in to sync across devices.',
            type: 'success',
          });
        }
        setGuestItems(nextGuest);
        saveGuestWishlist(nextGuest);
        setCurrentWishlist({
          id: 0,
          name: 'My Wishlist',
          items: nextGuest.map((it, idx) => ({ id: idx + 1, product_id: it.id, ...it })),
        });
      }
    } catch (err) {
      showToast({
        title: 'Wishlist Error',
        message: err.message || 'Unable to update wishlist.',
        type: 'error',
      });
    } finally {
      setIsLoading(false);
    }
  }

  async function addToWishlist(product, wishlistId) {
    if (!product?.id || !isAuthenticated) return false;

    const targetWishlist = wishlists.find(
      (wishlist) => String(wishlist.id) === String(wishlistId),
    );
    if (!targetWishlist) {
      showToast({
        title: 'Collection Unavailable',
        message: 'Please refresh your wishlists and try again.',
        type: 'error',
      });
      return false;
    }

    setIsLoading(true);
    try {
      const updated = await addProductToWishlist({
        wishlist_name: targetWishlist.name,
        product_id: product.id,
      });
      if (!updated || !Array.isArray(updated.items)) {
        throw new Error('The wishlist response was invalid.');
      }

      setActiveWishlistId(targetWishlist.id);
      applyWishlistDetails(updated, updated.items);
      showToast({
        title: 'Saved to Wishlist',
        message: `Item saved to "${targetWishlist.name}".`,
        type: 'success',
      });
      return true;
    } catch (err) {
      showToast({
        title: 'Wishlist Error',
        message: err.message || 'Unable to update wishlist.',
        type: 'error',
      });
      return false;
    } finally {
      setIsLoading(false);
    }
  }

  // Create new named wishlist
  // Create new named wishlist
  async function handleCreateWishlist(name) {
    if (!name?.trim()) return;
    setIsLoading(true);
    try {
      const created = await createWishlist({ name: name.trim() });
      const lists = await getWishlists();
      setWishlists(lists || []);
      if (created?.id) {
        await selectWishlist(created.id);
      }
      showToast({
        title: 'Wishlist Created',
        message: `Created wishlist "${name.trim()}".`,
        type: 'success',
      });
      return created;
    } catch (err) {
      showToast({
        title: 'Creation Failed',
        message: err.message || 'Unable to create wishlist.',
        type: 'error',
      });
      throw err;
    } finally {
      setIsLoading(false);
    }
  }

  // Delete an entire wishlist collection
  async function handleDeleteWishlist(wishlistId) {
    const targetId = wishlistId || activeWishlistId || currentWishlist?.id;
    if (!targetId) return;

    setIsLoading(true);
    try {
      if (isAuthenticated) {
        await deleteWishlistApi(targetId);
        const remaining = await getWishlists();
        const safeLists = remaining || [];
        setWishlists(safeLists);

        if (safeLists.length > 0) {
          const nextActive = safeLists[0];
          setActiveWishlistId(nextActive.id);
          const details = await getWishlistDetails(nextActive.id);
          if (details) setCurrentWishlist(details);
        } else {
          // If no wishlists left, create default
          const defaultList = await createWishlist({ name: 'My Wishlist' });
          if (defaultList) {
            setWishlists([defaultList]);
            setActiveWishlistId(defaultList.id);
            setCurrentWishlist(defaultList);
          } else {
            setActiveWishlistId(null);
            setCurrentWishlist({ id: 0, name: 'Default', items: [] });
          }
        }
      } else {
        clearGuestWishlist();
        setGuestItems([]);
        setCurrentWishlist({ id: 0, name: 'Guest Wishlist', items: [] });
      }

      showToast({
        title: 'Wishlist Deleted',
        message: 'The wishlist collection was successfully deleted.',
        type: 'info',
      });
    } catch (err) {
      showToast({
        title: 'Delete Failed',
        message: err.message || 'Unable to delete wishlist.',
        type: 'error',
      });
      throw err;
    } finally {
      setIsLoading(false);
    }
  }

  // Clear all items from a wishlist
  async function handleClearWishlist(wishlistId) {
    const targetId = wishlistId || activeWishlistId || currentWishlist?.id;
    setIsLoading(true);
    try {
      if (isAuthenticated && targetId) {
        const updated = await clearWishlistApi(targetId);
        applyWishlistDetails(updated, []);
      } else {
        clearGuestWishlist();
        setGuestItems([]);
        setCurrentWishlist({ id: 0, name: 'Guest Wishlist', items: [] });
      }

      showToast({
        title: 'Wishlist Cleared',
        message: 'All items removed from this collection.',
        type: 'info',
      });
    } catch (err) {
      showToast({
        title: 'Clear Failed',
        message: err.message || 'Unable to clear wishlist items.',
        type: 'error',
      });
      throw err;
    } finally {
      setIsLoading(false);
    }
  }

  // Remove specific item from wishlist
  async function removeItem(itemId, productId) {
    setIsLoading(true);
    try {
      if (isAuthenticated && currentWishlist?.id) {
        // Resolve target item ID
        let targetItemId = itemId;
        if (!targetItemId || targetItemId <= 0) {
          const matched = currentWishlist.items?.find(
            (it) => it.product_id === productId || it.id === itemId,
          );
          if (matched) targetItemId = matched.id;
        }

        if (targetItemId) {
          const updated = await removeWishlistItem(currentWishlist.id, targetItemId);
          const fallbackItems = (currentWishlist.items || []).filter(
            (it) => it.id !== targetItemId,
          );
          applyWishlistDetails(updated, fallbackItems);
        }

        showToast({
          title: 'Removed from Wishlist',
          message: 'Item removed from your collection.',
          type: 'info',
        });
      } else {
        // Guest mode removal
        const targetId = productId || itemId;
        const nextGuest = getGuestWishlist().filter((it) => {
          if (typeof it === 'object') {
            return it.id !== targetId && it.product_id !== targetId;
          }
          return it !== targetId;
        });

        saveGuestWishlist(nextGuest);
        setGuestItems(nextGuest);
        setCurrentWishlist({
          id: 0,
          name: 'My Wishlist',
          items: nextGuest.map((it, idx) =>
            typeof it === 'object'
              ? { id: idx + 1, product_id: it.id || it.product_id, ...it }
              : { id: idx + 1, product_id: it, product_name: 'Saved product' },
          ),
        });

        showToast({
          title: 'Removed from Wishlist',
          message: 'Item removed from your guest wishlist.',
          type: 'info',
        });
      }
    } catch (err) {
      showToast({
        title: 'Error',
        message: err.message || 'Unable to remove item from wishlist.',
        type: 'error',
      });
    } finally {
      setIsLoading(false);
    }
  }

  const value = useMemo(
    () => ({
      wishlists,
      currentWishlist,
      activeWishlistId,
      isAuthenticated,
      isInWishlist,
      toggleWishlist,
      addToWishlist,
      selectWishlist,
      createWishlist: handleCreateWishlist,
      deleteWishlist: handleDeleteWishlist,
      clearWishlist: handleClearWishlist,
      removeItem,
      isLoading,
    }),
    [wishlists, currentWishlist, activeWishlistId, guestItems, isLoading, isAuthenticated],
  );

  return <WishlistContext.Provider value={value}>{children}</WishlistContext.Provider>;
}

export function useWishlist() {
  const context = useContext(WishlistContext);
  if (!context) {
    return {
      wishlists: [],
      currentWishlist: { id: 0, name: 'My Wishlist', items: [] },
      activeWishlistId: null,
      isAuthenticated: false,
      isInWishlist: () => false,
      toggleWishlist: async () => {},
      addToWishlist: async () => false,
      selectWishlist: async () => {},
      createWishlist: async () => {},
      deleteWishlist: async () => {},
      clearWishlist: async () => {},
      removeItem: async () => {},
      isLoading: false,
    };
  }
  return context;
}
