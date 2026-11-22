import { useEffect, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { ArrowLeft, Check, Package, Save } from 'lucide-react';
import {
  createProduct,
  getProductById,
  listBrands,
  listCategories,
  updateProduct,
} from '../../../api/adminCatalogService';
import { Button } from '../../../components/common/Button';
import { useToast } from '../../../context/ToastContext';
import './AdminProductsPage.css';

export function AdminProductFormPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { showToast } = useToast();

  const isEdit = Boolean(id);

  const [formData, setFormData] = useState({
    name: '',
    short_description: '',
    description: '',
    category_id: '',
    brand_id: '',
  });

  const [categories, setCategories] = useState([]);
  const [brands, setBrands] = useState([]);
  const [errors, setErrors] = useState({});
  const [serverError, setServerError] = useState('');
  const [isLoading, setIsLoading] = useState(isEdit);
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    async function init() {
      try {
        const [catsRes, brandsRes] = await Promise.allSettled([
          listCategories({ page: 1, limit: 100 }),
          listBrands({ page: 1, limit: 100 }),
        ]);

        if (catsRes.status === 'fulfilled' && catsRes.value?.categories) {
          setCategories(catsRes.value.categories.filter((c) => c.is_active));
        }
        if (brandsRes.status === 'fulfilled' && brandsRes.value?.brands) {
          setBrands(brandsRes.value.brands.filter((b) => b.is_active));
        }

        if (isEdit) {
          const product = await getProductById(id);
          if (product) {
            setFormData({
              name: product.name || '',
              short_description: product.short_description || '',
              description: product.description || '',
              category_id: String(product.category_id || ''),
              brand_id: String(product.brand_id || ''),
            });
          }
        }
      } catch (err) {
        showToast({
          type: 'error',
          title: 'Error Loading Form',
          message: err?.response?.data?.message || err?.message || 'Failed to initialize form data.',
        });
      } finally {
        setIsLoading(false);
      }
    }

    init();
  }, [id, isEdit]);

  function handleChange(field, value) {
    setFormData((prev) => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: null }));
    }
    if (serverError) {
      setServerError('');
    }
  }

  function validate() {
    const errs = {};
    if (!formData.name.trim()) {
      errs.name = 'Product name is required.';
    }
    if (!formData.category_id) {
      errs.category_id = 'Category selection is required.';
    }
    if (!formData.brand_id) {
      errs.brand_id = 'Brand selection is required.';
    }
    setErrors(errs);
    return Object.keys(errs).length === 0;
  }

  async function handleSubmit(e) {
    e.preventDefault();
    if (isSubmitting) return;

    setServerError('');
    if (!validate()) return;

    setIsSubmitting(true);
    try {
      const payload = {
        name: formData.name.trim(),
        category_id: Number(formData.category_id),
        brand_id: Number(formData.brand_id),
        short_description: formData.short_description.trim() || undefined,
        description: formData.description.trim() || undefined,
      };

      if (isEdit) {
        await updateProduct(id, payload);
        showToast({
          type: 'success',
          title: 'Product Updated',
          message: `Product "${formData.name}" has been updated.`,
        });
        navigate(`/admin/products/${id}`);
      } else {
        const created = await createProduct(payload);
        showToast({
          type: 'success',
          title: 'Product Created',
          message: `Product "${formData.name}" has been created. You can now add variants and upload images.`,
        });
        if (created?.id) {
          navigate(`/admin/products/${created.id}`);
        } else {
          navigate('/admin/products');
        }
      }
    } catch (err) {
      const msg =
        err?.response?.data?.message ||
        err?.message ||
        'Failed to save product. Please verify all inputs.';
      setServerError(msg);
    } finally {
      setIsSubmitting(false);
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

  return (
    <div className="admin-catalog-page">
      <div className="admin-page-header">
        <div className="admin-page-title-group">
          <Link
            to={isEdit ? `/admin/products/${id}` : '/admin/products'}
            className="admin-back-link"
          >
            <ArrowLeft size={14} />
            <span>{isEdit ? 'Back to Product Details' : 'Back to Products'}</span>
          </Link>
          <h1 className="admin-page-title">
            {isEdit ? `Edit Product: ${formData.name || ''}` : 'Create New Product'}
          </h1>
          <p className="admin-page-subtitle">
            Configure product catalog details. Variants and inventory are managed on the product profile.
          </p>
        </div>
      </div>

      <div className="admin-detail-card">
        {serverError && (
          <div
            style={{
              background: '#fef2f2',
              border: '1px solid #fecaca',
              color: '#ef4444',
              borderRadius: 8,
              padding: '12px 16px',
              fontSize: 13,
              marginBottom: 20,
            }}
          >
            {serverError}
          </div>
        )}

        <form onSubmit={handleSubmit} className="admin-form-group">
          <div className="admin-form-field">
            <label className="admin-form-label" htmlFor="prod-name">
              Product Title <span className="required">*</span>
            </label>
            <input
              id="prod-name"
              type="text"
              className="admin-form-input"
              placeholder="e.g. Silk Georgette Trench Coat"
              value={formData.name}
              onChange={(e) => handleChange('name', e.target.value)}
              required
            />
            {errors.name && <span className="admin-form-error">{errors.name}</span>}
          </div>

          <div className="admin-product-select-grid">
            <div className="admin-form-field">
              <label className="admin-form-label" htmlFor="prod-category">
                Category <span className="required">*</span>
              </label>
              <select
                id="prod-category"
                className="admin-form-select"
                value={formData.category_id}
                onChange={(e) => handleChange('category_id', e.target.value)}
                required
              >
                <option value="">Select Category</option>
                {categories.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
              {errors.category_id && (
                <span className="admin-form-error">{errors.category_id}</span>
              )}
            </div>

            <div className="admin-form-field">
              <label className="admin-form-label" htmlFor="prod-brand">
                Brand <span className="required">*</span>
              </label>
              <select
                id="prod-brand"
                className="admin-form-select"
                value={formData.brand_id}
                onChange={(e) => handleChange('brand_id', e.target.value)}
                required
              >
                <option value="">Select Brand</option>
                {brands.map((b) => (
                  <option key={b.id} value={b.id}>
                    {b.name}
                  </option>
                ))}
              </select>
              {errors.brand_id && (
                <span className="admin-form-error">{errors.brand_id}</span>
              )}
            </div>
          </div>

          <div className="admin-form-field">
            <label className="admin-form-label" htmlFor="prod-short-desc">
              Short Description (Brief summary)
            </label>
            <input
              id="prod-short-desc"
              type="text"
              className="admin-form-input"
              placeholder="For example: Cotton T-shirt, relaxed fit..."
              value={formData.short_description}
              onChange={(e) => handleChange('short_description', e.target.value)}
            />
          </div>

          <div className="admin-form-field">
            <label className="admin-form-label" htmlFor="prod-desc">
              Full Description
            </label>
            <textarea
              id="prod-desc"
              className="admin-form-textarea"
              style={{ minHeight: 120 }}
              placeholder="Describe the product, its materials, fit, and care instructions..."
              value={formData.description}
              onChange={(e) => handleChange('description', e.target.value)}
            />
          </div>

          <div className="admin-form-actions">
            <Link
              to={isEdit ? `/admin/products/${id}` : '/admin/products'}
              className="drape-btn drape-btn--secondary drape-btn--md admin-button-link"
            >
              Cancel
            </Link>
            <Button
              type="submit"
              variant="primary"
              size="md"
              leftIcon={<Save size={16} />}
              isLoading={isSubmitting}
            >
              {isEdit ? 'Save Changes' : 'Create Product'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
