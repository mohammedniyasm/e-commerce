import { useEffect, useRef, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import {
  ArrowDown,
  ArrowLeft,
  ArrowUp,
  Check,
  Crop,
  Edit2,
  Image as ImageIcon,
  Layers,
  Package,
  Plus,
  Power,
  Star,
  Tags,
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
import { ProductImageEditor } from '../../../components/admin/ProductImageEditor';
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
  const [draggedImageId, setDraggedImageId] = useState(null);
  const [dropTargetImageId, setDropTargetImageId] = useState(null);
  const [assignmentImageId, setAssignmentImageId] = useState(null);
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
  const [editingImageIndex, setEditingImageIndex] = useState(null);
  const [pendingImageVariants, setPendingImageVariants] = useState({});
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
    e.target.value = '';
    if (!files.length) return;

    setSelectedImageFiles((prev) => [...prev, ...files]);

    const newPreviews = files.map((f) => URL.createObjectURL(f));
    setImagePreviews((prev) => [...prev, ...newPreviews]);
  }

  function removeSelectedFile(index) {
    if (imagePreviews[index]) {
      URL.revokeObjectURL(imagePreviews[index]);
    }
    setSelectedImageFiles((prev) => prev.filter((_, i) => i !== index));
    setImagePreviews((prev) => prev.filter((_, i) => i !== index));
    setEditingImageIndex((current) =>
      current === index ? null : current > index ? current - 1 : current,
    );
  }

  function handleApplyImageEdit(index, editedFile) {
    const previewUrl = URL.createObjectURL(editedFile);
    if (imagePreviews[index]) {
      URL.revokeObjectURL(imagePreviews[index]);
    }
    setSelectedImageFiles((current) =>
      current.map((file, fileIndex) => (fileIndex === index ? editedFile : file)),
    );
    setImagePreviews((current) =>
      current.map((preview, previewIndex) =>
        previewIndex === index ? previewUrl : preview,
      ),
    );
    setEditingImageIndex(null);
  }

  function closeUploadModal() {
    setIsUploadImageOpen(false);
    setEditingImageIndex(null);
    setSelectedImageFiles([]);
    imagePreviews.forEach((preview) => URL.revokeObjectURL(preview));
    setImagePreviews([]);
    setTargetVariantId('');
  }

  async function handleUploadImages() {
    if (editingImageIndex !== null) {
      showToast({
        type: 'warning',
        title: 'Finish Image Crop',
        message: 'Apply or cancel the current image crop before uploading.',
      });
      return;
    }

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
      imagePreviews.forEach((preview) => URL.revokeObjectURL(preview));
      setImagePreviews([]);
      setEditingImageIndex(null);
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

  async function handleAssignImageToVariant(img) {
    const variantId = pendingImageVariants[img.id];
    if (variantId === undefined || isActionLoading) return;

    setIsActionLoading(true);
    try {
      const updatedImage = await updateProductImage(img.id, {
        variant_id: variantId ? Number(variantId) : null,
      });
      setImages((current) =>
        current.map((image) => (image.id === img.id ? updatedImage : image)),
      );
      setPendingImageVariants((current) => {
        const next = { ...current };
        delete next[img.id];
        return next;
      });
      setAssignmentImageId(null);
      showToast({
        type: 'success',
        title: 'Image Assignment Updated',
        message: variantId
          ? 'Image is now assigned to the selected variant.'
          : 'Image is now available at the product level.',
      });
    } catch (err) {
      showToast({
        type: 'error',
        title: 'Assignment Failed',
        message: err?.response?.data?.message || err?.message || 'Could not assign image to variant.',
      });
    } finally {
      setIsActionLoading(false);
    }
  }

  async function handleReorderImages(sourceIndex, targetIndex) {
    if (
      isActionLoading ||
      sourceIndex < 0 ||
      targetIndex < 0 ||
      sourceIndex >= images.length ||
      targetIndex >= images.length ||
      sourceIndex === targetIndex
    ) {
      return;
    }

    let reorderedImages = [...images];
    const direction = targetIndex > sourceIndex ? 1 : -1;
    setIsActionLoading(true);
    try {
      let currentIndex = sourceIndex;
      while (currentIndex !== targetIndex) {
        const nextIndex = currentIndex + direction;
        const movingImage = reorderedImages[currentIndex];
        const displacedImage = reorderedImages[nextIndex];

        await updateProductImage(movingImage.id, {
          display_order: displacedImage.display_order,
        });

        reorderedImages[currentIndex] = {
          ...displacedImage,
          display_order: movingImage.display_order,
        };
        reorderedImages[nextIndex] = {
          ...movingImage,
          display_order: displacedImage.display_order,
        };
        currentIndex = nextIndex;
      }

      setImages(reorderedImages);
      showToast({
        type: 'success',
        title: 'Image Order Updated',
        message: 'Product gallery order has been updated.',
      });
    } catch (err) {
      let message =
        err?.response?.data?.message || err?.message || 'Could not reorder product images.';
      try {
        const refreshed = await listProductImages(product.id);
        setImages(refreshed);
      } catch (refreshError) {
        setImages(reorderedImages);
        message = `${message} The gallery could not be refreshed: ${
          refreshError?.message || 'refresh failed'
        }`;
      }
      showToast({
        type: 'error',
        title: 'Reorder Failed',
        message,
      });
    } finally {
      setIsActionLoading(false);
      setDraggedImageId(null);
      setDropTargetImageId(null);
    }
  }

  function handleImageDragStart(event, image) {
    event.dataTransfer.effectAllowed = 'move';
    event.dataTransfer.setData('text/plain', String(image.id));
    setDraggedImageId(image.id);
  }

  function handleImageDrop(event, targetIndex) {
    event.preventDefault();
    const draggedId = event.dataTransfer.getData('text/plain') || draggedImageId;
    const sourceIndex = images.findIndex((image) => String(image.id) === String(draggedId));
    setDropTargetImageId(null);
    void handleReorderImages(sourceIndex, targetIndex);
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
              Drag to reorder images. Hover a photo to assign its variant, set it as primary, or remove it.
            </p>
            <Button
              variant="primary"
              size="sm"
              leftIcon={<Upload size={14} />}
              onClick={() => {
                imagePreviews.forEach((preview) => URL.revokeObjectURL(preview));
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
              {images.map((img, index) => (
                <div key={img.id} className="admin-managed-image">
                  <div
                    className={`admin-image-card ${
                      draggedImageId === img.id ? 'is-dragging' : ''
                    } ${dropTargetImageId === img.id ? 'is-drop-target' : ''}`}
                    onDragOver={(event) => {
                      event.preventDefault();
                      event.dataTransfer.dropEffect = 'move';
                      setDropTargetImageId((current) =>
                        current === img.id ? current : img.id,
                      );
                    }}
                    onDrop={(event) => handleImageDrop(event, index)}
                  >
                    {img.is_primary && (
                      <span className="admin-primary-tag">
                        <Star size={10} style={{ display: 'inline', marginRight: 2 }} /> Primary
                      </span>
                    )}
                    <img
                      src={img.image_url}
                      alt={product.name}
                      draggable={!isActionLoading}
                      onDragStart={(event) => handleImageDragStart(event, img)}
                      onDragEnd={() => {
                        setDraggedImageId(null);
                        setDropTargetImageId(null);
                      }}
                    />

                    <div className="admin-image-overlay">
                      <div className="admin-image-overlay__actions">
                        {variants.length > 0 && (
                          <button
                            type="button"
                            className={`admin-image-assign-trigger ${
                              assignmentImageId === img.id ? 'is-active' : ''
                            }`}
                            aria-label={`Assign ${product.name} image to a variant or product level`}
                            aria-expanded={assignmentImageId === img.id}
                            onClick={(event) => {
                              event.stopPropagation();
                              setAssignmentImageId((current) =>
                                current === img.id ? null : img.id,
                              );
                            }}
                          >
                            <Tags size={14} />
                            <span>Assign</span>
                          </button>
                        )}
                        <button
                          type="button"
                          className="admin-icon-btn"
                          aria-label={`Move image ${index + 1} up`}
                          title="Move Up"
                          disabled={isActionLoading || index === 0}
                          onClick={() => void handleReorderImages(index, index - 1)}
                        >
                          <ArrowUp size={14} />
                        </button>
                        <button
                          type="button"
                          className="admin-icon-btn"
                          aria-label={`Move image ${index + 1} down`}
                          title="Move Down"
                          disabled={isActionLoading || index === images.length - 1}
                          onClick={() => void handleReorderImages(index, index + 1)}
                        >
                          <ArrowDown size={14} />
                        </button>
                        {!img.is_primary && (
                          <button
                            type="button"
                            className="admin-icon-btn"
                            aria-label={`Set ${product.name} image as primary`}
                            title="Set as Primary"
                            disabled={isActionLoading}
                            onClick={() => handleSetPrimaryImage(img)}
                          >
                            <Star size={14} />
                          </button>
                        )}
                        <button
                          type="button"
                          className="admin-icon-btn danger"
                          aria-label={`Delete ${product.name} image`}
                          title="Delete Image"
                          disabled={isActionLoading}
                          onClick={() => {
                            setImageToDelete(img);
                            setIsDeleteImageOpen(true);
                          }}
                        >
                          <Trash2 size={14} />
                        </button>
                      </div>

                      {variants.length > 0 && assignmentImageId === img.id && (
                        <div className="admin-image-variant-assignment">
                          <label htmlFor={`image-variant-${img.id}`}>
                            ASSIGN IMAGE TO
                          </label>
                          <select
                            id={`image-variant-${img.id}`}
                            className="admin-form-select"
                            aria-label={`Assign image ${index + 1} to a variant or product level`}
                            value={
                              pendingImageVariants[img.id] ??
                              (img.variant_id ? String(img.variant_id) : '')
                            }
                            disabled={isActionLoading}
                            onChange={(event) =>
                              setPendingImageVariants((current) => ({
                                ...current,
                                [img.id]: event.target.value,
                              }))
                            }
                          >
                            <option value="">Product-level (all variants)</option>
                            {variants.map((variant) => (
                              <option key={variant.id} value={String(variant.id)}>
                                {variant.sku}
                                {variant.color ? ` — ${variant.color}` : ''}
                                {variant.size ? ` / ${variant.size}` : ''}
                              </option>
                            ))}
                          </select>
                          <Button
                            type="button"
                            variant="secondary"
                            size="sm"
                            disabled={
                              isActionLoading ||
                              pendingImageVariants[img.id] === undefined ||
                              pendingImageVariants[img.id] === String(img.variant_id || '')
                            }
                            onClick={() => void handleAssignImageToVariant(img)}
                          >
                            Assign
                          </Button>
                        </div>
                      )}
                    </div>
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
        onClose={closeUploadModal}
        title="Upload Product Images"
        eyebrow="Asset Management"
        description="Crop images to the storefront's 3:4 product-card ratio, then upload and assign them."
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
            accept="image/jpeg,image/png,image/webp"
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
            <div className="admin-upload-previews">
              {imagePreviews.map((src, idx) => (
                <div key={`${src}-${idx}`} className="admin-upload-preview">
                  <img src={src} alt={`Selected image ${idx + 1}`} />
                  <button
                    type="button"
                    className="admin-upload-preview__crop"
                    aria-label={`Crop image ${idx + 1} for product listings`}
                    onClick={() => setEditingImageIndex(idx)}
                  >
                    <Crop size={12} />
                    Crop
                  </button>
                  <button
                    type="button"
                    className="admin-upload-preview__remove"
                    aria-label={`Remove image ${idx + 1}`}
                    onClick={() => removeSelectedFile(idx)}
                  >
                    <X size={12} />
                  </button>
                </div>
              ))}
            </div>
          )}

          {editingImageIndex !== null &&
            selectedImageFiles[editingImageIndex] &&
            imagePreviews[editingImageIndex] && (
              <ProductImageEditor
                key={`${editingImageIndex}-${imagePreviews[editingImageIndex]}`}
                file={selectedImageFiles[editingImageIndex]}
                src={imagePreviews[editingImageIndex]}
                onCancel={() => setEditingImageIndex(null)}
                onApply={(editedFile) =>
                  handleApplyImageEdit(editingImageIndex, editedFile)
                }
              />
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
