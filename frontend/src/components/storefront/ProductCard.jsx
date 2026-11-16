import { Link } from 'react-router-dom';
import { ArrowUpRight } from 'lucide-react';
import { formatCurrencyINR } from '../../utils/formatters';
import { StatusPill } from '../common/Badge';
import './ProductCard.css';

export function ProductCard({ product }) {
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

  const productUrl = slug ? `/products/${slug}` : `/products/id/${id}`;
  const imageUrl = primary_image?.image_url || '/placeholder-product.svg';

  const priceDisplay =
    min_price > 0
      ? min_price === max_price || max_price === 0
        ? formatCurrencyINR(min_price)
        : `${formatCurrencyINR(min_price)} – ${formatCurrencyINR(max_price)}`
      : 'Inquire Price';

  return (
    <article className="drape-product-card" data-testid={`product-card-${id}`}>
      <Link to={productUrl} className="drape-product-card__image-link" tabIndex="-1" aria-hidden="true">
        <div className="drape-product-card__image-wrapper">
          <img
            src={imageUrl}
            alt={name}
            className="drape-product-card__image"
            loading="lazy"
            onError={(e) => {
              // Graceful fallback to styled neutral placeholder
              e.currentTarget.style.display = 'none';
              e.currentTarget.parentElement.classList.add('is-fallback');
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
        </div>
      </Link>

      <div className="drape-product-card__content">
        <div className="drape-product-card__meta">
          <span className="drape-product-card__brand">
            {brand_name || 'DRAPE ATELIER'}
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
    </article>
  );
}
