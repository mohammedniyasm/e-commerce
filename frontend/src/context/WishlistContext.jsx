import { createContext, useContext, useEffect, useMemo, useState } from 'react';
import {
  addProductToWishlist,
  clearGuestWishlist,
  createWishlist,
  deleteWishlist,
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
  const [currentWishlist, setCurrentWishlist] = useState({ id: 0, name: 'Default', items: [] });
  const [guestItems, setGuestItems] = useState([]);
  const [isLoading, setIsLoading] = useState(false);

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
              name: 'Guest Wishlist',
              items: local.map((item, index) =>
                typeof item === 'object'
                  ? { id: index + 1, product_id: item.id, ...item }
                  : { id: index + 1, product_id: item, product_name: 'Atelier Piece' },
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
            await removeWishlistItem(currentWishlist.id, item.id);
            setCurrentWishlist((prev) => ({
              ...prev,
              items: prev.items.filter((it) => it.id !== item.id),
            }));
            showToast({
              title: 'Removed from Wishlist',
              message: 'Item removed from your wishlist.',
              type: 'info',
            });
          }
        } else {
          await addProductToWishlist({
            wishlist_name: targetList,
            product_id: product.id,
          });
          if (currentWishlist.id) {
            const updated = await getWishlistDetails(currentWishlist.id);
            if (updated) setCurrentWishlist(updated);
          }
          showToast({
            title: 'Saved to Wishlist',
            message: 'Item saved to your curated wishlist.',
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
          name: 'Guest Wishlist',
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

  // Create new named wishlist
  async function handleCreateWishlist(name) {
    if (!name?.trim()) return;
    setIsLoading(true);
    try {
      const created = await createWishlist({ name: name.trim() });
      const lists = await getWishlists();
      setWishlists(lists);
      if (created?.id) {
        await selectWishlist(created.id);
      }
      showToast({
        title: 'Wishlist Created',
        message: `Created wishlist "${name.trim()}".`,
        type: 'success',
      });
    } finally {
      setIsLoading(false);
    }
  }

  // Remove specific item from wishlist
  async function removeItem(itemId, productId) {
    if (isAuthenticated && currentWishlist.id) {
      setIsLoading(true);
      try {
        await removeWishlistItem(currentWishlist.id, itemId);
        setCurrentWishlist((prev) => ({
          ...prev,
          items: prev.items.filter((it) => it.id !== itemId),
        }));
      } finally {
        setIsLoading(false);
      }
    } else {
      toggleWishlist({ id: productId });
    }
  }

  const value = useMemo(
    () => ({
      wishlists,
      currentWishlist,
      activeWishlistId,
      isInWishlist,
      toggleWishlist,
      selectWishlist,
      createWishlist: handleCreateWishlist,
      removeItem,
      isLoading,
    }),
    [wishlists, currentWishlist, activeWishlistId, guestItems, isLoading],
  );

  return <WishlistContext.Provider value={value}>{children}</WishlistContext.Provider>;
}

export function useWishlist() {
  const context = useContext(WishlistContext);
  if (!context) {
    throw new Error('useWishlist must be used within a WishlistProvider');
  }
  return context;
}
