import { createServer } from 'vite';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import fs from 'fs';
import path from 'path';

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
  console.log('PHASE 10: DRAPE FINAL INTEGRATION, VERIFICATION & AUDIT');
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

    // ----------------------------------------------------
    // 1. RECURSIVE MODULE & IMPORT INTEGRITY AUDIT
    // ----------------------------------------------------
    console.log('[1/7] Performing Full Codebase Module & Import Audit...');

    function getAllSourceFiles(dir) {
      const entries = fs.readdirSync(dir, { withFileTypes: true });
      let files = [];
      for (const entry of entries) {
        const fullPath = path.join(dir, entry.name);
        if (entry.isDirectory()) {
          files = files.concat(getAllSourceFiles(fullPath));
        } else if (entry.name.endsWith('.js') || entry.name.endsWith('.jsx')) {
          files.push(fullPath);
        }
      }
      return files;
    }

    const allSourceFiles = getAllSourceFiles('src')
      .filter((f) => !f.endsWith('main.jsx'))
      .map((f) => '/' + f.replace(/\\/g, '/'));

    let loadErrors = 0;
    for (const file of allSourceFiles) {
      try {
        await vite.ssrLoadModule(file);
      } catch (err) {
        console.error(`  Error loading ${file}:`, err.message);
        loadErrors++;
      }
    }
    assert(loadErrors === 0, `All ${allSourceFiles.length} source modules compile and evaluate without syntax/import errors`);

    // ----------------------------------------------------
    // 2. AUTHENTICATION, AUTHORIZATION & ROLE ISOLATION
    // ----------------------------------------------------
    console.log('\n[2/7] Testing Authentication, Authorization & Role Boundaries...');
    const { setAccessToken, clearAccessToken, getSession } = await vite.ssrLoadModule('/src/api/tokenStore.js');
    const { RequireAuth } = await vite.ssrLoadModule('/src/routes/RequireAuth.jsx');
    const { AuthProvider } = await vite.ssrLoadModule('/src/context/AuthContext.jsx');
    const { ToastProvider } = await vite.ssrLoadModule('/src/context/ToastContext.jsx');
    const { CartProvider } = await vite.ssrLoadModule('/src/context/CartContext.jsx');
    const { WishlistProvider } = await vite.ssrLoadModule('/src/context/WishlistContext.jsx');
    const { AppRoutes } = await vite.ssrLoadModule('/src/routes/AppRoutes.jsx');

    clearAccessToken();

    function renderPathWithAuth(pathStr) {
      return renderToString(
        React.createElement(
          MemoryRouter,
          { initialEntries: [pathStr] },
          React.createElement(
            ToastProvider,
            null,
            React.createElement(
              AuthProvider,
              null,
              React.createElement(
                CartProvider,
                null,
                React.createElement(
                  WishlistProvider,
                  null,
                  React.createElement(AppRoutes, null)
                )
              )
            )
          )
        )
      );
    }

    // Unauthenticated user going to /account -> blocked
    const unauthAccountHtml = renderPathWithAuth('/account');
    assert(!unauthAccountHtml.includes('AccountOverviewPage'), 'Unauthenticated user cannot access /account');

    // Unauthenticated user going to /admin -> blocked from admin portal
    const unauthAdminHtml = renderPathWithAuth('/admin');
    assert(!unauthAdminHtml.includes('admin-sidebar'), 'Unauthenticated user is blocked from accessing Admin console');

    const adminLoginHtml = renderPathWithAuth('/admin/login');
    assert(adminLoginHtml.includes('Admin Email') || adminLoginHtml.includes('admin-login-card'), 'Admin Login portal resolves correctly');

    // Customer Token
    setAccessToken(createMockToken('customer'));
    assert(getSession()?.role === 'customer', 'Token store correctly decodes customer role');

    // Customer going to /account -> allowed
    const customerAccountHtml = renderPathWithAuth('/account');
    assert(customerAccountHtml.includes('account-nav') || customerAccountHtml.includes('Account Overview'), 'Customer user can access /account');

    // Customer going to /admin -> strictly blocked
    const customerAdminHtml = renderPathWithAuth('/admin');
    assert(!customerAdminHtml.includes('admin-sidebar') && !customerAdminHtml.includes('Dashboard Overview'), 'Customer session is strictly blocked from Admin layout & dashboard');

    // Admin Token
    setAccessToken(createMockToken('admin'));
    assert(getSession()?.role === 'admin', 'Token store correctly decodes admin role');

    // Admin going to /admin -> allowed
    const adminHtml = renderPathWithAuth('/admin');
    assert(adminHtml.includes('admin-sidebar') && adminHtml.includes('Dashboard Overview'), 'Admin user is granted access to Admin dashboard');

    clearAccessToken();
    assert(getSession() === null, 'clearAccessToken terminates session and clears user context');

    // ----------------------------------------------------
    // 3. STOREFRONT & CATALOG ARCHITECTURE
    // ----------------------------------------------------
    console.log('\n[3/7] Testing Storefront & Catalog Architecture...');
    const ProductService = await vite.ssrLoadModule('/src/api/productService.js');
    assert(typeof ProductService.getStoreProducts === 'function', 'exports getStoreProducts');
    assert(typeof ProductService.getStoreProductBySlug === 'function', 'exports getStoreProductBySlug');
    assert(typeof ProductService.getStoreProductById === 'function', 'exports getStoreProductById');
    assert(typeof ProductService.getProductReviews === 'function', 'exports getProductReviews');
    assert(typeof ProductService.getRelatedProducts === 'function', 'exports getRelatedProducts');

    // Product Card component
    const { ProductCard } = await vite.ssrLoadModule('/src/components/storefront/ProductCard.jsx');
    const mockProduct = {
      id: 101,
      name: 'Cashmere Double-Breasted Coat',
      slug: 'cashmere-double-breasted-coat',
      category_name: 'Outerwear',
      brand_name: 'Drape Atelier',
      min_price: 35000,
      images: [{ id: 1, image_url: 'https://images.unsplash.com/photo-1.jpg' }],
      in_stock: true,
    };

    const cardHtml = renderToString(
      React.createElement(
        MemoryRouter,
        null,
        React.createElement(
          ToastProvider,
          null,
          React.createElement(
            AuthProvider,
            null,
            React.createElement(ProductCard, { product: mockProduct })
          )
        )
      )
    );
    assert(cardHtml.includes('Cashmere Double-Breasted Coat'), 'ProductCard renders product title');
    assert(cardHtml.includes('₹35,000') || cardHtml.includes('35,000'), 'ProductCard renders formatted INR price');
    assert(cardHtml.includes('Outerwear'), 'ProductCard renders category');

    // ----------------------------------------------------
    // 4. CART & WISHLIST SYNCHRONIZATION
    // ----------------------------------------------------
    console.log('\n[4/7] Testing Cart & Wishlist Synchronization...');
    const CartService = await vite.ssrLoadModule('/src/api/cartService.js');
    assert(typeof CartService.getCustomerCart === 'function', 'exports getCustomerCart');
    assert(typeof CartService.addCustomerCartItem === 'function', 'exports addCustomerCartItem');
    assert(typeof CartService.validateGuestCart === 'function', 'exports validateGuestCart');
    assert(typeof CartService.mergeGuestCart === 'function', 'exports mergeGuestCart');

    const WishlistService = await vite.ssrLoadModule('/src/api/wishlistService.js');
    assert(typeof WishlistService.getWishlists === 'function', 'exports getWishlists');
    assert(typeof WishlistService.addProductToWishlist === 'function', 'exports addProductToWishlist');
    assert(typeof WishlistService.mergeWishlistItems === 'function', 'exports mergeWishlistItems');

    // ----------------------------------------------------
    // 5. CHECKOUT & PAYMENT INTEGRITY RULES
    // ----------------------------------------------------
    console.log('\n[5/7] Testing Checkout & Payment Verification Rules...');
    const CheckoutService = await vite.ssrLoadModule('/src/api/checkoutService.js');
    assert(typeof CheckoutService.placeOrder === 'function', 'exports placeOrder');
    assert(typeof CheckoutService.createRazorpayOrder === 'function', 'exports createRazorpayOrder');
    assert(typeof CheckoutService.verifyRazorpayPayment === 'function', 'exports verifyRazorpayPayment');
    assert(typeof CheckoutService.getUserAddresses === 'function', 'exports getUserAddresses');
    assert(typeof CheckoutService.validateCoupon === 'function', 'exports validateCoupon');
    assert(typeof CheckoutService.getUserWallet === 'function', 'exports getUserWallet');

    // Verify terminal payment failure handling in CheckoutPage:
    // When order_status = "confirmed" and payment_status = "failed", pending order must be cleared.
    const checkoutSource = fs.readFileSync('src/pages/checkout/CheckoutPage.jsx', 'utf8');
    assert(
      checkoutSource.includes("orderStatus === 'confirmed' &&\n                paymentStatus === 'failed'") ||
      (checkoutSource.includes("orderStatus === 'confirmed'") && checkoutSource.includes("paymentStatus === 'failed'")),
      'CheckoutPage handles terminal confirmed+payment_failed state without offering infinite retry'
    );
    assert(
      checkoutSource.includes('sessionStorage.removeItem') || checkoutSource.includes('PENDING_ORDER_STORAGE_KEY'),
      'CheckoutPage clears sessionStorage pending order key on terminal failure'
    );

    // ----------------------------------------------------
    // 6. ADMIN OPERATIONS, CATALOG & GOVERNANCE RULES
    // ----------------------------------------------------
    console.log('\n[6/7] Testing Admin Operations & Business Rules...');
    const AdminOps = await vite.ssrLoadModule('/src/api/adminOperationsService.js');

    // Order status machine
    const orderDetailsSource = fs.readFileSync('src/pages/admin/orders/AdminOrderDetailsPage.jsx', 'utf8');
    assert(orderDetailsSource.includes('pending: [\'confirmed\', \'cancelled\']'), 'Enforces pending -> confirmed/cancelled transition');
    assert(orderDetailsSource.includes('confirmed: [\'processing\', \'cancelled\']'), 'Enforces confirmed -> processing/cancelled transition');
    assert(orderDetailsSource.includes('processing: [\'shipped\', \'cancelled\']'), 'Enforces processing -> shipped/cancelled transition');
    assert(orderDetailsSource.includes('shipped: [\'delivered\', \'returned\']'), 'Enforces shipped -> delivered/returned transition');
    assert(orderDetailsSource.includes('delivered: []'), 'Delivered orders are terminal in the admin UI');

    // Review Moderation Policy: Admins cannot delete customer reviews (visibility only)
    assert(typeof AdminOps.deleteAdminReview === 'undefined', 'Admin review deletion is prohibited; only visibility moderation is supported');
    assert(typeof AdminOps.setReviewVisibility === 'function', 'Admin review visibility toggle endpoint is available');

    // Customer wallet refund: direct endpoint POST /admin/orders/:id/refund
    assert(typeof AdminOps.processWalletRefund === 'function', 'processWalletRefund handles direct wallet credit');

    // ----------------------------------------------------
    // 7. DESIGN TOKENS, ACCESSIBILITY & RESPONSIVENESS
    // ----------------------------------------------------
    console.log('\n[7/7] Testing Design Tokens, Accessibility & Responsiveness...');
    const tokensContent = fs.readFileSync('src/styles/tokens.css', 'utf8');
    assert(tokensContent.includes('--store-header-height: 64px'), 'Stores 64px header height token');
    assert(tokensContent.includes('--admin-sidebar-width: 260px'), 'Stores 260px admin sidebar width token');
    assert(tokensContent.includes('--admin-card-radius: 12px'), 'Stores 12px admin card radius token');
    assert(tokensContent.includes('--color-success: #059669'), 'Stores semantic green success token');
    assert(tokensContent.includes('--color-danger: #ef4444'), 'Stores semantic red danger token');

    // Accessible Components check
    const modalSource = fs.readFileSync('src/components/common/Modal.jsx', 'utf8');
    assert(modalSource.includes('role="dialog"') && modalSource.includes('aria-modal="true"'), 'Modal declares role="dialog" and aria-modal="true"');

    const buttonSource = fs.readFileSync('src/components/common/Button.jsx', 'utf8');
    assert(buttonSource.includes('aria-busy={isLoading}'), 'Button communicates loading state via aria-busy');

    const inputSource = fs.readFileSync('src/components/common/Input.jsx', 'utf8');
    assert(inputSource.includes('aria-invalid={hasError}'), 'Input communicates validation state via aria-invalid');

    // Summary
    const total = assertions.length;
    const passed = assertions.filter((a) => a.pass).length;
    const failed = total - passed;

    console.log('\n====================================================');
    console.log(`PHASE 10 TEST SUMMARY: ${passed}/${total} PASSED (${failed} FAILED)`);
    console.log('====================================================\n');

    if (failed > 0) {
      process.exit(1);
    }
  } finally {
    await vite.close();
  }
}

run().catch((err) => {
  console.error('Fatal error running Phase 10 tests:', err);
  process.exit(1);
});
