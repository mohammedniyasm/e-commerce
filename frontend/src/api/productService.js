import { apiClient } from './client';

/**
 * Storefront Product & Discovery API Services
 * Corresponds to docs/openapi.yaml Storefront tag
 */

export async function getStoreProducts(params = {}) {
  const cleanParams = {};
  if (params.search) cleanParams.search = params.search;
  if (params.category_id) cleanParams.category_id = params.category_id;
  if (params.brand_id) cleanParams.brand_id = params.brand_id;
  if (params.min_price != null && params.min_price !== '') cleanParams.min_price = params.min_price;
  if (params.max_price != null && params.max_price !== '') cleanParams.max_price = params.max_price;
  if (params.sort) cleanParams.sort = params.sort;
  if (params.page != null) cleanParams.page = params.page;
  if (params.limit != null) cleanParams.limit = params.limit;

  const response = await apiClient.get('/products', {
    params: cleanParams,
    skipAuth: true,
  });
  return response?.data?.data;
}

export async function getStoreProductBySlug(slug) {
  const response = await apiClient.get(`/products/slug/${encodeURIComponent(slug)}`, {
    skipAuth: true,
  });
  return response?.data?.data;
}

export async function getStoreProductById(id) {
  const response = await apiClient.get(`/products/id/${id}`, {
    skipAuth: true,
  });
  return response?.data?.data;
}

export async function getRelatedProducts(productId, { limit = 4 } = {}) {
  const response = await apiClient.get(`/products/${productId}/recommendations`, {
    params: { limit },
    skipAuth: true,
  });
  return response?.data?.data || [];
}

export async function getProductReviews(productId, { page = 1, limit = 10 } = {}) {
  const response = await apiClient.get(`/products/${productId}/reviews`, {
    params: { page, limit },
    skipAuth: true,
  });
  return response?.data?.data;
}

export async function getStoreCategories() {
  const response = await apiClient.get('/categories', {
    skipAuth: true,
  });
  return response?.data?.data || [];
}

export async function getStoreBrands() {
  const response = await apiClient.get('/brands', {
    skipAuth: true,
  });
  return response?.data?.data || [];
}
