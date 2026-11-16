import { apiClient } from './client';

// ==========================================
// MODULE A & B: ORDERS & PAYMENTS
// ==========================================

export async function listAdminOrders(params = {}) {
  const queryParams = {
    page: params.page || 1,
    limit: params.limit || 10,
    ...(params.search ? { search: params.search } : {}),
    ...(params.order_status ? { order_status: params.order_status } : {}),
    ...(params.payment_status ? { payment_status: params.payment_status } : {}),
    ...(params.sort ? { sort: params.sort } : {}),
    ...(params.user_id ? { user_id: params.user_id } : {}),
  };
  const response = await apiClient.get('/admin/orders', {
    params: queryParams,
    requiresAuth: true,
  });
  return response?.data?.data ?? { orders: [], total_count: 0, page: 1, limit: 10, total_pages: 0 };
}

export async function getAdminOrderById(orderId) {
  const response = await apiClient.get(`/admin/orders/${orderId}`, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function updateOrderStatus(orderId, status, note = null) {
  const payload = {
    status,
    ...(note ? { note } : {}),
  };
  const response = await apiClient.patch(`/admin/orders/${orderId}/status`, payload, { requiresAuth: true });
  return response?.data ?? null;
}

export async function updateCODPaymentStatus(orderId, status) {
  const response = await apiClient.patch(
    `/admin/orders/${orderId}/payment-status`,
    { status },
    { requiresAuth: true }
  );
  return response?.data ?? null;
}

export async function listRefundPendingOrders(params = {}) {
  const queryParams = {
    page: params.page || 1,
    limit: params.limit || 10,
    ...(params.search ? { search: params.search } : {}),
  };
  const response = await apiClient.get('/admin/orders/refund-pending', {
    params: queryParams,
    requiresAuth: true,
  });
  return response?.data?.data ?? { orders: [], total_count: 0, page: 1, limit: 10, total_pages: 0 };
}

export async function processWalletRefund(orderId) {
  const response = await apiClient.post(`/admin/orders/${orderId}/refund`, undefined, { requiresAuth: true });
  return response?.data ?? null;
}

// ==========================================
// MODULE C: SHIPMENTS
// ==========================================

export async function createShipment(payload) {
  const response = await apiClient.post('/admin/shipments', payload, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function getShipmentById(shipmentId) {
  const response = await apiClient.get(`/admin/shipments/${shipmentId}`, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function getShipmentByOrderId(orderId) {
  const response = await apiClient.get(`/admin/orders/${orderId}/shipment`, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function updateShipment(shipmentId, payload) {
  const response = await apiClient.patch(`/admin/shipments/${shipmentId}`, payload, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function updateShippingStatus(shipmentId, payload) {
  const response = await apiClient.patch(`/admin/shipments/${shipmentId}/status`, payload, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function getShipmentTrackingHistory(shipmentId) {
  const response = await apiClient.get(`/admin/shipments/${shipmentId}/history`, { requiresAuth: true });
  return response?.data?.data ?? [];
}

// ==========================================
// MODULE D: RETURNS & EXCHANGES
// ==========================================

export async function listReturnRequests(params = {}) {
  const queryParams = {
    page: params.page || 1,
    limit: params.limit || 10,
    ...(params.order_id ? { order_id: params.order_id } : {}),
    ...(params.status ? { status: params.status } : {}),
  };
  const response = await apiClient.get('/admin/return-requests', {
    params: queryParams,
    requiresAuth: true,
  });
  return response?.data?.data ?? { items: [], total: 0, page: 1, limit: 10 };
}

export async function getReturnRequestById(id) {
  const response = await apiClient.get(`/admin/return-requests/${id}`, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function updateReturnRequestStatus(id, payload) {
  const response = await apiClient.patch(`/admin/return-requests/${id}/status`, payload, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function createExchangeOrder(returnRequestId) {
  const response = await apiClient.post(`/admin/return-requests/${returnRequestId}/exchange`, undefined, {
    requiresAuth: true,
  });
  return response?.data?.data ?? null;
}

export async function createReturnShipment(returnRequestId) {
  const response = await apiClient.post(
    `/admin/return-requests/${returnRequestId}/return-shipment`,
    undefined,
    { requiresAuth: true }
  );
  return response?.data?.data ?? null;
}

// ==========================================
// MODULE E: COUPONS
// ==========================================

export async function listAdminCoupons(params = {}) {
  const queryParams = {
    page: params.page || 1,
    limit: params.limit || 10,
    ...(params.search ? { search: params.search } : {}),
    ...(params.active_only !== undefined ? { active_only: params.active_only } : {}),
    ...(params.sort ? { sort: params.sort } : {}),
  };
  const response = await apiClient.get('/admin/coupons', {
    params: queryParams,
    requiresAuth: true,
  });
  return response?.data?.data ?? { coupons: [], page: 1, limit: 10, total: 0, total_pages: 0 };
}

export async function getAdminCouponById(id) {
  const response = await apiClient.get(`/admin/coupons/${id}`, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function createAdminCoupon(payload) {
  const response = await apiClient.post('/admin/coupons', payload, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function updateAdminCoupon(id, payload) {
  const response = await apiClient.put(`/admin/coupons/${id}`, payload, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function deleteAdminCoupon(id) {
  const response = await apiClient.delete(`/admin/coupons/${id}`, { requiresAuth: true });
  return response?.data ?? null;
}

export async function toggleAdminCouponActive(id) {
  const response = await apiClient.patch(`/admin/coupons/${id}/toggle`, undefined, { requiresAuth: true });
  return response?.data ?? null;
}

export async function listCouponUsages(couponId, params = {}) {
  const queryParams = {
    page: params.page || 1,
    limit: params.limit || 10,
  };
  const response = await apiClient.get(`/admin/coupons/${couponId}/usages`, {
    params: queryParams,
    requiresAuth: true,
  });
  return response?.data?.data ?? { usages: [], page: 1, limit: 10, total: 0, total_pages: 0 };
}

// ==========================================
// MODULE F: REVIEWS
// ==========================================

export async function listAdminReviews(params = {}) {
  const queryParams = {
    page: params.page || 1,
    limit: params.limit || 10,
  };
  const response = await apiClient.get('/admin/reviews', {
    params: queryParams,
    requiresAuth: true,
  });
  return response?.data?.data ?? { reviews: [], page: 1, limit: 10, total: 0, total_pages: 0 };
}

export async function setReviewVisibility(reviewId, isVisible) {
  const response = await apiClient.patch(
    `/admin/reviews/${reviewId}/visibility`,
    { is_visible: isVisible },
    { requiresAuth: true }
  );
  return response?.data ?? null;
}

// ==========================================
// MODULE G: CUSTOMERS / USERS
// ==========================================

export async function listAdminUsers(params = {}) {
  const queryParams = {
    page: params.page || 1,
    limit: params.limit || 10,
    ...(params.search ? { search: params.search } : {}),
  };
  const response = await apiClient.get('/admin/users', {
    params: queryParams,
    requiresAuth: true,
  });
  return response?.data?.data ?? { users: [], page: 1, limit: 10, total: 0, total_pages: 0 };
}

export async function getAdminUserById(userId) {
  const response = await apiClient.get(`/admin/users/${userId}`, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function addAdminCustomer(payload) {
  const response = await apiClient.post('/admin/users', payload, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function blockAdminUser(userId) {
  const response = await apiClient.patch(`/admin/users/${userId}/block`, undefined, { requiresAuth: true });
  return response?.data ?? null;
}

export async function unblockAdminUser(userId) {
  const response = await apiClient.patch(`/admin/users/${userId}/unblock`, undefined, { requiresAuth: true });
  return response?.data ?? null;
}

export async function updateAdminCustomer(userId, payload) {
  const response = await apiClient.put(`/admin/users/${userId}`, payload, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function deleteAdminCustomer(userId) {
  const response = await apiClient.delete(`/admin/users/${userId}`, { requiresAuth: true });
  return response?.data ?? null;
}

export async function getUserWallet(userId) {
  const response = await apiClient.get(`/admin/wallets/${userId}`, { requiresAuth: true });
  return response?.data?.data ?? null;
}

// ==========================================
// MODULE H: REPORTS & EXPORTS
// ==========================================

export async function getSalesReport(params = {}) {
  const queryParams = {
    group_by: params.group_by || 'day',
    ...(params.start_date ? { start_date: params.start_date } : {}),
    ...(params.end_date ? { end_date: params.end_date } : {}),
  };
  const response = await apiClient.get('/admin/reports/sales', {
    params: queryParams,
    requiresAuth: true,
  });
  return response?.data?.data ?? null;
}

export async function getOrdersReport(params = {}) {
  const queryParams = {
    group_by: params.group_by || 'day',
    ...(params.start_date ? { start_date: params.start_date } : {}),
    ...(params.end_date ? { end_date: params.end_date } : {}),
  };
  const response = await apiClient.get('/admin/reports/orders', {
    params: queryParams,
    requiresAuth: true,
  });
  return response?.data?.data ?? null;
}

export async function getProductsReport(params = {}) {
  const queryParams = {
    ...(params.start_date ? { start_date: params.start_date } : {}),
    ...(params.end_date ? { end_date: params.end_date } : {}),
  };
  const response = await apiClient.get('/admin/reports/products', {
    params: queryParams,
    requiresAuth: true,
  });
  return response?.data?.data ?? null;
}

export async function downloadReportFile(endpointPath, params = {}, filename = 'report') {
  const response = await apiClient.get(endpointPath, {
    params,
    requiresAuth: true,
    responseType: 'blob',
  });

  const blob = new Blob([response.data]);
  const downloadUrl = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = downloadUrl;
  link.setAttribute('download', filename);
  document.body.appendChild(link);
  link.click();
  link.remove();
  window.URL.revokeObjectURL(downloadUrl);
}

// ==========================================
// SETTINGS
// ==========================================

export async function getReferralSettings() {
  const response = await apiClient.get('/admin/referrals/settings', { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function updateReferralSettings(payload) {
  const response = await apiClient.put('/admin/referrals/settings', payload, { requiresAuth: true });
  return response?.data?.data ?? null;
}
