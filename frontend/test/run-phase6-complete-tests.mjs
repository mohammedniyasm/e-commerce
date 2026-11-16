import { createServer } from 'vite';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter, Routes, Route, Outlet } from 'react-router-dom';
import fs from 'node:fs';

async function run() {
  console.log('====================================================');
  console.log('PHASE 6: COMPLETE SUITE — CUSTOMER ACCOUNT MANAGEMENT');
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
    console.log('[1/8] Testing Phase 6 API Services Contracts...');
    const ReturnService = await vite.ssrLoadModule('/src/api/returnService.js');
    assert(typeof ReturnService.createReturnRequest === 'function', 'returnService exports createReturnRequest');
    assert(typeof ReturnService.getMyReturnRequests === 'function', 'returnService exports getMyReturnRequests');
    assert(typeof ReturnService.getReturnRequestById === 'function', 'returnService exports getReturnRequestById');
    assert(typeof ReturnService.getReturnRequestShipments === 'function', 'returnService exports getReturnRequestShipments');

    const WalletService = await vite.ssrLoadModule('/src/api/walletService.js');
    assert(typeof WalletService.getWallet === 'function', 'walletService exports getWallet');
    assert(typeof WalletService.getWalletTransactions === 'function', 'walletService exports getWalletTransactions');
    assert(typeof WalletService.getTransactionById === 'function', 'walletService exports getTransactionById');
    assert(typeof WalletService.createTopUpOrder === 'function', 'walletService exports createTopUpOrder');
    assert(typeof WalletService.verifyTopUpPayment === 'function', 'walletService exports verifyTopUpPayment');

    const ReferralService = await vite.ssrLoadModule('/src/api/referralService.js');
    assert(typeof ReferralService.getReferralDashboard === 'function', 'referralService exports getReferralDashboard');

    // 2. HTTP Endpoint Mappings
    console.log('\n[2/8] Testing HTTP Endpoint Mappings & Payloads...');
    const { apiClient } = await vite.ssrLoadModule('/src/api/client.js');
    let lastRequest = null;
    function mockApi(method, mockData) {
      const original = apiClient[method];
      apiClient[method] = async (url, data, config) => {
        lastRequest = { method, url, data, config };
        return { data: { success: true, data: mockData } };
      };
      return () => {
        apiClient[method] = original;
      };
    }

    // Return request mapping
    {
      const restore = mockApi('post', { id: 10 });
      await ReturnService.createReturnRequest({
        order_id: '123e4567-e89b-12d3-a456-426614174000',
        order_item_id: 1,
        request_type: 'return',
        quantity: 2,
        reason: 'Sizing adjustment required for jacket',
      });
      assert(lastRequest?.url === '/user/return-requests', 'createReturnRequest calls POST /user/return-requests');
      assert(lastRequest?.data?.order_id === '123e4567-e89b-12d3-a456-426614174000', 'createReturnRequest transmits order_id');
      assert(lastRequest?.data?.request_type === 'return', 'createReturnRequest transmits request_type');
      assert(lastRequest?.data?.quantity === 2, 'createReturnRequest transmits quantity');
      restore();
    }

    // Wallet top-up mapping
    {
      const restore = mockApi('post', { gateway_order_id: 'order_9842', amount: 500000 });
      await WalletService.createTopUpOrder(5000);
      assert(lastRequest?.url === '/user/wallet/top-up/order', 'createTopUpOrder calls POST /user/wallet/top-up/order');
      assert(lastRequest?.data?.amount === 5000, 'createTopUpOrder transmits numeric amount');
      restore();
    }

    // Wallet verify payment mapping
    {
      const restore = mockApi('post', { success: true });
      await WalletService.verifyTopUpPayment({
        gateway_order_id: 'order_9842',
        gateway_payment_id: 'pay_9842',
        razorpay_signature: 'sig_9842',
      });
      assert(lastRequest?.url === '/user/wallet/top-up/verify', 'verifyTopUpPayment calls POST /user/wallet/top-up/verify');
      assert(lastRequest?.data?.razorpay_signature === 'sig_9842', 'verifyTopUpPayment transmits razorpay_signature matching backend DTO');
      restore();
    }

    // 3. ProfilePage SSR Rendering
    console.log('\n[3/8] Testing ProfilePage (Milestone 2) SSR Rendering...');
    const { ProfilePage } = await vite.ssrLoadModule('/src/pages/account/ProfilePage.jsx');
    const { ToastProvider } = await vite.ssrLoadModule('/src/context/ToastContext.jsx');

    function OutletWrapper({ context, children }) {
      return React.createElement('div', null, React.createElement(Outlet, { context }), children);
    }

    const mockProfileContext = {
      profile: {
        id: 1,
        name: 'Arjun Nair',
        email: 'arjun.nair@example.com',
        phone: '+91 98470 12345',
        profile_image: null,
        email_verified_at: '2026-11-10T10:00:00Z',
      },
      refreshData: () => {},
    };

    const profileHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/account/profile'] },
        React.createElement(
          ToastProvider,
          null,
          React.createElement(
            Routes,
            null,
            React.createElement(
              Route,
              {
                path: '/account',
                element: React.createElement(OutletWrapper, { context: mockProfileContext }),
              },
              React.createElement(Route, {
                path: 'profile',
                element: React.createElement(ProfilePage),
              }),
            ),
          ),
        ),
      ),
    );

    assert(profileHtml.includes('Profile Portrait'), 'ProfilePage renders Profile Portrait section');
    assert(profileHtml.includes('UPLOAD NEW PHOTO'), 'ProfilePage renders Upload Photo action');
    assert(profileHtml.includes('Personal Information'), 'ProfilePage renders Personal Information section');
    assert(profileHtml.includes('value="Arjun Nair"'), 'ProfilePage pre-populates name field');
    assert(profileHtml.includes('value="+91 98470 12345"'), 'ProfilePage pre-populates phone field');
    assert(profileHtml.includes('SAVE CHANGES'), 'ProfilePage renders Save Changes submit button');
    assert(profileHtml.includes('Email Address'), 'ProfilePage renders Email Address section');
    assert(profileHtml.includes('VERIFIED'), 'ProfilePage renders verified status badge');
    assert(profileHtml.includes('CHANGE EMAIL'), 'ProfilePage renders Change Email action');
    assert(profileHtml.includes('Password &amp; Security') || profileHtml.includes('Password & Security'), 'ProfilePage renders Password & Security section');
    assert(profileHtml.includes('CURRENT PASSWORD'), 'ProfilePage renders Current Password field');
    assert(profileHtml.includes('NEW PASSWORD'), 'ProfilePage renders New Password field');
    assert(profileHtml.includes('CONFIRM NEW PASSWORD'), 'ProfilePage renders Confirm New Password field');
    assert(profileHtml.includes('SECURITY REQUIREMENTS:'), 'ProfilePage renders password complexity requirements checklist');

    // 4. AddressesPage SSR Rendering
    console.log('\n[4/8] Testing AddressesPage (Milestone 3) SSR Rendering...');
    const { AddressesPage } = await vite.ssrLoadModule('/src/pages/account/AddressesPage.jsx');
    const mockAddressesContext = {
      addresses: [
        {
          id: 101,
          name: 'Arjun Nair',
          phone: '+91 98470 12345',
          address_line1: 'Bungalow 4, Princess Street',
          address_line_2: 'Near Heritage Hotel',
          city: 'Fort Kochi',
          state: 'Kerala',
          postal_code: '682001',
          country: 'India',
          is_default: true,
        },
        {
          id: 102,
          name: 'Arjun Nair (Studio)',
          phone: '+91 98470 99999',
          address_line1: '88 MG Road, 3rd Floor',
          city: 'Kochi',
          state: 'Kerala',
          postal_code: '682016',
          country: 'India',
          is_default: false,
        },
      ],
      isLoading: false,
      refreshData: () => {},
    };

    const addressesHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/account/addresses'] },
        React.createElement(
          ToastProvider,
          null,
          React.createElement(
            Routes,
            null,
            React.createElement(
              Route,
              {
                path: '/account',
                element: React.createElement(OutletWrapper, { context: mockAddressesContext }),
              },
              React.createElement(Route, {
                path: 'addresses',
                element: React.createElement(AddressesPage),
              }),
            ),
          ),
        ),
      ),
    );

    assert(addressesHtml.includes('Address Book'), 'AddressesPage renders header title');
    assert(addressesHtml.includes('ADD NEW ADDRESS'), 'AddressesPage renders Add New Address button');
    assert(addressesHtml.includes('Bungalow 4, Princess Street'), 'AddressesPage renders address line 1');
    assert(addressesHtml.includes('DEFAULT'), 'AddressesPage renders DEFAULT badge for primary address');
    assert(addressesHtml.includes('Edit'), 'AddressesPage renders Edit button on address card');
    assert(addressesHtml.includes('Delete'), 'AddressesPage renders Delete button on address card');
    assert(addressesHtml.includes('Set As Default'), 'AddressesPage renders Set As Default on secondary address card');

    // 5. OrdersPage SSR Rendering
    console.log('\n[5/8] Testing OrdersPage (Milestone 4) SSR Rendering...');
    const { OrdersPage } = await vite.ssrLoadModule('/src/pages/account/OrdersPage.jsx');
    const ordersHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/account/orders'] },
        React.createElement(
          ToastProvider,
          null,
          React.createElement(OrdersPage),
        ),
      ),
    );

    assert(ordersHtml.includes('My Orders'), 'OrdersPage renders My Orders heading');
    assert(ordersHtml.includes('All Orders'), 'OrdersPage renders All Orders tab');
    assert(ordersHtml.includes('In Transit'), 'OrdersPage renders In Transit filter tab');
    assert(ordersHtml.includes('Delivered'), 'OrdersPage renders Delivered filter tab');
    assert(ordersHtml.includes('Cancelled'), 'OrdersPage renders Cancelled filter tab');

    // 6. OrderDetailsPage SSR Rendering
    console.log('\n[6/8] Testing OrderDetailsPage (Milestones 5 & 6) SSR Rendering...');
    const { OrderDetailsPage } = await vite.ssrLoadModule('/src/pages/account/OrderDetailsPage.jsx');
    const orderDetailsHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/account/orders/test-id'] },
        React.createElement(
          ToastProvider,
          null,
          React.createElement(OrderDetailsPage),
        ),
      ),
    );
    assert(orderDetailsHtml.includes('account-order-details') || orderDetailsHtml.includes('order-details'), 'OrderDetailsPage renders container');

    // 7. ReturnsPage & WalletPage & ReferralPage SSR Rendering
    console.log('\n[7/8] Testing Returns, Wallet & Referral Pages (Milestones 7, 8, 9) SSR Rendering...');
    const { ReturnsPage } = await vite.ssrLoadModule('/src/pages/account/ReturnsPage.jsx');
    const returnsHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/account/returns'] },
        React.createElement(ToastProvider, null, React.createElement(ReturnsPage)),
      ),
    );
    assert(returnsHtml.includes('Returns &amp; Exchanges') || returnsHtml.includes('Returns & Exchanges'), 'ReturnsPage renders title');
    assert(returnsHtml.includes('Atelier Returns &amp; Wallet Refunds') || returnsHtml.includes('Atelier Returns & Wallet Refunds'), 'ReturnsPage renders returns guide card');

    const { AuthProvider } = await vite.ssrLoadModule('/src/context/AuthContext.jsx');
    const { WalletPage } = await vite.ssrLoadModule('/src/pages/account/WalletPage.jsx');
    const walletHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/account/wallet'] },
        React.createElement(
          ToastProvider,
          null,
          React.createElement(
            AuthProvider,
            null,
            React.createElement(WalletPage),
          ),
        ),
      ),
    );
    assert(walletHtml.includes('Atelier Wallet'), 'WalletPage renders title');
    assert(walletHtml.includes('AVAILABLE BALANCE'), 'WalletPage renders available balance card');
    assert(walletHtml.includes('TOP UP WALLET'), 'WalletPage renders top up action button');
    assert(walletHtml.includes('Transaction History'), 'WalletPage renders transaction history section');

    const { ReferralPage } = await vite.ssrLoadModule('/src/pages/account/ReferralPage.jsx');
    const referralHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/account/referrals'] },
        React.createElement(
          ToastProvider,
          null,
          React.createElement(ReferralPage),
        ),
      ),
    );
    assert(referralHtml.includes('Referral Program'), 'ReferralPage renders title');
    assert(referralHtml.includes('Share Drape with Friends'), 'ReferralPage renders referral hero title');
    assert(referralHtml.includes('COPY CODE'), 'ReferralPage renders Copy Code button');
    assert(referralHtml.includes('COPY INVITATION LINK'), 'ReferralPage renders Copy Invitation Link button');
    assert(referralHtml.includes('TOTAL REFERRALS'), 'ReferralPage renders Total Referrals metric');
    assert(referralHtml.includes('TOTAL REWARDS EARNED'), 'ReferralPage renders Total Rewards metric');
    assert(referralHtml.includes('How The Program Works'), 'ReferralPage renders how program works section');

    // 8. Verification that No Placeholder Pages Remain in AppRoutes
    console.log('\n[8/8] Verifying AppRoutes Route Integration & Zero Placeholders...');
    const appRoutesSource = fs.readFileSync(new URL('../src/routes/AppRoutes.jsx', import.meta.url), 'utf8');

    assert(!appRoutesSource.includes('AccountPlaceholderPage'), 'AppRoutes contains NO AccountPlaceholderPage occurrences');
    assert(appRoutesSource.includes('path="profile" element={<ProfilePage />}'), 'AppRoutes registers ProfilePage');
    assert(appRoutesSource.includes('path="addresses" element={<AddressesPage />}'), 'AppRoutes registers AddressesPage');
    assert(appRoutesSource.includes('path="orders" element={<OrdersPage />}'), 'AppRoutes registers OrdersPage');
    assert(appRoutesSource.includes('path="orders/:orderId" element={<OrderDetailsPage />}'), 'AppRoutes registers OrderDetailsPage');
    assert(appRoutesSource.includes('path="returns" element={<ReturnsPage />}'), 'AppRoutes registers ReturnsPage');
    assert(appRoutesSource.includes('path="wallet" element={<WalletPage />}'), 'AppRoutes registers WalletPage');
    assert(appRoutesSource.includes('path="referrals" element={<ReferralPage />}'), 'AppRoutes registers ReferralPage');

    // Summary
    const passed = assertions.filter((a) => a.pass).length;
    const total = assertions.length;
    console.log('\n====================================================');
    console.log(`PHASE 6 COMPLETE SUITE RESULTS: ${passed}/${total} PASSED`);
    console.log('====================================================');

    if (passed < total) {
      process.exit(1);
    }
  } finally {
    await vite.close();
  }
}

run().catch((err) => {
  console.error('Fatal test error:', err);
  process.exit(1);
});
