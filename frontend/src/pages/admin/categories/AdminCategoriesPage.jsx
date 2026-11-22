import { useEffect, useState } from 'react';
import {
  Edit2,
  FolderPlus,
  Power,
  RotateCcw,
  Search,
  Trash2,
} from 'lucide-react';
import {
  createCategory,
  deleteCategory,
  listCategories,
  listDeletedCategories,
  restoreCategory,
  toggleCategoryActive,
  updateCategory,
} from '../../../api/adminCatalogService';
import { Badge } from '../../../components/common/Badge';
import { Button } from '../../../components/common/Button';
import { DataTable } from '../../../components/common/DataTable';
import { Modal } from '../../../components/common/Modal';
import { useToast } from '../../../context/ToastContext';
import { formatDate } from '../../../utils/formatters';
import './AdminCategoriesPage.css';

export function AdminCategoriesPage() {
  const { showToast } = useToast();

  const [activeTab, setActiveTab] = useState('active'); // 'active' | 'deleted'
  const [categories, setCategories] = useState([]);
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [totalPages, setTotalPages] = useState(1);
  const [totalItems, setTotalItems] = useState(0);
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [isLoading, setIsLoading] = useState(true);

  // Modals state
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isEditOpen, setIsEditOpen] = useState(false);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [isActionLoading, setIsActionLoading] = useState(false);

  // Active item for Edit/Delete
  const [selectedCategory, setSelectedCategory] = useState(null);
  const [formData, setFormData] = useState({ name: '', description: '' });
  const [formErrors, setFormErrors] = useState({});

  async function fetchCategories() {
    setIsLoading(true);
    try {
      const fetchFn = activeTab === 'active' ? listCategories : listDeletedCategories;
      const res = await fetchFn({ page, limit, search });
      setCategories(res.categories || []);
      setTotalPages(res.total_pages || 1);
      setTotalItems(res.total || 0);
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Error Loading Categories',
        message: err?.response?.data?.message || err?.message || 'Failed to fetch categories.',
      });
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    fetchCategories();
  }, [activeTab, page, search]);

  function handleSearchSubmit(e) {
    e.preventDefault();
    setPage(1);
    setSearch(searchInput.trim());
  }

  function openCreateModal() {
    setFormData({ name: '', description: '' });
    setFormErrors({});
    setIsCreateOpen(true);
  }

  function openEditModal(category) {
    setSelectedCategory(category);
    setFormData({
      name: category.name || '',
      description: category.description || '',
    });
    setFormErrors({});
    setIsEditOpen(true);
  }

  function openDeleteModal(category) {
    setSelectedCategory(category);
    setIsDeleteOpen(true);
  }

  function validateForm() {
    const errors = {};
    if (!formData.name.trim()) {
      errors.name = 'Category name is required.';
    }
    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  }

  async function handleCreateCategory() {
    if (!validateForm()) return;

    setIsActionLoading(true);
    try {
      await createCategory({
        name: formData.name.trim(),
        description: formData.description.trim() || undefined,
      });
      showToast({
        type: 'success',
        title: 'Category Created',
        message: `Category "${formData.name}" has been created successfully.`,
      });
      setIsCreateOpen(false);
      fetchCategories();
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Creation Failed',
        message: err?.response?.data?.message || err?.message || 'Could not create category.',
      });
    } finally {
      setIsActionLoading(false);
    }
  }

  async function handleUpdateCategory() {
    if (!validateForm() || !selectedCategory) return;

    setIsActionLoading(true);
    try {
      await updateCategory(selectedCategory.id, {
        name: formData.name.trim(),
        description: formData.description.trim() || undefined,
      });
      showToast({
        type: 'success',
        title: 'Category Updated',
        message: `Category "${formData.name}" was updated successfully.`,
      });
      setIsEditOpen(false);
      fetchCategories();
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Update Failed',
        message: err?.response?.data?.message || err?.message || 'Could not update category.',
      });
    } finally {
      setIsActionLoading(false);
    }
  }

  async function handleDeleteCategory() {
    if (!selectedCategory) return;

    setIsActionLoading(true);
    try {
      await deleteCategory(selectedCategory.id);
      showToast({
        type: 'success',
        title: 'Category Deleted',
        message: `Category "${selectedCategory.name}" has been deleted.`,
      });
      setIsDeleteOpen(false);
      fetchCategories();
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Delete Failed',
        message:
          err?.response?.data?.message ||
          'Failed to delete category. Ensure no active products belong to this category.',
      });
    } finally {
      setIsActionLoading(false);
    }
  }

  async function handleToggleActive(category) {
    try {
      await toggleCategoryActive(category.id);
      showToast({
        type: 'success',
        title: 'Status Updated',
        message: `Category "${category.name}" is now ${category.is_active ? 'inactive' : 'active'}.`,
      });
      fetchCategories();
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Toggle Failed',
        message: err?.response?.data?.message || 'Could not toggle category status.',
      });
    }
  }

  async function handleRestore(category) {
    try {
      await restoreCategory(category.id);
      showToast({
        type: 'success',
        title: 'Category Restored',
        message: `Category "${category.name}" has been restored.`,
      });
      fetchCategories();
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Restore Failed',
        message: err?.response?.data?.message || 'Could not restore category.',
      });
    }
  }

  const columns = [
    {
      key: 'name',
      header: 'Category Name',
      sortable: false,
      render: (_, cat) => (
        <div className="admin-row-name">
          <span>{cat.name}</span>
        </div>
      ),
    },
    {
      key: 'description',
      header: 'Description',
      sortable: false,
      render: (desc) => (
        <span className="admin-row-description" title={desc}>
          {desc || '—'}
        </span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      align: 'center',
      render: (_, cat) => (
        <button
          type="button"
          className="admin-status-pill"
          onClick={() => activeTab === 'active' && handleToggleActive(cat)}
          title={activeTab === 'active' ? 'Click to toggle active status' : undefined}
          aria-label={
            activeTab === 'active'
              ? `Toggle active status for category ${cat.name}`
              : `Category ${cat.name} is ${cat.is_active ? 'active' : 'inactive'}`
          }
          disabled={activeTab !== 'active'}
        >
          <Badge variant={cat.is_active ? 'success' : 'neutral'} size="sm">
            {cat.is_active ? 'Active' : 'Inactive'}
          </Badge>
        </button>
      ),
    },
    {
      key: 'created_at',
      header: 'Created Date',
      render: (date) => <span>{formatDate(date)}</span>,
    },
    {
      key: 'actions',
      header: 'Actions',
      align: 'right',
      render: (_, cat) => (
        <div className="admin-actions-cell">
          {activeTab === 'active' ? (
            <>
              <button
                type="button"
                className="admin-icon-btn"
                title={cat.is_active ? 'Deactivate Category' : 'Activate Category'}
                aria-label={`${cat.is_active ? 'Deactivate' : 'Activate'} category ${cat.name}`}
                onClick={() => handleToggleActive(cat)}
              >
                <Power size={14} color={cat.is_active ? '#059669' : '#94a3b8'} />
              </button>
              <button
                type="button"
                className="admin-icon-btn"
                title="Edit Category"
                aria-label={`Edit category ${cat.name}`}
                onClick={() => openEditModal(cat)}
              >
                <Edit2 size={14} />
              </button>
              <button
                type="button"
                className="admin-icon-btn danger"
                title="Delete Category"
                aria-label={`Delete category ${cat.name}`}
                onClick={() => openDeleteModal(cat)}
              >
                <Trash2 size={14} />
              </button>
            </>
          ) : (
            <button
              type="button"
              className="admin-icon-btn success"
              title="Restore Category"
              aria-label={`Restore category ${cat.name}`}
              onClick={() => handleRestore(cat)}
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
      {/* Page Header */}
      <div className="admin-page-header">
        <div className="admin-page-title-group">
          <h1 className="admin-page-title">Categories</h1>
          <p className="admin-page-subtitle">
            Manage product taxonomy and catalog classification structure.
          </p>
        </div>

        <div className="admin-header-actions">
          <Button
            variant="primary"
            size="md"
            className="admin-category-create-btn"
            leftIcon={<FolderPlus size={16} />}
            onClick={openCreateModal}
          >
            New Category
          </Button>
        </div>
      </div>

      {/* Toolbar / Search / Tabs */}
      <div className="admin-toolbar-card admin-category-toolbar">
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
            Active Catalog
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
              placeholder="Search categories..."
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
            />
          </div>
          <Button type="submit" variant="secondary" size="sm">
            Search
          </Button>
        </form>
      </div>

      {/* Data Table */}
      <DataTable
        columns={columns}
        data={categories}
        keyField="id"
        className="admin-category-table"
        isLoading={isLoading}
        emptyMessage={
          search
            ? `No categories match "${search}".`
            : activeTab === 'active'
            ? 'No categories found. Create your first category above.'
            : 'No deleted categories found.'
        }
        pagination={{
          page,
          totalPages,
          totalItems,
          limit,
          onPageChange: (newPage) => setPage(newPage),
        }}
      />

      {/* Create Category Modal */}
      <Modal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        title="Create Category"
        eyebrow="Category Management"
        description="Add a new product category to the Drape catalog."
        confirmText="Create Category"
        cancelText="Cancel"
        isConfirmLoading={isActionLoading}
        onConfirm={handleCreateCategory}
      >
        <div className="admin-form-group">
          <div className="admin-form-field">
            <label className="admin-form-label" htmlFor="create-cat-name">
              Category Name <span className="required">*</span>
            </label>
            <input
              id="create-cat-name"
              type="text"
              className="admin-form-input"
              placeholder="e.g. Silk Shirts, Outerwear, Footwear"
              value={formData.name}
              onChange={(e) => {
                setFormData((prev) => ({ ...prev, name: e.target.value }));
                if (formErrors.name) setFormErrors((prev) => ({ ...prev, name: null }));
              }}
              required
            />
            {formErrors.name && <span className="admin-form-error">{formErrors.name}</span>}
          </div>

          <div className="admin-form-field">
            <label className="admin-form-label" htmlFor="create-cat-desc">
              Description (Optional)
            </label>
            <textarea
              id="create-cat-desc"
              className="admin-form-textarea"
              placeholder="Describe the apparel or collection category..."
              value={formData.description}
              onChange={(e) =>
                setFormData((prev) => ({ ...prev, description: e.target.value }))
              }
            />
          </div>
        </div>
      </Modal>

      {/* Edit Category Modal */}
      <Modal
        isOpen={isEditOpen}
        onClose={() => setIsEditOpen(false)}
        title="Edit Category"
        eyebrow="Category Management"
        description="Update category name and collection details."
        confirmText="Save Changes"
        cancelText="Cancel"
        isConfirmLoading={isActionLoading}
        onConfirm={handleUpdateCategory}
      >
        <div className="admin-form-group">
          <div className="admin-form-field">
            <label className="admin-form-label" htmlFor="edit-cat-name">
              Category Name <span className="required">*</span>
            </label>
            <input
              id="edit-cat-name"
              type="text"
              className="admin-form-input"
              value={formData.name}
              onChange={(e) => {
                setFormData((prev) => ({ ...prev, name: e.target.value }));
                if (formErrors.name) setFormErrors((prev) => ({ ...prev, name: null }));
              }}
              required
            />
            {formErrors.name && <span className="admin-form-error">{formErrors.name}</span>}
          </div>

          <div className="admin-form-field">
            <label className="admin-form-label" htmlFor="edit-cat-desc">
              Description (Optional)
            </label>
            <textarea
              id="edit-cat-desc"
              className="admin-form-textarea"
              value={formData.description}
              onChange={(e) =>
                setFormData((prev) => ({ ...prev, description: e.target.value }))
              }
            />
          </div>
        </div>
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={isDeleteOpen}
        onClose={() => setIsDeleteOpen(false)}
        title="Delete Category"
        eyebrow="Destructive Action"
        description={`Are you sure you want to delete category "${selectedCategory?.name}"? If active products are assigned to this category, the server will reject deletion.`}
        confirmText="Delete Category"
        cancelText="Cancel"
        confirmVariant="danger"
        isConfirmLoading={isActionLoading}
        onConfirm={handleDeleteCategory}
      />
    </div>
  );
}
