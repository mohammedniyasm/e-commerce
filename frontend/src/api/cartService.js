import { apiClient } from './client';

const GUEST_CART_KEY = 'drape_guest_cart';

/**
 * Local Guest Cart Persistence Helpers
 */
export function getGuestCart() {
  try {
    const raw = localStorage.getItem(GUEST_CART_KEY);
    return raw ? JSON.parse(raw) : [];
  } catch {
    return [];
  }
}

export function saveGuestCart(items) {
  try {
    localStorage.setItem(GUEST_CART_KEY, JSON.stringify(items));
  } catch {
    // Ignore storage quota errors
  }
}

export function clearGuestCart() {
  try {
    localStorage.removeItem(GUEST_CART_KEY);
  } catch {
    // Ignore
  }
}

/**
 * Cart API Calls
 */

export async function getCustomerCart() {
  const response = await apiClient.get('/user/cart', { requiresAuth: true });
  return response?.data?.data;
}

export async function addCustomerCartItem({ variant_id, quantity }) {
  const response = await apiClient.post(
    '/user/cart/items',
    { variant_id, quantity },
    { requiresAuth: true },
  );
  return response?.data?.data;
}

export async function updateCustomerCartItem(itemId, { quantity }) {
  const response = await apiClient.patch(
    `/user/cart/items/${itemId}`,
    { quantity },
    { requiresAuth: true },
  );
  return response?.data?.data;
}

export async function removeCustomerCartItem(itemId) {
  const response = await apiClient.delete(`/user/cart/items/${itemId}`, {
    requiresAuth: true,
  });
  return response?.data?.data;
}

export async function clearCustomerCart() {
  const response = await apiClient.delete('/user/cart', { requiresAuth: true });
  return response?.data?.data;
}

export async function mergeGuestCart(items) {
  if (!items || items.length === 0) return null;
  const payload = {
    items: items.map((item) => ({
      variant_id: item.variant_id,
      quantity: item.quantity,
    })),
  };
  const response = await apiClient.post('/user/cart/merge', payload, {
    requiresAuth: true,
  });
  return response?.data?.data;
}

// Aliases for unified naming
export const getUserCart = getCustomerCart;
export const addToUserCart = addCustomerCartItem;
export const updateCartItemQuantity = updateCustomerCartItem;
export const removeCartItem = removeCustomerCartItem;
export const clearUserCart = clearCustomerCart;
export const mergeCartItems = mergeGuestCart;

export async function validateGuestCart(items) {
  if (!items || items.length === 0) {
    return [];
  }
  const payload = {
    items: items.map((item) => ({
      variant_id: item.variant_id,
      quantity: item.quantity,
    })),
  };
  const response = await apiClient.post('/guest-cart/validate', payload, {
    skipAuth: true,
  });
  return response?.data?.data || [];
}
