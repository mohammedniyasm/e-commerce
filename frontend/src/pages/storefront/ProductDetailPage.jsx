import { useEffect, useMemo, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import {
  ArrowLeft,
  Check,
  ChevronRight,
  Heart,
  Minus,
  Plus,
  RefreshCw,
  Share2,
  ShieldCheck,
  ShoppingBag,
  Star,
  Truck,
} from 'lucide-react';
import {
  getProductReviews,
  getRelatedProducts,
  getStoreProductById,
  getStoreProductBySlug,
} from '../../api/productService';
import { useCart } from '../../context/CartContext';
import { useWishlist } from '../../context/WishlistContext';
import { useToast } from '../../context/ToastContext';
import { formatCurrencyINR, formatDate } from '../../utils/formatters';
import { StatusPill } from '../../components/common/Badge';
import { Button } from '../../components/common/Button';
import { SkeletonLoader } from '../../components/common/SkeletonLoader';
import { ProductCard } from '../../components/storefront/ProductCard';
import './ProductDetailPage.css';

export function ProductDetailPage() {
  const { slug, id } = useParams();
  const navigate = useNavigate();
  const { showToast } = useToast();
  const { addToCart } = useCart();
  const { isInWishlist, toggleWishlist } = useWishlist();

  const [product, setProduct] = useState(null);
  const [reviewsData, setReviewsData] = useState({ reviews: [], total: 0 });
  const [relatedProducts, setRelatedProducts] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  // Active Variant Selection
  const [selectedColor, setSelectedColor] = useState(null);
  const [selectedSize, setSelectedSize] = useState(null);
  const [quantity, setQuantity] = useState(1);
  const [activeImageIndex, setActiveImageIndex] = useState(0);
  const [activeTab, setActiveTab] = useState('details'); // 'details' | 'fabrication' | 'shipping'
  const [isAddingToCart, setIsAddingToCart] = useState(false);

  // Fetch product by slug or id
  useEffect(() => {
    let isMounted = true;

    async function loadProduct() {
      try {
        setIsLoading(true);
        setError(null);
        let data = null;

        if (slug) {
          data = await getStoreProductBySlug(slug);
        } else if (id) {
          data = await getStoreProductById(id);
        }

        if (isMounted && data) {
          setProduct(data);

          // Auto-select initial color & size if available
          const firstVariant = data.variants?.[0];
          if (firstVariant) {
            setSelectedColor(firstVariant.color || null);
            setSelectedSize(firstVariant.size || null);
          }

          // Fetch reviews & recommendations
          Promise.allSettled([
            getProductReviews(data.id, { limit: 6 }),
            getRelatedProducts(data.id, { limit: 4 }),
          ]).then(([revs, rels]) => {
            if (isMounted) {
              if (revs.status === 'fulfilled' && revs.value) {
                setReviewsData(revs.value);
              }
              if (rels.status === 'fulfilled' && Array.isArray(rels.value)) {
                setRelatedProducts(rels.value);
              }
            }
          });
        }
      } catch (err) {
        if (isMounted) {
          setError(err.message || 'Product not found or unavailable.');
        }
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    }

    loadProduct();

    return () => {
      isMounted = false;
    };
  }, [slug, id]);

  // Distinct colors and sizes from variants
  const availableColors = useMemo(() => {
    if (!product?.variants) return [];
    const colors = new Set();
    product.variants.forEach((v) => {
      if (v.color) colors.add(v.color);
    });
    return Array.from(colors);
  }, [product]);

  const availableSizes = useMemo(() => {
    if (!product?.variants) return [];
    const sizes = new Set();
    product.variants.forEach((v) => {
      if (v.size) sizes.add(v.size);
    });
    return Array.from(sizes);
  }, [product]);

  // Find exact matching variant
  const selectedVariant = useMemo(() => {
    if (!product?.variants || product.variants.length === 0) return null;

    // If both color and size exist, match both
    if (availableColors.length > 0 && availableSizes.length > 0) {
      return (
        product.variants.find(
          (v) => v.color === selectedColor && v.size === selectedSize,
        ) || null
      );
    }

    // If only color exists
    if (availableColors.length > 0) {
      return product.variants.find((v) => v.color === selectedColor) || null;
    }

    // If only size exists
    if (availableSizes.length > 0) {
      return product.variants.find((v) => v.size === selectedSize) || null;
    }

    // Single default variant without color/size attributes
    return product.variants[0] || null;
  }, [product, selectedColor, selectedSize, availableColors, availableSizes]);

  // Unified Gallery Images
  const allImages = useMemo(() => {
    if (!product) return [];
    const imgs = [];
    if (product.images && product.images.length > 0) {
      imgs.push(...product.images);
    }
    // Also include any variant specific images
    if (product.variants) {
      product.variants.forEach((v) => {
        if (v.images) {
          v.images.forEach((img) => {
            if (!imgs.some((existing) => existing.image_url === img.image_url)) {
              imgs.push(img);
            }
          });
        }
      });
    }
    return imgs.length > 0 ? imgs : [{ id: 0, image_url: '/placeholder-product.svg' }];
  }, [product]);

  useEffect(() => {
    const variantImages = selectedVariant?.images?.filter((image) => image.image_url) || [];
    const productImage =
      product?.images?.find((image) => image.is_primary && image.image_url) ||
      product?.images?.find((image) => image.image_url);
    const imageToSelect =
      variantImages.find((image) => image.is_primary) ||
      variantImages[0] ||
      productImage ||
      allImages[0];

    if (!imageToSelect) return;

    const imageIndex = allImages.findIndex(
      (image) => image.image_url === imageToSelect.image_url,
    );
    if (imageIndex >= 0) {
      setActiveImageIndex(imageIndex);
    }
  }, [selectedVariant, product, allImages]);

  // Reset quantity if stock changes
  useEffect(() => {
    setQuantity(1);
  }, [selectedVariant]);

  // Add to cart handler
  async function handleAddToCart() {
    if (!selectedVariant) {
      showToast({
        title: 'Variant Required',
        message: 'Please select an available size and color before adding to bag.',
        type: 'warning',
      });
      return;
    }

    if (!selectedVariant.in_stock || selectedVariant.stock <= 0) {
      showToast({
        title: 'Item Unavailable',
        message: 'The selected option is currently out of stock.',
        type: 'error',
      });
      return;
    }

    setIsAddingToCart(true);
    try {
      await addToCart({
        variant_id: selectedVariant.id,
        quantity,
        productDetails: {
          id: product.id,
          name: product.name,
          slug: product.slug,
          category_name: product.category?.name,
          brand_name: product.brand?.name,
          sku: selectedVariant.sku,
          size: selectedVariant.size,
          color: selectedVariant.color,
          selling_price: selectedVariant.selling_price,
          min_price: selectedVariant.selling_price,
          image_url: allImages[0]?.image_url,
        },
      });
    } finally {
      setIsAddingToCart(false);
    }
  }

  if (isLoading) {
    return (
      <div className="pdp-page" data-testid="pdp-loading">
        <div className="pdp-breadcrumbs">
          <SkeletonLoader height="16px" width="220px" />
        </div>
        <div className="pdp-grid">
          <div className="pdp-gallery-column">
            <SkeletonLoader className="pdp-gallery-skeleton" height="560px" />
          </div>
          <div className="pdp-details-column">
            <SkeletonLoader height="14px" width="30%" />
            <SkeletonLoader height="36px" width="80%" style={{ margin: '12px 0' }} />
            <SkeletonLoader height="24px" width="40%" style={{ marginBottom: '24px' }} />
            <SkeletonLoader height="120px" />
            <SkeletonLoader height="48px" style={{ marginTop: '24px' }} />
          </div>
        </div>
      </div>
    );
  }

  if (error || !product) {
    return (
      <div className="pdp-page pdp-error-page">
        <h2 className="pdp-error-title">Product Not Found</h2>
        <p className="pdp-error-desc">
          {error || 'This product is no longer available.'}
        </p>
        <Link to="/catalog">
          <Button variant="primary" size="md">
            Back to Products
          </Button>
        </Link>
      </div>
    );
  }

  // Calculate pricing display
  const hasVariant = Boolean(selectedVariant);
  const currentSellingPrice = hasVariant
    ? selectedVariant.selling_price
    : product.min_price;
  const currentMrp = hasVariant ? selectedVariant.mrp : product.max_price;
  const hasDiscount = currentMrp > currentSellingPrice;
  const discountPercent = hasDiscount
    ? Math.round(((currentMrp - currentSellingPrice) / currentMrp) * 100)
    : 0;

  const isCurrentVariantInStock = hasVariant
    ? selectedVariant.in_stock && selectedVariant.stock > 0
    : product.in_stock;

  return (
    <div className="pdp-page" data-testid="pdp-page">
      {/* Breadcrumbs */}
      <nav className="pdp-breadcrumbs" aria-label="Breadcrumb">
        <Link to="/">Home</Link>
        <span className="pdp-breadcrumbs__sep">/</span>
        <Link to="/catalog">Collection</Link>
        {product.category?.name && (
          <>
            <span className="pdp-breadcrumbs__sep">/</span>
            <Link to={`/catalog?category_id=${product.category.id}`}>
              {product.category.name}
            </Link>
          </>
        )}
        <span className="pdp-breadcrumbs__sep">/</span>
        <span className="pdp-breadcrumbs__current">{product.name}</span>
      </nav>

      {/* Main 2-Column Product Detail Layout */}
      <div className="pdp-grid">
        {/* Left Column: Interactive Image Gallery */}
        <div className="pdp-gallery-column">
          {/* Main Large Image */}
          <div className="pdp-main-image-wrapper">
            <img
              src={allImages[activeImageIndex]?.image_url || '/placeholder-product.svg'}
              alt={product.name}
              className="pdp-main-image"
              onError={(e) => {
                e.currentTarget.style.display = 'none';
                e.currentTarget.parentElement.classList.add('is-fallback');
              }}
            />
            {!isCurrentVariantInStock && (
              <div className="pdp-stock-overlay-badge">
                <StatusPill variant="danger" label="Out of Stock" />
              </div>
            )}
          </div>

          {/* Thumbnails Strip */}
          {allImages.length > 1 && (
            <div className="pdp-thumbnails-strip" aria-label="Product image thumbnails">
              {allImages.map((img, idx) => (
                <button
                  key={img.id || idx}
                  type="button"
                  className={`pdp-thumbnail-btn ${
                    idx === activeImageIndex ? 'is-active' : ''
                  }`}
                  onClick={() => setActiveImageIndex(idx)}
                  aria-label={`View image ${idx + 1}`}
                >
                  <img
                    src={img.image_url}
                    alt=""
                    className="pdp-thumbnail-img"
                    loading="lazy"
                  />
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Right Column: Product Specs & Variant Selection */}
        <div className="pdp-details-column">
          {/* Brand & Category Header */}
          <div className="pdp-brand-row">
            <span className="pdp-brand-name">
              {product.brand?.name || 'DRAPE'}
            </span>
            {product.category?.name && (
              <span className="pdp-category-pill">{product.category.name}</span>
            )}
          </div>

          {/* Product Title */}
          <h1 className="pdp-title">{product.name}</h1>

          {/* Ratings Summary */}
          {reviewsData.total > 0 && (
            <div className="pdp-rating-summary">
              <div className="pdp-stars">
                {[1, 2, 3, 4, 5].map((s) => (
                  <Star key={s} size={14} fill="#eab308" color="#eab308" />
                ))}
              </div>
              <span className="pdp-review-count">
                ({reviewsData.total} verified reviews)
              </span>
            </div>
          )}

          {/* Price Section */}
          <div className="pdp-price-section">
            <span className="pdp-selling-price">
              {formatCurrencyINR(currentSellingPrice)}
            </span>
            {hasDiscount && (
              <>
                <span className="pdp-mrp-price">{formatCurrencyINR(currentMrp)}</span>
                <span className="pdp-discount-badge">{discountPercent}% OFF</span>
              </>
            )}
            {!hasVariant && product.min_price !== product.max_price && (
              <span className="pdp-price-note">
                (Range across variants: {formatCurrencyINR(product.min_price)} –{' '}
                {formatCurrencyINR(product.max_price)})
              </span>
            )}
          </div>

          <p className="pdp-tax-note">Inclusive of all duties and domestic taxes</p>

          <hr className="pdp-divider" />

          {/* Variant Selection 1: Color */}
          {availableColors.length > 0 && (
            <div className="pdp-variant-group">
              <div className="pdp-variant-label-row">
                <span className="pdp-variant-label">COLOR:</span>
                <span className="pdp-variant-value">{selectedColor || 'Select'}</span>
              </div>
              <div className="pdp-color-options" role="radiogroup" aria-label="Color options">
                {availableColors.map((color) => (
                  <button
                    key={color}
                    type="button"
                    className={`pdp-color-pill ${selectedColor === color ? 'is-selected' : ''}`}
                    onClick={() => setSelectedColor(color)}
                    role="radio"
                    aria-checked={selectedColor === color}
                  >
                    <span>{color}</span>
                    {selectedColor === color && <Check size={12} />}
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Variant Selection 2: Size */}
          {availableSizes.length > 0 && (
            <div className="pdp-variant-group">
              <div className="pdp-variant-label-row">
                <span className="pdp-variant-label">SIZE:</span>
                <span className="pdp-variant-value">{selectedSize || 'Select'}</span>
              </div>
              <div className="pdp-size-options" role="radiogroup" aria-label="Size options">
                {availableSizes.map((size) => {
                  const variantForSize = product.variants?.find(
                    (v) =>
                      v.size === size && (!selectedColor || v.color === selectedColor),
                  );
                  const isAvailable = variantForSize && variantForSize.stock > 0;

                  return (
                    <button
                      key={size}
                      type="button"
                      className={`pdp-size-btn ${
                        selectedSize === size ? 'is-selected' : ''
                      } ${!isAvailable ? 'is-out-of-stock' : ''}`}
                      onClick={() => setSelectedSize(size)}
                      role="radio"
                      aria-checked={selectedSize === size}
                      disabled={!isAvailable}
                    >
                      <span>{size}</span>
                    </button>
                  );
                })}
              </div>
            </div>
          )}

          {/* Stock Availability Indicator */}
          <div className="pdp-stock-status">
            {selectedVariant ? (
              selectedVariant.in_stock && selectedVariant.stock > 0 ? (
                <div className="pdp-stock-indicator is-in-stock">
                  <span className="pdp-stock-dot" />
                  <span>
                    In Stock — {selectedVariant.stock <= 5
                      ? `Only ${selectedVariant.stock} left`
                      : `${selectedVariant.stock} units available`}
                  </span>
                </div>
              ) : (
                <div className="pdp-stock-indicator is-out-of-stock">
                  <span className="pdp-stock-dot" />
                  <span>Currently Out of Stock</span>
                </div>
              )
            ) : (
              <span className="pdp-stock-hint">
                Please select your preferred color and size above.
              </span>
            )}
          </div>

          {/* Quantity Selector & Action Buttons */}
          <div className="pdp-actions-row">
            {/* Quantity */}
            <div className="pdp-quantity-selector" aria-label="Select quantity">
              <button
                type="button"
                className="pdp-qty-btn"
                onClick={() => setQuantity((q) => Math.max(1, q - 1))}
                disabled={quantity <= 1 || !isCurrentVariantInStock}
                aria-label="Decrease quantity"
              >
                <Minus size={14} />
              </button>
              <span className="pdp-qty-number">{quantity}</span>
              <button
                type="button"
                className="pdp-qty-btn"
                onClick={() =>
                  setQuantity((q) =>
                    selectedVariant?.stock ? Math.min(selectedVariant.stock, q + 1) : q + 1,
                  )
                }
                disabled={
                  !isCurrentVariantInStock ||
                  (selectedVariant?.stock && quantity >= selectedVariant.stock)
                }
                aria-label="Increase quantity"
              >
                <Plus size={14} />
              </button>
            </div>

            {/* Add to Bag Button */}
            <Button
              type="button"
              variant="primary"
              size="lg"
              fullWidth
              isLoading={isAddingToCart}
              disabled={!isCurrentVariantInStock}
              onClick={handleAddToCart}
              leftIcon={<ShoppingBag size={16} />}
            >
              {!isCurrentVariantInStock ? 'OUT OF STOCK' : 'ADD TO BAG'}
            </Button>

            {/* Wishlist Toggle Button */}
            <button
              type="button"
              className={`pdp-wishlist-toggle-btn ${isInWishlist(product.id) ? 'is-active' : ''}`}
              onClick={() => toggleWishlist(product)}
              aria-label={isInWishlist(product.id) ? 'Remove from wishlist' : 'Add to wishlist'}
            >
              <Heart
                size={20}
                fill={isInWishlist(product.id) ? '#000000' : 'none'}
                color={isInWishlist(product.id) ? '#000000' : 'currentColor'}
              />
            </button>
          </div>

          {/* Value Highlights */}
          <div className="pdp-highlights">
            <div className="pdp-highlight-item">
              <Truck size={16} />
              <span>Insured tracking on all domestic dispatches</span>
            </div>
            <div className="pdp-highlight-item">
              <ShieldCheck size={16} />
              <span>Return &amp; exchange requests supported on delivered items</span>
            </div>
          </div>

          <hr className="pdp-divider" />

          {/* Product Tabs (Details / Fabrication / Shipping) */}
          <div className="pdp-tabs">
            <div className="pdp-tabs-nav">
              <button
                type="button"
                className={`pdp-tab-btn ${activeTab === 'details' ? 'is-active' : ''}`}
                onClick={() => setActiveTab('details')}
              >
                DESCRIPTION
              </button>
              <button
                type="button"
                className={`pdp-tab-btn ${activeTab === 'fabrication' ? 'is-active' : ''}`}
                onClick={() => setActiveTab('fabrication')}
              >
                MATERIALS & CARE
              </button>
              <button
                type="button"
                className={`pdp-tab-btn ${activeTab === 'shipping' ? 'is-active' : ''}`}
                onClick={() => setActiveTab('shipping')}
              >
                DELIVERY &amp; CARE
              </button>
            </div>

            <div className="pdp-tab-content">
              {activeTab === 'details' && (
                <div className="pdp-prose">
                  <p>{product.description || product.short_description || 'Product details are not available yet.'}</p>
                  {selectedVariant?.sku && (
                    <p className="pdp-sku-meta">SKU: {selectedVariant.sku}</p>
                  )}
                </div>
              )}
              {activeTab === 'fabrication' && (
                <div className="pdp-prose">
                  <p>Check the product description for material and care details.</p>
                </div>
              )}
              {activeTab === 'shipping' && (
                <div className="pdp-prose">
                  <p>Delivery estimates are shown at checkout. Tracking details will be available after your order ships.</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Product Reviews Section */}
      <section className="pdp-reviews-section" aria-labelledby="product-reviews-heading">
        <div className="pdp-reviews-header">
          <div>
            <h2 id="product-reviews-heading" className="pdp-section-heading">
              Customer Reviews
            </h2>
            <p className="pdp-section-subtitle">
              Reviews from customers who purchased this product
            </p>
          </div>
        </div>

        {reviewsData.reviews?.length > 0 ? (
          <div className="pdp-reviews-grid">
            {reviewsData.reviews.map((rev) => (
              <div key={rev.id} className="pdp-review-card">
                <div className="pdp-review-card__header">
                  <div className="pdp-stars">
                    {Array.from({ length: 5 }, (_, i) => (
                      <Star
                        key={i}
                        size={12}
                        fill={i < rev.rating ? '#eab308' : '#e5e7eb'}
                        color={i < rev.rating ? '#eab308' : '#e5e7eb'}
                      />
                    ))}
                  </div>
                  {rev.created_at && (
                    <span className="pdp-review-date">{formatDate(rev.created_at)}</span>
                  )}
                </div>
                <p className="pdp-review-comment">“{rev.comment}”</p>
                <span className="pdp-review-author">— {rev.user_name || 'Verified Client'}</span>
              </div>
            ))}
          </div>
        ) : (
          <div className="pdp-empty-reviews">
            <p>No reviews yet. Customers can review this product after purchase.</p>
          </div>
        )}
      </section>

      {/* Recommendations / You May Also Like */}
      {relatedProducts.length > 0 && (
        <section className="pdp-recommendations-section" aria-labelledby="recommended-products">
          <div className="pdp-section-header">
            <h2             id="recommended-products" className="pdp-section-heading">
              You May Also Like
            </h2>
            <p className="pdp-section-subtitle">
              More products to explore
            </p>
          </div>

          <div className="pdp-recommendations-grid">
            {relatedProducts.map((rel) => (
              <ProductCard
                key={rel.id}
                product={{
                  id: rel.id,
                  name: rel.name,
                  slug: rel.slug,
                  brand_name: rel.brand_name,
                  category_name: rel.category_name,
                  min_price: rel.starting_price,
                  max_price: rel.starting_price,
                  primary_image: rel.primary_image_url
                    ? { image_url: rel.primary_image_url }
                    : null,
                  in_stock: rel.is_in_stock,
                }}
              />
            ))}
          </div>
        </section>
      )}
    </div>
  );
}
