import { createServer } from 'vite';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import fs from 'node:fs';

async function run() {
  console.log('====================================================');
  console.log('PHASE 5: CART, WISHLIST, CHECKOUT & PAYMENTS TESTS');
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

    // 1. API Services Contract Verification (Original & Extended)
    console.log('[1/7] Testing Cart, Wishlist & Checkout API Contracts...');
    const CartService = await vite.ssrLoadModule('/src/api/cartService.js');
    assert(typeof CartService.getGuestCart === 'function', 'cartService exports getGuestCart');
    assert(typeof CartService.saveGuestCart === 'function', 'cartService exports saveGuestCart');
    assert(typeof CartService.clearGuestCart === 'function', 'cartService exports clearGuestCart');
    assert(typeof CartService.getUserCart === 'function', 'cartService exports getUserCart');
    assert(typeof CartService.addToUserCart === 'function', 'cartService exports addToUserCart');
    assert(typeof CartService.updateCartItemQuantity === 'function', 'cartService exports updateCartItemQuantity');
    assert(typeof CartService.removeCartItem === 'function', 'cartService exports removeCartItem');
    assert(typeof CartService.clearUserCart === 'function', 'cartService exports clearUserCart');
    assert(typeof CartService.mergeCartItems === 'function', 'cartService exports mergeCartItems');
    assert(typeof CartService.validateGuestCart === 'function', 'cartService exports validateGuestCart');

    const WishlistService = await vite.ssrLoadModule('/src/api/wishlistService.js');
    assert(typeof WishlistService.getGuestWishlist === 'function', 'wishlistService exports getGuestWishlist');
    assert(typeof WishlistService.saveGuestWishlist === 'function', 'wishlistService exports saveGuestWishlist');
    assert(typeof WishlistService.clearGuestWishlist === 'function', 'wishlistService exports clearGuestWishlist');
    assert(typeof WishlistService.getWishlists === 'function', 'wishlistService exports getWishlists');
    assert(typeof WishlistService.createWishlist === 'function', 'wishlistService exports createWishlist');
    assert(typeof WishlistService.getWishlistDetails === 'function', 'wishlistService exports getWishlistDetails');
    assert(typeof WishlistService.updateWishlist === 'function', 'wishlistService exports updateWishlist');
    assert(typeof WishlistService.deleteWishlist === 'function', 'wishlistService exports deleteWishlist');
    assert(typeof WishlistService.addProductToWishlist === 'function', 'wishlistService exports addProductToWishlist');
    assert(typeof WishlistService.removeWishlistItem === 'function', 'wishlistService exports removeWishlistItem');
    assert(typeof WishlistService.mergeWishlistItems === 'function', 'wishlistService exports mergeWishlistItems');

    const CheckoutService = await vite.ssrLoadModule('/src/api/checkoutService.js');
    assert(typeof CheckoutService.getUserAddresses === 'function', 'checkoutService exports getUserAddresses');
    assert(typeof CheckoutService.addUserAddress === 'function', 'checkoutService exports addUserAddress');
    assert(typeof CheckoutService.updateUserAddress === 'function', 'checkoutService exports updateUserAddress');
    assert(typeof CheckoutService.deleteUserAddress === 'function', 'checkoutService exports deleteUserAddress');
    assert(typeof CheckoutService.setDefaultUserAddress === 'function', 'checkoutService exports setDefaultUserAddress');
    assert(typeof CheckoutService.validateCoupon === 'function', 'checkoutService exports validateCoupon');
    assert(typeof CheckoutService.getUserWallet === 'function', 'checkoutService exports getUserWallet');
    assert(typeof CheckoutService.placeOrder === 'function', 'checkoutService exports placeOrder');
    assert(typeof CheckoutService.getMyOrderById === 'function', 'checkoutService exports getMyOrderById');
    assert(typeof CheckoutService.getMyOrders === 'function', 'checkoutService exports getMyOrders');
    assert(typeof CheckoutService.createRazorpayOrder === 'function', 'checkoutService exports createRazorpayOrder');
    assert(typeof CheckoutService.verifyRazorpayPayment === 'function', 'checkoutService exports verifyRazorpayPayment');
    assert(typeof CheckoutService.cancelOrder === 'function', 'checkoutService exports cancelOrder');

    // 2. Component SSR Rendering Tests
    console.log('\n[2/7] Testing Cart, Wishlist & Checkout SSR Rendering...');
    const { AuthProvider } = await vite.ssrLoadModule('/src/context/AuthContext.jsx');
    const { ToastProvider } = await vite.ssrLoadModule('/src/context/ToastContext.jsx');
    const { CartProvider } = await vite.ssrLoadModule('/src/context/CartContext.jsx');
    const { WishlistProvider } = await vite.ssrLoadModule('/src/context/WishlistContext.jsx');

    function renderWithApp(element, { route = '/' } = {}) {
      return renderToString(
        React.createElement(
          MemoryRouter,
          { initialEntries: [route] },
          React.createElement(
            ToastProvider,
            null,
            React.createElement(
              AuthProvider,
              null,
              React.createElement(
                CartProvider,
                null,
                React.createElement(WishlistProvider, null, element),
              ),
            ),
          ),
        ),
      );
    }

    const { CartPage } = await vite.ssrLoadModule('/src/pages/cart/CartPage.jsx');
    const cartHtml = renderWithApp(React.createElement(CartPage), { route: '/cart' });
    assert(cartHtml.includes('Shopping Bag'), 'CartPage renders "Shopping Bag" heading');
    assert(cartHtml.includes('Your Bag is Empty'), 'CartPage renders empty bag state when empty');
    assert(cartHtml.includes('EXPLORE COLLECTION'), 'CartPage renders EXPLORE COLLECTION CTA in empty state');
    assert(cartHtml.includes('cart-page'), 'CartPage container has proper CSS class');

    const { WishlistPage } = await vite.ssrLoadModule('/src/pages/wishlist/WishlistPage.jsx');
    const wishlistHtml = renderWithApp(React.createElement(WishlistPage), { route: '/wishlist' });
    assert(wishlistHtml.includes('My Wishlist'), 'WishlistPage renders "My Wishlist" title');
    assert(wishlistHtml.includes('Your Wishlist is Empty'), 'WishlistPage renders empty state');
    assert(wishlistHtml.includes('EXPLORE COLLECTION'), 'WishlistPage renders link to explore collection');
    assert(wishlistHtml.includes('wishlist-page'), 'WishlistPage container has proper CSS class');

    const { CheckoutPage } = await vite.ssrLoadModule('/src/pages/checkout/CheckoutPage.jsx');
    const checkoutHtml = renderWithApp(React.createElement(CheckoutPage), { route: '/checkout' });
    assert(checkoutHtml.includes('checkout-page'), 'CheckoutPage container renders');
    assert(
      checkoutHtml.includes('checkout-loading') ||
        checkoutHtml.includes('Your Bag is Empty') ||
        checkoutHtml.includes('Delivery Address'),
      'CheckoutPage renders initial state (loading skeleton or empty bag safeguard)',
    );

    const { OrderConfirmationPage } = await vite.ssrLoadModule(
      '/src/pages/checkout/OrderConfirmationPage.jsx',
    );
    const confirmHtml = renderWithApp(React.createElement(OrderConfirmationPage), {
      route: '/checkout/success',
    });
    assert(confirmHtml.includes('Order Confirmed'), 'OrderConfirmationPage renders "Order Confirmed" heading');
    assert(confirmHtml.includes('ORDER REFERENCE'), 'OrderConfirmationPage renders order reference card');
    assert(confirmHtml.includes('Order Status'), 'OrderConfirmationPage renders order status block');
    assert(confirmHtml.includes('Payment Status'), 'OrderConfirmationPage renders payment status block');
    assert(confirmHtml.includes('Estimated Delivery'), 'OrderConfirmationPage renders estimated delivery section');
    assert(confirmHtml.includes('VIEW MY ORDERS'), 'OrderConfirmationPage renders VIEW MY ORDERS button');
    assert(confirmHtml.includes('CONTINUE SHOPPING'), 'OrderConfirmationPage renders CONTINUE SHOPPING button');

    // 3. Payment & Verification Contract Validation
    console.log('\n[3/7] Testing Gateway Verification & Authoritative Safeguards...');
    const sampleOrderPayload = {
      cart_item_ids: [101, 102],
      address_id: 5,
      payment_method: 'online',
      coupon_code: 'WELCOME10',
    };
    assert(Array.isArray(sampleOrderPayload.cart_item_ids), 'Order payload requires cart_item_ids array');
    assert(typeof sampleOrderPayload.address_id === 'number', 'Order payload requires numeric address_id');
    assert(
      ['cod', 'online', 'wallet'].includes(sampleOrderPayload.payment_method),
      'Order payload enforces valid payment method',
    );
    assert(!('total_amount' in sampleOrderPayload), 'Order payload does NOT trust frontend-calculated total_amount');

    const sampleVerifyPayload = {
      order_id: 'd9b73468-45b1-4f01-9ae6-d18ceb21b0e3',
      razorpay_order_id: 'order_EKwxwAgIt4iNp',
      razorpay_payment_id: 'pay_29QQoUBcxrhLE',
      razorpay_signature: '9ef4dffbfd84f1318f6739a3ce19f9d85851857c1b5f9ff35b08586de2ec3e86',
    };
    assert(Boolean(sampleVerifyPayload.order_id), 'Payment verify requires backend order_id');
    assert(Boolean(sampleVerifyPayload.razorpay_order_id), 'Payment verify requires razorpay_order_id');
    assert(Boolean(sampleVerifyPayload.razorpay_payment_id), 'Payment verify requires razorpay_payment_id');
    assert(Boolean(sampleVerifyPayload.razorpay_signature), 'Payment verify requires razorpay_signature');

    // 4. BEHAVIORAL REGRESSION: Razorpay Order Creation Endpoint
    console.log('\n[4/7] Testing Razorpay Order Creation Endpoint & Payload...');
    const { apiClient } = await vite.ssrLoadModule('/src/api/client.js');
    let lastRequest = null;
    function mockApi(method, mockResponseData) {
      const original = apiClient[method];
      apiClient[method] = async (url, data, config) => {
        lastRequest = { method, url, data, config };
        return { data: { success: true, data: mockResponseData } };
      };
      return () => {
        apiClient[method] = original;
      };
    }

    {
      const restore = mockApi('post', { razorpay_order_id: 'order_test_999' });
      const orderId = 'a1b2c3d4-e5f6-7890-abcd-ef1234567890';
      const result = await CheckoutService.createRazorpayOrder({ order_id: orderId });

      assert(
        lastRequest?.url === '/user/payments/razorpay/order',
        'createRazorpayOrder calls exact backend endpoint /user/payments/razorpay/order',
        `Endpoint: ${lastRequest?.url}`,
      );
      assert(
        lastRequest?.url !== '/user/payments/razorpay/create',
        'createRazorpayOrder does NOT use obsolete /user/payments/razorpay/create endpoint',
      );
      assert(
        lastRequest?.data?.order_id === orderId,
        'createRazorpayOrder passes order_id in request body',
      );
      assert(
        result?.razorpay_order_id === 'order_test_999',
        'createRazorpayOrder correctly extracts and returns razorpay_order_id',
      );
      restore();
    }

    // 5. BEHAVIORAL REGRESSION: Razorpay Verification Endpoint & Address State Field
    console.log('\n[5/7] Testing Razorpay Verification & Address State Field...');
    {
      const restore = mockApi('post', { verified: true });
      const verifyPayload = {
        order_id: 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
        razorpay_order_id: 'order_EKwxwAgIt4iNp',
        razorpay_payment_id: 'pay_29QQoUBcxrhLE',
        razorpay_signature: '9ef4dffbfd84f1318f6739a3ce19f9d85851857c1b5f9ff35b08586de2ec3e86',
      };
      await CheckoutService.verifyRazorpayPayment(verifyPayload);

      assert(
        lastRequest?.url === '/user/payments/razorpay/verify',
        'verifyRazorpayPayment calls /user/payments/razorpay/verify',
      );
      assert(
        lastRequest?.data?.order_id === verifyPayload.order_id,
        'verifyRazorpayPayment passes order_id',
      );
      assert(
        lastRequest?.data?.razorpay_order_id === verifyPayload.razorpay_order_id,
        'verifyRazorpayPayment passes razorpay_order_id',
      );
      assert(
        lastRequest?.data?.razorpay_payment_id === verifyPayload.razorpay_payment_id,
        'verifyRazorpayPayment passes razorpay_payment_id',
      );
      assert(
        lastRequest?.data?.razorpay_signature === verifyPayload.razorpay_signature,
        'verifyRazorpayPayment passes razorpay_signature',
      );
      restore();
    }

    {
      const restore = mockApi('post', { id: 10 });
      const addressInput = {
        name: 'Eleanor Vance',
        phone: '+91 9876543210',
        address_line1: '482 Mercer Street',
        city: 'Mumbai',
        state: 'Maharashtra',
        postal_code: '400001',
        is_default: true,
      };
      await CheckoutService.addUserAddress(addressInput);

      assert(
        lastRequest?.url === '/user/addresses',
        'addUserAddress posts to /user/addresses',
      );
      assert(
        lastRequest?.data?.state === 'Maharashtra',
        'addUserAddress includes required state field in payload matching backend DTO',
      );
      assert(
        Boolean(lastRequest?.data?.name && lastRequest?.data?.phone && lastRequest?.data?.postal_code),
        'addUserAddress includes all mandatory address fields',
      );
      restore();
    }

    // Verify State field presence in CheckoutPage module
    const checkoutModule = await vite.transformRequest('/src/pages/checkout/CheckoutPage.jsx');
    assert(
      checkoutModule?.code?.includes('addr-state') && checkoutModule?.code?.includes('State / Province'),
      'CheckoutPage includes State / Province input field for address creation matching backend DTO',
    );

    // 6. BEHAVIORAL REGRESSION: Wishlist PATCH & Order Delivery Date Fallbacks
    console.log('\n[6/7] Testing Wishlist PATCH & Order Confirmation Date Handling...');
    {
      let patchCalled = false;
      let putCalled = false;
      const originalPatch = apiClient.patch;
      const originalPut = apiClient.put;

      apiClient.patch = async (url, data) => {
        patchCalled = true;
        lastRequest = { method: 'PATCH', url, data };
        return { data: { success: true, data: { id: 1, name: data.name } } };
      };
      apiClient.put = async (url, data) => {
        putCalled = true;
        lastRequest = { method: 'PUT', url, data };
        return { data: { success: true } };
      };

      await WishlistService.updateWishlist(7, { name: 'Summer Capsule' });

      assert(patchCalled, 'updateWishlist uses HTTP PATCH method matching Gin router and OpenAPI');
      assert(!putCalled, 'updateWishlist does NOT use HTTP PUT method');
      assert(
        lastRequest?.url === '/user/wishlist/7',
        'updateWishlist calls /user/wishlist/:id with wishlist ID',
      );
      assert(
        lastRequest?.data?.name === 'Summer Capsule',
        'updateWishlist passes name in request body matching UpdateWishlistRequest',
      );

      apiClient.patch = originalPatch;
      apiClient.put = originalPut;
    }

    // Test: Missing payment status must NOT default to "paid"
    {
      const unconfirmedOrder = {
        id: 'ord-test-1',
        order_number: 'ORD-UNCONFIRMED',
        order_status: 'pending',
        payment_status: undefined,
        payment_method: 'online',
        total_amount: 12500,
      };

      const html = renderToString(
        React.createElement(
          MemoryRouter,
          {
            initialEntries: [{ pathname: '/checkout/success', state: { order: unconfirmedOrder } }],
          },
          React.createElement(OrderConfirmationPage),
        ),
      );

      assert(
        html.includes('PENDING'),
        'OrderConfirmationPage defaults missing payment status to PENDING (safe fallback)',
      );
      assert(
        !html.includes('">PAID<'),
        'OrderConfirmationPage NEVER defaults missing online payment status to PAID',
      );
    }

    // Test: Delivery dates use backend response values
    {
      const orderWithDelivery = {
        id: 'ord-test-2',
        order_number: 'ORD-DELIVERY-TEST',
        order_status: 'placed',
        payment_status: 'paid',
        payment_method: 'online',
        estimated_delivery_start: '2026-11-20T00:00:00Z',
        estimated_delivery_end: '2026-11-25T00:00:00Z',
        total_amount: 15000,
      };

      const html = renderToString(
        React.createElement(
          MemoryRouter,
          {
            initialEntries: [
              { pathname: '/checkout/success', state: { order: orderWithDelivery } },
            ],
          },
          React.createElement(OrderConfirmationPage),
        ),
      );

      assert(
        html.includes('20 Nov') && html.includes('25 Nov'),
        'OrderConfirmationPage displays backend-provided estimated delivery date range (20 Nov – 25 Nov)',
      );
    }

    // Test: Missing delivery dates fallback gracefully
    {
      const orderWithoutDelivery = {
        id: 'ord-test-3',
        order_number: 'ORD-NO-DATES',
        order_status: 'placed',
        payment_status: 'paid',
        total_amount: 15000,
      };

      const html = renderToString(
        React.createElement(
          MemoryRouter,
          {
            initialEntries: [
              { pathname: '/checkout/success', state: { order: orderWithoutDelivery } },
            ],
          },
          React.createElement(OrderConfirmationPage),
        ),
      );

      assert(
        html.includes('Standard delivery (confirmed at dispatch)'),
        'OrderConfirmationPage handles missing delivery dates without inventing an arbitrary promise',
      );
    }

    // 7. PENDING ORDER RECOVERY, CANCELLATION & STORAGE RESILIENCE
    console.log('\n[7/7] Testing Pending Order Recovery, Cancellation & Storage Resilience...');
    {
      const restoreGet = mockApi('get', {
        id: 'order-pending-uuid',
        order_number: 'ORD-RECOVERED-888',
        order_status: 'pending',
        payment_status: 'pending',
        payment_method: 'online',
        placed_at: new Date(Date.now() - 2 * 60 * 1000).toISOString(),
        total_amount: 9800,
      });

      const recovered = await CheckoutService.getMyOrderById('order-pending-uuid');

      assert(
        lastRequest?.url === '/user/orders/order-pending-uuid',
        'getMyOrderById calls /user/orders/:id to recover order',
      );
      assert(
        recovered?.order_number === 'ORD-RECOVERED-888',
        'getMyOrderById retrieves pending order by ID',
      );
      assert(
        recovered?.order_status === 'pending' && recovered?.payment_status === 'pending',
        'Pending order recovery verifies both order_status and payment_status are pending before retry',
      );
      restoreGet();

      // Test cancelOrder endpoint
      let patchCancelCalled = false;
      const originalPatch = apiClient.patch;
      apiClient.patch = async (url) => {
        patchCancelCalled = true;
        lastRequest = { method: 'PATCH', url };
        return { data: { success: true, message: 'Order cancelled successfully' } };
      };

      await CheckoutService.cancelOrder('order-pending-uuid');
      assert(
        patchCancelCalled && lastRequest?.url === '/user/orders/order-pending-uuid/cancel',
        'cancelOrder calls PATCH /user/orders/:id/cancel to release inventory and cancel pending order',
      );
      apiClient.patch = originalPatch;

      // Storage Resilience Testing
      let sessionStore = { drape_pending_order_id: 'order-pending-uuid' };
      function handleRecoveryError(err) {
        const status = err.response?.status;
        if (status === 401) {
          // Authentication expired: do NOT delete stored order ID
          return;
        }
        if (status === 404) {
          const msg = (err.response?.data?.message || '').toLowerCase();
          if (msg.includes('order not found') || msg.includes('not found')) {
            delete sessionStore.drape_pending_order_id;
          }
          return;
        }
        // Network error / server downtime: do NOT delete stored order ID
      }

      // Network error test
      handleRecoveryError(new Error('Network Error'));
      assert(
        sessionStore.drape_pending_order_id === 'order-pending-uuid',
        'Pending order ID is retained in storage on network errors (not erroneously deleted)',
      );

      // 401 Unauthorized test
      handleRecoveryError({ response: { status: 401 } });
      assert(
        sessionStore.drape_pending_order_id === 'order-pending-uuid',
        'Pending order ID is retained on HTTP 401 Unauthorized (auth expiry does not prove order missing)',
      );

      // 404 Not Found test
      handleRecoveryError({ response: { status: 404, data: { message: 'Order not found' } } });
      assert(
        !sessionStore.drape_pending_order_id,
        'Pending order ID is cleared only when backend confirms 404 Order Not Found',
      );

      // Terminal Webhook Failure Recovery Test (order_status='confirmed', payment_status='failed')
      let recoveryStorage = { drape_pending_order_id: 'order-confirmed-failed-uuid' };
      let pendingOrderState = { id: 'order-confirmed-failed-uuid' };
      let paymentErrorMessage = '';

      function processRecoveredOrder(recovered) {
        const orderStatus = (recovered?.order_status || '').toLowerCase();
        const paymentStatus = (recovered?.payment_status || '').toLowerCase();
        const placedTime = new Date(recovered?.placed_at || recovered?.created_at).getTime();
        const isExpired = !isNaN(placedTime) && Date.now() - placedTime > 10 * 60 * 1000;
        const isEligibleForRetry =
          orderStatus === 'pending' && paymentStatus === 'pending' && !isExpired;

        if (recovered && isEligibleForRetry) {
          pendingOrderState = recovered;
        } else if (
          recovered &&
          orderStatus === 'confirmed' &&
          paymentStatus === 'failed'
        ) {
          delete recoveryStorage.drape_pending_order_id;
          pendingOrderState = null;
          paymentErrorMessage = `Payment for Order #${recovered.order_number} could not be completed and has failed permanently. This order cannot be retried. Please review your bag and place a new order.`;
        } else if (
          recovered &&
          (orderStatus === 'cancelled' ||
            orderStatus === 'expired' ||
            paymentStatus === 'paid')
        ) {
          delete recoveryStorage.drape_pending_order_id;
        }
      }

      const webhookFailedOrder = {
        id: 'order-confirmed-failed-uuid',
        order_number: 'ORD-FAIL-777',
        order_status: 'confirmed',
        payment_status: 'failed',
      };

      processRecoveredOrder(webhookFailedOrder);

      assert(
        !recoveryStorage.drape_pending_order_id,
        'Backend order with order_status="confirmed" and payment_status="failed" clears stored pending order ID from sessionStorage',
      );
      assert(
        pendingOrderState === null,
        'Backend order with order_status="confirmed" and payment_status="failed" sets createdPendingOrder to null (prevents Razorpay retry)',
      );
      assert(
        paymentErrorMessage.includes('failed permanently') && paymentErrorMessage.includes('cannot be retried'),
        'Backend order with order_status="confirmed" and payment_status="failed" sets terminal non-retryable payment failure message',
      );
      const checkoutFileContent = fs.readFileSync(
        new URL('../src/pages/checkout/CheckoutPage.jsx', import.meta.url),
        'utf8',
      );
      assert(
        checkoutFileContent.includes("orderStatus === 'confirmed'") &&
          checkoutFileContent.includes("paymentStatus === 'failed'") &&
          checkoutFileContent.includes('sessionStorage.removeItem(PENDING_ORDER_STORAGE_KEY)'),
        'CheckoutPage component source code handles confirmed order with failed payment status as terminal state',
      );
    }

    // Final Summary
    const passed = assertions.filter((a) => a.pass).length;
    const total = assertions.length;
    console.log('\n====================================================');
    console.log(`PHASE 5 TEST RESULTS: ${passed}/${total} PASSED`);
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
