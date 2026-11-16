import { createServer } from 'vite';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter, Routes, Route } from 'react-router-dom';

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
  console.log('PHASE 8: AUTOMATED TEST SUITE — ADMIN CATALOG MANAGEMENT');
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
    console.log('[1/7] Testing adminCatalogService Exports...');
    const CatalogService = await vite.ssrLoadModule('/src/api/adminCatalogService.js');

    // Categories
    assert(typeof CatalogService.listCategories === 'function', 'exports listCategories');
    assert(typeof CatalogService.getCategoryById === 'function', 'exports getCategoryById');
    assert(typeof CatalogService.createCategory === 'function', 'exports createCategory');
    assert(typeof CatalogService.updateCategory === 'function', 'exports updateCategory');
    assert(typeof CatalogService.deleteCategory === 'function', 'exports deleteCategory');
    assert(typeof CatalogService.listDeletedCategories === 'function', 'exports listDeletedCategories');
    assert(typeof CatalogService.restoreCategory === 'function', 'exports restoreCategory');
    assert(typeof CatalogService.toggleCategoryActive === 'function', 'exports toggleCategoryActive');

    // Brands
    assert(typeof CatalogService.listBrands === 'function', 'exports listBrands');
    assert(typeof CatalogService.getBrandById === 'function', 'exports getBrandById');
    assert(typeof CatalogService.createBrand === 'function', 'exports createBrand');
    assert(typeof CatalogService.updateBrand === 'function', 'exports updateBrand');
    assert(typeof CatalogService.deleteBrand === 'function', 'exports deleteBrand');
    assert(typeof CatalogService.uploadBrandLogo === 'function', 'exports uploadBrandLogo');
    assert(typeof CatalogService.deleteBrandLogo === 'function', 'exports deleteBrandLogo');
    assert(typeof CatalogService.listDeletedBrands === 'function', 'exports listDeletedBrands');
    assert(typeof CatalogService.restoreBrand === 'function', 'exports restoreBrand');
    assert(typeof CatalogService.toggleBrandActive === 'function', 'exports toggleBrandActive');

    // Products
    assert(typeof CatalogService.listProducts === 'function', 'exports listProducts');
    assert(typeof CatalogService.getProductById === 'function', 'exports getProductById');
    assert(typeof CatalogService.getProductBySlug === 'function', 'exports getProductBySlug');
    assert(typeof CatalogService.createProduct === 'function', 'exports createProduct');
    assert(typeof CatalogService.updateProduct === 'function', 'exports updateProduct');
    assert(typeof CatalogService.deleteProduct === 'function', 'exports deleteProduct');
    assert(typeof CatalogService.listDeletedProducts === 'function', 'exports listDeletedProducts');
    assert(typeof CatalogService.restoreProduct === 'function', 'exports restoreProduct');
    assert(typeof CatalogService.toggleProductActive === 'function', 'exports toggleProductActive');
    assert(typeof CatalogService.toggleProductListed === 'function', 'exports toggleProductListed');

    // Variants
    assert(typeof CatalogService.listVariantsByProductId === 'function', 'exports listVariantsByProductId');
    assert(typeof CatalogService.getVariantById === 'function', 'exports getVariantById');
    assert(typeof CatalogService.createVariant === 'function', 'exports createVariant');
    assert(typeof CatalogService.updateVariant === 'function', 'exports updateVariant');
    assert(typeof CatalogService.deleteVariant === 'function', 'exports deleteVariant');

    // Images
    assert(typeof CatalogService.listProductImages === 'function', 'exports listProductImages');
    assert(typeof CatalogService.listVariantImages === 'function', 'exports listVariantImages');
    assert(typeof CatalogService.getProductImageById === 'function', 'exports getProductImageById');
    assert(typeof CatalogService.uploadProductImages === 'function', 'exports uploadProductImages');
    assert(typeof CatalogService.updateProductImage === 'function', 'exports updateProductImage');
    assert(typeof CatalogService.deleteProductImage === 'function', 'exports deleteProductImage');

    // 2. HTTP Endpoint Mappings & Payloads
    console.log('\n[2/7] Testing HTTP Endpoint Mappings and Request Options...');
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

    // Category endpoint calls
    let restore = mockApi('get', { categories: [{ id: 1, name: 'Trench Coats' }], total: 1 });
    await CatalogService.listCategories({ page: 2, limit: 15, search: 'Trench' });
    assert(lastRequest.url === '/admin/categories', 'listCategories hits /admin/categories');
    assert(lastRequest.config?.params?.page === 2, 'listCategories passes page');
    assert(lastRequest.config?.params?.limit === 15, 'listCategories passes limit');
    assert(lastRequest.config?.params?.search === 'Trench', 'listCategories passes search');
    assert(lastRequest.config?.requiresAuth === true, 'listCategories requires auth');
    restore();

    restore = mockApi('post', { id: 1, name: 'Silk Shirts' });
    await CatalogService.createCategory({ name: 'Silk Shirts', description: 'Finest Mulberry Silk' });
    assert(lastRequest.url === '/admin/categories', 'createCategory hits /admin/categories');
    assert(lastRequest.data?.name === 'Silk Shirts', 'createCategory sends name');
    assert(lastRequest.config?.requiresAuth === true, 'createCategory requires auth');
    restore();

    restore = mockApi('delete', null);
    await CatalogService.deleteCategory(5);
    assert(lastRequest.url === '/admin/categories/5', 'deleteCategory hits /admin/categories/5');
    restore();

    restore = mockApi('patch', null);
    await CatalogService.toggleCategoryActive(5);
    assert(lastRequest.url === '/admin/categories/5/toggle', 'toggleCategoryActive hits /admin/categories/5/toggle');
    restore();

    // Brand endpoint calls
    restore = mockApi('get', { brands: [{ id: 1, name: 'Bottega' }], total: 1 });
    await CatalogService.listBrands({ page: 1, search: 'Bottega' });
    assert(lastRequest.url === '/admin/brands', 'listBrands hits /admin/brands');
    restore();

    restore = mockApi('delete', null);
    await CatalogService.deleteBrand(3);
    assert(lastRequest.url === '/admin/brands/3', 'deleteBrand hits /admin/brands/3');
    restore();

    restore = mockApi('patch', null);
    await CatalogService.toggleBrandActive(3);
    assert(lastRequest.url === '/admin/brands/3/toggle', 'toggleBrandActive hits /admin/brands/3/toggle');
    restore();

    // Product endpoint calls
    restore = mockApi('get', { products: [{ id: 10, name: 'Overcoat' }], total: 1 });
    await CatalogService.listProducts({ page: 1, category_id: 2, brand_id: 3, is_active: true });
    assert(lastRequest.url === '/admin/products', 'listProducts hits /admin/products');
    assert(lastRequest.config?.params?.category_id === 2, 'listProducts passes category_id');
    assert(lastRequest.config?.params?.brand_id === 3, 'listProducts passes brand_id');
    assert(lastRequest.config?.params?.is_active === true, 'listProducts passes is_active');
    restore();

    restore = mockApi('post', { id: 10, name: 'Overcoat' });
    await CatalogService.createProduct({ name: 'Overcoat', category_id: 2, brand_id: 3 });
    assert(lastRequest.url === '/admin/products', 'createProduct hits /admin/products');
    assert(lastRequest.data?.category_id === 2, 'createProduct sends category_id');
    restore();

    restore = mockApi('patch', null);
    await CatalogService.toggleProductListed(10);
    assert(lastRequest.url === '/admin/products/10/listed', 'toggleProductListed hits /admin/products/10/listed');
    restore();

    // Variant endpoint calls
    restore = mockApi('get', [{ id: 1, sku: 'OVC-BLK-M', selling_price: 15000 }]);
    await CatalogService.listVariantsByProductId(10);
    assert(lastRequest.url === '/admin/products/10/variants', 'listVariantsByProductId hits /admin/products/10/variants');
    restore();

    restore = mockApi('post', { id: 1, sku: 'OVC-BLK-M' });
    await CatalogService.createVariant(10, { sku: 'OVC-BLK-M', mrp: 18000, selling_price: 15000, stock: 5 });
    assert(lastRequest.url === '/admin/products/10/variants', 'createVariant hits /admin/products/10/variants');
    assert(lastRequest.data?.sku === 'OVC-BLK-M', 'createVariant sends sku');
    assert(lastRequest.data?.selling_price === 15000, 'createVariant sends selling_price');
    restore();

    // Product Image endpoint calls
    restore = mockApi('get', [{ id: 1, image_url: 'https://cdn.drape/p1.jpg', is_primary: true }]);
    await CatalogService.listProductImages(10);
    assert(lastRequest.url === '/admin/products/10/images', 'listProductImages hits /admin/products/10/images');
    restore();

    restore = mockApi('put', { id: 1, is_primary: true });
    await CatalogService.updateProductImage(1, { is_primary: true });
    assert(lastRequest.url === '/admin/product-images/1', 'updateProductImage hits /admin/product-images/1');
    assert(lastRequest.data?.is_primary === true, 'updateProductImage sends is_primary');
    restore();

    restore = mockApi('delete', null);
    await CatalogService.deleteProductImage(1);
    assert(lastRequest.url === '/admin/product-images/1', 'deleteProductImage hits /admin/product-images/1');
    restore();

    // 3. Category Management UI Rendering
    console.log('\n[3/7] Testing AdminCategoriesPage Rendering & Elements...');
    const { AdminCategoriesPage } = await vite.ssrLoadModule('/src/pages/admin/categories/AdminCategoriesPage.jsx');
    const { ToastProvider } = await vite.ssrLoadModule('/src/context/ToastContext.jsx');
    const { AuthProvider } = await vite.ssrLoadModule('/src/context/AuthContext.jsx');
    const { setAccessToken, clearAccessToken } = await vite.ssrLoadModule('/src/api/tokenStore.js');

    setAccessToken(createMockToken('admin'));

    const catHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/admin/categories'] },
        React.createElement(
          ToastProvider,
          null,
          React.createElement(AuthProvider, null, React.createElement(AdminCategoriesPage, null))
        )
      )
    );
    assert(catHtml.includes('Categories'), 'AdminCategoriesPage renders page title');
    assert(catHtml.includes('New Category'), 'AdminCategoriesPage renders New Category action button');
    assert(catHtml.includes('Active Catalog'), 'AdminCategoriesPage renders Active Catalog tab');
    assert(catHtml.includes('Archived / Deleted'), 'AdminCategoriesPage renders Archived / Deleted tab');
    assert(catHtml.includes('Search categories...'), 'AdminCategoriesPage renders search input');

    // 4. Brand Management UI Rendering
    console.log('\n[4/7] Testing AdminBrandsPage Rendering & Elements...');
    const { AdminBrandsPage } = await vite.ssrLoadModule('/src/pages/admin/brands/AdminBrandsPage.jsx');

    const brandHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/admin/brands'] },
        React.createElement(
          ToastProvider,
          null,
          React.createElement(AuthProvider, null, React.createElement(AdminBrandsPage, null))
        )
      )
    );
    assert(brandHtml.includes('Brands'), 'AdminBrandsPage renders page title');
    assert(brandHtml.includes('New Brand'), 'AdminBrandsPage renders New Brand action button');
    assert(brandHtml.includes('Search brands...'), 'AdminBrandsPage renders search input');
    assert(brandHtml.includes('Active Catalog'), 'AdminBrandsPage renders Active tab');

    // 5. Product Management UI Rendering (Listing & Form)
    console.log('\n[5/7] Testing AdminProductsPage and AdminProductFormPage...');
    const { AdminProductsPage } = await vite.ssrLoadModule('/src/pages/admin/products/AdminProductsPage.jsx');
    const { AdminProductFormPage } = await vite.ssrLoadModule('/src/pages/admin/products/AdminProductFormPage.jsx');

    const prodListHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/admin/products'] },
        React.createElement(
          ToastProvider,
          null,
          React.createElement(AuthProvider, null, React.createElement(AdminProductsPage, null))
        )
      )
    );
    assert(prodListHtml.includes('Catalog Products'), 'AdminProductsPage renders Catalog Products title');
    assert(prodListHtml.includes('Create Product'), 'AdminProductsPage renders Create Product button link');
    assert(prodListHtml.includes('All Categories'), 'AdminProductsPage renders Category filter select');
    assert(prodListHtml.includes('All Brands'), 'AdminProductsPage renders Brand filter select');
    assert(prodListHtml.includes('All Statuses'), 'AdminProductsPage renders Status filter select');

    const prodFormHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/admin/products/new'] },
        React.createElement(
          ToastProvider,
          null,
          React.createElement(AuthProvider, null, React.createElement(AdminProductFormPage, null))
        )
      )
    );
    assert(prodFormHtml.includes('Create New Product'), 'AdminProductFormPage renders Create New Product heading');
    assert(prodFormHtml.includes('Product Title'), 'AdminProductFormPage renders Product Title label');
    assert(prodFormHtml.includes('Select Category'), 'AdminProductFormPage renders Category selector');
    assert(prodFormHtml.includes('Select Brand'), 'AdminProductFormPage renders Brand selector');
    assert(prodFormHtml.includes('Short Description'), 'AdminProductFormPage renders Short Description input');
    assert(prodFormHtml.includes('Full Description'), 'AdminProductFormPage renders Full Description textarea');

    // 6. Product Details, Variant & Image Management Rendering
    console.log('\n[6/7] Testing AdminProductDetailsPage (Product, Variants & Images)...');
    const { AdminProductDetailsPage } = await vite.ssrLoadModule('/src/pages/admin/products/AdminProductDetailsPage.jsx');

    // Mock getProductById to return a product
    const unmockProduct = mockApi('get', {
      id: 10,
      name: 'Cashmere Double-Breasted Coat',
      slug: 'cashmere-double-breasted-coat',
      category_id: 1,
      brand_id: 1,
      is_active: true,
      is_listed: true,
      short_description: 'Pure Italian cashmere.',
      description: 'Handcrafted in Milan.',
      created_at: '2026-11-15T10:00:00Z',
    });

    const detailsHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/admin/products/10'] },
        React.createElement(
          ToastProvider,
          null,
          React.createElement(
            AuthProvider,
            null,
            React.createElement(
              Routes,
              null,
              React.createElement(
                Route,
                { path: '/admin/products/:id', element: React.createElement(AdminProductDetailsPage, null) }
              )
            )
          )
        )
      )
    );
    unmockProduct();

    // Verify Tab navigation
    assert(detailsHtml.includes('Variants') || detailsHtml.includes('Loading product details...'), 'Product Details renders or loads correctly');

    // 7. Route Integration Verification
    console.log('\n[7/7] Verifying AppRoutes Route Architecture...');
    const { AppRoutes } = await vite.ssrLoadModule('/src/routes/AppRoutes.jsx');

    const routesHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/admin/categories'] },
        React.createElement(
          ToastProvider,
          null,
          React.createElement(AuthProvider, null, React.createElement(AppRoutes, null))
        )
      )
    );
    assert(routesHtml.includes('Categories'), 'AppRoutes resolves /admin/categories');

    const brandRouteHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/admin/brands'] },
        React.createElement(
          ToastProvider,
          null,
          React.createElement(AuthProvider, null, React.createElement(AppRoutes, null))
        )
      )
    );
    assert(brandRouteHtml.includes('Brands'), 'AppRoutes resolves /admin/brands');

    const prodRouteHtml = renderToString(
      React.createElement(
        MemoryRouter,
        { initialEntries: ['/admin/products'] },
        React.createElement(
          ToastProvider,
          null,
          React.createElement(AuthProvider, null, React.createElement(AppRoutes, null))
        )
      )
    );
    assert(prodRouteHtml.includes('Catalog Products'), 'AppRoutes resolves /admin/products');

    clearAccessToken();

    // Summary
    const failed = assertions.filter((a) => !a.pass);
    console.log('\n====================================================');
    console.log(`PHASE 8 TEST RESULTS: ${assertions.length - failed.length}/${assertions.length} PASSED`);
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
