import { apiClient, refreshAccessToken } from './client';
import { clearAccessToken, getSession, setAccessToken } from './tokenStore';

function readAccessToken(response) {
  const accessToken = response?.data?.data?.access_token;

  if (!accessToken) {
    throw new Error('The server response did not include an access token.');
  }

  return accessToken;
}

async function signIn(path, credentials) {
  const response = await apiClient.post(path, credentials, { skipAuth: true });
  const accessToken = readAccessToken(response);
  setAccessToken(accessToken);
  return getSession();
}

export function loginCustomer(credentials) {
  return signIn('/auth/login', credentials);
}

export function registerCustomer(payload) {
  return apiClient.post('/auth/register', payload, { skipAuth: true });
}

export function loginWithGoogle(payload) {
  return signIn('/auth/google', payload);
}

export function forgotPassword({ email }) {
  return apiClient.post('/auth/forgot-password', { email }, { skipAuth: true });
}

export function verifyForgotPasswordOTP({ email, otp }) {
  return apiClient.post('/auth/forgot-password/verify-otp', { email, otp }, { skipAuth: true });
}

export function resendForgotPasswordOTP({ email }) {
  return apiClient.post('/auth/forgot-password/resend-otp', { email }, { skipAuth: true });
}

export function resetPassword({ new_password }) {
  return apiClient.post('/auth/reset-password', { new_password }, { skipAuth: true });
}

export function loginAdmin(credentials) {
  return signIn('/admin/auth/login', credentials);
}

export async function restoreSession() {
  try {
    await refreshAccessToken();
    return getSession();
  } catch {
    return null;
  }
}

export async function logout({ role } = {}) {
  const path = role === 'admin' ? '/admin/auth/logout' : '/auth/logout';

  try {
    await apiClient.post(path, undefined, {
      requiresAuth: true,
      skipSessionRefresh: true,
    });
  } finally {
    clearAccessToken();
  }
}
