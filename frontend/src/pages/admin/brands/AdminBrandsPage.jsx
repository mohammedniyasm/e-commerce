import { useEffect, useRef, useState } from 'react';
import {
  Award,
  Edit2,
  Image as ImageIcon,
  Power,
  RotateCcw,
  Search,
  Trash2,
  Upload,
  X,
} from 'lucide-react';
import {
  createBrand,
  deleteBrand,
  deleteBrandLogo,
  listBrands,
  listDeletedBrands,
  restoreBrand,
  toggleBrandActive,
  updateBrand,
  uploadBrandLogo,
} from '../../../api/adminCatalogService';
import { Badge } from '../../../components/common/Badge';
import { Button } from '../../../components/common/Button';
import { DataTable } from '../../../components/common/DataTable';
import { Modal } from '../../../components/common/Modal';
import { useToast } from '../../../context/ToastContext';
import { formatDate } from '../../../utils/formatters';
import '../categories/AdminCategoriesPage.css';
import './AdminBrandsPage.css';

export function AdminBrandsPage() {
  const { showToast } = useToast();

  const [activeTab, setActiveTab] = useState('active'); // 'active' | 'deleted'
  const [brands, setBrands] = useState([]);
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
  const [isLogoModalOpen, setIsLogoModalOpen] = useState(false);
  const [isActionLoading, setIsActionLoading] = useState(false);

  // Active item for Edit/Delete/Logo
  const [selectedBrand, setSelectedBrand] = useState(null);
  const [formData, setFormData] = useState({ name: '', description: '' });
  const [formErrors, setFormErrors] = useState({});

  // Logo file upload state
  const [logoFile, setLogoFile] = useState(null);
  const [logoPreview, setLogoPreview] = useState(null);
  const fileInputRef = useRef(null);

  async function fetchBrands() {
    setIsLoading(true);
    try {
      const fetchFn = activeTab === 'active' ? listBrands : listDeletedBrands;
      const res = await fetchFn({ page, limit, search });
      setBrands(res.brands || []);
      setTotalPages(res.total_pages || 1);
      setTotalItems(res.total || 0);
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Error Loading Brands',
        message: err?.response?.data?.message || err?.message || 'Failed to fetch brands.',
      });
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    fetchBrands();
  }, [activeTab, page, search]);

  function handleSearchSubmit(e) {
    e.preventDefault();
    setPage(1);
    setSearch(searchInput.trim());
  }

  function handleLogoFileChange(e) {
    const file = e.target.files?.[0];
    if (!file) return;

    if (!file.type.startsWith('image/')) {
      showToast({
        type: 'error',
        title: 'Invalid File',
        message: 'Please select an image file (PNG, JPG, SVG, WebP).',
      });
      return;
    }

    setLogoFile(file);
    const reader = new FileReader();
    reader.onload = (event) => setLogoPreview(event.target?.result);
    reader.readAsDataURL(file);
  }

  function clearLogoSelection() {
    setLogoFile(null);
    setLogoPreview(null);
    if (fileInputRef.current) fileInputRef.current.value = '';
  }

  function openCreateModal() {
    setFormData({ name: '', description: '' });
    clearLogoSelection();
    setFormErrors({});
    setIsCreateOpen(true);
  }

  function openEditModal(brand) {
    setSelectedBrand(brand);
    setFormData({
      name: brand.name || '',
      description: brand.description || '',
    });
    setFormErrors({});
    setIsEditOpen(true);
  }

  function openDeleteModal(brand) {
    setSelectedBrand(brand);
    setIsDeleteOpen(true);
  }

  function openLogoModal(brand) {
    setSelectedBrand(brand);
    setLogoPreview(brand.logo || null);
    setLogoFile(null);
    setIsLogoModalOpen(true);
  }

  function validateForm() {
    const errors = {};
    if (!formData.name.trim()) {
      errors.name = 'Brand name is required.';
    }
    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  }

  async function handleCreateBrand() {
    if (!validateForm()) return;

    setIsActionLoading(true);
    try {
      const created = await createBrand({
        name: formData.name.trim(),
        description: formData.description.trim() || undefined,
      });

      // If user also picked a logo file
      if (logoFile && created?.id) {
        try {
          await uploadBrandLogo(created.id, logoFile);
        } catch {
          showToast({
            type: 'warning',
            title: 'Brand Created',
            message: 'Brand was created, but logo upload encountered an issue.',
          });
        }
      }

      showToast({
        type: 'success',
        title: 'Brand Created',
        message: `Brand "${formData.name}" has been created successfully.`,
      });
      setIsCreateOpen(false);
      fetchBrands();
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Creation Failed',
        message: err?.response?.data?.message || err?.message || 'Could not create brand.',
      });
    } finally {
      setIsActionLoading(false);
    }
  }

  async function handleUpdateBrand() {
    if (!validateForm() || !selectedBrand) return;

    setIsActionLoading(true);
    try {
      await updateBrand(selectedBrand.id, {
        name: formData.name.trim(),
        description: formData.description.trim() || undefined,
      });
      showToast({
        type: 'success',
        title: 'Brand Updated',
        message: `Brand "${formData.name}" was updated successfully.`,
      });
      setIsEditOpen(false);
      fetchBrands();
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Update Failed',
        message: err?.response?.data?.message || err?.message || 'Could not update brand.',
      });
    } finally {
      setIsActionLoading(false);
    }
  }

  async function handleSaveBrandLogo() {
    if (!selectedBrand) return;

    setIsActionLoading(true);
    try {
      if (logoFile) {
        await uploadBrandLogo(selectedBrand.id, logoFile);
        showToast({
          type: 'success',
          title: 'Logo Updated',
          message: `Logo for "${selectedBrand.name}" has been updated.`,
        });
      }
      setIsLogoModalOpen(false);
      fetchBrands();
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Logo Upload Failed',
        message: err?.response?.data?.message || 'Could not update brand logo.',
      });
    } finally {
      setIsActionLoading(false);
    }
  }

  async function handleDeleteLogo() {
    if (!selectedBrand) return;

    setIsActionLoading(true);
    try {
      await deleteBrandLogo(selectedBrand.id);
      showToast({
        type: 'success',
        title: 'Logo Removed',
        message: `Logo for "${selectedBrand.name}" was deleted.`,
      });
      setLogoPreview(null);
      setLogoFile(null);
      setIsLogoModalOpen(false);
      fetchBrands();
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Action Failed',
        message: err?.response?.data?.message || 'Could not delete brand logo.',
      });
    } finally {
      setIsActionLoading(false);
    }
  }

  async function handleDeleteBrand() {
    if (!selectedBrand) return;

    setIsActionLoading(true);
    try {
      await deleteBrand(selectedBrand.id);
      showToast({
        type: 'success',
        title: 'Brand Deleted',
        message: `Brand "${selectedBrand.name}" has been deleted.`,
      });
      setIsDeleteOpen(false);
      fetchBrands();
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Delete Failed',
        message:
          err?.response?.data?.message ||
          'Failed to delete brand. Ensure no active products are associated with this brand.',
      });
    } finally {
      setIsActionLoading(false);
    }
  }

  async function handleToggleActive(brand) {
    try {
      await toggleBrandActive(brand.id);
      showToast({
        type: 'success',
        title: 'Status Updated',
        message: `Brand "${brand.name}" is now ${brand.is_active ? 'inactive' : 'active'}.`,
      });
      fetchBrands();
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Toggle Failed',
        message: err?.response?.data?.message || 'Could not toggle brand status.',
      });
    }
  }

  async function handleRestore(brand) {
    try {
      await restoreBrand(brand.id);
      showToast({
        type: 'success',
        title: 'Brand Restored',
        message: `Brand "${brand.name}" has been restored.`,
      });
      fetchBrands();
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Restore Failed',
        message: err?.response?.data?.message || 'Could not restore brand.',
      });
    }
  }

  const columns = [
    {
      key: 'name',
      header: 'Brand',
      sortable: false,
      render: (_, b) => (
        <div className="admin-brand-name-group">
          <div className="admin-brand-logo-thumb">
            {b.logo ? (
              <img src={b.logo} alt={b.name} className="admin-brand-logo-img" />
            ) : (
              <span className="admin-brand-logo-fallback">{b.name.slice(0, 2)}</span>
            )}
          </div>
          <div className="admin-row-name">
            <span>{b.name}</span>
          </div>
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
      render: (_, b) => (
        <button
          type="button"
          className="admin-status-pill"
          onClick={() => activeTab === 'active' && handleToggleActive(b)}
          title={activeTab === 'active' ? 'Click to toggle active status' : undefined}
          disabled={activeTab !== 'active'}
        >
          <Badge variant={b.is_active ? 'success' : 'neutral'} size="sm">
            {b.is_active ? 'Active' : 'Inactive'}
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
      render: (_, b) => (
        <div className="admin-actions-cell">
          {activeTab === 'active' ? (
            <>
              <button
                type="button"
                className="admin-icon-btn"
                title="Manage Brand Logo"
                onClick={() => openLogoModal(b)}
              >
                <ImageIcon size={14} />
              </button>
              <button
                type="button"
                className="admin-icon-btn"
                title={b.is_active ? 'Deactivate Brand' : 'Activate Brand'}
                onClick={() => handleToggleActive(b)}
              >
                <Power size={14} color={b.is_active ? '#059669' : '#94a3b8'} />
              </button>
              <button
                type="button"
                className="admin-icon-btn"
                title="Edit Brand"
                onClick={() => openEditModal(b)}
              >
                <Edit2 size={14} />
              </button>
              <button
                type="button"
                className="admin-icon-btn danger"
                title="Delete Brand"
                onClick={() => openDeleteModal(b)}
              >
                <Trash2 size={14} />
              </button>
            </>
          ) : (
            <button
              type="button"
              className="admin-icon-btn success"
              title="Restore Brand"
              onClick={() => handleRestore(b)}
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
          <h1 className="admin-page-title">Brands</h1>
          <p className="admin-page-subtitle">
            Manage luxury fashion houses, partner ateliers, and brand logos.
          </p>
        </div>

        <div className="admin-header-actions">
          <Button variant="primary" size="md" onClick={openCreateModal}>
            <Award size={16} />
            <span>New Brand</span>
          </Button>
        </div>
      </div>

      {/* Toolbar / Search / Tabs */}
      <div className="admin-toolbar-card">
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
              placeholder="Search brands..."
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
        data={brands}
        keyField="id"
        isLoading={isLoading}
        emptyMessage={
          search
            ? `No brands match "${search}".`
            : activeTab === 'active'
            ? 'No brands registered. Add your first brand above.'
            : 'No deleted brands found.'
        }
        pagination={{
          page,
          totalPages,
          totalItems,
          limit,
          onPageChange: (newPage) => setPage(newPage),
        }}
      />

      {/* Create Brand Modal */}
      <Modal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        title="Create Brand"
        eyebrow="Brand Management"
        description="Add a designer label or brand to the Drape luxury catalog."
        confirmText="Create Brand"
        cancelText="Cancel"
        isConfirmLoading={isActionLoading}
        onConfirm={handleCreateBrand}
      >
        <div className="admin-form-group">
          <div className="admin-form-field">
            <label className="admin-form-label" htmlFor="create-brand-name">
              Brand Name <span className="required">*</span>
            </label>
            <input
              id="create-brand-name"
              type="text"
              className="admin-form-input"
              placeholder="e.g. Saint Laurent, Brunello Cucinelli, Loro Piana"
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
            <label className="admin-form-label" htmlFor="create-brand-desc">
              Description (Optional)
            </label>
            <textarea
              id="create-brand-desc"
              className="admin-form-textarea"
              placeholder="Heritage, craftsmanship, origin details..."
              value={formData.description}
              onChange={(e) =>
                setFormData((prev) => ({ ...prev, description: e.target.value }))
              }
            />
          </div>

          <div className="admin-form-field">
            <label className="admin-form-label">Brand Logo (Optional)</label>
            <input
              type="file"
              ref={fileInputRef}
              accept="image/*"
              style={{ display: 'none' }}
              onChange={handleLogoFileChange}
            />

            {logoPreview ? (
              <div className="admin-logo-preview">
                <img src={logoPreview} alt="Logo preview" />
                <button
                  type="button"
                  className="admin-logo-remove-btn"
                  onClick={clearLogoSelection}
                  title="Remove selected logo"
                >
                  <X size={12} />
                </button>
              </div>
            ) : (
              <div
                className="admin-logo-upload-box"
                onClick={() => fileInputRef.current?.click()}
              >
                <Upload size={20} color="#64748b" />
                <span style={{ fontSize: 13, color: '#334155', fontWeight: 500 }}>
                  Click to choose brand logo
                </span>
                <span style={{ fontSize: 11, color: '#94a3b8' }}>
                  PNG, SVG, JPG, WebP supported
                </span>
              </div>
            )}
          </div>
        </div>
      </Modal>

      {/* Edit Brand Modal */}
      <Modal
        isOpen={isEditOpen}
        onClose={() => setIsEditOpen(false)}
        title="Edit Brand"
        eyebrow="Brand Management"
        description="Update brand name and description."
        confirmText="Save Changes"
        cancelText="Cancel"
        isConfirmLoading={isActionLoading}
        onConfirm={handleUpdateBrand}
      >
        <div className="admin-form-group">
          <div className="admin-form-field">
            <label className="admin-form-label" htmlFor="edit-brand-name">
              Brand Name <span className="required">*</span>
            </label>
            <input
              id="edit-brand-name"
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
            <label className="admin-form-label" htmlFor="edit-brand-desc">
              Description (Optional)
            </label>
            <textarea
              id="edit-brand-desc"
              className="admin-form-textarea"
              value={formData.description}
              onChange={(e) =>
                setFormData((prev) => ({ ...prev, description: e.target.value }))
              }
            />
          </div>
        </div>
      </Modal>

      {/* Logo Management Modal */}
      <Modal
        isOpen={isLogoModalOpen}
        onClose={() => setIsLogoModalOpen(false)}
        title={`Brand Logo — ${selectedBrand?.name || ''}`}
        eyebrow="Asset Management"
        description="Upload a high-resolution logo or delete the current brand emblem."
        confirmText="Save Logo"
        cancelText="Close"
        isConfirmLoading={isActionLoading}
        onConfirm={handleSaveBrandLogo}
      >
        <div className="admin-form-group">
          <input
            type="file"
            ref={fileInputRef}
            accept="image/*"
            style={{ display: 'none' }}
            onChange={handleLogoFileChange}
          />

          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            {logoPreview ? (
              <div className="admin-logo-preview">
                <img src={logoPreview} alt="Brand logo" />
              </div>
            ) : (
              <div className="admin-brand-logo-thumb" style={{ width: 80, height: 80 }}>
                <span className="admin-brand-logo-fallback" style={{ fontSize: 20 }}>
                  {selectedBrand?.name?.slice(0, 2)}
                </span>
              </div>
            )}

            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              <Button
                type="button"
                variant="secondary"
                size="sm"
                onClick={() => fileInputRef.current?.click()}
              >
                <Upload size={14} />
                <span>Upload New Logo</span>
              </Button>

              {selectedBrand?.logo && (
                <Button
                  type="button"
                  variant="danger"
                  size="sm"
                  onClick={handleDeleteLogo}
                  disabled={isActionLoading}
                >
                  <Trash2 size={14} />
                  <span>Remove Logo</span>
                </Button>
              )}
            </div>
          </div>
        </div>
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={isDeleteOpen}
        onClose={() => setIsDeleteOpen(false)}
        title="Delete Brand"
        eyebrow="Destructive Action"
        description={`Are you sure you want to delete brand "${selectedBrand?.name}"? The server will reject deletion if active products are assigned to this brand.`}
        confirmText="Delete Brand"
        cancelText="Cancel"
        confirmVariant="danger"
        isConfirmLoading={isActionLoading}
        onConfirm={handleDeleteBrand}
      />
    </div>
  );
}
