import { useState } from 'react';
import { Link } from 'react-router-dom';
import {
  ArrowRight,
  Heart,
  Plus,
  ShoppingBag,
  Trash2,
  FolderMinus,
  RefreshCw,
} from 'lucide-react';
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
    deleteWishlist,
    clearWishlist,
    removeItem,
    isLoading,
  } = useWishlist();
  const { addToCart } = useCart();
  const { showToast } = useToast();

  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isClearModalOpen, setIsClearModalOpen] = useState(false);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [newListName, setNewListName] = useState('');
  const [createError, setCreateError] = useState('');
  const [movingItemId, setMovingItemId] = useState(null);
  const [removingItemId, setRemovingItemId] = useState(null);

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

  async function handleRemoveProduct(item) {
    setRemovingItemId(item.id || item.product_id);
    try {
      await removeItem(item.id, item.product_id);
    } finally {
      setRemovingItemId(null);
    }
  }

  async function handleConfirmClearWishlist() {
    try {
      await clearWishlist(activeWishlistId || currentWishlist?.id);
    } finally {
      setIsClearModalOpen(false);
    }
  }

  async function handleConfirmDeleteWishlist() {
    try {
      await deleteWishlist(activeWishlistId || currentWishlist?.id);
    } finally {
      setIsDeleteModalOpen(false);
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
      {/* Header with Title and Collection Actions */}
      <div className="wishlist-header">
        <div className="wishlist-header__titles">
          <div className="wishlist-header__eyebrow">CURATED SILHOUETTES</div>
          <h1 className="wishlist-title">
            My Wishlist{currentWishlist?.name && currentWishlist.name !== 'My Wishlist' ? ` — ${currentWishlist.name}` : ''}
          </h1>
          <p className="wishlist-subtitle">
            {isEmpty
              ? 'No pieces currently saved in this collection.'
              : `${items.length} ${items.length === 1 ? 'piece' : 'pieces'} curated in this collection.`}
          </p>
        </div>

        <div className="wishlist-header__actions">
          {!isEmpty && (
            <Button
              variant="outline"
              size="sm"
              leftIcon={<Trash2 size={14} />}
              onClick={() => setIsClearModalOpen(true)}
              data-testid="clear-wishlist-btn"
            >
              CLEAR COLLECTION
            </Button>
          )}

          {isAuthenticated && wishlists.length > 0 && (
            <Button
              variant="outline"
              size="sm"
              leftIcon={<FolderMinus size={14} />}
              onClick={() => setIsDeleteModalOpen(true)}
              data-testid="delete-wishlist-btn"
            >
              DELETE COLLECTION
            </Button>
          )}

          {isAuthenticated && (
            <Button
              variant="primary"
              size="sm"
              leftIcon={<Plus size={16} />}
              onClick={() => setIsCreateModalOpen(true)}
              data-testid="create-wishlist-btn"
            >
              NEW COLLECTION
            </Button>
          )}
        </div>
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
                key={item.id || item.product_id}
                className={`wishlist-card ${isUnavailable ? 'is-unavailable' : ''}`}
                data-testid={`wishlist-card-${item.id || item.product_id}`}
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
                    onClick={() => handleRemoveProduct(item)}
                    aria-label={`Remove ${item.product_name} from wishlist`}
                    title="Remove piece"
                    disabled={isLoading || removingItemId === (item.id || item.product_id)}
                    data-testid={`remove-item-${item.id || item.product_id}`}
                  >
                    {removingItemId === (item.id || item.product_id) ? (
                      <RefreshCw size={14} className="spin" />
                    ) : (
                      <Trash2 size={16} />
                    )}
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
        title="Create New Collection"
        eyebrow="ATELIER WISHLISTS"
        size="sm"
      >
        <form onSubmit={handleCreateSubmit} className="wishlist-modal-form">
          <FormField
            label="COLLECTION NAME"
            htmlFor="new-wishlist-name"
            error={createError}
            hint="e.g. Summer Capsule, Evening Silhouettes, Work Essentials"
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
              CREATE COLLECTION
            </Button>
          </div>
        </form>
      </Modal>

      {/* Clear Wishlist Confirmation Modal */}
      <Modal
        isOpen={isClearModalOpen}
        onClose={() => setIsClearModalOpen(false)}
        title="Clear Collection"
        eyebrow="COLLECTION MANAGEMENT"
        size="sm"
      >
        <div className="wishlist-modal-confirm" data-testid="clear-wishlist-modal">
          <p className="wishlist-modal-confirm__desc">
            Are you sure you wish to remove all pieces from{' '}
            <strong>"{currentWishlist?.name || 'this collection'}"</strong>? This will clear all
            curated pieces.
          </p>
          <div className="wishlist-modal-confirm__actions">
            <Button
              type="button"
              variant="outline"
              size="md"
              onClick={() => setIsClearModalOpen(false)}
              disabled={isLoading}
            >
              KEEP PIECES
            </Button>
            <Button
              type="button"
              variant="primary"
              size="md"
              onClick={handleConfirmClearWishlist}
              loading={isLoading}
              data-testid="confirm-clear-wishlist-btn"
            >
              CLEAR COLLECTION
            </Button>
          </div>
        </div>
      </Modal>

      {/* Delete Wishlist Confirmation Modal */}
      <Modal
        isOpen={isDeleteModalOpen}
        onClose={() => setIsDeleteModalOpen(false)}
        title="Delete Collection"
        eyebrow="COLLECTION MANAGEMENT"
        size="sm"
      >
        <div className="wishlist-modal-confirm" data-testid="delete-wishlist-modal">
          <p className="wishlist-modal-confirm__desc">
            Are you sure you wish to permanently delete the collection{' '}
            <strong>"{currentWishlist?.name || 'this collection'}"</strong>? All pieces within it
            will be removed.
          </p>
          <div className="wishlist-modal-confirm__actions">
            <Button
              type="button"
              variant="outline"
              size="md"
              onClick={() => setIsDeleteModalOpen(false)}
              disabled={isLoading}
            >
              CANCEL
            </Button>
            <Button
              type="button"
              variant="primary"
              size="md"
              onClick={handleConfirmDeleteWishlist}
              loading={isLoading}
              data-testid="confirm-delete-wishlist-btn"
            >
              DELETE COLLECTION
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
