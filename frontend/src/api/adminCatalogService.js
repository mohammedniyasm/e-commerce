import { apiClient } from './client';

// ==========================================
// CATEGORIES
// ==========================================

export async function listCategories(params = {}) {
  const queryParams = {
    page: params.page || 1,
    limit: params.limit || 10,
    ...(params.search ? { search: params.search } : {}),
  };
  const response = await apiClient.get('/admin/categories', {
    params: queryParams,
    requiresAuth: true,
  });
  return response?.data?.data ?? { categories: [], page: 1, limit: 10, total: 0, total_pages: 0 };
}

export async function getCategoryById(id) {
  const response = await apiClient.get(`/admin/categories/${id}`, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function createCategory(payload) {
  const response = await apiClient.post('/admin/categories', payload, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function updateCategory(id, payload) {
  const response = await apiClient.put(`/admin/categories/${id}`, payload, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function deleteCategory(id) {
  const response = await apiClient.delete(`/admin/categories/${id}`, { requiresAuth: true });
  return response?.data ?? null;
}

export async function listDeletedCategories(params = {}) {
  const queryParams = {
    page: params.page || 1,
    limit: params.limit || 10,
    ...(params.search ? { search: params.search } : {}),
  };
  const response = await apiClient.get('/admin/categories/deleted', {
    params: queryParams,
    requiresAuth: true,
  });
  return response?.data?.data ?? { categories: [], page: 1, limit: 10, total: 0, total_pages: 0 };
}

export async function restoreCategory(id) {
  const response = await apiClient.patch(`/admin/categories/${id}/restore`, undefined, { requiresAuth: true });
  return response?.data ?? null;
}

export async function toggleCategoryActive(id) {
  const response = await apiClient.patch(`/admin/categories/${id}/toggle`, undefined, { requiresAuth: true });
  return response?.data ?? null;
}

// ==========================================
// BRANDS
// ==========================================

export async function listBrands(params = {}) {
  const queryParams = {
    page: params.page || 1,
    limit: params.limit || 10,
    ...(params.search ? { search: params.search } : {}),
  };
  const response = await apiClient.get('/admin/brands', {
    params: queryParams,
    requiresAuth: true,
  });
  return response?.data?.data ?? { brands: [], page: 1, limit: 10, total: 0, total_pages: 0 };
}

export async function getBrandById(id) {
  const response = await apiClient.get(`/admin/brands/${id}`, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function createBrand(payload) {
  const response = await apiClient.post('/admin/brands', payload, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function updateBrand(id, payload) {
  const response = await apiClient.put(`/admin/brands/${id}`, payload, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function deleteBrand(id) {
  const response = await apiClient.delete(`/admin/brands/${id}`, { requiresAuth: true });
  return response?.data ?? null;
}

export async function uploadBrandLogo(id, file) {
  const formData = new FormData();
  formData.append('logo', file);
  const response = await apiClient.put(`/admin/brands/${id}/logo`, formData, {
    requiresAuth: true,
    headers: { 'Content-Type': 'multipart/form-data' },
  });
  return response?.data?.data ?? null;
}

export async function deleteBrandLogo(id) {
  const response = await apiClient.delete(`/admin/brands/${id}/logo`, { requiresAuth: true });
  return response?.data ?? null;
}

export async function listDeletedBrands(params = {}) {
  const queryParams = {
    page: params.page || 1,
    limit: params.limit || 10,
    ...(params.search ? { search: params.search } : {}),
  };
  const response = await apiClient.get('/admin/brands/deleted', {
    params: queryParams,
    requiresAuth: true,
  });
  return response?.data?.data ?? { brands: [], page: 1, limit: 10, total: 0, total_pages: 0 };
}

export async function restoreBrand(id) {
  const response = await apiClient.patch(`/admin/brands/${id}/restore`, undefined, { requiresAuth: true });
  return response?.data ?? null;
}

export async function toggleBrandActive(id) {
  const response = await apiClient.patch(`/admin/brands/${id}/toggle`, undefined, { requiresAuth: true });
  return response?.data ?? null;
}

// ==========================================
// PRODUCTS
// ==========================================

export async function listProducts(params = {}) {
  const queryParams = {
    page: params.page || 1,
    limit: params.limit || 10,
    ...(params.search ? { search: params.search } : {}),
    ...(params.is_active !== undefined ? { is_active: params.is_active } : {}),
    ...(params.is_listed !== undefined ? { is_listed: params.is_listed } : {}),
    ...(params.category_id ? { category_id: params.category_id } : {}),
    ...(params.brand_id ? { brand_id: params.brand_id } : {}),
  };
  const response = await apiClient.get('/admin/products', {
    params: queryParams,
    requiresAuth: true,
  });
  return response?.data?.data ?? { products: [], page: 1, limit: 10, total: 0, total_pages: 0 };
}

export async function getProductById(id) {
  const response = await apiClient.get(`/admin/products/${id}`, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function getProductBySlug(slug) {
  const response = await apiClient.get(`/admin/products/slug/${slug}`, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function createProduct(payload) {
  const response = await apiClient.post('/admin/products', payload, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function updateProduct(id, payload) {
  const response = await apiClient.put(`/admin/products/${id}`, payload, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function deleteProduct(id) {
  const response = await apiClient.delete(`/admin/products/${id}`, { requiresAuth: true });
  return response?.data ?? null;
}

export async function listDeletedProducts(params = {}) {
  const queryParams = {
    page: params.page || 1,
    limit: params.limit || 10,
    ...(params.search ? { search: params.search } : {}),
  };
  const response = await apiClient.get('/admin/products/deleted', {
    params: queryParams,
    requiresAuth: true,
  });
  return response?.data?.data ?? { products: [], page: 1, limit: 10, total: 0, total_pages: 0 };
}

export async function restoreProduct(id) {
  const response = await apiClient.patch(`/admin/products/${id}/restore`, undefined, { requiresAuth: true });
  return response?.data ?? null;
}

export async function toggleProductActive(id) {
  const response = await apiClient.patch(`/admin/products/${id}/active`, undefined, { requiresAuth: true });
  return response?.data ?? null;
}

export async function toggleProductListed(id) {
  const response = await apiClient.patch(`/admin/products/${id}/listed`, undefined, { requiresAuth: true });
  return response?.data ?? null;
}

// ==========================================
// VARIANTS
// ==========================================

export async function listVariantsByProductId(productId) {
  const response = await apiClient.get(`/admin/products/${productId}/variants`, { requiresAuth: true });
  return response?.data?.data ?? [];
}

export async function getVariantById(id) {
  const response = await apiClient.get(`/admin/variants/${id}`, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function createVariant(productId, payload) {
  const response = await apiClient.post(`/admin/products/${productId}/variants`, payload, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function updateVariant(id, payload) {
  const response = await apiClient.put(`/admin/variants/${id}`, payload, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function deleteVariant(id) {
  const response = await apiClient.delete(`/admin/variants/${id}`, { requiresAuth: true });
  return response?.data ?? null;
}

// ==========================================
// PRODUCT IMAGES
// ==========================================

export async function listProductImages(productId) {
  const response = await apiClient.get(`/admin/products/${productId}/images`, { requiresAuth: true });
  return response?.data?.data ?? [];
}

export async function listVariantImages(variantId) {
  const response = await apiClient.get(`/admin/variants/${variantId}/images`, { requiresAuth: true });
  return response?.data?.data ?? [];
}

export async function getProductImageById(id) {
  const response = await apiClient.get(`/admin/product-images/${id}`, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function uploadProductImages(productId, files, variantId = null) {
  const formData = new FormData();
  if (variantId) {
    formData.append('variant_id', String(variantId));
  }
  for (const file of files) {
    formData.append('images', file);
  }
  const response = await apiClient.post(`/admin/products/${productId}/images`, formData, {
    requiresAuth: true,
    headers: { 'Content-Type': 'multipart/form-data' },
  });
  return response?.data?.data ?? [];
}

export async function updateProductImage(id, payload) {
  const response = await apiClient.put(`/admin/product-images/${id}`, payload, { requiresAuth: true });
  return response?.data?.data ?? null;
}

export async function deleteProductImage(id) {
  const response = await apiClient.delete(`/admin/product-images/${id}`, { requiresAuth: true });
  return response?.data ?? null;
}
