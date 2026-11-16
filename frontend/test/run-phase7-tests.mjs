import { createServer } from 'vite';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter, Routes, Route } from 'react-router-dom';

function createMockToken(role = 'customer') {
  const header = Buffer.from(JSON.stringify({ alg: 'HS256', typ: 'JWT' })).toString('base64');
  const payload = Buffer.from(
    JSON.stringify({
      sub: 1,
      role: role,
      exp: Math.floor(Date.now() / 1000) + 3600,
    })
  ).toString('base64');
  return `${header}.${payload}.mockSignature`;
}

async function run() {
  console.log('====================================================');
  console.log('PHASE 7: AUTOMATED TEST SUITE — ADMIN AUTH & DASHBOARD');
  console.log('====================================================\n');

  const vite = await createServer({
    server: { middlewareMode: true },
    appType: 'custom',
    root: process.cwd(),
  });

  try {
    const assertions = [];

    function assert(condition, message, details = '') {
      if (condition) {
        assertions.push({ pass: true, message });
        console.log(`  ✓ ${message} ${details ? `(${details})` : ''}`);
      } else {
        assertions.push({ pass: false, message });
        console.error(`  ✗ FAIL: ${message} ${details ? `(${details})` : ''}`);
      }
    }

    // 1. API Services Contract Verification
    console.log('[1/6] Testing adminService API Functions and Endpoints...');
    const AdminService = await vite.ssrLoadModule('/src/api/adminService.js');
    assert(typeof AdminService.adminLogin === 'function', 'adminService exports adminLogin');
    assert(typeof AdminService.getAdminProfile === 'function', 'adminService exports getAdminProfile');
    assert(typeof AdminService.getDashboardOverview === 'function', 'adminService exports getDashboardOverview');
    assert(typeof AdminService.getDashboardSales === 'function', 'adminService exports getDashboardSales');
    assert(typeof AdminService.getDashboardOrderStats === 'function', 'adminService exports getDashboardOrderStats');
    assert(typeof AdminService.getDashboardCustomerStats === 'function', 'adminService exports getDashboardCustomerStats');
    assert(typeof AdminService.getDashboardTopProducts === 'function', 'adminService exports getDashboardTopProducts');
    assert(typeof AdminService.getDashboardTopCategories === 'function', 'adminService exports getDashboardTopCategories');
    assert(typeof AdminService.getDashboardTopBrands === 'function', 'adminService exports getDashboardTopBrands');
    assert(typeof AdminService.getAdminRecentOrders === 'function', 'adminService exports getAdminRecentOrders');

    // 2. HTTP Endpoint Mappings & Payloads
    console.log('\n[2/6] Testing HTTP Endpoint Mappings and Request Options...');
    const { apiClient } = await vite.ssrLoadModule('/src/api/client.js');
    let lastRequest = null;
    function mockApi(method, mockData) {
      const original = apiClient[method];
      apiClient[method] = async (url, dataOrConfig, maybeConfig) => {
        if (method === 'get' || method === 'delete') {
          lastRequest = { method, url, config: dataOrConfig };
        } else {
          lastRequest = { method, url, data: dataOrConfig, config: maybeConfig };
        }
        return { data: { success: true, data: mockData } };
      };
      return () => {
        apiClient[method] = original;
      };
    }

    // Overview
    let restore = mockApi('get', { total_revenue: 500000, total_orders: 120 });
    await AdminService.getDashboardOverview();
    assert(lastRequest.url === '/admin/dashboard/overview', 'getDashboardOverview requests /admin/dashboard/overview');
    assert(lastRequest.config?.requiresAuth === true, 'getDashboardOverview requires authentication');
    restore();

    // Sales
    restore = mockApi('get', [{ date: '2026-11-14', total_revenue: 25000, total_orders: 10 }]);
    await AdminService.getDashboardSales({ group_by: 'day', start_date: '2026-11-01', end_date: '2026-11-14' });
    assert(lastRequest.url === '/admin/dashboard/sales', 'getDashboardSales requests /admin/dashboard/sales');
    assert(lastRequest.config?.params?.group_by === 'day', 'getDashboardSales passes group_by param');
    assert(lastRequest.config?.params?.start_date === '2026-11-01', 'getDashboardSales passes start_date');
    assert(lastRequest.config?.requiresAuth === true, 'getDashboardSales requires authentication');
    restore();

    // Orders Stats
    restore = mockApi('get', { total_orders: 120, delivered_orders: 80 });
    await AdminService.getDashboardOrderStats();
    assert(lastRequest.url === '/admin/dashboard/orders', 'getDashboardOrderStats requests /admin/dashboard/orders');
    restore();

    // Customers Stats
    restore = mockApi('get', { total_customers: 340, verified_customers: 300 });
    await AdminService.getDashboardCustomerStats();
    assert(lastRequest.url === '/admin/dashboard/customers', 'getDashboardCustomerStats requests /admin/dashboard/customers');
    restore();

    // Top Products
    restore = mockApi('get', [{ product_id: 1, product_name: 'Silk Shirt', revenue: 45000 }]);
    await AdminService.getDashboardTopProducts(5);
    assert(lastRequest.url === '/admin/dashboard/top-products', 'getDashboardTopProducts requests /admin/dashboard/top-products');
    assert(lastRequest.config?.params?.limit === 5, 'getDashboardTopProducts passes limit param');
    restore();

    // Recent Orders
    restore = mockApi('get', { orders: [{ id: 1, order_number: 'ORD-001', total_amount: 3500 }], total_count: 1 });
    await AdminService.getAdminRecentOrders({ limit: 5, page: 1 });
    assert(lastRequest.url === '/admin/orders', 'getAdminRecentOrders requests /admin/orders');
    assert(lastRequest.config?.params?.limit === 5, 'getAdminRecentOrders passes limit');
    restore();

    // 3. Role-Based Route Protection (RequireAuth)
    console.log('\n[3/6] Testing RequireAuth Role-Based Access Isolation...');
    const { RequireAuth } = await vite.ssrLoadModule('/src/routes/RequireAuth.jsx');
    const { AuthProvider } = await vite.ssrLoadModule('/src/context/AuthContext.jsx');
    const { setAccessToken, clearAccessToken } = await vite.ssrLoadModule('/src/api/tokenStore.js');

    clearAccessToken();

    // Anonymous User accessing Admin Protected Route
    const anonHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/admin'] },
        React.createElement(AuthProvider, null, React.createElement(RequireAuth, { role: 'admin' }, 'SECRET_ADMIN_PORTAL'))
      )
    );
    assert(!anonHtml.includes('SECRET_ADMIN_PORTAL'), 'Anonymous user is blocked from admin route');

    // Customer Session accessing Admin Protected Route
    setAccessToken(createMockToken('customer'));
    const customerHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/admin'] },
        React.createElement(AuthProvider, null, React.createElement(RequireAuth, { role: 'admin' }, 'SECRET_ADMIN_PORTAL'))
      )
    );
    assert(!customerHtml.includes('SECRET_ADMIN_PORTAL'), 'Customer role is strictly prevented from accessing admin route');

    // Admin Session accessing Admin Route
    setAccessToken(createMockToken('admin'));
    const adminHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/admin'] },
        React.createElement(AuthProvider, null, React.createElement(RequireAuth, { role: 'admin' }, 'SECRET_ADMIN_PORTAL'))
      )
    );
    assert(adminHtml.includes('SECRET_ADMIN_PORTAL'), 'Admin role is granted access to admin route');
    clearAccessToken();

    // 4. Admin Login Page Rendering
    console.log('\n[4/6] Testing AdminLoginPage Layout & Structure...');
    const { AdminLoginPage } = await vite.ssrLoadModule('/src/pages/admin/AdminLoginPage.jsx');
    const { ToastProvider } = await vite.ssrLoadModule('/src/context/ToastContext.jsx');

    const loginHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/admin/login'] },
        React.createElement(
          ToastProvider,
          null,
          React.createElement(AuthProvider, null, React.createElement(AdminLoginPage, null))
        )
      )
    );
    assert(loginHtml.includes('DRAPE'), 'AdminLoginPage renders DRAPE brand logo');
    assert(loginHtml.includes('Admin'), 'AdminLoginPage renders Admin badge');
    assert(loginHtml.includes('admin-email'), 'AdminLoginPage renders admin email input');
    assert(loginHtml.includes('admin-password'), 'AdminLoginPage renders admin password input');
    assert(loginHtml.includes('Enter Admin Console'), 'AdminLoginPage renders submit button');
    assert(loginHtml.includes('Return to Storefront'), 'AdminLoginPage renders storefront return link');

    // 5. Admin Layout, Sidebar & Header Components
    console.log('\n[5/6] Testing AdminLayout, AdminSidebar and AdminHeader...');
    const { AdminSidebar } = await vite.ssrLoadModule('/src/components/admin/AdminSidebar.jsx');
    const { AdminHeader } = await vite.ssrLoadModule('/src/components/admin/AdminHeader.jsx');
    const { AdminLayout } = await vite.ssrLoadModule('/src/layouts/AdminLayout.jsx');

    setAccessToken(createMockToken('admin'));

    const sidebarHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/admin'] },
        React.createElement(AuthProvider, null, React.createElement(AdminSidebar, { isOpen: true }))
      )
    );
    assert(sidebarHtml.includes('DRAPE'), 'AdminSidebar renders DRAPE wordmark');
    assert(sidebarHtml.includes('Dashboard'), 'AdminSidebar contains Dashboard nav item');
    assert(sidebarHtml.includes('Products'), 'AdminSidebar contains Products nav item');
    assert(sidebarHtml.includes('Categories'), 'AdminSidebar contains Categories nav item');
    assert(sidebarHtml.includes('Brands'), 'AdminSidebar contains Brands nav item');
    assert(sidebarHtml.includes('Orders'), 'AdminSidebar contains Orders nav item');
    assert(sidebarHtml.includes('Payments'), 'AdminSidebar contains Payments nav item');
    assert(sidebarHtml.includes('Returns'), 'AdminSidebar contains Returns nav item');
    assert(sidebarHtml.includes('Coupons'), 'AdminSidebar contains Coupons nav item');
    assert(sidebarHtml.includes('Reviews'), 'AdminSidebar contains Reviews nav item');
    assert(sidebarHtml.includes('Customers'), 'AdminSidebar contains Customers nav item');
    assert(sidebarHtml.includes('Reports'), 'AdminSidebar contains Reports nav item');
    assert(sidebarHtml.includes('Settings'), 'AdminSidebar contains Settings nav item');
    assert(sidebarHtml.includes('Sign Out'), 'AdminSidebar contains Sign Out button');

    const headerHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/admin'] },
        React.createElement(AdminHeader, null)
      )
    );
    assert(headerHtml.includes('Admin'), 'AdminHeader renders breadcrumb root');
    assert(headerHtml.includes('Dashboard'), 'AdminHeader renders active section in breadcrumb');
    assert(headerHtml.includes('View Store'), 'AdminHeader provides link to View Store');

    const layoutHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/admin'] },
        React.createElement(
          ToastProvider,
          null,
          React.createElement(AuthProvider, null, React.createElement(AdminLayout, null))
        )
      )
    );
    assert(layoutHtml.includes('admin-shell'), 'AdminLayout renders admin-shell container');
    assert(layoutHtml.includes('admin-content'), 'AdminLayout renders admin-content main wrapper');

    // 6. Admin Dashboard Page Component
    console.log('\n[6/6] Testing AdminDashboardPage Structure and Widgets...');
    const { AdminDashboardPage } = await vite.ssrLoadModule('/src/pages/admin/AdminDashboardPage.jsx');

    const dashboardHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/admin'] },
        React.createElement(
          ToastProvider,
          null,
          React.createElement(AuthProvider, null, React.createElement(AdminDashboardPage, null))
        )
      )
    );
    assert(dashboardHtml.includes('Dashboard Overview'), 'AdminDashboardPage renders Dashboard Overview heading');
    assert(dashboardHtml.includes('Total Orders'), 'AdminDashboardPage renders Total Orders KPI card');
    assert(dashboardHtml.includes('Total Revenue'), 'AdminDashboardPage renders Total Revenue KPI card');
    assert(dashboardHtml.includes('Customers'), 'AdminDashboardPage renders Total Customers KPI card');
    assert(dashboardHtml.includes('Catalog Products'), 'AdminDashboardPage renders Catalog Products KPI card');
    assert(dashboardHtml.includes('Revenue Analytics'), 'AdminDashboardPage renders Revenue Analytics section');
    assert(dashboardHtml.includes('7 Days'), 'AdminDashboardPage renders 7 Days time filter button');
    assert(dashboardHtml.includes('30 Days'), 'AdminDashboardPage renders 30 Days time filter button');
    assert(dashboardHtml.includes('1 Year'), 'AdminDashboardPage renders 1 Year time filter button');
    assert(dashboardHtml.includes('Recent Orders'), 'AdminDashboardPage renders Recent Orders table section');
    assert(dashboardHtml.includes('Top Selling Products'), 'AdminDashboardPage renders Top Selling Products section');
    assert(dashboardHtml.includes('Order Status Distribution'), 'AdminDashboardPage renders Order Status Distribution');
    assert(dashboardHtml.includes('Customer Account Health'), 'AdminDashboardPage renders Customer Account Health');

    clearAccessToken();

    // Summary
    const failed = assertions.filter((a) => !a.pass);
    console.log('\n====================================================');
    console.log(`PHASE 7 TEST RESULTS: ${assertions.length - failed.length}/${assertions.length} PASSED`);
    console.log('====================================================\n');

    if (failed.length > 0) {
      console.error(`FAILED ${failed.length} assertions:`);
      failed.forEach((f) => console.error(` - ${f.message}`));
      process.exit(1);
    }
  } catch (err) {
    console.error('Test execution error:', err);
    process.exit(1);
  } finally {
    await vite.close();
  }
}

run();
