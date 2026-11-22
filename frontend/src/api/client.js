import axios from 'axios';
import { toApiError } from './errors';
import { clearAccessToken, getAccessToken, setAccessToken } from './tokenStore';

const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';

export const apiClient = axios.create({
  baseURL,
  withCredentials: true,
  headers: {
    Accept: 'application/json',
  },
});

const refreshClient = axios.create({
  baseURL,
  withCredentials: true,
  headers: {
    Accept: 'application/json',
  },
});

const refreshPromises = new Map();
const expiredSessionListeners = new Set();

function notifyExpiredSession(error) {
  expiredSessionListeners.forEach((listener) => listener(error));
}

function extractAccessToken(response) {
  const accessToken = response?.data?.data?.access_token;

  if (!accessToken) {
    throw new Error('The refresh response did not include an access token.');
  }

  return accessToken;
}

export function subscribeToExpiredSession(listener) {
  expiredSessionListeners.add(listener);
  return () => expiredSessionListeners.delete(listener);
}

export function getRefreshEndpoint(requestUrl) {
  const path = typeof requestUrl === 'string' ? requestUrl.split('?')[0] : '';
  return path === '/admin' || path.startsWith('/admin/')
    ? '/admin/auth/refresh-token'
    : '/auth/refresh-token';
}

export function refreshAccessToken(endpoint = '/auth/refresh-token') {
  if (!refreshPromises.has(endpoint)) {
    const refreshPromise = refreshClient
      .post(endpoint)
      .then(extractAccessToken)
      .then((accessToken) => {
        setAccessToken(accessToken);
        return accessToken;
      })
      .catch((error) => {
        const apiError = toApiError(error);
        clearAccessToken();
        notifyExpiredSession(apiError);
        throw apiError;
      })
      .finally(() => {
        refreshPromises.delete(endpoint);
      });
    refreshPromises.set(endpoint, refreshPromise);
  }

  return refreshPromises.get(endpoint);
}

apiClient.interceptors.request.use((config) => {
  const accessToken = getAccessToken();

  if (config.requiresAuth && accessToken && !config.skipAuth) {
    config.headers = config.headers ?? {};
    config.headers.Authorization = `Bearer ${accessToken}`;
  }

  return config;
});

apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    const shouldRefresh =
      error.response?.status === 401 &&
      originalRequest?.requiresAuth &&
      !originalRequest._retry &&
      !originalRequest.skipSessionRefresh;

    if (!shouldRefresh) {
      return Promise.reject(toApiError(error));
    }

    originalRequest._retry = true;

    try {
      const accessToken = await refreshAccessToken(getRefreshEndpoint(originalRequest.url));
      originalRequest.headers = originalRequest.headers ?? {};
      originalRequest.headers.Authorization = `Bearer ${accessToken}`;
      return apiClient(originalRequest);
    } catch (refreshError) {
      return Promise.reject(toApiError(refreshError));
    }
  },
);
