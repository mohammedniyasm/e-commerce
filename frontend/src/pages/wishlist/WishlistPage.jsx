import { useState } from 'react';
import { Link } from 'react-router-dom';
import { ArrowRight, Heart, Plus, ShoppingBag, Trash2 } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { useCart } from '../../context/CartContext';
import { useToast } from '../../context/ToastContext';
import { useWishlist } from '../../context/WishlistContext';
import { formatCurrencyINR } from '../../utils/formatters';
import { Badge } from '../../components/common/Badge';
import { Button } from '../../components/common/Button';
import { FormField } from '../../components/common/FormField';
import { Input } from '../../components/common/Input';
import { Modal } from '../../components/common/Modal';
import './WishlistPage.css';

export function WishlistPage() {
  const { isAuthenticated } = useAuth();
  const {
    wishlists,
    currentWishlist,
    activeWishlistId,
    selectWishlist,
    createWishlist,
    removeItem,
    isLoading,
  } = useWishlist();
  const { addToCart } = useCart();
  const { showToast } = useToast();

  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [newListName, setNewListName] = useState('');
  const [createError, setCreateError] = useState('');
  const [movingItemId, setMovingItemId] = useState(null);

  const items = currentWishlist?.items || [];
  const isEmpty = items.length === 0;

  async function handleCreateSubmit(e) {
    e.preventDefault();
    if (!newListName.trim()) {
      setCreateError('Please enter a wishlist name.');
      return;
    }
    setCreateError('');
    try {
      await createWishlist(newListName.trim());
      setNewListName('');
      setIsCreateModalOpen(false);
    } catch (err) {
      setCreateError(err.message || 'Failed to create wishlist.');
    }
  }

  async function handleMoveToBag(item) {
    setMovingItemId(item.id);
    try {
      // If item has default_variant_id or variant_id, add to cart directly
      const variantId = item.variant_id || item.default_variant_id;
      if (variantId) {
        await addToCart({ variantId, quantity: 1 });
        await removeItem(item.id, item.product_id);
        showToast({
          title: 'Moved to Bag',
          message: `${item.product_name || 'Item'} has been moved to your shopping bag.`,
          type: 'success',
        });
      } else {
        // Product requires size/color selection at PDP
        showToast({
          title: 'Select Size & Color',
          message: 'Please choose your preferred variant from the product page.',
          type: 'info',
        });
      }
    } catch (err) {
      showToast({
        title: 'Error',
        message: err.message || 'Unable to move item to bag.',
        type: 'error',
      });
    } finally {
      setMovingItemId(null);
    }
  }

  return (
    <div className="wishlist-page" data-testid="wishlist-page">
      <div className="wishlist-header">
        <div className="wishlist-header__titles">
          <h1 className="wishlist-title">My Wishlist</h1>
          <p className="wishlist-subtitle">
            {isAuthenticated
              ? 'Organize your favorite silhouettes across custom curated collections.'
              : 'Save pieces to your temporary wishlist. Sign in to sync across devices.'}
          </p>
        </div>

        {isAuthenticated && (
          <Button
            variant="outline"
            size="sm"
            leftIcon={<Plus size={16} />}
            onClick={() => setIsCreateModalOpen(true)}
            data-testid="create-wishlist-btn"
          >
            NEW LIST
          </Button>
        )}
      </div>

      {/* Multi-wishlist selector tabs */}
      {isAuthenticated && wishlists.length > 1 && (
        <nav className="wishlist-tabs" aria-label="Wishlists">
          {wishlists.map((list) => {
            const isActive = list.id === activeWishlistId;
            return (
              <button
                key={list.id}
                type="button"
                className={`wishlist-tab ${isActive ? 'is-active' : ''}`}
                onClick={() => selectWishlist(list.id)}
                aria-pressed={isActive}
                data-testid={`wishlist-tab-${list.id}`}
              >
                <span>{list.name}</span>
                {list.item_count !== undefined && (
                  <span className="wishlist-tab__count">({list.item_count})</span>
                )}
              </button>
            );
          })}
        </nav>
      )}

      {/* Empty State */}
      {isEmpty ? (
        <div className="wishlist-empty-state" data-testid="wishlist-empty-state">
          <div className="wishlist-empty-icon-wrapper" aria-hidden="true">
            <Heart size={44} />
          </div>
          <h2 className="wishlist-empty-title">Your Wishlist is Empty</h2>
          <p className="wishlist-empty-desc">
            Curate your personal collection of timeless tailoring and quiet luxury silhouettes.
          </p>
          <Link to="/catalog">
            <Button variant="primary" size="lg" rightIcon={<ArrowRight size={14} />}>
              EXPLORE COLLECTION
            </Button>
          </Link>
        </div>
      ) : (
        <div className="wishlist-grid" data-testid="wishlist-grid">
          {items.map((item) => {
            const itemImg = item.image_url || '/placeholder-product.svg';
            const itemPrice = item.starting_price || item.min_price || item.price || 0;
            const isUnavailable = item.is_available === false;
            const productHref = item.product_slug
              ? `/products/${item.product_slug}`
              : `/products/id/${item.product_id}`;

            return (
              <article
                key={item.id}
                className={`wishlist-card ${isUnavailable ? 'is-unavailable' : ''}`}
                data-testid={`wishlist-card-${item.id}`}
              >
                {/* Image & Badges */}
                <div className="wishlist-card__image-container">
                  <Link to={productHref} tabIndex="-1" aria-hidden="true">
                    <img
                      src={itemImg}
                      alt={item.product_name}
                      className="wishlist-card__image"
                      loading="lazy"
                      onError={(e) => {
                        e.currentTarget.style.display = 'none';
                      }}
                    />
                  </Link>

                  {/* Remove Button */}
                  <button
                    type="button"
                    className="wishlist-card__remove-btn"
                    onClick={() => removeItem(item.id, item.product_id)}
                    aria-label={`Remove ${item.product_name} from wishlist`}
                    title="Remove item"
                    data-testid={`remove-item-${item.id}`}
                  >
                    <Trash2 size={16} />
                  </button>

                  {isUnavailable && (
                    <div className="wishlist-card__badge-overlay">
                      <Badge variant="warning" size="sm">
                        OUT OF STOCK
                      </Badge>
                    </div>
                  )}
                </div>

                {/* Content */}
                <div className="wishlist-card__body">
                  <h3 className="wishlist-card__title">
                    <Link to={productHref} className="wishlist-card__title-link">
                      {item.product_name}
                    </Link>
                  </h3>

                  <div className="wishlist-card__price">
                    {itemPrice > 0 ? formatCurrencyINR(itemPrice) : 'Price on selection'}
                  </div>

                  {/* Move to Bag or View Product */}
                  <div className="wishlist-card__actions">
                    {item.variant_id || item.default_variant_id ? (
                      <Button
                        variant="primary"
                        size="sm"
                        fullWidth
                        leftIcon={<ShoppingBag size={14} />}
                        disabled={isUnavailable || movingItemId === item.id || isLoading}
                        loading={movingItemId === item.id}
                        onClick={() => handleMoveToBag(item)}
                      >
                        MOVE TO BAG
                      </Button>
                    ) : (
                      <Link to={productHref} className="wishlist-card__view-btn">
                        <Button variant="secondary" size="sm" fullWidth>
                          SELECT SIZE
                        </Button>
                      </Link>
                    )}
                  </div>
                </div>
              </article>
            );
          })}
        </div>
      )}

      {/* Create Wishlist Modal */}
      <Modal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        title="Create New Wishlist"
      >
        <form onSubmit={handleCreateSubmit} className="wishlist-modal-form">
          <FormField
            label="Wishlist Name"
            htmlFor="new-wishlist-name"
            error={createError}
            hint="e.g. Summer Tailoring, Evening Silhouettes, Work Essentials"
          >
            <Input
              id="new-wishlist-name"
              type="text"
              placeholder="e.g. Autumn Capsule"
              value={newListName}
              onChange={(e) => setNewListName(e.target.value)}
              autoFocus
              required
            />
          </FormField>

          <div className="wishlist-modal-actions">
            <Button
              type="button"
              variant="outline"
              size="md"
              onClick={() => setIsCreateModalOpen(false)}
            >
              CANCEL
            </Button>
            <Button type="submit" variant="primary" size="md">
              CREATE WISHLIST
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
