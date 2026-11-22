import { createServer } from 'vite';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';

async function run() {
  console.log('====================================================');
  console.log('PHASE 4: STOREFRONT & PRODUCT DISCOVERY TESTS');
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

    // 1. API Services Contract Check
    console.log('[1/4] Testing Product & Storefront API Services...');
    const ProductService = await vite.ssrLoadModule('/src/api/productService.js');
    assert(typeof ProductService.getStoreProducts === 'function', 'getStoreProducts is exported');
    assert(typeof ProductService.getStoreProductBySlug === 'function', 'getStoreProductBySlug is exported');
    assert(typeof ProductService.getStoreProductById === 'function', 'getStoreProductById is exported');
    assert(typeof ProductService.getRelatedProducts === 'function', 'getRelatedProducts is exported');
    assert(typeof ProductService.getProductReviews === 'function', 'getProductReviews is exported');
    assert(typeof ProductService.getStoreCategories === 'function', 'getStoreCategories is exported');
    assert(typeof ProductService.getStoreBrands === 'function', 'getStoreBrands is exported');

    // 2. Providers helper
    const { AuthProvider } = await vite.ssrLoadModule('/src/context/AuthContext.jsx');
    const { ToastProvider } = await vite.ssrLoadModule('/src/context/ToastContext.jsx');
    const { CartProvider } = await vite.ssrLoadModule('/src/context/CartContext.jsx');
    const { WishlistProvider } = await vite.ssrLoadModule('/src/context/WishlistContext.jsx');
    const { setAccessToken, clearAccessToken } = await vite.ssrLoadModule('/src/api/tokenStore.js');

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

    // 3. StorefrontHeader & Footer Rendering
    console.log('\n[2/4] Testing Storefront Header & Footer...');
    const { StorefrontHeader } = await vite.ssrLoadModule('/src/components/storefront/StorefrontHeader.jsx');
    const { StorefrontFooter } = await vite.ssrLoadModule('/src/components/storefront/StorefrontFooter.jsx');

    clearAccessToken();
    const unauthHeaderHtml = renderWithApp(React.createElement(StorefrontHeader));
    assert(unauthHeaderHtml.includes('DRAPE'), 'StorefrontHeader renders DRAPE brand logo');
    assert(unauthHeaderHtml.includes('COLLECTION'), 'StorefrontHeader renders COLLECTION nav link');
    assert(unauthHeaderHtml.includes('Search collection'), 'StorefrontHeader renders search input placeholder');
    assert(unauthHeaderHtml.includes('SIGN IN'), 'StorefrontHeader shows SIGN IN for unauthenticated visitor');

    // Test Authenticated User in Header
    function createMockToken(role, name = 'Eleni Vance') {
      const header = Buffer.from(JSON.stringify({ alg: 'HS256', typ: 'JWT' })).toString('base64url');
      const payload = Buffer.from(
        JSON.stringify({
          user_id: 1,
          name: name,
          role: role,
          exp: Math.floor(Date.now() / 1000) + 3600,
        })
      ).toString('base64url');
      return `${header}.${payload}.sig`;
    }

    setAccessToken(createMockToken('customer'));
    const authHeaderHtml = renderWithApp(React.createElement(StorefrontHeader));
    assert(authHeaderHtml.includes('Customer profile menu'), 'StorefrontHeader displays customer avatar menu when authenticated');
    clearAccessToken();

    // StorefrontFooter
    const footerHtml = renderWithApp(React.createElement(StorefrontFooter));
    assert(footerHtml.includes('ABOUT US'), 'StorefrontFooter renders ABOUT US column');
    assert(footerHtml.includes('QUICK LINKS'), 'StorefrontFooter renders QUICK LINKS column');
    assert(footerHtml.includes('CUSTOMER SERVICE'), 'StorefrontFooter renders CUSTOMER SERVICE column');
    assert(footerHtml.includes('CONTACT US'), 'StorefrontFooter renders CONTACT US column');
    assert(footerHtml.includes('PRICES IN INDIAN RUPEES'), 'StorefrontFooter renders INR currency notice');

    // 4. ProductCard Component Verification
    console.log('\n[3/4] Testing ProductCard Component & Business Rules...');
    const { ProductCard } = await vite.ssrLoadModule('/src/components/storefront/ProductCard.jsx');

    const sampleProduct = {
      id: 101,
      name: 'Relaxed Linen Overshirt',
      slug: 'relaxed-linen-overshirt',
      brand_name: 'ATELIER DRAPE',
      category_name: 'Overshirts',
      min_price: 6499,
      max_price: 6499,
      primary_image: { id: 1, image_url: 'https://cdn.example.com/overshirt.jpg' },
      in_stock: true,
    };

    const productCardHtml = renderWithApp(React.createElement(ProductCard, { product: sampleProduct }));
    assert(productCardHtml.includes('Relaxed Linen Overshirt'), 'ProductCard renders product title');
    assert(productCardHtml.includes('ATELIER DRAPE'), 'ProductCard renders brand name');
    assert(productCardHtml.includes('Overshirts'), 'ProductCard renders category badge');
    assert(productCardHtml.includes('₹6,499'), 'ProductCard formats currency in Indian Rupees (INR)');
    assert(productCardHtml.includes('/products/relaxed-linen-overshirt'), 'ProductCard generates correct slug link');

    // Out of stock product
    const outOfStockProduct = {
      ...sampleProduct,
      id: 102,
      slug: 'sold-out-suit',
      in_stock: false,
    };
    const outOfStockCardHtml = renderWithApp(React.createElement(ProductCard, { product: outOfStockProduct }));
    assert(outOfStockCardHtml.includes('Out of Stock'), 'ProductCard displays Out of Stock status pill when in_stock is false');

    // Price range product
    const rangeProduct = {
      ...sampleProduct,
      id: 103,
      min_price: 4999,
      max_price: 7999,
    };
    const rangeCardHtml = renderWithApp(React.createElement(ProductCard, { product: rangeProduct }));
    assert(rangeCardHtml.includes('₹4,999') && rangeCardHtml.includes('₹7,999'), 'ProductCard renders price range when min_price != max_price');

    // 5. Storefront Pages (Home, Catalog, PDP)
    console.log('\n[4/4] Testing HomePage, CatalogPage & ProductDetailPage...');

    // HomePage
    const { HomePage } = await vite.ssrLoadModule('/src/pages/storefront/HomePage.jsx');
    const homeHtml = renderWithApp(React.createElement(HomePage));
    assert(homeHtml.includes('Clothes for Every Day'), 'HomePage renders campaign headline');
    assert(homeHtml.includes('SHOP COLLECTION'), 'HomePage renders primary shop CTA');
    assert(homeHtml.includes('INSURED SHIPPING'), 'HomePage renders value features');
    assert(homeHtml.includes('Drape brings together comfortable clothing'), 'HomePage renders brand message');

    // CatalogPage
    const { CatalogPage } = await vite.ssrLoadModule('/src/pages/storefront/CatalogPage.jsx');
    const catalogHtml = renderWithApp(React.createElement(CatalogPage), { route: '/catalog?search=linen' });
    assert(catalogHtml.includes('Shop All Products'), 'CatalogPage renders product listing title');
    assert(catalogHtml.includes('Collection'), 'CatalogPage renders breadcrumbs');
    assert(catalogHtml.includes('KEYWORD SEARCH'), 'CatalogPage renders keyword search filter group');
    assert(catalogHtml.includes('PRICE RANGE (₹)'), 'CatalogPage renders price range filter group');
    assert(catalogHtml.includes('linen') && catalogHtml.includes('catalog-active-filters'), 'CatalogPage renders active filter chip for search term');

    // ProductDetailPage
    const { ProductDetailPage } = await vite.ssrLoadModule('/src/pages/storefront/ProductDetailPage.jsx');
    const pdpHtml = renderWithApp(React.createElement(ProductDetailPage), { route: '/products/relaxed-linen-overshirt' });
    assert(pdpHtml.includes('pdp-page') || pdpHtml.includes('pdp-breadcrumbs'), 'ProductDetailPage mounts with responsive container');

    // Summary
    const failed = assertions.filter((a) => !a.pass);
    console.log(`\n====================================================`);
    console.log(`STOREFRONT TESTS SUMMARY: ${assertions.length - failed.length}/${assertions.length} PASSED`);
    console.log(`====================================================\n`);

    if (failed.length > 0) {
      process.exit(1);
    } else {
      process.exit(0);
    }
  } catch (err) {
    console.error('Storefront Test Failure:', err);
    process.exit(1);
  } finally {
    await vite.close();
  }
}

run();
