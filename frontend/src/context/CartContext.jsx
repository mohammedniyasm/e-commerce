import { createContext, useContext, useEffect, useMemo, useState } from 'react';
import {
  addCustomerCartItem,
  clearCustomerCart,
  clearGuestCart,
  getCustomerCart,
  getGuestCart,
  mergeGuestCart,
  removeCustomerCartItem,
  saveGuestCart,
  updateCustomerCartItem,
  validateGuestCart,
} from '../api/cartService';
import { useAuth } from './AuthContext';
import { useToast } from './ToastContext';

const CartContext = createContext(null);

export function CartProvider({ children }) {
  const { isAuthenticated, status } = useAuth();
  const { showToast } = useToast();

  const [cart, setCart] = useState({
    id: 0,
    items: [],
    item_count: 0,
    total_quantity: 0,
    available_subtotal: 0,
    has_unavailable_items: false,
  });
  const [isLoading, setIsLoading] = useState(false);

  // Load or merge cart on auth status change
  useEffect(() => {
    let isMounted = true;

    async function syncCart() {
      if (status === 'initializing') return;

      setIsLoading(true);
      try {
        if (isAuthenticated) {
          // Merge guest cart if items exist
          const guestItems = getGuestCart();
          if (guestItems.length > 0) {
            try {
              await mergeGuestCart(guestItems);
              clearGuestCart();
            } catch {
              // Ignore merge failure and continue fetching customer cart
            }
          }
          const customerCart = await getCustomerCart();
          if (isMounted && customerCart) {
            setCart(customerCart);
          }
        } else {
          // Guest Cart: read from localStorage and validate with backend
          const guestItems = getGuestCart();
          if (guestItems.length > 0) {
            try {
              const validationResults = await validateGuestCart(guestItems);
              // Build guest cart response structure
              let totalQty = 0;
              let subtotal = 0;
              let hasUnavailable = false;

              const validatedItems = guestItems.map((item, index) => {
                const validation = validationResults.find(
                  (v) => v.variant_id === item.variant_id,
                );
                const isAvail = validation ? validation.is_available : true;
                const price = validation ? validation.current_price : item.unit_price || 0;
                const stock = validation ? validation.available_stock : 99;

                if (!isAvail) hasUnavailable = true;
                totalQty += item.quantity;
                subtotal += isAvail ? price * item.quantity : 0;

                return {
                  id: item.id || index + 1,
                  variant_id: item.variant_id,
                  product_id: item.product_id,
                  product_name: item.product_name,
                  product_slug: item.product_slug,
                  category_name: item.category_name,
                  brand_name: item.brand_name,
                  sku: item.sku,
                  size: item.size,
                  color: item.color,
                  quantity: item.quantity,
                  unit_price: price,
                  line_total: price * item.quantity,
                  available_stock: stock,
                  is_available: isAvail,
                  availability_reason: validation?.availability_reason || '',
                  image_url: item.image_url,
                };
              });

              if (isMounted) {
                setCart({
                  id: 0,
                  items: validatedItems,
                  item_count: validatedItems.length,
                  total_quantity: totalQty,
                  available_subtotal: subtotal,
                  has_unavailable_items: hasUnavailable,
                });
              }
            } catch {
              // Fallback to unvalidated local guest cart
              let totalQty = 0;
              let subtotal = 0;
              guestItems.forEach((it) => {
                totalQty += it.quantity;
                subtotal += (it.unit_price || 0) * it.quantity;
              });
              if (isMounted) {
                setCart({
                  id: 0,
                  items: guestItems,
                  item_count: guestItems.length,
                  total_quantity: totalQty,
                  available_subtotal: subtotal,
                  has_unavailable_items: false,
                });
              }
            }
          } else {
            if (isMounted) {
              setCart({
                id: 0,
                items: [],
                item_count: 0,
                total_quantity: 0,
                available_subtotal: 0,
                has_unavailable_items: false,
              });
            }
          }
        }
      } catch {
        // Handle gracefully
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    }

    syncCart();

    return () => {
      isMounted = false;
    };
  }, [isAuthenticated, status]);

  // Add Item to Cart
  async function addToCart({ variant_id, quantity = 1, productDetails = {} }) {
    if (!variant_id || quantity <= 0) return;

    setIsLoading(true);
    try {
      if (isAuthenticated) {
        const updated = await addCustomerCartItem({ variant_id, quantity });
        if (updated) {
          setCart(updated);
        }
      } else {
        // Guest Cart
        const guestItems = getGuestCart();
        const existingIdx = guestItems.findIndex((it) => it.variant_id === variant_id);

        if (existingIdx !== -1) {
          guestItems[existingIdx].quantity += quantity;
        } else {
          guestItems.push({
            id: Date.now(),
            variant_id,
            product_id: productDetails.id,
            product_name: productDetails.name,
            product_slug: productDetails.slug,
            category_name: productDetails.category_name,
            brand_name: productDetails.brand_name,
            sku: productDetails.sku,
            size: productDetails.size,
            color: productDetails.color,
            quantity,
            unit_price: productDetails.selling_price || productDetails.min_price || 0,
            image_url: productDetails.image_url,
            is_available: true,
          });
        }
        saveGuestCart(guestItems);

        // Update local cart state
        let totalQty = 0;
        let subtotal = 0;
        guestItems.forEach((it) => {
          totalQty += it.quantity;
          subtotal += (it.unit_price || 0) * it.quantity;
        });
        setCart({
          id: 0,
          items: guestItems,
          item_count: guestItems.length,
          total_quantity: totalQty,
          available_subtotal: subtotal,
          has_unavailable_items: false,
        });
      }

      showToast({
        title: 'Bag Updated',
        message: 'Item successfully added to your shopping bag.',
        type: 'success',
      });
    } catch (err) {
      showToast({
        title: 'Unable to Add Item',
        message: err.message || 'Failed to update shopping bag.',
        type: 'error',
      });
      throw err;
    } finally {
      setIsLoading(false);
    }
  }

  // Update Item Quantity
  async function updateQuantity(itemId, quantity, variantId) {
    if (quantity <= 0) {
      return removeItem(itemId, variantId);
    }

    setIsLoading(true);
    try {
      if (isAuthenticated) {
        const updated = await updateCustomerCartItem(itemId, { quantity });
        if (updated) setCart(updated);
      } else {
        const guestItems = getGuestCart();
        const target = guestItems.find(
          (it) => it.id === itemId || it.variant_id === variantId,
        );
        if (target) {
          target.quantity = quantity;
          saveGuestCart(guestItems);
          let totalQty = 0;
          let subtotal = 0;
          guestItems.forEach((it) => {
            totalQty += it.quantity;
            subtotal += (it.unit_price || 0) * it.quantity;
          });
          setCart({
            id: 0,
            items: guestItems,
            item_count: guestItems.length,
            total_quantity: totalQty,
            available_subtotal: subtotal,
            has_unavailable_items: false,
          });
        }
      }
    } catch (err) {
      showToast({
        title: 'Update Failed',
        message: err.message || 'Unable to update item quantity.',
        type: 'error',
      });
    } finally {
      setIsLoading(false);
    }
  }

  // Remove Item
  async function removeItem(itemId, variantId) {
    setIsLoading(true);
    try {
      if (isAuthenticated) {
        const updated = await removeCustomerCartItem(itemId);
        if (updated) setCart(updated);
      } else {
        const guestItems = getGuestCart().filter(
          (it) => it.id !== itemId && it.variant_id !== variantId,
        );
        saveGuestCart(guestItems);
        let totalQty = 0;
        let subtotal = 0;
        guestItems.forEach((it) => {
          totalQty += it.quantity;
          subtotal += (it.unit_price || 0) * it.quantity;
        });
        setCart({
          id: 0,
          items: guestItems,
          item_count: guestItems.length,
          total_quantity: totalQty,
          available_subtotal: subtotal,
          has_unavailable_items: false,
        });
      }
      showToast({
        title: 'Item Removed',
        message: 'The item was removed from your bag.',
        type: 'info',
      });
    } catch (err) {
      showToast({
        title: 'Removal Failed',
        message: err.message || 'Unable to remove item from bag.',
        type: 'error',
      });
    } finally {
      setIsLoading(false);
    }
  }

  // Clear Cart
  async function clearCart() {
    setIsLoading(true);
    try {
      if (isAuthenticated) {
        await clearCustomerCart();
      } else {
        clearGuestCart();
      }
      setCart({
        id: 0,
        items: [],
        item_count: 0,
        total_quantity: 0,
        available_subtotal: 0,
        has_unavailable_items: false,
      });
    } finally {
      setIsLoading(false);
    }
  }

  // Refresh Cart
  async function refreshCart() {
    if (isAuthenticated) {
      try {
        const customerCart = await getCustomerCart();
        if (customerCart) setCart(customerCart);
      } catch {
        // Ignore
      }
    }
  }

  const value = useMemo(
    () => ({
      cart,
      totalItems: cart?.total_quantity || cart?.items?.reduce((sum, i) => sum + i.quantity, 0) || 0,
      isLoading,
      addToCart,
      updateQuantity,
      removeItem,
      clearCart,
      refreshCart,
    }),
    [cart, isLoading],
  );

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>;
}

export function useCart() {
  const context = useContext(CartContext);
  if (!context) {
    throw new Error('useCart must be used within a CartProvider');
  }
  return context;
}
