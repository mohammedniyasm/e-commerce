import { createServer } from 'vite';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';

function createMockToken(role = 'admin') {
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
  console.log('PHASE 9: AUTOMATED TEST SUITE — ADMIN OPERATIONS & REPORTING');
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
    console.log('[1/5] Testing adminOperationsService Exports...');
    const OpsService = await vite.ssrLoadModule('/src/api/adminOperationsService.js');

    // Orders & Payments
    assert(typeof OpsService.listAdminOrders === 'function', 'exports listAdminOrders');
    assert(typeof OpsService.getAdminOrderById === 'function', 'exports getAdminOrderById');
    assert(typeof OpsService.updateOrderStatus === 'function', 'exports updateOrderStatus');
    assert(typeof OpsService.updateCODPaymentStatus === 'function', 'exports updateCODPaymentStatus');
    assert(typeof OpsService.listRefundPendingOrders === 'function', 'exports listRefundPendingOrders');
    assert(typeof OpsService.processWalletRefund === 'function', 'exports processWalletRefund');

    // Shipments
    assert(typeof OpsService.createShipment === 'function', 'exports createShipment');
    assert(typeof OpsService.getShipmentById === 'function', 'exports getShipmentById');
    assert(typeof OpsService.getShipmentByOrderId === 'function', 'exports getShipmentByOrderId');
    assert(typeof OpsService.updateShipment === 'function', 'exports updateShipment');
    assert(typeof OpsService.updateShippingStatus === 'function', 'exports updateShippingStatus');
    assert(typeof OpsService.getShipmentTrackingHistory === 'function', 'exports getShipmentTrackingHistory');

    // Returns & Exchanges
    assert(typeof OpsService.listReturnRequests === 'function', 'exports listReturnRequests');
    assert(typeof OpsService.getReturnRequestById === 'function', 'exports getReturnRequestById');
    assert(typeof OpsService.updateReturnRequestStatus === 'function', 'exports updateReturnRequestStatus');
    assert(typeof OpsService.createExchangeOrder === 'function', 'exports createExchangeOrder');
    assert(typeof OpsService.createReturnShipment === 'function', 'exports createReturnShipment');

    // Coupons
    assert(typeof OpsService.listAdminCoupons === 'function', 'exports listAdminCoupons');
    assert(typeof OpsService.getAdminCouponById === 'function', 'exports getAdminCouponById');
    assert(typeof OpsService.createAdminCoupon === 'function', 'exports createAdminCoupon');
    assert(typeof OpsService.updateAdminCoupon === 'function', 'exports updateAdminCoupon');
    assert(typeof OpsService.deleteAdminCoupon === 'function', 'exports deleteAdminCoupon');
    assert(typeof OpsService.toggleAdminCouponActive === 'function', 'exports toggleAdminCouponActive');
    assert(typeof OpsService.listCouponUsages === 'function', 'exports listCouponUsages');

    // Reviews
    assert(typeof OpsService.listAdminReviews === 'function', 'exports listAdminReviews');
    assert(typeof OpsService.setReviewVisibility === 'function', 'exports setReviewVisibility');

    // Customers / Users
    assert(typeof OpsService.listAdminUsers === 'function', 'exports listAdminUsers');
    assert(typeof OpsService.getAdminUserById === 'function', 'exports getAdminUserById');
    assert(typeof OpsService.addAdminCustomer === 'function', 'exports addAdminCustomer');
    assert(typeof OpsService.blockAdminUser === 'function', 'exports blockAdminUser');
    assert(typeof OpsService.unblockAdminUser === 'function', 'exports unblockAdminUser');
    assert(typeof OpsService.updateAdminCustomer === 'function', 'exports updateAdminCustomer');
    assert(typeof OpsService.deleteAdminCustomer === 'function', 'exports deleteAdminCustomer');
    assert(typeof OpsService.getUserWallet === 'function', 'exports getUserWallet');

    // Reports & Exports
    assert(typeof OpsService.getSalesReport === 'function', 'exports getSalesReport');
    assert(typeof OpsService.getOrdersReport === 'function', 'exports getOrdersReport');
    assert(typeof OpsService.getProductsReport === 'function', 'exports getProductsReport');
    assert(typeof OpsService.downloadReportFile === 'function', 'exports downloadReportFile');

    // Settings
    assert(typeof OpsService.getReferralSettings === 'function', 'exports getReferralSettings');
    assert(typeof OpsService.updateReferralSettings === 'function', 'exports updateReferralSettings');

    // Also verify re-exports in adminService
    const AdminService = await vite.ssrLoadModule('/src/api/adminService.js');
    assert(typeof AdminService.listAdminOrders === 'function', 'adminService re-exports listAdminOrders');
    assert(typeof AdminService.createShipment === 'function', 'adminService re-exports createShipment');
    assert(typeof AdminService.listReturnRequests === 'function', 'adminService re-exports listReturnRequests');
    assert(typeof AdminService.listAdminCoupons === 'function', 'adminService re-exports listAdminCoupons');
    assert(typeof AdminService.listAdminReviews === 'function', 'adminService re-exports listAdminReviews');
    assert(typeof AdminService.listAdminUsers === 'function', 'adminService re-exports listAdminUsers');
    assert(typeof AdminService.getSalesReport === 'function', 'adminService re-exports getSalesReport');
    assert(typeof AdminService.getReferralSettings === 'function', 'adminService re-exports getReferralSettings');

    // 2. HTTP Endpoint Mappings & Payloads
    console.log('\n[2/5] Testing HTTP Endpoint Mappings and Request Options...');
    const { apiClient } = await vite.ssrLoadModule('/src/api/client.js');
    let lastRequest = null;
    function mockApi(method, mockData) {
      const original = apiClient[method];
      apiClient[method] = async (url, ...args) => {
        lastRequest = { method, url, args };
        return { data: { success: true, data: mockData } };
      };
      return () => {
        apiClient[method] = original;
      };
    }

    // Orders Endpoint Check
    let restore = mockApi('get', { orders: [], total_count: 0 });
    await OpsService.listAdminOrders({ page: 2, limit: 15, search: 'ORD-101' });
    assert(lastRequest.url === '/admin/orders', 'listAdminOrders calls /admin/orders');
    assert(lastRequest.args[0]?.params?.search === 'ORD-101', 'passes search param correctly');
    assert(lastRequest.args[0]?.requiresAuth === true, 'sets requiresAuth: true on listAdminOrders');
    restore();

    // Order Status Update
    restore = mockApi('patch', { status: 'shipped' });
    await OpsService.updateOrderStatus(99, 'shipped', 'Express air delivery');
    assert(lastRequest.url === '/admin/orders/99/status', 'updateOrderStatus calls /admin/orders/:id/status');
    assert(lastRequest.args[0]?.status === 'shipped', 'passes target status correctly');
    assert(lastRequest.args[0]?.note === 'Express air delivery', 'passes transition note correctly');
    restore();

    // COD Payment Update
    restore = mockApi('patch', { status: 'paid' });
    await OpsService.updateCODPaymentStatus(99, 'paid');
    assert(lastRequest.url === '/admin/orders/99/payment-status', 'updateCODPaymentStatus calls /admin/orders/:id/payment-status');
    assert(lastRequest.args[0]?.status === 'paid', 'passes status: paid');
    restore();

    // Wallet Refund
    restore = mockApi('post', { success: true });
    await OpsService.processWalletRefund(99);
    assert(lastRequest.url === '/admin/orders/99/refund', 'processWalletRefund calls /admin/orders/:id/refund');
    assert(lastRequest.args[1]?.requiresAuth === true, 'sets requiresAuth: true on processWalletRefund');
    restore();

    // Shipment creation
    restore = mockApi('post', { id: 10 });
    await OpsService.createShipment({ order_id: 99, carrier: 'BlueDart', tracking_number: 'BD123' });
    assert(lastRequest.url === '/admin/shipments', 'createShipment calls /admin/shipments');
    assert(lastRequest.args[0]?.carrier === 'BlueDart', 'passes carrier correctly');
    restore();

    // Shipment status update
    restore = mockApi('patch', { status: 'delivered' });
    await OpsService.updateShippingStatus(10, { status: 'delivered', location: 'Delhi Hub' });
    assert(lastRequest.url === '/admin/shipments/10/status', 'updateShippingStatus calls /admin/shipments/:id/status');
    assert(lastRequest.args[0]?.location === 'Delhi Hub', 'passes location correctly');
    restore();

    // Return request status
    restore = mockApi('patch', { status: 'approved' });
    await OpsService.updateReturnRequestStatus(5, { status: 'approved', admin_note: 'Approved for refund' });
    assert(lastRequest.url === '/admin/return-requests/5/status', 'updateReturnRequestStatus calls /admin/return-requests/:id/status');
    restore();

    // Exchange order creation
    restore = mockApi('post', { order_id: 105 });
    await OpsService.createExchangeOrder(5);
    assert(lastRequest.url === '/admin/return-requests/5/exchange', 'createExchangeOrder calls /admin/return-requests/:id/exchange');
    restore();

    // Return shipment creation
    restore = mockApi('post', { shipment_id: 20 });
    await OpsService.createReturnShipment(5);
    assert(lastRequest.url === '/admin/return-requests/5/return-shipment', 'createReturnShipment calls /admin/return-requests/:id/return-shipment');
    restore();

    // Coupon Toggle
    restore = mockApi('patch', { is_active: false });
    await OpsService.toggleAdminCouponActive(12);
    assert(lastRequest.url === '/admin/coupons/12/toggle', 'toggleAdminCouponActive calls /admin/coupons/:id/toggle');
    restore();

    // Coupon Usages
    restore = mockApi('get', { usages: [] });
    await OpsService.listCouponUsages(12);
    assert(lastRequest.url === '/admin/coupons/12/usages', 'listCouponUsages calls /admin/coupons/:id/usages');
    restore();

    // Review Visibility Moderation
    restore = mockApi('patch', { is_visible: false });
    await OpsService.setReviewVisibility(45, false);
    assert(lastRequest.url === '/admin/reviews/45/visibility', 'setReviewVisibility calls /admin/reviews/:id/visibility');
    assert(lastRequest.args[0]?.is_visible === false, 'passes is_visible: false correctly');
    restore();

    // Customer Block & Unblock
    restore = mockApi('patch', { is_blocked: true });
    await OpsService.blockAdminUser(77);
    assert(lastRequest.url === '/admin/users/77/block', 'blockAdminUser calls /admin/users/:id/block');
    restore();

    restore = mockApi('patch', { is_blocked: false });
    await OpsService.unblockAdminUser(77);
    assert(lastRequest.url === '/admin/users/77/unblock', 'unblockAdminUser calls /admin/users/:id/unblock');
    restore();

    // Customer Wallet lookup
    restore = mockApi('get', { balance: 1500, transactions: [] });
    await OpsService.getUserWallet(77);
    assert(lastRequest.url === '/admin/wallets/77', 'getUserWallet calls /admin/wallets/:userId');
    restore();

    // Reports
    restore = mockApi('get', { sales: [], total_revenue: 50000 });
    await OpsService.getSalesReport({ group_by: 'week', start_date: '2026-01-01', end_date: '2026-01-31' });
    assert(lastRequest.url === '/admin/reports/sales', 'getSalesReport calls /admin/reports/sales');
    assert(lastRequest.args[0]?.params?.group_by === 'week', 'passes group_by param');
    restore();

    restore = mockApi('get', { orders: [] });
    await OpsService.getOrdersReport({ group_by: 'month' });
    assert(lastRequest.url === '/admin/reports/orders', 'getOrdersReport calls /admin/reports/orders');
    restore();

    restore = mockApi('get', { products: [] });
    await OpsService.getProductsReport();
    assert(lastRequest.url === '/admin/reports/products', 'getProductsReport calls /admin/reports/products');
    restore();

    // Referral Settings
    restore = mockApi('get', { referrer_amount: 500, referee_amount: 250 });
    await OpsService.getReferralSettings();
    assert(lastRequest.url === '/admin/referrals/settings', 'getReferralSettings calls /admin/referrals/settings');
    restore();

    restore = mockApi('put', { referrer_amount: 600 });
    await OpsService.updateReferralSettings({ referrer_amount: 600, referee_amount: 300 });
    assert(lastRequest.url === '/admin/referrals/settings', 'updateReferralSettings calls PUT /admin/referrals/settings');
    assert(lastRequest.args[0]?.referrer_amount === 600, 'passes payload to updateReferralSettings');
    restore();

    // 3. UI Component Renders (SSR)
    console.log('\n[3/5] Testing Phase 9 Page Component Rendering...');
    const { AuthProvider } = await vite.ssrLoadModule('/src/context/AuthContext.jsx');
    const { ToastProvider } = await vite.ssrLoadModule('/src/context/ToastContext.jsx');

    function renderWithContext(element) {
      return renderToString(
        React.createElement(
          MemoryRouter,
          null,
          React.createElement(
            AuthProvider,
            null,
            React.createElement(ToastProvider, null, element)
          )
        )
      );
    }

    // AdminOrdersPage
    const { AdminOrdersPage } = await vite.ssrLoadModule('/src/pages/admin/orders/AdminOrdersPage.jsx');
    const ordersHtml = renderWithContext(React.createElement(AdminOrdersPage));
    assert(ordersHtml.includes('Orders Management'), 'AdminOrdersPage renders heading');
    assert(ordersHtml.includes('Fulfillment &amp; Operations'), 'AdminOrdersPage renders eyebrow');
    assert(ordersHtml.includes('All Order Statuses'), 'AdminOrdersPage renders status dropdown');

    // AdminOrderDetailsPage
    const { AdminOrderDetailsPage } = await vite.ssrLoadModule('/src/pages/admin/orders/AdminOrderDetailsPage.jsx');
    const orderDetailsHtml = renderWithContext(React.createElement(AdminOrderDetailsPage));
    assert(orderDetailsHtml.includes('admin-order-details-page'), 'AdminOrderDetailsPage renders skeleton/container');

    // AdminPaymentsPage
    const { AdminPaymentsPage } = await vite.ssrLoadModule('/src/pages/admin/payments/AdminPaymentsPage.jsx');
    const paymentsHtml = renderWithContext(React.createElement(AdminPaymentsPage));
    assert(paymentsHtml.includes('Payment Management'), 'AdminPaymentsPage renders heading');
    assert(paymentsHtml.includes('All Transactions'), 'AdminPaymentsPage renders transactions tab');
    assert(paymentsHtml.includes('Refund Pending Orders'), 'AdminPaymentsPage renders refund pending tab');

    // AdminReturnsPage
    const { AdminReturnsPage } = await vite.ssrLoadModule('/src/pages/admin/returns/AdminReturnsPage.jsx');
    const returnsHtml = renderWithContext(React.createElement(AdminReturnsPage));
    assert(returnsHtml.includes('Returns &amp; Exchanges'), 'AdminReturnsPage renders heading');
    assert(returnsHtml.includes('Customer Care &amp; Reverse Logistics'), 'AdminReturnsPage renders eyebrow');

    // AdminCouponsPage
    const { AdminCouponsPage } = await vite.ssrLoadModule('/src/pages/admin/coupons/AdminCouponsPage.jsx');
    const couponsHtml = renderWithContext(React.createElement(AdminCouponsPage));
    assert(couponsHtml.includes('Coupons &amp; Campaigns'), 'AdminCouponsPage renders heading');
    assert(couponsHtml.includes('Create Coupon'), 'AdminCouponsPage renders create button');

    // AdminReviewsPage
    const { AdminReviewsPage } = await vite.ssrLoadModule('/src/pages/admin/reviews/AdminReviewsPage.jsx');
    const reviewsHtml = renderWithContext(React.createElement(AdminReviewsPage));
    assert(reviewsHtml.includes('Reviews &amp; Feedback'), 'AdminReviewsPage renders heading');
    assert(reviewsHtml.includes('Review Moderation Policy'), 'AdminReviewsPage renders domain rule policy banner');

    // AdminCustomersPage
    const { AdminCustomersPage } = await vite.ssrLoadModule('/src/pages/admin/customers/AdminCustomersPage.jsx');
    const customersHtml = renderWithContext(React.createElement(AdminCustomersPage));
    assert(customersHtml.includes('Customer Management'), 'AdminCustomersPage renders heading');
    assert(customersHtml.includes('Add Customer'), 'AdminCustomersPage renders add customer button');

    // AdminReportsPage
    const { AdminReportsPage } = await vite.ssrLoadModule('/src/pages/admin/reports/AdminReportsPage.jsx');
    const reportsHtml = renderWithContext(React.createElement(AdminReportsPage));
    assert(reportsHtml.includes('Executive Reports &amp; Exports'), 'AdminReportsPage renders heading');
    assert(reportsHtml.includes('Sales &amp; Revenue'), 'AdminReportsPage renders sales tab');
    assert(reportsHtml.includes('Export Excel'), 'AdminReportsPage renders Excel export action');

    // AdminSettingsPage
    const { AdminSettingsPage } = await vite.ssrLoadModule('/src/pages/admin/settings/AdminSettingsPage.jsx');
    const settingsHtml = renderWithContext(React.createElement(AdminSettingsPage));
    assert(settingsHtml.includes('admin-settings-page'), 'AdminSettingsPage renders container');

    // 4. Domain Rules & Business Validation
    console.log('\n[4/5] Testing Domain Rules Verification...');
    // Verify Review Deletion restriction in Admin: no deleteAdminReview should exist
    assert(typeof OpsService.deleteAdminReview === 'undefined', 'Admin cannot delete reviews per domain policy (visibility moderation only)');
    // Verify Wallet refund is direct POST /admin/orders/:id/refund without inventing third-party Razorpay refund endpoints
    assert(typeof OpsService.refundRazorpay === 'undefined', 'No artificial gateway refund invented; uses direct wallet credit');

    // 5. Route Integration Check
    console.log('\n[5/5] Testing Route Integration in AppRoutes.jsx...');
    const { setAccessToken, clearAccessToken } = await vite.ssrLoadModule('/src/api/tokenStore.js');
    setAccessToken(createMockToken('admin'));

    const { AppRoutes } = await vite.ssrLoadModule('/src/routes/AppRoutes.jsx');

    function renderAdminRoute(path) {
      return renderToString(
        React.createElement(
          MemoryRouter,
          { initialEntries: [path] },
          React.createElement(
            ToastProvider,
            null,
            React.createElement(AuthProvider, null, React.createElement(AppRoutes, null))
          )
        )
      );
    }

    const ordersRouteHtml = renderAdminRoute('/admin/orders');
    assert(ordersRouteHtml.includes('Orders Management'), 'AppRoutes resolves /admin/orders');

    const paymentsRouteHtml = renderAdminRoute('/admin/payments');
    assert(paymentsRouteHtml.includes('Payment Management'), 'AppRoutes resolves /admin/payments');

    const returnsRouteHtml = renderAdminRoute('/admin/returns');
    assert(returnsRouteHtml.includes('Returns &amp; Exchanges'), 'AppRoutes resolves /admin/returns');

    const couponsRouteHtml = renderAdminRoute('/admin/coupons');
    assert(couponsRouteHtml.includes('Coupons &amp; Campaigns'), 'AppRoutes resolves /admin/coupons');

    const reviewsRouteHtml = renderAdminRoute('/admin/reviews');
    assert(reviewsRouteHtml.includes('Reviews &amp; Feedback'), 'AppRoutes resolves /admin/reviews');

    const customersRouteHtml = renderAdminRoute('/admin/customers');
    assert(customersRouteHtml.includes('Customer Management'), 'AppRoutes resolves /admin/customers');

    const reportsRouteHtml = renderAdminRoute('/admin/reports');
    assert(reportsRouteHtml.includes('Executive Reports &amp; Exports'), 'AppRoutes resolves /admin/reports');

    const settingsRouteHtml = renderAdminRoute('/admin/settings');
    assert(settingsRouteHtml.includes('admin-settings-page') || settingsRouteHtml.includes('Settings'), 'AppRoutes resolves /admin/settings');

    clearAccessToken();

    // Summary
    const total = assertions.length;
    const passed = assertions.filter((a) => a.pass).length;
    const failed = total - passed;

    console.log('\n====================================================');
    console.log(`PHASE 9 TEST SUMMARY: ${passed}/${total} PASSED (${failed} FAILED)`);
    console.log('====================================================\n');

    if (failed > 0) {
      process.exit(1);
    }
  } finally {
    await vite.close();
  }
}

run().catch((err) => {
  console.error('Fatal error running Phase 9 tests:', err);
  process.exit(1);
});
