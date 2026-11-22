import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  Eye,
  Package,
  PackagePlus,
  Power,
  RotateCcw,
  Search,
  Tag,
  Trash2,
} from 'lucide-react';
import {
  deleteProduct,
  listBrands,
  listCategories,
  listDeletedProducts,
  listProducts,
  restoreProduct,
  toggleProductActive,
  toggleProductListed,
} from '../../../api/adminCatalogService';
import { Badge } from '../../../components/common/Badge';
import { Button } from '../../../components/common/Button';
import { DataTable } from '../../../components/common/DataTable';
import { Modal } from '../../../components/common/Modal';
import { useToast } from '../../../context/ToastContext';
import { formatDate } from '../../../utils/formatters';
import '../categories/AdminCategoriesPage.css';
import './AdminProductsPage.css';

export function AdminProductsPage() {
  const { showToast } = useToast();

  const [activeTab, setActiveTab] = useState('active'); // 'active' | 'deleted'
  const [products, setProducts] = useState([]);
  const [categories, setCategories] = useState([]);
  const [brands, setBrands] = useState([]);

  // Filters
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [totalPages, setTotalPages] = useState(1);
  const [totalItems, setTotalItems] = useState(0);
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [selectedCategory, setSelectedCategory] = useState('');
  const [selectedBrand, setSelectedBrand] = useState('');
  const [selectedStatus, setSelectedStatus] = useState(''); // 'active' | 'inactive' | 'listed' | 'unlisted'

  const [isLoading, setIsLoading] = useState(true);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [productToDelete, setProductToDelete] = useState(null);
  const [isActionLoading, setIsActionLoading] = useState(false);

  // Load Categories & Brands for filtering dropdowns
  useEffect(() => {
    async function loadFilterOptions() {
      try {
        const [catRes, brandRes] = await Promise.allSettled([
          listCategories({ page: 1, limit: 100 }),
          listBrands({ page: 1, limit: 100 }),
        ]);
        if (catRes.status === 'fulfilled' && catRes.value?.categories) {
          setCategories(catRes.value.categories);
        }
        if (brandRes.status === 'fulfilled' && brandRes.value?.brands) {
          setBrands(brandRes.value.brands);
        }
      } catch {
        // filter fallback
      }
    }
    loadFilterOptions();
  }, []);

  async function fetchProducts() {
    setIsLoading(true);
    try {
      const fetchFn = activeTab === 'active' ? listProducts : listDeletedProducts;
      const params = {
        page,
        limit,
        search,
        ...(selectedCategory ? { category_id: Number(selectedCategory) } : {}),
        ...(selectedBrand ? { brand_id: Number(selectedBrand) } : {}),
        ...(selectedStatus === 'active' ? { is_active: true } : {}),
        ...(selectedStatus === 'inactive' ? { is_active: false } : {}),
        ...(selectedStatus === 'listed' ? { is_listed: true } : {}),
        ...(selectedStatus === 'unlisted' ? { is_listed: false } : {}),
      };

      const res = await fetchFn(params);
      setProducts(res.products || []);
      setTotalPages(res.total_pages || 1);
      setTotalItems(res.total || 0);
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Error Loading Products',
        message: err?.response?.data?.message || err?.message || 'Failed to fetch catalog products.',
      });
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    fetchProducts();
  }, [activeTab, page, search, selectedCategory, selectedBrand, selectedStatus]);

  function handleSearchSubmit(e) {
    e.preventDefault();
    setPage(1);
    setSearch(searchInput.trim());
  }

  async function handleToggleActive(product) {
    try {
      await toggleProductActive(product.id);
      showToast({
        type: 'success',
        title: 'Status Updated',
        message: `Product "${product.name}" active status updated.`,
      });
      fetchProducts();
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Update Failed',
        message: err?.response?.data?.message || 'Could not toggle product status.',
      });
    }
  }

  async function handleToggleListed(product) {
    try {
      await toggleProductListed(product.id);
      showToast({
        type: 'success',
        title: 'Listing Updated',
        message: `Product "${product.name}" listed status updated.`,
      });
      fetchProducts();
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Update Failed',
        message: err?.response?.data?.message || 'Could not toggle product listing.',
      });
    }
  }

  async function handleDeleteProduct() {
    if (!productToDelete) return;

    setIsActionLoading(true);
    try {
      await deleteProduct(productToDelete.id);
      showToast({
        type: 'success',
        title: 'Product Deleted',
        message: `Product "${productToDelete.name}" has been deleted.`,
      });
      setIsDeleteOpen(false);
      fetchProducts();
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Delete Failed',
        message: err?.response?.data?.message || 'Could not delete product.',
      });
    } finally {
      setIsActionLoading(false);
    }
  }

  async function handleRestoreProduct(product) {
    try {
      await restoreProduct(product.id);
      showToast({
        type: 'success',
        title: 'Product Restored',
        message: `Product "${product.name}" has been restored to catalog.`,
      });
      fetchProducts();
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Restore Failed',
        message: err?.response?.data?.message || 'Could not restore product.',
      });
    }
  }

  const categoryMap = new Map(categories.map((c) => [c.id, c.name]));
  const brandMap = new Map(brands.map((b) => [b.id, b.name]));

  const columns = [
    {
      key: 'name',
      header: 'Product',
      sortable: false,
      render: (_, p) => (
        <div className="admin-product-name-col">
          <div className="admin-product-thumb">
            <Package size={20} color="#94a3b8" />
          </div>
          <div className="admin-product-title-group">
            <Link to={`/admin/products/${p.id}`} className="admin-product-title">
              {p.name}
            </Link>
            <span className="admin-product-slug">{p.slug || `#${p.id}`}</span>
          </div>
        </div>
      ),
    },
    {
      key: 'category_id',
      header: 'Category',
      render: (catId) => (
        <Badge variant="neutral" size="sm">
          {categoryMap.get(catId) || `Category #${catId}`}
        </Badge>
      ),
    },
    {
      key: 'brand_id',
      header: 'Brand',
      render: (brandId) => (
        <span style={{ fontSize: 13, color: '#334155', fontWeight: 500 }}>
          {brandMap.get(brandId) || `Brand #${brandId}`}
        </span>
      ),
    },
    {
      key: 'is_active',
      header: 'Active',
      align: 'center',
      render: (_, p) => (
        <button
          type="button"
          className="admin-status-pill"
          onClick={() => activeTab === 'active' && handleToggleActive(p)}
          title={activeTab === 'active' ? 'Click to toggle active' : undefined}
          aria-label={`${p.is_active ? 'Deactivate' : 'Activate'} ${p.name}`}
          disabled={activeTab !== 'active'}
        >
          <Badge variant={p.is_active ? 'success' : 'neutral'} size="sm">
            {p.is_active ? 'Active' : 'Inactive'}
          </Badge>
        </button>
      ),
    },
    {
      key: 'is_listed',
      header: 'Listed',
      align: 'center',
      render: (_, p) => (
        <button
          type="button"
          className="admin-status-pill"
          onClick={() => activeTab === 'active' && handleToggleListed(p)}
          title={activeTab === 'active' ? 'Click to toggle listed' : undefined}
          aria-label={`${p.is_listed ? 'Hide' : 'List'} ${p.name} ${p.is_listed ? 'from' : 'on'} storefront`}
          disabled={activeTab !== 'active'}
        >
          <Badge variant={p.is_listed ? 'info' : 'warning'} size="sm">
            {p.is_listed ? 'Storefront' : 'Hidden'}
          </Badge>
        </button>
      ),
    },
    {
      key: 'created_at',
      header: 'Created',
      render: (d) => <span>{formatDate(d)}</span>,
    },
    {
      key: 'actions',
      header: 'Actions',
      align: 'right',
      render: (_, p) => (
        <div className="admin-actions-cell">
          <Link
            to={`/admin/products/${p.id}`}
            className="admin-icon-btn"
            title="View Details & Variants"
            aria-label={`View ${p.name} details and variants`}
          >
            <Eye size={14} />
          </Link>

          {activeTab === 'active' ? (
            <>
              <button
                type="button"
                className="admin-icon-btn"
                title={p.is_active ? 'Deactivate Product' : 'Activate Product'}
                aria-label={`${p.is_active ? 'Deactivate' : 'Activate'} ${p.name}`}
                onClick={() => handleToggleActive(p)}
              >
                <Power size={14} color={p.is_active ? '#059669' : '#94a3b8'} />
              </button>
              <button
                type="button"
                className="admin-icon-btn danger"
                title="Delete Product"
                aria-label={`Delete ${p.name}`}
                onClick={() => {
                  setProductToDelete(p);
                  setIsDeleteOpen(true);
                }}
              >
                <Trash2 size={14} />
              </button>
            </>
          ) : (
            <button
              type="button"
              className="admin-icon-btn success"
              title="Restore Product"
              aria-label={`Restore ${p.name}`}
              onClick={() => handleRestoreProduct(p)}
            >
              <RotateCcw size={14} />
            </button>
          )}
        </div>
      ),
    },
  ];

  return (
    <div className="admin-catalog-page">
      {/* Header */}
      <div className="admin-page-header">
        <div className="admin-page-title-group">
          <h1 className="admin-page-title">Catalog Products</h1>
          <p className="admin-page-subtitle">
            Manage product collections, variant attributes, inventory and pricing.
          </p>
        </div>

        <div className="admin-header-actions">
          <Link
            to="/admin/products/new"
            className="drape-btn drape-btn--primary drape-btn--md admin-button-link"
          >
            <PackagePlus size={16} />
            <span>Create Product</span>
          </Link>
        </div>
      </div>

      {/* Toolbar / Filters */}
      <div className="admin-toolbar-card admin-product-filter-card">
        <div className="admin-tabs-group" role="tablist">
          <button
            type="button"
            role="tab"
            aria-selected={activeTab === 'active'}
            className={`admin-tab-btn ${activeTab === 'active' ? 'active' : ''}`}
            onClick={() => {
              setActiveTab('active');
              setPage(1);
            }}
          >
            Active Products
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={activeTab === 'deleted'}
            className={`admin-tab-btn ${activeTab === 'deleted' ? 'active' : ''}`}
            onClick={() => {
              setActiveTab('deleted');
              setPage(1);
            }}
          >
            Archived / Deleted
          </button>
        </div>

        <form className="admin-filter-controls" onSubmit={handleSearchSubmit}>
          <div className="admin-search-wrapper">
            <Search size={16} className="admin-search-icon" />
            <input
              type="search"
              className="admin-search-input"
              placeholder="Search products..."
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
            />
          </div>

          <select
            className="admin-filter-select"
            value={selectedCategory}
            onChange={(e) => {
              setSelectedCategory(e.target.value);
              setPage(1);
            }}
          >
            <option value="">All Categories</option>
            {categories.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>

          <select
            className="admin-filter-select"
            value={selectedBrand}
            onChange={(e) => {
              setSelectedBrand(e.target.value);
              setPage(1);
            }}
          >
            <option value="">All Brands</option>
            {brands.map((b) => (
              <option key={b.id} value={b.id}>
                {b.name}
              </option>
            ))}
          </select>

          <select
            className="admin-filter-select"
            value={selectedStatus}
            onChange={(e) => {
              setSelectedStatus(e.target.value);
              setPage(1);
            }}
          >
            <option value="">All Statuses</option>
            <option value="active">Active</option>
            <option value="inactive">Inactive</option>
            <option value="listed">Storefront Listed</option>
            <option value="unlisted">Hidden</option>
          </select>

          <Button type="submit" variant="secondary" size="sm">
            Filter
          </Button>
        </form>
      </div>

      {/* Table */}
      <DataTable
        className="admin-products-table"
        columns={columns}
        data={products}
        keyField="id"
        isLoading={isLoading}
        emptyMessage={
          search
            ? `No products match "${search}".`
            : activeTab === 'active'
            ? 'No catalog products found. Create your first product above.'
            : 'No deleted products found.'
        }
        pagination={{
          page,
          totalPages,
          totalItems,
          limit,
          onPageChange: (newPage) => setPage(newPage),
        }}
      />

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={isDeleteOpen}
        onClose={() => setIsDeleteOpen(false)}
        title="Delete Product"
        eyebrow="Destructive Action"
        description={`Are you sure you want to delete product "${productToDelete?.name}"? It will be archived and removed from the active storefront catalog.`}
        confirmText="Delete Product"
        cancelText="Cancel"
        confirmVariant="danger"
        isConfirmLoading={isActionLoading}
        onConfirm={handleDeleteProduct}
      />
    </div>
  );
}
