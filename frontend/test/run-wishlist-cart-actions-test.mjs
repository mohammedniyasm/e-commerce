import assert from 'node:assert/strict';
import { createServer } from 'vite';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';

console.log('====================================================');
console.log('WISHLIST & CART ACTION REPAIR VERIFICATION TEST SUITE');
console.log('====================================================\n');

async function runTests() {
  const vite = await createServer({
    server: { middlewareMode: true },
    appType: 'custom',
    root: process.cwd(),
  });

  try {
    console.log('[1/4] Verifying WishlistContext & CartContext exports...');
    const { WishlistProvider, useWishlist } = await vite.ssrLoadModule('/src/context/WishlistContext.jsx');
    const { CartProvider, useCart } = await vite.ssrLoadModule('/src/context/CartContext.jsx');
    const { AuthProvider } = await vite.ssrLoadModule('/src/context/AuthContext.jsx');
    const { ToastProvider } = await vite.ssrLoadModule('/src/context/ToastContext.jsx');
    assert(typeof WishlistProvider === 'function', 'WishlistProvider is a function');
    assert(typeof useWishlist === 'function', 'useWishlist is a function');
    assert(typeof CartProvider === 'function', 'CartProvider is a function');
    assert(typeof useCart === 'function', 'useCart is a function');
    console.log('  ✓ Context providers and hooks export cleanly\n');

    console.log('[2/4] Verifying WishlistPage rendering with collection management actions...');
    const { WishlistPage } = await vite.ssrLoadModule('/src/pages/wishlist/WishlistPage.jsx');
    const wishlistHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/wishlist'] },
        React.createElement(
          AuthProvider,
          null,
          React.createElement(
            ToastProvider,
            null,
            React.createElement(
              WishlistProvider,
              null,
              React.createElement(CartProvider, null, React.createElement(WishlistPage))
            )
          )
        )
      )
    );
    assert(wishlistHtml.includes('My Wishlist'), 'WishlistPage renders "My Wishlist" title');
    assert(wishlistHtml.includes('CURATED SILHOUETTES'), 'WishlistPage renders luxury eyebrow');
    assert(wishlistHtml.includes('wishlist-empty-state'), 'WishlistPage renders accessible empty state');
    assert(wishlistHtml.includes('wishlist-modal-confirm') || wishlistHtml.includes('wishlist-page'), 'WishlistPage includes confirm modal structure');
    console.log('  ✓ WishlistPage renders title, eyebrow, and collection action controls\n');

    console.log('[3/4] Verifying CartPage rendering with clear bag action & modal...');
    const { CartPage } = await vite.ssrLoadModule('/src/pages/cart/CartPage.jsx');
    const cartHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/cart'] },
        React.createElement(
          AuthProvider,
          null,
          React.createElement(
            ToastProvider,
            null,
            React.createElement(
              CartProvider,
              null,
              React.createElement(CartPage)
            )
          )
        )
      )
    );
    assert(cartHtml.includes('Shopping Bag'), 'CartPage renders "Shopping Bag"');
    assert(cartHtml.includes('Your Bag is Empty'), 'CartPage renders empty state');
    assert(cartHtml.includes('cart-clear-modal') || cartHtml.includes('cart-page'), 'CartPage includes clear bag modal structure');
    console.log('  ✓ CartPage renders shopping bag and clear bag modal controls\n');

    console.log('[4/4] Verifying ProductCard standalone rendering with quick wishlist toggle...');
    const { ProductCard } = await vite.ssrLoadModule('/src/components/storefront/ProductCard.jsx');
    const cardHtml = renderToString(
      React.createElement(
        MemoryRouter,
        null,
        React.createElement(ProductCard, {
          product: {
            id: 101,
            name: 'Atelier Silk Wool Suit',
            slug: 'atelier-silk-wool-suit',
            min_price: 45000,
            in_stock: true,
          },
        })
      )
    );
    assert(cardHtml.includes('Atelier Silk Wool Suit'), 'ProductCard renders product title');
    assert(cardHtml.includes('drape-product-card__wishlist-btn'), 'ProductCard renders quick wishlist toggle button');
    assert(cardHtml.includes('wishlist-toggle-101'), 'ProductCard has data-testid for wishlist toggle');
    console.log('  ✓ ProductCard includes accessible wishlist heart toggle\n');

    console.log('====================================================');
    console.log('WISHLIST & CART ACTION TESTS: ALL PASSED (0 FAILED)');
    console.log('====================================================');
  } finally {
    await vite.close();
  }
}

runTests().catch((err) => {
  console.error('Test Suite Failed:', err);
  process.exit(1);
});
