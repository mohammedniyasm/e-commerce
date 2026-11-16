import { useEffect, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import {
  ChevronLeft,
  ChevronRight,
  Filter,
  RefreshCw,
  SlidersHorizontal,
  X,
} from 'lucide-react';
import {
  getStoreBrands,
  getStoreCategories,
  getStoreProducts,
} from '../../api/productService';
import { Button } from '../../components/common/Button';
import { FormField } from '../../components/common/FormField';
import { Input } from '../../components/common/Input';
import { Select } from '../../components/common/Select';
import { SkeletonLoader } from '../../components/common/SkeletonLoader';
import { Badge } from '../../components/common/Badge';
import { ProductCard } from '../../components/storefront/ProductCard';
import './CatalogPage.css';

const SORT_OPTIONS = [
  { value: '', label: 'Featured / Default' },
  { value: 'price_asc', label: 'Price: Low to High' },
  { value: 'price_desc', label: 'Price: High to Low' },
  { value: 'name_asc', label: 'Name: A to Z' },
  { value: 'name_desc', label: 'Name: Z to A' },
];

export function CatalogPage() {
  const [searchParams, setSearchParams] = useSearchParams();

  // URL state
  const currentSearch = searchParams.get('search') || '';
  const currentCategoryId = searchParams.get('category_id') || '';
  const currentBrandId = searchParams.get('brand_id') || '';
  const currentMinPrice = searchParams.get('min_price') || '';
  const currentMaxPrice = searchParams.get('max_price') || '';
  const currentSort = searchParams.get('sort') || '';
  const currentPage = parseInt(searchParams.get('page') || '1', 10);

  // Local filter states for price inputs before applying
  const [minPriceInput, setMinPriceInput] = useState(currentMinPrice);
  const [maxPriceInput, setMaxPriceInput] = useState(currentMaxPrice);
  const [isMobileFiltersOpen, setIsMobileFiltersOpen] = useState(false);

  // Data states
  const [productsData, setProductsData] = useState({
    products: [],
    page: 1,
    limit: 12,
    total: 0,
    total_pages: 1,
  });
  const [categories, setCategories] = useState([]);
  const [brands, setBrands] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  // Load filter reference metadata (categories and brands) once
  useEffect(() => {
    let isMounted = true;
    Promise.allSettled([getStoreCategories(), getStoreBrands()]).then(([cats, brs]) => {
      if (isMounted) {
        if (cats.status === 'fulfilled' && Array.isArray(cats.value)) {
          setCategories(cats.value);
        }
        if (brs.status === 'fulfilled' && Array.isArray(brs.value)) {
          setBrands(brs.value);
        }
      }
    });
    return () => {
      isMounted = false;
    };
  }, []);

  // Sync inputs with URL
  useEffect(() => {
    setMinPriceInput(currentMinPrice);
    setMaxPriceInput(currentMaxPrice);
  }, [currentMinPrice, currentMaxPrice]);

  // Fetch products whenever search params change
  useEffect(() => {
    let isMounted = true;

    async function fetchProducts() {
      try {
        setIsLoading(true);
        setError(null);

        const data = await getStoreProducts({
          search: currentSearch || undefined,
          category_id: currentCategoryId ? parseInt(currentCategoryId, 10) : undefined,
          brand_id: currentBrandId ? parseInt(currentBrandId, 10) : undefined,
          min_price: currentMinPrice ? parseFloat(currentMinPrice) : undefined,
          max_price: currentMaxPrice ? parseFloat(currentMaxPrice) : undefined,
          sort: currentSort || undefined,
          page: currentPage,
          limit: 12,
        });

        if (isMounted && data) {
          setProductsData(data);
        }
      } catch (err) {
        if (isMounted) {
          setError(err.message || 'Unable to load products. Please check connection.');
        }
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    }

    fetchProducts();

    return () => {
      isMounted = false;
    };
  }, [
    currentSearch,
    currentCategoryId,
    currentBrandId,
    currentMinPrice,
    currentMaxPrice,
    currentSort,
    currentPage,
  ]);

  function updateFilter(key, value) {
    const nextParams = new URLSearchParams(searchParams);
    if (value) {
      nextParams.set(key, value);
    } else {
      nextParams.delete(key);
    }
    // Reset to page 1 whenever filters change
    if (key !== 'page') {
      nextParams.delete('page');
    }
    setSearchParams(nextParams);
  }

  function handlePriceFilterSubmit(e) {
    e.preventDefault();
    const nextParams = new URLSearchParams(searchParams);
    if (minPriceInput) nextParams.set('min_price', minPriceInput);
    else nextParams.delete('min_price');

    if (maxPriceInput) nextParams.set('max_price', maxPriceInput);
    else nextParams.delete('max_price');

    nextParams.delete('page');
    setSearchParams(nextParams);
  }

  function resetAllFilters() {
    setMinPriceInput('');
    setMaxPriceInput('');
    setSearchParams(new URLSearchParams());
  }

  const activeCategoryName = categories.find(
    (c) => String(c.id) === currentCategoryId,
  )?.name;
  const activeBrandName = brands.find(
    (b) => String(b.id) === currentBrandId,
  )?.name;

  const hasActiveFilters = Boolean(
    currentSearch ||
      currentCategoryId ||
      currentBrandId ||
      currentMinPrice ||
      currentMaxPrice ||
      currentSort,
  );

  return (
    <div className="catalog-page" data-testid="catalog-page">
      {/* Breadcrumb Header */}
      <div className="catalog-header">
        <nav className="catalog-breadcrumbs" aria-label="Breadcrumb">
          <Link to="/">Home</Link>
          <span className="catalog-breadcrumbs__separator">/</span>
          <span className="catalog-breadcrumbs__current">Collection</span>
        </nav>

        <div className="catalog-header__title-row">
          <div>
            <h1 className="catalog-title">Curated Collection</h1>
            <p className="catalog-subtitle">
              {isLoading
                ? 'Discovering silhouettes...'
                : `Showing ${productsData.products?.length || 0} of ${productsData.total || 0} pieces`}
            </p>
          </div>

          {/* Mobile Filter Toggle & Desktop Sort */}
          <div className="catalog-header__actions">
            <button
              type="button"
              className="catalog-mobile-filter-btn"
              onClick={() => setIsMobileFiltersOpen((prev) => !prev)}
              aria-label="Toggle Filters"
            >
              <SlidersHorizontal size={16} />
              <span>Filters {hasActiveFilters && '•'}</span>
            </button>

            <div className="catalog-sort-wrapper">
              <span className="catalog-sort-label">SORT BY:</span>
              <Select
                value={currentSort}
                onChange={(e) => updateFilter('sort', e.target.value)}
                options={SORT_OPTIONS}
                className="catalog-sort-select"
                aria-label="Sort collection"
              />
            </div>
          </div>
        </div>

        {/* Active Filter Chips */}
        {hasActiveFilters && (
          <div className="catalog-active-filters" aria-label="Active filters">
            <span className="catalog-active-filters__label">ACTIVE:</span>
            {currentSearch && (
              <Badge
                variant="neutral"
                onRemove={() => updateFilter('search', '')}
              >
                Search: “{currentSearch}”
              </Badge>
            )}
            {activeCategoryName && (
              <Badge
                variant="neutral"
                onRemove={() => updateFilter('category_id', '')}
              >
                Category: {activeCategoryName}
              </Badge>
            )}
            {activeBrandName && (
              <Badge
                variant="neutral"
                onRemove={() => updateFilter('brand_id', '')}
              >
                Brand: {activeBrandName}
              </Badge>
            )}
            {(currentMinPrice || currentMaxPrice) && (
              <Badge
                variant="neutral"
                onRemove={() => {
                  setMinPriceInput('');
                  setMaxPriceInput('');
                  const next = new URLSearchParams(searchParams);
                  next.delete('min_price');
                  next.delete('max_price');
                  setSearchParams(next);
                }}
              >
                Price: ₹{currentMinPrice || '0'} – ₹{currentMaxPrice || '∞'}
              </Badge>
            )}
            {currentSort && (
              <Badge
                variant="neutral"
                onRemove={() => updateFilter('sort', '')}
              >
                Sort: {SORT_OPTIONS.find((s) => s.value === currentSort)?.label}
              </Badge>
            )}
            <button
              type="button"
              className="catalog-clear-all-btn"
              onClick={resetAllFilters}
            >
              Reset All
            </button>
          </div>
        )}
      </div>

      {/* Main Catalog Layout */}
      <div className="catalog-body">
        {/* Sidebar Filters */}
        <aside
          className={`catalog-sidebar ${isMobileFiltersOpen ? 'is-open' : ''}`}
          aria-label="Filter products"
        >
          <div className="catalog-sidebar__header">
            <h2 className="catalog-sidebar__title">Filters</h2>
            {isMobileFiltersOpen && (
              <button
                type="button"
                className="catalog-sidebar__close-btn"
                onClick={() => setIsMobileFiltersOpen(false)}
                aria-label="Close filters"
              >
                <X size={20} />
              </button>
            )}
          </div>

          {/* Search in Catalog */}
          <div className="catalog-filter-group">
            <h3 className="catalog-filter-group__title">KEYWORD SEARCH</h3>
            <Input
              type="search"
              size="sm"
              placeholder="Filter by keyword..."
              value={currentSearch}
              onChange={(e) => updateFilter('search', e.target.value)}
            />
          </div>

          {/* Category Filter */}
          {categories.length > 0 && (
            <div className="catalog-filter-group">
              <h3 className="catalog-filter-group__title">CATEGORIES</h3>
              <div className="catalog-filter-options">
                <button
                  type="button"
                  className={`catalog-filter-option ${!currentCategoryId ? 'is-active' : ''}`}
                  onClick={() => updateFilter('category_id', '')}
                >
                  All Categories
                </button>
                {categories.map((cat) => (
                  <button
                    key={cat.id}
                    type="button"
                    className={`catalog-filter-option ${
                      currentCategoryId === String(cat.id) ? 'is-active' : ''
                    }`}
                    onClick={() =>
                      updateFilter(
                        'category_id',
                        currentCategoryId === String(cat.id) ? '' : String(cat.id),
                      )
                    }
                  >
                    <span>{cat.name}</span>
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Brand Filter */}
          {brands.length > 0 && (
            <div className="catalog-filter-group">
              <h3 className="catalog-filter-group__title">BRANDS</h3>
              <div className="catalog-filter-options">
                <button
                  type="button"
                  className={`catalog-filter-option ${!currentBrandId ? 'is-active' : ''}`}
                  onClick={() => updateFilter('brand_id', '')}
                >
                  All Brands
                </button>
                {brands.map((b) => (
                  <button
                    key={b.id}
                    type="button"
                    className={`catalog-filter-option ${
                      currentBrandId === String(b.id) ? 'is-active' : ''
                    }`}
                    onClick={() =>
                      updateFilter(
                        'brand_id',
                        currentBrandId === String(b.id) ? '' : String(b.id),
                      )
                    }
                  >
                    <span>{b.name}</span>
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Price Range Filter */}
          <div className="catalog-filter-group">
            <h3 className="catalog-filter-group__title">PRICE RANGE (₹)</h3>
            <form onSubmit={handlePriceFilterSubmit} className="catalog-price-form">
              <div className="catalog-price-inputs">
                <Input
                  type="number"
                  size="sm"
                  placeholder="Min"
                  value={minPriceInput}
                  onChange={(e) => setMinPriceInput(e.target.value)}
                  min="0"
                />
                <span className="catalog-price-sep">–</span>
                <Input
                  type="number"
                  size="sm"
                  placeholder="Max"
                  value={maxPriceInput}
                  onChange={(e) => setMaxPriceInput(e.target.value)}
                  min="0"
                />
              </div>
              <Button type="submit" variant="secondary" size="sm" fullWidth>
                Apply Price
              </Button>
            </form>
          </div>

          {/* Clear Filters */}
          {hasActiveFilters && (
            <div className="catalog-filter-group">
              <Button
                type="button"
                variant="outline"
                size="sm"
                fullWidth
                onClick={resetAllFilters}
              >
                Reset All Filters
              </Button>
            </div>
          )}
        </aside>

        {/* Product Grid Area */}
        <section className="catalog-content" aria-label="Product listings">
          {isLoading ? (
            <div className="catalog-grid">
              {[1, 2, 3, 4, 5, 6].map((i) => (
                <div key={i} className="catalog-skeleton-card">
                  <SkeletonLoader height="340px" />
                  <div style={{ marginTop: '12px' }}>
                    <SkeletonLoader height="12px" width="30%" />
                    <SkeletonLoader height="16px" width="70%" style={{ marginTop: '8px' }} />
                    <SkeletonLoader height="14px" width="40%" style={{ marginTop: '8px' }} />
                  </div>
                </div>
              ))}
            </div>
          ) : error ? (
            <div className="catalog-state-card">
              <p className="catalog-state-card__msg">{error}</p>
              <Button
                variant="secondary"
                size="sm"
                onClick={() => updateFilter('page', String(currentPage))}
              >
                Try Again
              </Button>
            </div>
          ) : productsData.products?.length > 0 ? (
            <>
              <div className="catalog-grid">
                {productsData.products.map((product) => (
                  <ProductCard key={product.id} product={product} />
                ))}
              </div>

              {/* Pagination Bar */}
              {productsData.total_pages > 1 && (
                <nav className="catalog-pagination" aria-label="Pagination">
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={currentPage <= 1}
                    onClick={() => {
                      updateFilter('page', String(currentPage - 1));
                      window.scrollTo({ top: 0, behavior: 'smooth' });
                    }}
                    leftIcon={<ChevronLeft size={16} />}
                  >
                    Previous
                  </Button>

                  <div className="catalog-pagination__pages">
                    {Array.from(
                      { length: productsData.total_pages },
                      (_, i) => i + 1,
                    ).map((p) => (
                      <button
                        key={p}
                        type="button"
                        className={`catalog-pagination__page-btn ${
                          p === currentPage ? 'is-active' : ''
                        }`}
                        onClick={() => {
                          updateFilter('page', String(p));
                          window.scrollTo({ top: 0, behavior: 'smooth' });
                        }}
                        aria-current={p === currentPage ? 'page' : undefined}
                      >
                        {p}
                      </button>
                    ))}
                  </div>

                  <Button
                    variant="outline"
                    size="sm"
                    disabled={currentPage >= productsData.total_pages}
                    onClick={() => {
                      updateFilter('page', String(currentPage + 1));
                      window.scrollTo({ top: 0, behavior: 'smooth' });
                    }}
                    rightIcon={<ChevronRight size={16} />}
                  >
                    Next
                  </Button>
                </nav>
              )}
            </>
          ) : (
            <div className="catalog-state-card">
              <h3 className="catalog-state-card__title">No Silhouettes Found</h3>
              <p className="catalog-state-card__msg">
                We couldn't find any products matching your specific combination of filters.
              </p>
              <Button variant="primary" size="md" onClick={resetAllFilters}>
                Clear All Filters
              </Button>
            </div>
          )}
        </section>
      </div>
    </div>
  );
}
