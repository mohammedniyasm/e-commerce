import { apiClient } from './client';
import { loginAdmin as authLoginAdmin } from './authService';

/**
 * Admin Authentication
 */
export async function adminLogin(credentials) {
  return authLoginAdmin(credentials);
}

/**
 * Admin Profile
 * GET /api/v1/admin/profile
 */
export async function getAdminProfile() {
  const response = await apiClient.get('/admin/profile', { requiresAuth: true });
  return response?.data?.data ?? null;
}

/**
 * Dashboard KPI Overview
 * GET /api/v1/admin/dashboard/overview
 * Returns: { total_revenue, total_orders, total_customers, total_products, ... }
 */
export async function getDashboardOverview() {
  const response = await apiClient.get('/admin/dashboard/overview', { requiresAuth: true });
  const overview = response?.data?.data;
  if (!overview) return null;

  return {
    total_revenue: overview.total_revenue ?? overview.TotalRevenue,
    total_orders: overview.total_orders ?? overview.TotalOrders,
    total_customers: overview.total_customers ?? overview.TotalCustomers,
    total_products: overview.total_products ?? overview.TotalProducts,
    pending_orders: overview.pending_orders ?? overview.PendingOrders,
    confirmed_orders: overview.confirmed_orders ?? overview.ConfirmedOrders,
    processing_orders: overview.processing_orders ?? overview.ProcessingOrders,
    shipped_orders: overview.shipped_orders ?? overview.ShippedOrders,
    delivered_orders: overview.delivered_orders ?? overview.DeliveredOrders,
    cancelled_orders: overview.cancelled_orders ?? overview.CancelledOrders,
    returned_orders: overview.returned_orders ?? overview.ReturnedOrders,
  };
}

/**
 * Dashboard Sales Time Series
 * GET /api/v1/admin/dashboard/sales?group_by=day&start_date=YYYY-MM-DD&end_date=YYYY-MM-DD
 * Returns: Array of { date, total_revenue, total_orders }
 */
export async function getDashboardSales(params = {}) {
  const queryParams = {
    group_by: params.group_by || 'day',
    ...(params.start_date ? { start_date: params.start_date } : {}),
    ...(params.end_date ? { end_date: params.end_date } : {}),
  };
  const response = await apiClient.get('/admin/dashboard/sales', {
    params: queryParams,
    requiresAuth: true,
  });
  return response?.data?.data ?? [];
}

/**
 * Dashboard Order Statistics
 * GET /api/v1/admin/dashboard/orders
 * Returns: { total_orders, pending_orders, confirmed_orders, processing_orders, shipped_orders, delivered_orders, cancelled_orders, returned_orders }
 */
export async function getDashboardOrderStats() {
  const response = await apiClient.get('/admin/dashboard/orders', { requiresAuth: true });
  return response?.data?.data ?? null;
}

/**
 * Dashboard Customer Statistics
 * GET /api/v1/admin/dashboard/customers
 * Returns: { total_customers, blocked_customers, verified_customers, unverified_customers }
 */
export async function getDashboardCustomerStats() {
  const response = await apiClient.get('/admin/dashboard/customers', { requiresAuth: true });
  return response?.data?.data ?? null;
}

/**
 * Top Selling Products
 * GET /api/v1/admin/dashboard/top-products?limit=5
 * Returns: Array of { product_id, product_name, units_sold, total_orders, revenue }
 */
export async function getDashboardTopProducts(limit = 5) {
  const response = await apiClient.get('/admin/dashboard/top-products', {
    params: { limit },
    requiresAuth: true,
  });
  return response?.data?.data ?? [];
}

/**
 * Top Categories
 * GET /api/v1/admin/dashboard/top-categories?limit=5
 * Returns: Array of { category_id, category_name, units_sold, revenue }
 */
export async function getDashboardTopCategories(limit = 5) {
  const response = await apiClient.get('/admin/dashboard/top-categories', {
    params: { limit },
    requiresAuth: true,
  });
  return response?.data?.data ?? [];
}

/**
 * Top Brands
 * GET /api/v1/admin/dashboard/top-brands?limit=5
 * Returns: Array of { brand_id, brand_name, units_sold, revenue }
 */
export async function getDashboardTopBrands(limit = 5) {
  const response = await apiClient.get('/admin/dashboard/top-brands', {
    params: { limit },
    requiresAuth: true,
  });
  return response?.data?.data ?? [];
}

/**
 * Admin Orders List (Recent Orders)
 * GET /api/v1/admin/orders?page=1&limit=5
 * Returns: { orders: [...], total_count, page, limit, total_pages }
 */
export async function getAdminRecentOrders(params = { limit: 5, page: 1 }) {
  const response = await apiClient.get('/admin/orders', {
    params,
    requiresAuth: true,
  });
  return response?.data?.data ?? { orders: [], total_count: 0, page: 1, limit: 5, total_pages: 0 };
}

// Re-export Catalog Services for convenient unified access
export * from './adminCatalogService';

// Re-export Operations Services for convenient unified access
export * from './adminOperationsService';

