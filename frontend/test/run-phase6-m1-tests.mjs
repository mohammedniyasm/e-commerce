import { createServer } from 'vite';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter, Routes, Route, Outlet } from 'react-router-dom';
import fs from 'node:fs';

async function run() {
  console.log('====================================================');
  console.log('PHASE 6: MILESTONE 1 — ACCOUNT LAYOUT & OVERVIEW');
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
    console.log('[1/5] Testing Profile & Order API Services Contracts...');
    const ProfileService = await vite.ssrLoadModule('/src/api/profileService.js');
    assert(typeof ProfileService.getProfile === 'function', 'profileService exports getProfile');
    assert(typeof ProfileService.updateProfile === 'function', 'profileService exports updateProfile');
    assert(typeof ProfileService.changePassword === 'function', 'profileService exports changePassword');
    assert(typeof ProfileService.sendEmailChangeOTP === 'function', 'profileService exports sendEmailChangeOTP');
    assert(typeof ProfileService.verifyEmailChangeOTP === 'function', 'profileService exports verifyEmailChangeOTP');
    assert(typeof ProfileService.uploadProfileImage === 'function', 'profileService exports uploadProfileImage');
    assert(typeof ProfileService.deleteProfileImage === 'function', 'profileService exports deleteProfileImage');

    const OrderService = await vite.ssrLoadModule('/src/api/orderService.js');
    assert(typeof OrderService.getMyOrders === 'function', 'orderService exports getMyOrders');
    assert(typeof OrderService.getMyOrderById === 'function', 'orderService exports getMyOrderById');
    assert(typeof OrderService.getMyOrderByNumber === 'function', 'orderService exports getMyOrderByNumber');
    assert(typeof OrderService.cancelOrder === 'function', 'orderService exports cancelOrder');
    assert(typeof OrderService.downloadInvoice === 'function', 'orderService exports downloadInvoice');
    assert(typeof OrderService.getCustomerShipment === 'function', 'orderService exports getCustomerShipment');
    assert(typeof OrderService.getCustomerTrackingHistory === 'function', 'orderService exports getCustomerTrackingHistory');

    // 2. HTTP Endpoint & Payload Behavioral Verification
    console.log('\n[2/5] Testing Profile Service HTTP Mappings...');
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

    {
      const restore = mockApi('get', { id: 1, name: 'Arjun Nair' });
      await ProfileService.getProfile();
      assert(lastRequest?.url === '/user/profile', 'getProfile calls GET /user/profile');
      restore();
    }

    {
      const restore = mockApi('put', { id: 1, name: 'Arjun Nair', phone: '+91 98470 12345' });
      await ProfileService.updateProfile({ name: 'Arjun Nair', phone: '+91 98470 12345' });
      assert(lastRequest?.url === '/user/profile', 'updateProfile calls PUT /user/profile');
      assert(lastRequest?.data?.name === 'Arjun Nair' && lastRequest?.data?.phone === '+91 98470 12345', 'updateProfile transmits required name and phone fields');
      restore();
    }

    {
      const restore = mockApi('put', { message: 'Password updated' });
      await ProfileService.changePassword({ current_password: 'OldPassword1!', new_password: 'NewPassword2@' });
      assert(lastRequest?.url === '/user/profile/password', 'changePassword calls PUT /user/profile/password');
      assert(lastRequest?.data?.current_password === 'OldPassword1!' && lastRequest?.data?.new_password === 'NewPassword2@', 'changePassword transmits current and new passwords');
      restore();
    }

    // 3. Components SSR Rendering & Design Token Alignment
    console.log('\n[3/5] Testing ProfileHeaderRibbon & AccountNav SSR Rendering...');
    const { ProfileHeaderRibbon } = await vite.ssrLoadModule('/src/components/account/ProfileHeaderRibbon.jsx');
    const { AccountNav } = await vite.ssrLoadModule('/src/components/account/AccountNav.jsx');

    // ProfileHeaderRibbon tests
    {
      const mockProfile = {
        name: 'Arjun Nair',
        email: 'arjun.nair@example.com',
        phone: '+91 98470 12345',
      };
      const mockStats = {
        ordersCount: 4,
        wishlistCount: 12,
        addressesCount: 2,
      };

      const html = renderToString(
        React.createElement(
          MemoryRouter,
          null,
          React.createElement(ProfileHeaderRibbon, { profile: mockProfile, stats: mockStats }),
        ),
      );

      assert(html.includes('Arjun Nair'), 'ProfileHeaderRibbon renders customer name');
      assert(html.includes('AN'), 'ProfileHeaderRibbon calculates and renders correct avatar initials "AN"');
      assert(html.includes('arjun.nair@example.com'), 'ProfileHeaderRibbon renders email address');
      assert(html.includes('+91 98470 12345'), 'ProfileHeaderRibbon renders phone number');
      assert(html.includes('04') && html.includes('ORDERS'), 'ProfileHeaderRibbon renders ORDERS counter "04" matching Figma 32:10998');
      assert(html.includes('12') && html.includes('WISHLIST'), 'ProfileHeaderRibbon renders WISHLIST counter "12" matching Figma 32:10998');
      assert(html.includes('02') && html.includes('SAVED ADDRESSES'), 'ProfileHeaderRibbon renders SAVED ADDRESSES counter "02" matching Figma 32:10998');
      assert(html.includes('EDIT PROFILE'), 'ProfileHeaderRibbon renders EDIT PROFILE action button');
    }

    // AccountNav tests
    {
      const navHtml = renderToString(
        React.createElement(
          MemoryRouter,
          { initialEntries: ['/account'] },
          React.createElement(AccountNav, { onSignOutClick: () => {} }),
        ),
      );

      assert(navHtml.includes('Overview'), 'AccountNav renders Overview link');
      assert(navHtml.includes('Personal Profile'), 'AccountNav renders Personal Profile link');
      assert(navHtml.includes('My Orders'), 'AccountNav renders My Orders link');
      assert(navHtml.includes('Address Book'), 'AccountNav renders Address Book link');
      assert(navHtml.includes('Wallet'), 'AccountNav renders Wallet link');
      assert(navHtml.includes('Returns &amp; Exchanges') || navHtml.includes('Returns & Exchanges'), 'AccountNav renders Returns & Exchanges link');
      assert(navHtml.includes('Referral Program'), 'AccountNav renders Referral Program link');
      assert(navHtml.includes('Sign Out'), 'AccountNav renders Sign Out action button');
      assert(navHtml.includes('is-active'), 'AccountNav marks the active route with "is-active" class');
    }

    // 4. AccountOverviewPage & Context Integration
    console.log('\n[4/5] Testing AccountOverviewPage Sections & Fallbacks...');
    const { AccountOverviewPage } = await vite.ssrLoadModule('/src/pages/account/AccountOverviewPage.jsx');
    const { ToastProvider } = await vite.ssrLoadModule('/src/context/ToastContext.jsx');
    const { AuthProvider } = await vite.ssrLoadModule('/src/context/AuthContext.jsx');
    const { Modal } = await vite.ssrLoadModule('/src/components/common/Modal.jsx');

    const { AccountLayout } = await vite.ssrLoadModule('/src/layouts/AccountLayout.jsx');
    const layoutHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/account'] },
        React.createElement(
          ToastProvider,
          null,
          React.createElement(
            AuthProvider,
            null,
            React.createElement(AccountLayout),
          ),
        ),
      ),
    );

    assert(layoutHtml.includes('account-layout'), 'AccountLayout renders main container');
    assert(layoutHtml.includes('Home') && layoutHtml.includes('My Account'), 'AccountLayout renders breadcrumb navigation');
    assert(layoutHtml.includes('account-layout__sidebar'), 'AccountLayout contains sidebar container');
    assert(layoutHtml.includes('account-layout__content'), 'AccountLayout contains content container');

    const layoutSource = fs.readFileSync(new URL('../src/layouts/AccountLayout.jsx', import.meta.url), 'utf8');
    assert(
      layoutSource.includes('title="Sign Out"') &&
        layoutSource.includes('isSignOutModalOpen') &&
        layoutSource.includes('STAY SIGNED IN'),
      'AccountLayout component integrates Sign Out confirmation modal matching Figma 33:14649',
    );

    const modalHtml = renderToString(
      React.createElement(
        Modal,
        {
          isOpen: true,
          title: 'Sign Out',
          eyebrow: 'SESSION CONFIRMATION',
          onClose: () => {},
        },
        React.createElement(
          'div',
          { className: 'account-signout-modal' },
          React.createElement('button', null, 'STAY SIGNED IN'),
          React.createElement('button', null, 'SIGN OUT'),
        ),
      ),
    );
    assert(
      modalHtml.includes('Sign Out') &&
        modalHtml.includes('STAY SIGNED IN') &&
        modalHtml.includes('SIGN OUT'),
      'Sign Out confirmation modal renders properly matching Figma 33:14649',
    );

    // Test AccountOverviewPage with mock outlet context
    function OutletWrapper({ context }) {
      return React.createElement('div', null, React.createElement(Outlet, { context }));
    }

    const mockContext = {
      profile: {
        name: 'Arjun Nair',
        email: 'arjun.nair@example.com',
        phone: '+91 98470 12345',
        email_verified_at: '2026-11-10T10:00:00Z',
      },
      stats: { ordersCount: 4, wishlistCount: 12, addressesCount: 2 },
      recentOrders: [
        {
          id: 'ord-101',
          order_number: 'ORD-9842',
          order_status: 'confirmed',
          total_amount: 14500,
          placed_at: '2026-11-12T10:00:00Z',
          items: [{ id: 1, product_name: 'Tailored Wool Overcoat' }],
        },
      ],
      addresses: [
        {
          id: 1,
          name: 'Arjun Nair',
          phone: '+91 98470 12345',
          address_line1: 'Bungalow 4, Princess Street',
          city: 'Fort Kochi',
          state: 'Kerala',
          postal_code: '682001',
          is_default: true,
        },
      ],
      isLoading: false,
    };

    const overviewHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/account'] },
        React.createElement(
          Routes,
          null,
          React.createElement(
            Route,
            {
              path: '/account',
              element: React.createElement(OutletWrapper, { context: mockContext }),
            },
            React.createElement(Route, {
              index: true,
              element: React.createElement(AccountOverviewPage),
            }),
          ),
        ),
      ),
    );

    assert(overviewHtml.includes('Personal Details'), 'AccountOverviewPage renders Personal Details section (Figma Section 1)');
    assert(overviewHtml.includes('Arjun Nair'), 'AccountOverviewPage renders customer full name');
    assert(overviewHtml.includes('VERIFIED'), 'AccountOverviewPage renders email verified badge');
    assert(overviewHtml.includes('Recent Orders'), 'AccountOverviewPage renders Recent Orders section (Figma Section 2)');
    assert(overviewHtml.includes('ORD-9842'), 'AccountOverviewPage renders recent order number');
    assert(overviewHtml.includes('Saved Addresses'), 'AccountOverviewPage renders Saved Addresses section (Figma Section 3)');
    assert(overviewHtml.includes('Bungalow 4, Princess Street'), 'AccountOverviewPage renders default address street line');
    assert(overviewHtml.includes('Account Security'), 'AccountOverviewPage renders Account Security card');
    assert(overviewHtml.includes('Need Help?'), 'AccountOverviewPage renders customer support card');
    assert(overviewHtml.includes('Email Customer Support'), 'AccountOverviewPage renders support contact link');

    // 5. Route Protection & Architecture Verification
    console.log('\n[5/5] Testing Route Guards & Nested Route Architecture...');
    const appRoutesSource = fs.readFileSync(new URL('../src/routes/AppRoutes.jsx', import.meta.url), 'utf8');

    assert(appRoutesSource.includes('path="account" element={<AccountLayout />}'), 'AppRoutes registers AccountLayout at path="account"');
    assert(appRoutesSource.includes('index element={<AccountOverviewPage />}'), 'AppRoutes registers AccountOverviewPage as the index route of /account');
    assert(appRoutesSource.includes('path="profile"'), 'AppRoutes prepares nested route for profile');
    assert(appRoutesSource.includes('path="orders"'), 'AppRoutes prepares nested route for orders');
    assert(appRoutesSource.includes('path="addresses"'), 'AppRoutes prepares nested route for addresses');
    assert(appRoutesSource.includes('path="wallet"'), 'AppRoutes prepares nested route for wallet');
    assert(appRoutesSource.includes('path="returns"'), 'AppRoutes prepares nested route for returns');
    assert(appRoutesSource.includes('path="referrals"'), 'AppRoutes prepares nested route for referrals');

    // Summary
    const passed = assertions.filter((a) => a.pass).length;
    const total = assertions.length;
    console.log('\n====================================================');
    console.log(`MILESTONE 1 TEST RESULTS: ${passed}/${total} PASSED`);
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
