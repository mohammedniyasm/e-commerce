import { createServer } from 'vite';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import fs from 'node:fs';

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

    const {
      canMarkCodPaymentPaid,
      canProcessWalletRefund,
    } = await vite.ssrLoadModule('/src/utils/orderEligibility.js');
    assert(
      canMarkCodPaymentPaid({
        order_status: 'confirmed',
        payment_method: 'cod',
        payment_status: 'pending',
      }) &&
        !canMarkCodPaymentPaid({
          order_status: 'cancelled',
          payment_method: 'cod',
          payment_status: 'pending',
        }),
      'COD payment can only be marked paid on non-cancelled, pending orders',
    );
    assert(
      canProcessWalletRefund({ order_status: 'cancelled', payment_status: 'paid' }),
      'Wallet refund is available for paid cancelled orders',
    );
    assert(
      !canProcessWalletRefund({ order_status: 'cancelled', payment_status: 'pending' }) &&
        !canProcessWalletRefund({ order_status: 'cancelled', payment_status: 'failed' }) &&
        !canProcessWalletRefund({ order_status: 'returned', payment_status: 'paid' }),
      'Wallet refund is hidden for unpaid cancelled orders and unsupported returned orders',
    );

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
    assert(typeof OpsService.listAdminShipments === 'function', 'exports listAdminShipments');
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
    assert(typeof OpsService.getAdminReferralDashboard === 'function', 'exports getAdminReferralDashboard');
    assert(typeof OpsService.getTopAdminReferrers === 'function', 'exports getTopAdminReferrers');
    assert(typeof OpsService.listAdminReferrals === 'function', 'exports listAdminReferrals');

    // Also verify re-exports in adminService
    const AdminService = await vite.ssrLoadModule('/src/api/adminService.js');
    assert(typeof AdminService.listAdminOrders === 'function', 'adminService re-exports listAdminOrders');
    assert(typeof AdminService.createShipment === 'function', 'adminService re-exports createShipment');
    assert(typeof AdminService.listAdminShipments === 'function', 'adminService re-exports listAdminShipments');
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
    restore = mockApi('get', {
      shipments: [],
      total_count: 0,
      page: 2,
      limit: 25,
      total_pages: 0,
    });
    await OpsService.listAdminShipments({
      page: 2,
      limit: 25,
      search: 'BD-123',
      shipping_status: 'in_transit',
      type: 'delivery',
      courier_name: 'Blue Dart',
      order_id: '123e4567-e89b-12d3-a456-426614174000',
      return_request_id: 7,
      created_from: '2026-10-01',
      created_to: '2026-10-31',
      sort: 'oldest',
    });
    assert(lastRequest.url === '/admin/shipments', 'listAdminShipments calls /admin/shipments');
    assert(lastRequest.args[0]?.params?.shipping_status === 'in_transit', 'passes shipment status filter');
    assert(lastRequest.args[0]?.params?.return_request_id === 7, 'passes return request filter');
    assert(lastRequest.args[0]?.params?.sort === 'oldest', 'passes shipment sort order');
    assert(lastRequest.args[0]?.requiresAuth === true, 'requires admin authentication for shipment listing');
    restore();

    // Shipment creation
    restore = mockApi('post', { id: 10 });
    await OpsService.createShipment({
      order_id: '123e4567-e89b-12d3-a456-426614174000',
      courier_name: 'BlueDart',
      tracking_number: 'BD123',
    });
    assert(lastRequest.url === '/admin/shipments', 'createShipment calls /admin/shipments');
    assert(lastRequest.args[0]?.courier_name === 'BlueDart', 'passes courier_name from the shipment contract');
    assert(
      lastRequest.args[0]?.order_id === '123e4567-e89b-12d3-a456-426614174000',
      'preserves the order UUID when creating a shipment'
    );
    restore();

    // Shipment tracking details
    restore = mockApi('patch', { id: 10 });
    await OpsService.updateShipment(10, { courier_name: 'BlueDart', tracking_number: 'BD123' });
    assert(lastRequest.url === '/admin/shipments/10', 'updateShipment calls /admin/shipments/:id');
    assert(lastRequest.args[0]?.tracking_number === 'BD123', 'passes updated tracking details');
    restore();

    // Shipment status update
    restore = mockApi('patch', { status: 'delivered' });
    await OpsService.updateShippingStatus(10, {
      status: 'packed',
      location: 'Delhi Hub',
      description: 'Package prepared for dispatch',
    });
    assert(lastRequest.url === '/admin/shipments/10/status', 'updateShippingStatus calls /admin/shipments/:id/status');
    assert(lastRequest.args[0]?.location === 'Delhi Hub', 'passes location correctly');
    assert(lastRequest.args[0]?.description === 'Package prepared for dispatch', 'passes status description correctly');
    restore();

    // Return request status
    restore = mockApi('patch', { status: 'pickup_scheduled' });
    await OpsService.updateReturnRequestStatus(5, {
      status: 'pickup_scheduled',
      admin_note: 'Pickup arranged',
      courier_name: 'BlueDart',
      tracking_number: 'RET-123',
    });
    assert(lastRequest.url === '/admin/return-requests/5/status', 'updateReturnRequestStatus calls /admin/return-requests/:id/status');
    assert(lastRequest.args[0]?.courier_name === 'BlueDart', 'passes return pickup courier name');
    assert(lastRequest.args[0]?.tracking_number === 'RET-123', 'passes return pickup tracking number');
    restore();

    // Exchange order creation
    restore = mockApi('post', { order_id: 105 });
    await OpsService.createExchangeOrder(5);
    assert(lastRequest.url === '/admin/return-requests/5/exchange', 'createExchangeOrder calls /admin/return-requests/:id/exchange');
    restore();

    // Return shipment creation
    restore = mockApi('post', { shipment_id: 20 });
    await OpsService.createReturnShipment(5, {
      courier_name: 'BlueDart',
      tracking_number: 'RET-123',
    });
    assert(lastRequest.url === '/admin/return-requests/5/return-shipment', 'createReturnShipment calls /admin/return-requests/:id/return-shipment');
    assert(lastRequest.args[0]?.courier_name === 'BlueDart', 'passes return shipment courier name');
    assert(lastRequest.args[0]?.tracking_number === 'RET-123', 'passes return shipment tracking number');
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
    restore = mockApi('get', {
      summary: { total_sales: 50000, total_orders: 4, total_units_sold: 6 },
      breakdown: [{ date: '2026-01-01', sales: 12000, orders: 1, units_sold: 2 }],
    });
    const salesReport = await OpsService.getSalesReport({ group_by: 'week', start_date: '2026-01-01', end_date: '2026-01-31' });
    assert(lastRequest.url === '/admin/reports/sales', 'getSalesReport calls /admin/reports/sales');
    assert(lastRequest.args[0]?.params?.group_by === 'week', 'passes group_by param');
    assert(
      salesReport.summary.total_sales === 50000 && salesReport.breakdown[0].orders === 1,
      'getSalesReport preserves summary and breakdown data',
    );
    restore();

    restore = mockApi('get', {
      summary: { total_orders: 5, delivered_orders: 3, cancelled_orders: 1 },
      breakdown: [{ date: '2026-01-01', total_orders: 2 }],
    });
    const ordersReport = await OpsService.getOrdersReport({ group_by: 'month' });
    assert(lastRequest.url === '/admin/reports/orders', 'getOrdersReport calls /admin/reports/orders');
    assert(
      ordersReport.summary.delivered_orders === 3 && ordersReport.breakdown[0].total_orders === 2,
      'getOrdersReport preserves summary and breakdown data',
    );
    restore();

    restore = mockApi('get', {
      summary: { total_products_sold: 2, total_units_sold: 6, total_revenue: 18000 },
      breakdown: [{ product_id: 7, product_name: 'Sample', units_sold: 3, revenue: 9000 }],
    });
    const productsReport = await OpsService.getProductsReport();
    assert(lastRequest.url === '/admin/reports/products', 'getProductsReport calls /admin/reports/products');
    assert(
      productsReport.summary.total_products_sold === 2 && productsReport.breakdown[0].product_id === 7,
      'getProductsReport preserves summary and breakdown data',
    );
    restore();

    // Referral Settings
    const referralSettings = {
      referrer_reward: 500,
      referred_reward: 250,
      minimum_order_amount: 2000,
      is_active: true,
    };
    restore = mockApi('get', referralSettings);
    const fetchedReferralSettings = await OpsService.getReferralSettings();
    assert(lastRequest.url === '/admin/referrals/settings', 'getReferralSettings calls /admin/referrals/settings');
    assert(
      fetchedReferralSettings.referrer_reward === 500 &&
        fetchedReferralSettings.referred_reward === 250 &&
        fetchedReferralSettings.minimum_order_amount === 2000,
      'getReferralSettings returns fields matching the backend response',
    );
    restore();

    restore = mockApi('get', {
      total_referrals: 12,
      pending_referrals: 4,
      completed_referrals: 8,
      total_rewards_paid: 6000,
    });
    const referralDashboard = await OpsService.getAdminReferralDashboard();
    assert(lastRequest.url === '/admin/referrals/dashboard', 'getAdminReferralDashboard calls admin dashboard endpoint');
    assert(lastRequest.args[0]?.requiresAuth === true, 'admin referral dashboard requires authentication');
    assert(referralDashboard.total_referrals === 12, 'returns admin referral summary');
    restore();

    restore = mockApi('get', { referrers: [{ referrer_id: 7, name: 'Customer' }] });
    const topReferrers = await OpsService.getTopAdminReferrers(5);
    assert(lastRequest.url === '/admin/referrals/top-referrers', 'getTopAdminReferrers calls top-referrers endpoint');
    assert(lastRequest.args[0]?.params?.limit === 5, 'passes top-referrers limit');
    assert(topReferrers[0].referrer_id === 7, 'returns top referrer records');
    restore();

    restore = mockApi('get', {
      referrals: [{ id: 21, referrer_id: 7 }],
      pagination: { total: 1, limit: 10, offset: 0 },
    });
    const adminReferrals = await OpsService.listAdminReferrals({
      page: 2,
      limit: 10,
      status: 'completed',
      reward_status: 'pending',
      referrer_id: '7',
      start_date: '2026-01-01',
      end_date: '2026-01-31',
    });
    assert(lastRequest.url === '/admin/referrals', 'listAdminReferrals calls /admin/referrals');
    assert(
      lastRequest.args[0]?.params?.page === 2 &&
        lastRequest.args[0]?.params?.status === 'completed' &&
        lastRequest.args[0]?.params?.reward_status === 'pending' &&
        lastRequest.args[0]?.params?.referrer_id === '7',
      'passes referral filters and pagination',
    );
    assert(adminReferrals.referrals[0].id === 21, 'returns referral records and pagination');
    restore();

    const referralUpdate = {
      referrer_reward: 600,
      referred_reward: 300,
      minimum_order_amount: 2500,
      is_active: false,
    };
    restore = mockApi('put', referralSettings);
    await OpsService.updateReferralSettings(referralUpdate);
    assert(lastRequest.url === '/admin/referrals/settings', 'updateReferralSettings calls PUT /admin/referrals/settings');
    assert(
      JSON.stringify(lastRequest.args[0]) === JSON.stringify(referralUpdate),
      'passes backend-compatible referral settings payload',
    );
    restore();

    const adminSettingsSource = fs.readFileSync(
      new URL('../src/pages/admin/settings/AdminSettingsPage.jsx', import.meta.url),
      'utf8',
    );
    assert(
      adminSettingsSource.includes('ref.referrer_reward') &&
        adminSettingsSource.includes('ref.referred_reward') &&
        adminSettingsSource.includes('ref.minimum_order_amount') &&
        adminSettingsSource.includes('referrer_reward: parseFloat(referrerBonus)') &&
        adminSettingsSource.includes('referred_reward: parseFloat(refereeBonus)') &&
        adminSettingsSource.includes('minimum_order_amount: parseFloat(minOrderToReward)'),
      'AdminSettingsPage reads and submits the backend referral settings fields',
    );

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

    const { AdminReferralsPage } = await vite.ssrLoadModule('/src/pages/admin/referrals/AdminReferralsPage.jsx');
    const referralsHtml = renderWithContext(React.createElement(AdminReferralsPage));
    assert(referralsHtml.includes('Referral Program'), 'AdminReferralsPage renders heading');
    assert(referralsHtml.includes('Top Referrers'), 'AdminReferralsPage renders top-referrers section');
    assert(referralsHtml.includes('All Referrals'), 'AdminReferralsPage renders referral records section');
    assert(referralsHtml.includes('Program Settings'), 'AdminReferralsPage links to referral configuration');

    // AdminOrdersPage
    const { AdminOrdersPage } = await vite.ssrLoadModule('/src/pages/admin/orders/AdminOrdersPage.jsx');
    const ordersHtml = renderWithContext(React.createElement(AdminOrdersPage));
    assert(ordersHtml.includes('Orders Management'), 'AdminOrdersPage renders heading');
    assert(ordersHtml.includes('Fulfillment &amp; Operations'), 'AdminOrdersPage renders eyebrow');
    assert(ordersHtml.includes('All Order Statuses'), 'AdminOrdersPage renders status dropdown');
    const ordersPageSource = fs.readFileSync(
      new URL('../src/pages/admin/orders/AdminOrdersPage.jsx', import.meta.url),
      'utf8',
    );
    assert(
      ordersPageSource.includes("updateCODPaymentStatus(orderId, 'paid')") &&
        ordersPageSource.includes('canMarkCodPaymentPaid(item)'),
      'AdminOrdersPage exposes the supported pending COD payment update',
    );

    // AdminOrderDetailsPage
    const { AdminOrderDetailsPage } = await vite.ssrLoadModule('/src/pages/admin/orders/AdminOrderDetailsPage.jsx');
    const orderDetailsHtml = renderWithContext(React.createElement(AdminOrderDetailsPage));
    assert(orderDetailsHtml.includes('admin-order-details-page'), 'AdminOrderDetailsPage renders skeleton/container');
    const orderDetailsSource = fs.readFileSync(
      new URL('../src/pages/admin/orders/AdminOrderDetailsPage.jsx', import.meta.url),
      'utf8',
    );
    assert(
      orderDetailsSource.includes('const shippingAddress = order.address || order.shipping_address;') &&
        orderDetailsSource.includes('item.image_url || item.product?.thumbnail') &&
        orderDetailsSource.includes('item.variant_info') &&
        orderDetailsSource.includes('shippingAddress.country') &&
        orderDetailsSource.includes('order.user?.email'),
      'Admin order details use the backend address, image, and variant fields',
    );

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
    assert(returnsHtml.includes('value="requested"'), 'AdminReturnsPage filters by backend requested status');
    assert(returnsHtml.includes('value="inspection_failed"'), 'AdminReturnsPage filters by backend inspection status');
    const returnsSource = fs.readFileSync(
      new URL('../src/pages/admin/returns/AdminReturnsPage.jsx', import.meta.url),
      'utf8',
    );
    assert(
      returnsSource.includes('openDetailsModal(req)') &&
        returnsSource.includes('Requested Item') &&
        returnsSource.includes('Requested Exchange Variant') &&
        returnsSource.includes('variant.image_url') &&
        returnsSource.includes('RequestedExchangeVariantDetails') &&
        returnsSource.includes("normalizeStatus(selectedRequest.status) === 'inspection_pending'") &&
        returnsSource.includes('Item Received for Inspection') &&
        returnsSource.includes('Inspection Findings / Admin Notes') &&
        !returnsSource.includes("inspection_failed: ['closed']") &&
        returnsSource.includes('const req = record || {};') &&
        returnsSource.includes('const orderId = req.order_id || req.order?.id;') &&
        returnsSource.includes('const orderLabel ='),
      'AdminReturnsPage exposes original item and exchange variant details',
    );

    // AdminCouponsPage
    const { AdminCouponsPage } = await vite.ssrLoadModule('/src/pages/admin/coupons/AdminCouponsPage.jsx');
    const couponsHtml = renderWithContext(React.createElement(AdminCouponsPage));
    assert(couponsHtml.includes('Coupons &amp; Campaigns'), 'AdminCouponsPage renders heading');
    assert(couponsHtml.includes('Create Coupon'), 'AdminCouponsPage renders create button');
    const couponsSource = fs.readFileSync(
      new URL('../src/pages/admin/coupons/AdminCouponsPage.jsx', import.meta.url),
      'utf8',
    );
    assert(
      couponsSource.includes('leftIcon={<Edit2 size={16} />}') &&
        couponsSource.includes('leftIcon={<Power size={16} />}') &&
        couponsSource.includes('leftIcon={<Trash2 size={16} />}'),
      'AdminCouponsPage renders icons for coupon row actions',
    );
    assert(
      couponsSource.includes('minimum_order_amount: minOrderAmount') &&
        couponsSource.includes('maximum_discount: maxDiscountAmount') &&
        couponsSource.includes('usage_limit_per_user:') &&
        couponsSource.includes('start_at: new Date(validFrom)') &&
        couponsSource.includes('expires_at: new Date(validUntil)') &&
        couponsSource.includes('coupon.minimum_order_amount') &&
        couponsSource.includes('coupon.maximum_discount') &&
        couponsSource.includes('coupon.usage_limit_per_user') &&
        couponsSource.includes('coupon.start_at') &&
        couponsSource.includes('coupon.expires_at'),
      'AdminCouponsPage reads and submits fields matching the coupon API',
    );

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
    const reportsSource = fs.readFileSync(
      new URL('../src/pages/admin/reports/AdminReportsPage.jsx', import.meta.url),
      'utf8',
    );
    assert(
      reportsSource.includes('Array.isArray(reportData?.breakdown)') &&
        reportsSource.includes('summary.total_sales') &&
        reportsSource.includes('summary.total_units_sold') &&
        reportsSource.includes("key: 'orders'") &&
        reportsSource.includes("key: 'sales'") &&
        reportsSource.includes("key: 'product_id'") &&
        reportsSource.includes('summary.total_revenue'),
      'AdminReportsPage renders report summaries and breakdowns from the backend schema',
    );

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
    const { getShipmentStatusTransitions } = await vite.ssrLoadModule('/src/utils/shipmentStatus.js');

    assert(
      getShipmentStatusTransitions('pending').includes('packed')
        && !getShipmentStatusTransitions('pending').includes('in_transit'),
      'Shipment status options follow allowed backend transitions'
    );
    assert(
      getShipmentStatusTransitions('delivered').length === 0,
      'Terminal shipment statuses cannot transition further'
    );

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

    const shipmentsRouteHtml = renderAdminRoute('/admin/shipments');
    assert(shipmentsRouteHtml.includes('Shipment Management'), 'AppRoutes resolves /admin/shipments');

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

    const referralsRouteHtml = renderAdminRoute('/admin/referrals');
    assert(referralsRouteHtml.includes('Referral Program'), 'AppRoutes resolves /admin/referrals');

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
