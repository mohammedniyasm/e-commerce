import { apiClient } from './client';

const GUEST_WISHLIST_KEY = 'drape_guest_wishlist';

/**
 * Local Guest Wishlist Persistence
 */
export function getGuestWishlist() {
  try {
    const raw = localStorage.getItem(GUEST_WISHLIST_KEY);
    return raw ? JSON.parse(raw) : [];
  } catch {
    return [];
  }
}

export function saveGuestWishlist(items) {
  try {
    localStorage.setItem(GUEST_WISHLIST_KEY, JSON.stringify(items));
  } catch {
    // Ignore
  }
}

export function clearGuestWishlist() {
  try {
    localStorage.removeItem(GUEST_WISHLIST_KEY);
  } catch {
    // Ignore
  }
}

/**
 * Customer Wishlist API Calls
 */

export async function getWishlists() {
  const response = await apiClient.get('/user/wishlist', { requiresAuth: true });
  return response?.data?.data || [];
}

export async function createWishlist({ name }) {
  const response = await apiClient.post(
    '/user/wishlist',
    { name },
    { requiresAuth: true },
  );
  return response?.data?.data;
}

export async function getWishlistDetails(wishlistId) {
  const response = await apiClient.get(`/user/wishlist/${wishlistId}`, {
    requiresAuth: true,
  });
  return response?.data?.data;
}

export async function updateWishlist(wishlistId, { name }) {
  const response = await apiClient.patch(
    `/user/wishlist/${wishlistId}`,
    { name },
    { requiresAuth: true },
  );
  return response?.data?.data;
}

export async function deleteWishlist(wishlistId) {
  const response = await apiClient.delete(`/user/wishlist/${wishlistId}`, {
    requiresAuth: true,
  });
  return response?.data?.data;
}

export async function addProductToWishlist({ wishlist_name, product_id }) {
  const response = await apiClient.post(
    '/user/wishlist/items',
    { wishlist_name, product_id },
    { requiresAuth: true },
  );
  return response?.data?.data;
}

export async function removeWishlistItem(wishlistId, itemId) {
  const response = await apiClient.delete(
    `/user/wishlist/${wishlistId}/items/${itemId}`,
    { requiresAuth: true },
  );
  return response?.data?.data;
}

export async function clearWishlist(wishlistId) {
  const response = await apiClient.delete(`/user/wishlist/${wishlistId}/items`, {
    requiresAuth: true,
  });
  return response?.data?.data;
}

export async function mergeWishlistItems({ wishlist_name, product_ids }) {
  if (!product_ids || product_ids.length === 0) return null;
  const response = await apiClient.post(
    '/user/wishlist/merge',
    { wishlist_name, product_ids },
    { requiresAuth: true },
  );
  return response?.data?.data;
}
