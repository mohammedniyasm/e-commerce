import { apiClient } from './client';

/**
 * Customer Wallet API Service
 */

export async function getWallet() {
  const response = await apiClient.get('/user/wallet', {
    requiresAuth: true,
  });
  return response.data?.data || response.data;
}

export async function getWalletTransactions(params = {}) {
  const response = await apiClient.get('/user/wallet/transactions', {
    params,
    requiresAuth: true,
  });
  return response.data?.data || response.data;
}

export async function getTransactionById(id) {
  const response = await apiClient.get(`/user/wallet/transactions/${id}`, {
    requiresAuth: true,
  });
  return response.data?.data || response.data;
}

export async function createTopUpOrder(amount) {
  const response = await apiClient.post(
    '/user/wallet/top-up/order',
    { amount: Number(amount) },
    { requiresAuth: true },
  );
  return response.data?.data || response.data;
}

export async function verifyTopUpPayment({
  gateway_order_id,
  gateway_payment_id,
  razorpay_signature,
}) {
  const response = await apiClient.post(
    '/user/wallet/top-up/verify',
    {
      gateway_order_id,
      gateway_payment_id,
      razorpay_signature,
    },
    { requiresAuth: true },
  );
  return response.data?.data || response.data;
}
