import { useEffect, useRef, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import {
  ArrowLeft,
  Check,
  Edit2,
  Image as ImageIcon,
  Layers,
  Package,
  Plus,
  Power,
  Star,
  Trash2,
  Upload,
  X,
} from 'lucide-react';
import {
  createVariant,
  deleteProduct,
  deleteProductImage,
  deleteVariant,
  getBrandById,
  getCategoryById,
  getProductById,
  listProductImages,
  listVariantsByProductId,
  toggleProductActive,
  toggleProductListed,
  updateProductImage,
  updateVariant,
  uploadProductImages,
} from '../../../api/adminCatalogService';
import { Badge } from '../../../components/common/Badge';
import { Button } from '../../../components/common/Button';
import { DataTable } from '../../../components/common/DataTable';
import { Modal } from '../../../components/common/Modal';
import { useToast } from '../../../context/ToastContext';
import { formatCurrency, formatDate } from '../../../utils/formatters';
import './AdminProductsPage.css';

export function AdminProductDetailsPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { showToast } = useToast();

  const [product, setProduct] = useState(null);
  const [category, setCategory] = useState(null);
  const [brand, setBrand] = useState(null);
  const [variants, setVariants] = useState([]);
  const [images, setImages] = useState([]);
  const [activeTab, setActiveTab] = useState('variants'); // 'variants' | 'images' | 'details'

  const [isLoading, setIsLoading] = useState(true);

  // Variant Modal States
  const [isVariantModalOpen, setIsVariantModalOpen] = useState(false);
  const [editingVariant, setEditingVariant] = useState(null);
  const [isDeleteVariantOpen, setIsDeleteVariantOpen] = useState(false);
  const [variantToDelete, setVariantToDelete] = useState(null);
  const [variantForm, setVariantForm] = useState({
    sku: '',
    size: '',
    color: '',
    mrp: '',
    selling_price: '',
    stock: '',
    is_active: true,
  });
  const [variantErrors, setVariantErrors] = useState({});

  // Image Upload Modal States
  const [isUploadImageOpen, setIsUploadImageOpen] = useState(false);
  const [selectedImageFiles, setSelectedImageFiles] = useState([]);
  const [imagePreviews, setImagePreviews] = useState([]);
  const [targetVariantId, setTargetVariantId] = useState('');
  const [isDeleteImageOpen, setIsDeleteImageOpen] = useState(false);
  const [imageToDelete, setImageToDelete] = useState(null);

  const [isActionLoading, setIsActionLoading] = useState(false);
  const fileInputRef = useRef(null);

  async function loadAllData() {
    setIsLoading(true);
    try {
      const prod = await getProductById(id);
      if (!prod) {
        showToast({ type: 'error', title: 'Product Not Found', message: 'Unable to locate product.' });
        navigate('/admin/products');
        return;
      }
      setProduct(prod);

      // Load related category, brand, variants, and images in parallel
      const [catRes, brandRes, varRes, imgRes] = await Promise.allSettled([
        prod.category_id ? getCategoryById(prod.category_id) : Promise.resolve(null),
        prod.brand_id ? getBrandById(prod.brand_id) : Promise.resolve(null),
        listVariantsByProductId(prod.id),
        listProductImages(prod.id),
      ]);

      if (catRes.status === 'fulfilled' && catRes.value) setCategory(catRes.value);
      if (brandRes.status === 'fulfilled' && brandRes.value) setBrand(brandRes.value);
      if (varRes.status === 'fulfilled' && Array.isArray(varRes.value)) setVariants(varRes.value);
      if (imgRes.status === 'fulfilled' && Array.isArray(imgRes.value)) setImages(imgRes.value);
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Error Loading Product',
        message: err?.response?.data?.message || err?.message || 'Failed to load product profile.',
      });
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    loadAllData();
  }, [id]);

  // Quick Toggles
  async function handleToggleActive() {
    try {
      await toggleProductActive(product.id);
      showToast({
        type: 'success',
        title: 'Status Updated',
        message: `Product is now ${product.is_active ? 'inactive' : 'active'}.`,
      });
      loadAllData();
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Toggle Failed',
        message: err?.response?.data?.message || 'Could not toggle product active status.',
      });
    }
  }

  async function handleToggleListed() {
    try {
      await toggleProductListed(product.id);
      showToast({
        type: 'success',
        title: 'Listing Updated',
        message: `Product is now ${product.is_listed ? 'hidden' : 'visible on storefront'}.`,
      });
      loadAllData();
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Toggle Failed',
        message: err?.response?.data?.message || 'Could not toggle product listed status.',
      });
    }
  }

  // -------------------------------------------------------------
  // VARIANT CRUD
  // -------------------------------------------------------------
  function openCreateVariantModal() {
    setEditingVariant(null);
    setVariantForm({
      sku: '',
      size: '',
      color: '',
      mrp: '',
      selling_price: '',
      stock: '0',
      is_active: true,
    });
    setVariantErrors({});
    setIsVariantModalOpen(true);
  }

  function openEditVariantModal(v) {
    setEditingVariant(v);
    setVariantForm({
      sku: v.sku || '',
      size: v.size || '',
      color: v.color || '',
      mrp: String(v.mrp || ''),
      selling_price: String(v.selling_price || ''),
      stock: String(v.stock != null ? v.stock : '0'),
      is_active: v.is_active ?? true,
    });
    setVariantErrors({});
    setIsVariantModalOpen(true);
  }

  function validateVariantForm() {
    const errs = {};
    if (!variantForm.sku.trim()) {
      errs.sku = 'Variant SKU is required.';
    }
    const mrpNum = Number(variantForm.mrp);
    if (isNaN(mrpNum) || mrpNum <= 0) {
      errs.mrp = 'MRP must be greater than zero.';
    }
    const spNum = Number(variantForm.selling_price);
    if (isNaN(spNum) || spNum <= 0) {
      errs.selling_price = 'Selling price must be greater than zero.';
    }
    if (!isNaN(mrpNum) && !isNaN(spNum) && spNum > mrpNum) {
      errs.selling_price = 'Selling price cannot be greater than MRP.';
    }
    const stockNum = Number(variantForm.stock);
    if (isNaN(stockNum) || stockNum < 0) {
      errs.stock = 'Stock cannot be negative.';
    }

    setVariantErrors(errs);
    return Object.keys(errs).length === 0;
  }

  async function handleSaveVariant() {
    if (!validateVariantForm()) return;

    setIsActionLoading(true);
    try {
      const payload = {
        sku: variantForm.sku.trim(),
        size: variantForm.size.trim() || undefined,
        color: variantForm.color.trim() || undefined,
        mrp: Number(variantForm.mrp),
        selling_price: Number(variantForm.selling_price),
        stock: parseInt(variantForm.stock, 10) || 0,
        ...(editingVariant ? { is_active: Boolean(variantForm.is_active) } : {}),
      };

      if (editingVariant) {
        await updateVariant(editingVariant.id, payload);
        showToast({
          type: 'success',
          title: 'Variant Updated',
          message: `Variant SKU "${payload.sku}" updated successfully.`,
        });
      } else {
        await createVariant(product.id, payload);
        showToast({
          type: 'success',
          title: 'Variant Created',
          message: `Variant SKU "${payload.sku}" created successfully.`,
        });
      }

      setIsVariantModalOpen(false);
      const updated = await listVariantsByProductId(product.id);
      setVariants(updated);
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Variant Save Failed',
        message: err?.response?.data?.message || err?.message || 'Could not save variant.',
      });
    } finally {
      setIsActionLoading(false);
    }
  }

  async function handleDeleteVariant() {
    if (!variantToDelete) return;

    setIsActionLoading(true);
    try {
      await deleteVariant(variantToDelete.id);
      showToast({
        type: 'success',
        title: 'Variant Deleted',
        message: `Variant SKU "${variantToDelete.sku}" deleted.`,
      });
      setIsDeleteVariantOpen(false);
      const updated = await listVariantsByProductId(product.id);
      setVariants(updated);
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Delete Failed',
        message: err?.response?.data?.message || 'Could not delete variant.',
      });
    } finally {
      setIsActionLoading(false);
    }
  }

  // -------------------------------------------------------------
  // PRODUCT IMAGE CRUD
  // -------------------------------------------------------------
  function handleImageFileSelect(e) {
    const files = Array.from(e.target.files || []);
    if (!files.length) return;

    setSelectedImageFiles((prev) => [...prev, ...files]);

    const newPreviews = files.map((f) => URL.createObjectURL(f));
    setImagePreviews((prev) => [...prev, ...newPreviews]);
  }

  function removeSelectedFile(index) {
    setSelectedImageFiles((prev) => prev.filter((_, i) => i !== index));
    setImagePreviews((prev) => prev.filter((_, i) => i !== index));
  }

  async function handleUploadImages() {
    if (!selectedImageFiles.length) {
      showToast({ type: 'warning', title: 'No Files', message: 'Please choose at least one image.' });
      return;
    }

    setIsActionLoading(true);
    try {
      const vId = targetVariantId ? Number(targetVariantId) : null;
      await uploadProductImages(product.id, selectedImageFiles, vId);
      showToast({
        type: 'success',
        title: 'Images Uploaded',
        message: `${selectedImageFiles.length} image(s) uploaded successfully.`,
      });

      setSelectedImageFiles([]);
      setImagePreviews([]);
      setTargetVariantId('');
      setIsUploadImageOpen(false);

      const refreshed = await listProductImages(product.id);
      setImages(refreshed);
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Upload Failed',
        message: err?.response?.data?.message || err?.message || 'Failed to upload images.',
      });
    } finally {
      setIsActionLoading(false);
    }
  }

  async function handleSetPrimaryImage(img) {
    try {
      await updateProductImage(img.id, { is_primary: true });
      showToast({
        type: 'success',
        title: 'Primary Image Set',
        message: 'This image is now the primary showcase photo for the product.',
      });
      const refreshed = await listProductImages(product.id);
      setImages(refreshed);
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Action Failed',
        message: err?.response?.data?.message || 'Could not set image as primary.',
      });
    }
  }

  async function handleDeleteImage() {
    if (!imageToDelete) return;

    setIsActionLoading(true);
    try {
      await deleteProductImage(imageToDelete.id);
      showToast({
        type: 'success',
        title: 'Image Deleted',
        message: 'Product image removed from gallery.',
      });
      setIsDeleteImageOpen(false);
      const refreshed = await listProductImages(product.id);
      setImages(refreshed);
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Delete Failed',
        message: err?.response?.data?.message || 'Could not delete product image.',
      });
    } finally {
      setIsActionLoading(false);
    }
  }

  if (isLoading) {
    return (
      <div className="admin-catalog-page">
        <div className="admin-detail-card" style={{ padding: 40, textAlign: 'center' }}>
          <p style={{ color: '#64748b' }}>Loading product details...</p>
        </div>
      </div>
    );
  }

  if (!product) return null;

  const variantColumns = [
    {
      key: 'sku',
      header: 'SKU',
      render: (sku) => (
        <span style={{ fontFamily: 'monospace', fontWeight: 600, color: '#0f172a' }}>
          {sku}
        </span>
      ),
    },
    {
      key: 'attributes',
      header: 'Attributes',
      render: (_, v) => (
        <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
          {v.size && <Badge variant="neutral" size="sm">Size: {v.size}</Badge>}
          {v.color && <Badge variant="neutral" size="sm">Color: {v.color}</Badge>}
          {!v.size && !v.color && <span style={{ color: '#94a3b8', fontSize: 12 }}>Standard</span>}
        </div>
      ),
    },
    {
      key: 'mrp',
      header: 'MRP',
      render: (mrp) => (
        <span style={{ color: '#64748b', textDecoration: 'line-through', fontSize: 12 }}>
          {formatCurrency(mrp)}
        </span>
      ),
    },
    {
      key: 'selling_price',
      header: 'Selling Price',
      render: (sp) => (
        <span style={{ fontWeight: 600, color: '#0f172a' }}>{formatCurrency(sp)}</span>
      ),
    },
    {
      key: 'stock',
      header: 'Stock',
      align: 'center',
      render: (stk) => (
        <Badge variant={stk > 5 ? 'success' : stk > 0 ? 'warning' : 'danger'} size="sm">
          {stk} in stock
        </Badge>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      align: 'center',
      render: (_, v) => (
        <Badge variant={v.is_active ? 'success' : 'neutral'} size="sm">
          {v.is_active ? 'Active' : 'Inactive'}
        </Badge>
      ),
    },
    {
      key: 'actions',
      header: 'Actions',
      align: 'right',
      render: (_, v) => (
        <div className="admin-actions-cell">
          <button
            type="button"
            className="admin-icon-btn"
            title="Edit Variant"
            aria-label={`Edit variant ${v.sku}`}
            onClick={() => openEditVariantModal(v)}
          >
            <Edit2 size={14} />
          </button>
          <button
            type="button"
            className="admin-icon-btn danger"
            title="Delete Variant"
            aria-label={`Delete variant ${v.sku}`}
            onClick={() => {
              setVariantToDelete(v);
              setIsDeleteVariantOpen(true);
            }}
          >
            <Trash2 size={14} />
          </button>
        </div>
      ),
    },
  ];

  return (
    <div className="admin-catalog-page">
      {/* Header */}
      <div className="admin-page-header">
        <div className="admin-page-title-group">
          <Link
            to="/admin/products"
            className="admin-back-link"
          >
            <ArrowLeft size={14} />
            <span>Back to Products</span>
          </Link>
          <div className="admin-product-heading">
            <h1 className="admin-page-title">{product.name}</h1>
            <Badge variant={product.is_active ? 'success' : 'neutral'} size="md">
              {product.is_active ? 'Active' : 'Inactive'}
            </Badge>
            <Badge variant={product.is_listed ? 'info' : 'warning'} size="md">
              {product.is_listed ? 'Storefront Listed' : 'Hidden'}
            </Badge>
          </div>
          <p className="admin-page-subtitle">
            Slug: <code style={{ fontFamily: 'monospace' }}>{product.slug}</code>
          </p>
        </div>

        <div className="admin-header-actions">
          <Button
            variant="secondary"
            size="md"
            leftIcon={<Power size={14} />}
            onClick={handleToggleActive}
          >
            {product.is_active ? 'Deactivate' : 'Activate'}
          </Button>

          <Button variant="secondary" size="md" onClick={handleToggleListed}>
            {product.is_listed ? 'Hide from Store' : 'Publish to Store'}
          </Button>

          <Link
            to={`/admin/products/${product.id}/edit`}
            className="drape-btn drape-btn--primary drape-btn--md admin-button-link"
          >
            <Edit2 size={14} />
            <span>Edit Product</span>
          </Link>
        </div>
      </div>

      {/* Tabs */}
      <div className="admin-toolbar-card admin-product-tabs">
        <div className="admin-tabs-group" role="tablist">
          <button
            type="button"
            role="tab"
            aria-selected={activeTab === 'variants'}
            className={`admin-tab-btn ${activeTab === 'variants' ? 'active' : ''}`}
            onClick={() => setActiveTab('variants')}
          >
            Variants ({variants.length})
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={activeTab === 'images'}
            className={`admin-tab-btn ${activeTab === 'images' ? 'active' : ''}`}
            onClick={() => setActiveTab('images')}
          >
            Product Images ({images.length})
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={activeTab === 'details'}
            className={`admin-tab-btn ${activeTab === 'details' ? 'active' : ''}`}
            onClick={() => setActiveTab('details')}
          >
            Overview & Specifications
          </button>
        </div>
      </div>

      {/* TAB 1: VARIANTS */}
      {activeTab === 'variants' && (
        <div className="admin-product-section">
          <div className="admin-product-section-toolbar">
            <p style={{ margin: 0, fontSize: 13, color: '#64748b' }}>
              Each variant specifies unique SKU, pricing, inventory quantity, size, and color.
            </p>
            <Button
              variant="primary"
              size="sm"
              leftIcon={<Plus size={14} />}
              onClick={openCreateVariantModal}
            >
              Add Variant
            </Button>
          </div>

          <DataTable
            className="admin-variants-table"
            columns={variantColumns}
            data={variants}
            keyField="id"
            emptyMessage="No variants configured for this product. Add at least one variant to enable purchasing."
          />
        </div>
      )}

      {/* TAB 2: PRODUCT IMAGES */}
      {activeTab === 'images' && (
        <div className="admin-product-section">
          <div className="admin-product-section-toolbar">
            <p style={{ margin: 0, fontSize: 13, color: '#64748b' }}>
              Upload gallery images and designate a primary thumbnail image for storefront listings.
            </p>
            <Button
              variant="primary"
              size="sm"
              leftIcon={<Upload size={14} />}
              onClick={() => {
                setSelectedImageFiles([]);
                setImagePreviews([]);
                setTargetVariantId('');
                setIsUploadImageOpen(true);
              }}
            >
              Upload Images
            </Button>
          </div>

          {images.length > 0 ? (
            <div className="admin-images-grid">
              {images.map((img) => (
                <div key={img.id} className="admin-image-card">
                  {img.is_primary && (
                    <span className="admin-primary-tag">
                      <Star size={10} style={{ display: 'inline', marginRight: 2 }} /> Primary
                    </span>
                  )}
                  <img src={img.image_url} alt={product.name} />

                  <div className="admin-image-overlay">
                    {!img.is_primary && (
                      <button
                        type="button"
                        className="admin-icon-btn"
                        aria-label={`Set ${product.name} image as primary`}
                        style={{ background: '#ffffff', color: '#0f172a' }}
                        title="Set as Primary"
                        onClick={() => handleSetPrimaryImage(img)}
                      >
                        <Star size={14} />
                      </button>
                    )}
                    <button
                      type="button"
                      className="admin-icon-btn danger"
                      aria-label={`Delete ${product.name} image`}
                      style={{ background: '#ffffff' }}
                      title="Delete Image"
                      onClick={() => {
                        setImageToDelete(img);
                        setIsDeleteImageOpen(true);
                      }}
                    >
                      <Trash2 size={14} />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="admin-detail-card" style={{ textAlign: 'center', padding: 40 }}>
              <ImageIcon size={32} color="#94a3b8" style={{ marginBottom: 12 }} />
              <p style={{ margin: 0, color: '#64748b', fontSize: 13 }}>
                No images uploaded yet. Click "Upload Images" to showcase this product.
              </p>
            </div>
          )}
        </div>
      )}

      {/* TAB 3: DETAILS & SPECIFICATIONS */}
      {activeTab === 'details' && (
        <div className="admin-detail-card">
          <h2 style={{ fontSize: 16, fontWeight: 600, color: '#0f172a', margin: '0 0 16px' }}>
            Product Information
          </h2>

          <div className="admin-meta-list" style={{ borderTop: 'none', paddingTop: 0 }}>
            <div className="admin-meta-item">
              <span className="admin-meta-label">Category</span>
              <span className="admin-meta-val">{category?.name || `ID #${product.category_id}`}</span>
            </div>
            <div className="admin-meta-item">
              <span className="admin-meta-label">Brand</span>
              <span className="admin-meta-val">{brand?.name || `ID #${product.brand_id}`}</span>
            </div>
            <div className="admin-meta-item">
              <span className="admin-meta-label">Storefront Listing</span>
              <span className="admin-meta-val">{product.is_listed ? 'Publicly Listed' : 'Hidden'}</span>
            </div>
            <div className="admin-meta-item">
              <span className="admin-meta-label">Created At</span>
              <span className="admin-meta-val">{formatDate(product.created_at)}</span>
            </div>
          </div>

          <div style={{ marginTop: 20 }}>
            <span className="admin-meta-label">Short Description</span>
            <p style={{ margin: '4px 0 0', fontSize: 13, color: '#334155' }}>
              {product.short_description || '—'}
            </p>
          </div>

          <div style={{ marginTop: 20 }}>
            <span className="admin-meta-label">Full Description</span>
            <p
              style={{
                margin: '4px 0 0',
                fontSize: 13,
                color: '#334155',
                lineHeight: 1.6,
                whiteSpace: 'pre-wrap',
              }}
            >
              {product.description || '—'}
            </p>
          </div>
        </div>
      )}

      {/* Variant Modal (Create/Edit) */}
      <Modal
        isOpen={isVariantModalOpen}
        onClose={() => setIsVariantModalOpen(false)}
        title={editingVariant ? `Edit Variant SKU: ${editingVariant.sku}` : 'Add Product Variant'}
        eyebrow="Variant Management"
        description="Configure pricing, SKU code, stock count and optional size/color attributes."
        confirmText={editingVariant ? 'Save Variant' : 'Create Variant'}
        cancelText="Cancel"
        isConfirmLoading={isActionLoading}
        onConfirm={handleSaveVariant}
      >
        <div className="admin-form-group">
          <div className="admin-form-field">
            <label className="admin-form-label" htmlFor="var-sku">
              SKU Code <span className="required">*</span>
            </label>
            <input
              id="var-sku"
              type="text"
              className="admin-form-input"
              placeholder="e.g. SLK-TRENCH-BLK-M"
              value={variantForm.sku}
              onChange={(e) => {
                setVariantForm((prev) => ({ ...prev, sku: e.target.value }));
                if (variantErrors.sku) setVariantErrors((prev) => ({ ...prev, sku: null }));
              }}
              required
            />
            {variantErrors.sku && <span className="admin-form-error">{variantErrors.sku}</span>}
          </div>

          <div className="admin-variant-optional-fields">
            <div className="admin-form-field">
              <label className="admin-form-label" htmlFor="var-size">
                Size (Optional)
              </label>
              <input
                id="var-size"
                type="text"
                className="admin-form-input"
                placeholder="e.g. 38, 40, S, M, L"
                value={variantForm.size}
                onChange={(e) => setVariantForm((prev) => ({ ...prev, size: e.target.value }))}
              />
            </div>

            <div className="admin-form-field">
              <label className="admin-form-label" htmlFor="var-color">
                Color (Optional)
              </label>
              <input
                id="var-color"
                type="text"
                className="admin-form-input"
                placeholder="e.g. Midnight Black, Ivory"
                value={variantForm.color}
                onChange={(e) => setVariantForm((prev) => ({ ...prev, color: e.target.value }))}
              />
            </div>
          </div>

          <div className="admin-variant-pricing-fields">
            <div className="admin-form-field">
              <label className="admin-form-label" htmlFor="var-mrp">
                MRP (₹) <span className="required">*</span>
              </label>
              <input
                id="var-mrp"
                type="number"
                min="0.01"
                step="0.01"
                className="admin-form-input"
                placeholder="14500"
                value={variantForm.mrp}
                onChange={(e) => {
                  setVariantForm((prev) => ({ ...prev, mrp: e.target.value }));
                  if (variantErrors.mrp) setVariantErrors((prev) => ({ ...prev, mrp: null }));
                }}
                required
              />
              {variantErrors.mrp && <span className="admin-form-error">{variantErrors.mrp}</span>}
            </div>

            <div className="admin-form-field">
              <label className="admin-form-label" htmlFor="var-sp">
                Selling Price (₹) <span className="required">*</span>
              </label>
              <input
                id="var-sp"
                type="number"
                min="0.01"
                step="0.01"
                className="admin-form-input"
                placeholder="12500"
                value={variantForm.selling_price}
                onChange={(e) => {
                  setVariantForm((prev) => ({ ...prev, selling_price: e.target.value }));
                  if (variantErrors.selling_price)
                    setVariantErrors((prev) => ({ ...prev, selling_price: null }));
                }}
                required
              />
              {variantErrors.selling_price && (
                <span className="admin-form-error">{variantErrors.selling_price}</span>
              )}
            </div>

            <div className="admin-form-field">
              <label className="admin-form-label" htmlFor="var-stock">
                Stock Quantity <span className="required">*</span>
              </label>
              <input
                id="var-stock"
                type="number"
                min="0"
                step="1"
                className="admin-form-input"
                placeholder="10"
                value={variantForm.stock}
                onChange={(e) => {
                  setVariantForm((prev) => ({ ...prev, stock: e.target.value }));
                  if (variantErrors.stock) setVariantErrors((prev) => ({ ...prev, stock: null }));
                }}
                required
              />
              {variantErrors.stock && (
                <span className="admin-form-error">{variantErrors.stock}</span>
              )}
            </div>
          </div>

          {editingVariant && (
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginTop: 4 }}>
              <input
                id="var-active"
                type="checkbox"
                checked={variantForm.is_active}
                onChange={(e) =>
                  setVariantForm((prev) => ({ ...prev, is_active: e.target.checked }))
                }
              />
              <label htmlFor="var-active" style={{ fontSize: 13, color: '#334155' }}>
                Variant is active and available for customer purchases
              </label>
            </div>
          )}
        </div>
      </Modal>

      {/* Delete Variant Confirmation */}
      <Modal
        isOpen={isDeleteVariantOpen}
        onClose={() => setIsDeleteVariantOpen(false)}
        title="Delete Variant"
        eyebrow="Destructive Action"
        description={`Are you sure you want to delete variant SKU "${variantToDelete?.sku}"?`}
        confirmText="Delete Variant"
        cancelText="Cancel"
        confirmVariant="danger"
        isConfirmLoading={isActionLoading}
        onConfirm={handleDeleteVariant}
      />

      {/* Upload Product Images Modal */}
      <Modal
        isOpen={isUploadImageOpen}
        onClose={() => setIsUploadImageOpen(false)}
        title="Upload Product Images"
        eyebrow="Asset Management"
        description="Select image files from your computer. You may optionally associate them with a specific variant."
        confirmText="Upload"
        cancelText="Cancel"
        isConfirmLoading={isActionLoading}
        onConfirm={handleUploadImages}
      >
        <div className="admin-form-group">
          {variants.length > 0 && (
            <div className="admin-form-field">
              <label className="admin-form-label">Variant Association (Optional)</label>
              <select
                className="admin-form-select"
                value={targetVariantId}
                onChange={(e) => setTargetVariantId(e.target.value)}
              >
                <option value="">Product-Level (All Variants)</option>
                {variants.map((v) => (
                  <option key={v.id} value={v.id}>
                    SKU: {v.sku} {v.size ? `(${v.size})` : ''} {v.color ? `(${v.color})` : ''}
                  </option>
                ))}
              </select>
            </div>
          )}

          <input
            type="file"
            ref={fileInputRef}
            multiple
            accept="image/*"
            style={{ display: 'none' }}
            onChange={handleImageFileSelect}
          />

          <div
            className="admin-logo-upload-box"
            onClick={() => fileInputRef.current?.click()}
          >
            <Upload size={24} color="#64748b" />
            <span style={{ fontSize: 13, fontWeight: 500, color: '#0f172a' }}>
              Click to browse or drop product images
            </span>
            <span style={{ fontSize: 11, color: '#94a3b8' }}>
              JPG, PNG, WebP supported. Multiple selection allowed.
            </span>
          </div>

          {imagePreviews.length > 0 && (
            <div style={{ display: 'flex', gap: 10, flexWrap: 'wrap', marginTop: 10 }}>
              {imagePreviews.map((src, idx) => (
                <div
                  key={idx}
                  style={{
                    position: 'relative',
                    width: 70,
                    height: 70,
                    borderRadius: 6,
                    border: '1px solid #e2e8f0',
                    overflow: 'hidden',
                  }}
                >
                  <img src={src} alt="Preview" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                  <button
                    type="button"
                    onClick={() => removeSelectedFile(idx)}
                    style={{
                      position: 'absolute',
                      top: 2,
                      right: 2,
                      background: 'rgba(0,0,0,0.6)',
                      color: '#fff',
                      border: 'none',
                      borderRadius: '50%',
                      width: 18,
                      height: 18,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      cursor: 'pointer',
                    }}
                  >
                    <X size={10} />
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      </Modal>

      {/* Delete Image Confirmation */}
      <Modal
        isOpen={isDeleteImageOpen}
        onClose={() => setIsDeleteImageOpen(false)}
        title="Delete Product Image"
        eyebrow="Destructive Action"
        description="Are you sure you want to remove this image from the gallery? If this is the primary image, another image will need to be designated."
        confirmText="Delete Image"
        cancelText="Cancel"
        confirmVariant="danger"
        isConfirmLoading={isActionLoading}
        onConfirm={handleDeleteImage}
      />
    </div>
  );
}
