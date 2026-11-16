import { apiClient } from './client';

/**
 * Customer Addresses API
 */

export async function getUserAddresses() {
  const response = await apiClient.get('/user/addresses', { requiresAuth: true });
  return response?.data?.data || [];
}

export async function addUserAddress(payload) {
  const response = await apiClient.post('/user/addresses', payload, {
    requiresAuth: true,
  });
  return response?.data?.data;
}

export async function updateUserAddress(id, payload) {
  const response = await apiClient.put(`/user/addresses/${id}`, payload, {
    requiresAuth: true,
  });
  return response?.data?.data;
}

export async function deleteUserAddress(id) {
  const response = await apiClient.delete(`/user/addresses/${id}`, {
    requiresAuth: true,
  });
  return response?.data?.data;
}

export async function setDefaultUserAddress(id) {
  const response = await apiClient.patch(`/user/addresses/${id}/default`, undefined, {
    requiresAuth: true,
  });
  return response?.data?.data;
}

/**
 * Coupon Validation API
 */

export async function validateCoupon({ code, order_amount }) {
  const response = await apiClient.post(
    '/user/coupon/validate',
    { code: code.trim().toUpperCase(), order_amount: Number(order_amount) },
    { requiresAuth: true },
  );
  return response?.data?.data;
}

/**
 * Customer Wallet API
 */

export async function getUserWallet() {
  const response = await apiClient.get('/user/wallet', { requiresAuth: true });
  return response?.data?.data;
}

/**
 * Order Placement & Query API
 */

export async function placeOrder({ cart_item_ids, address_id, payment_method, coupon_code }) {
  const payload = {
    cart_item_ids,
    address_id,
    payment_method,
    coupon_code: coupon_code ? coupon_code.trim().toUpperCase() : undefined,
  };
  const response = await apiClient.post('/user/orders', payload, {
    requiresAuth: true,
  });
  return response?.data?.data;
}

export async function getMyOrderById(orderId) {
  const response = await apiClient.get(`/user/orders/${orderId}`, {
    requiresAuth: true,
  });
  return response?.data?.data;
}

export async function cancelOrder(orderId) {
  const response = await apiClient.patch(
    `/user/orders/${orderId}/cancel`,
    undefined,
    { requiresAuth: true },
  );
  return response?.data?.data;
}

export async function getMyOrders(params = {}) {
  const response = await apiClient.get('/user/orders', {
    params,
    requiresAuth: true,
  });
  return response?.data?.data;
}

/**
 * Razorpay Payment Gateway APIs
 */

export async function createRazorpayOrder({ order_id }) {
  const response = await apiClient.post(
    '/user/payments/razorpay/order',
    { order_id },
    { requiresAuth: true },
  );
  return response?.data?.data;
}

export async function verifyRazorpayPayment({
  order_id,
  razorpay_order_id,
  razorpay_payment_id,
  razorpay_signature,
}) {
  const response = await apiClient.post(
    '/user/payments/razorpay/verify',
    {
      order_id,
      razorpay_order_id,
      razorpay_payment_id,
      razorpay_signature,
    },
    { requiresAuth: true },
  );
  return response?.data;
}
