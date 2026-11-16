import { apiClient } from './client';

/**
 * Customer Referral API Service
 */

export async function getReferralDashboard(params = {}) {
  const response = await apiClient.get('/user/referrals', {
    params,
    requiresAuth: true,
  });
  return response.data?.data || response.data;
}
