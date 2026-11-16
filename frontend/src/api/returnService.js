import { apiClient } from './client';

/**
 * Return & Exchange API Service
 */

export async function createReturnRequest({
  order_id,
  order_item_id,
  request_type,
  requested_variant_id,
  quantity,
  reason,
}) {
  const payload = {
    order_id,
    order_item_id: Number(order_item_id),
    request_type, // 'return' or 'exchange'
    quantity: Number(quantity),
    reason: reason.trim(),
  };

  if (requested_variant_id) {
    payload.requested_variant_id = Number(requested_variant_id);
  }

  const response = await apiClient.post('/user/return-requests', payload, {
    requiresAuth: true,
  });
  return response.data?.data || response.data;
}

export async function getMyReturnRequests(params = {}) {
  const response = await apiClient.get('/user/return-requests', {
    params,
    requiresAuth: true,
  });
  return response.data?.data || response.data;
}

export async function getReturnRequestById(id) {
  const response = await apiClient.get(`/user/return-requests/${id}`, {
    requiresAuth: true,
  });
  return response.data?.data || response.data;
}

export async function getReturnRequestShipments(id) {
  const response = await apiClient.get(`/user/return-requests/${id}/shipments`, {
    requiresAuth: true,
  });
  return response.data?.data || response.data;
}
