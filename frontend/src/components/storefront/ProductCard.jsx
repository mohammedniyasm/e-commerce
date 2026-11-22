import { useState } from 'react';
import { Link } from 'react-router-dom';
import { ArrowUpRight, Heart } from 'lucide-react';
import { createPortal } from 'react-dom';
import { useWishlist } from '../../context/WishlistContext';
import { formatCurrencyINR } from '../../utils/formatters';
import { Button } from '../common/Button';
import { Modal } from '../common/Modal';
import { StatusPill } from '../common/Badge';
import './ProductCard.css';

export function ProductCard({ product }) {
  const {
    isAuthenticated,
    wishlists,
    activeWishlistId,
    isInWishlist,
    toggleWishlist,
    addToWishlist,
    isLoading,
  } = useWishlist();
  const [isCollectionPickerOpen, setIsCollectionPickerOpen] = useState(false);
  const [selectedWishlistId, setSelectedWishlistId] = useState('');

  if (!product) return null;

  const {
    id,
    name,
    slug,
    brand_name,
    category_name,
    min_price = 0,
    max_price = 0,
    primary_image,
    in_stock = true,
  } = product;

  const isWishlisted = isInWishlist(id);
  const hasCollectionChoice = isAuthenticated && !isWishlisted && wishlists.length > 1;
  const productUrl = slug ? `/products/${slug}` : `/products/id/${id}`;
  const imageUrl = primary_image?.image_url || '/placeholder-product.svg';

  const priceDisplay =
    min_price > 0
      ? min_price === max_price || max_price === 0
        ? formatCurrencyINR(min_price)
        : `${formatCurrencyINR(min_price)} – ${formatCurrencyINR(max_price)}`
      : 'Inquire Price';

  function handleWishlistClick(event) {
    event.preventDefault();
    event.stopPropagation();

    if (isWishlisted || !isAuthenticated || wishlists.length <= 1) {
      void toggleWishlist(product);
      return;
    }

    const activeListExists = wishlists.some(
      (wishlist) => String(wishlist.id) === String(activeWishlistId),
    );
    setSelectedWishlistId(
      String(activeListExists ? activeWishlistId : wishlists[0].id),
    );
    setIsCollectionPickerOpen(true);
  }

  async function handleCollectionSubmit(event) {
    event.preventDefault();
    if (!selectedWishlistId) return;

    const added = await addToWishlist(product, Number(selectedWishlistId));
    if (added) setIsCollectionPickerOpen(false);
  }

  return (
    <article className="drape-product-card" data-testid={`product-card-${id}`}>
      <div className="drape-product-card__image-wrapper">
        <Link to={productUrl} className="drape-product-card__image-link" tabIndex="-1" aria-hidden="true">
          <img
            src={imageUrl}
            alt={name}
            className="drape-product-card__image"
            loading="lazy"
            onError={(e) => {
              // Graceful fallback to styled neutral placeholder
              e.currentTarget.style.display = 'none';
              e.currentTarget.closest('.drape-product-card__image-wrapper')?.classList.add('is-fallback');
            }}
          />
          <div className="drape-product-card__badge-overlay">
            {!in_stock ? (
              <StatusPill variant="danger" label="Out of Stock" />
            ) : (
              category_name && (
                <span className="drape-product-card__category-tag">
                  {category_name}
                </span>
              )
            )}
          </div>
        </Link>

        <button
          type="button"
          className={`drape-product-card__wishlist-btn ${isWishlisted ? 'is-wishlisted' : ''}`}
          onClick={handleWishlistClick}
          aria-label={isWishlisted ? `Remove ${name} from wishlist` : `Save ${name} to wishlist`}
          aria-haspopup={hasCollectionChoice ? 'dialog' : undefined}
          aria-expanded={hasCollectionChoice ? isCollectionPickerOpen : undefined}
          title={isWishlisted ? 'Saved to your wishlist' : 'Save to wishlist'}
          data-testid={`wishlist-toggle-${id}`}
        >
          <Heart
            size={16}
            fill={isWishlisted ? '#b48325' : 'none'}
            color={isWishlisted ? '#b48325' : 'currentColor'}
          />
        </button>
      </div>

      <div className="drape-product-card__content">
        <div className="drape-product-card__meta">
          <span className="drape-product-card__brand">
            {brand_name || 'DRAPE'}
          </span>
        </div>

        <h3 className="drape-product-card__title">
          <Link to={productUrl} className="drape-product-card__title-link">
            {name}
          </Link>
        </h3>

        <div className="drape-product-card__footer">
          <span className="drape-product-card__price">{priceDisplay}</span>
          <Link
            to={productUrl}
            className="drape-product-card__action-btn"
            aria-label={`View details for ${name}`}
          >
            <span>View</span>
            <ArrowUpRight size={14} />
          </Link>
        </div>
      </div>

      {isCollectionPickerOpen && createPortal(
        <Modal
          isOpen={isCollectionPickerOpen}
          onClose={() => setIsCollectionPickerOpen(false)}
          title="Choose a collection"
          eyebrow="SAVE PRODUCT"
          description={`Choose where to save ${name}.`}
          size="sm"
        >
          <form className="wishlist-collection-form" onSubmit={handleCollectionSubmit}>
            <label className="wishlist-collection-form__field">
              <span>Collection</span>
              <select
                value={selectedWishlistId}
                onChange={(event) => setSelectedWishlistId(event.target.value)}
                disabled={isLoading}
                required
              >
                {wishlists.map((wishlist) => (
                  <option key={wishlist.id} value={wishlist.id}>
                    {wishlist.name}
                    {wishlist.item_count > 0 ? ` (${wishlist.item_count} items)` : ''}
                  </option>
                ))}
              </select>
            </label>
            <div className="wishlist-collection-form__actions">
              <Button
                type="button"
                variant="outline"
                onClick={() => setIsCollectionPickerOpen(false)}
                disabled={isLoading}
              >
                Cancel
              </Button>
              <Button
                type="submit"
                variant="primary"
                isLoading={isLoading}
                disabled={!selectedWishlistId}
              >
                Save to collection
              </Button>
            </div>
          </form>
        </Modal>,
        document.body,
      )}
    </article>
  );
}
