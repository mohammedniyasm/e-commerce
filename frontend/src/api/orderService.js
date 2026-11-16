import { apiClient } from './client';

export async function getMyOrders(params = {}) {
  const response = await apiClient.get('/user/orders', {
    params,
    requiresAuth: true,
  });
  return response.data?.data || response.data;
}

export async function getMyOrderById(orderId) {
  const response = await apiClient.get(`/user/orders/${orderId}`, {
    requiresAuth: true,
  });
  return response.data?.data || response.data;
}

export async function getMyOrderByNumber(orderNumber) {
  const response = await apiClient.get(`/user/orders/number/${orderNumber}`, {
    requiresAuth: true,
  });
  return response.data?.data || response.data;
}

export async function cancelOrder(orderId) {
  const response = await apiClient.patch(`/user/orders/${orderId}/cancel`, undefined, {
    requiresAuth: true,
  });
  return response.data?.data || response.data;
}

export async function downloadInvoice(orderId) {
  const response = await apiClient.get(`/user/orders/${orderId}/invoice`, {
    responseType: 'blob',
    requiresAuth: true,
  });
  return response.data;
}

export async function getCustomerShipment(orderId) {
  const response = await apiClient.get(`/user/orders/${orderId}/shipment`, {
    requiresAuth: true,
  });
  return response.data?.data || response.data;
}

export async function getCustomerTrackingHistory(orderId) {
  const response = await apiClient.get(`/user/orders/${orderId}/shipment/history`, {
    requiresAuth: true,
  });
  return response.data?.data || response.data;
}
