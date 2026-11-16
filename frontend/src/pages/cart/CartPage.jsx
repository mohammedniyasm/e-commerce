import { Link, useNavigate } from 'react-router-dom';
import {
  AlertTriangle,
  ArrowRight,
  Minus,
  Plus,
  ShoppingBag,
  Trash2,
} from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { useCart } from '../../context/CartContext';
import { formatCurrencyINR } from '../../utils/formatters';
import { StatusPill } from '../../components/common/Badge';
import { Button } from '../../components/common/Button';
import './CartPage.css';

export function CartPage() {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();
  const { cart, updateQuantity, removeItem, clearCart, isLoading } = useCart();

  const items = cart?.items || [];
  const isEmpty = items.length === 0;

  function handleProceedToCheckout() {
    if (cart.has_unavailable_items) {
      return;
    }
    if (!isAuthenticated) {
      navigate('/login', { state: { from: '/checkout' } });
    } else {
      navigate('/checkout');
    }
  }

  return (
    <div className="cart-page" data-testid="cart-page">
      <div className="cart-header">
        <h1 className="cart-title">Shopping Bag</h1>
        <p className="cart-subtitle">
          {isEmpty
            ? 'Your shopping bag is empty'
            : `${cart.total_quantity || items.length} item(s) selected`}
        </p>
      </div>

      {isEmpty ? (
        <div className="cart-empty-state">
          <div className="cart-empty-icon-wrapper" aria-hidden="true">
            <ShoppingBag size={48} />
          </div>
          <h2 className="cart-empty-title">Your Bag is Empty</h2>
          <p className="cart-empty-desc">
            Explore our curated collections and timeless tailoring to add silhouettes to your personal wardrobe.
          </p>
          <Link to="/catalog">
            <Button variant="primary" size="lg" rightIcon={<ArrowRight size={14} />}>
              EXPLORE COLLECTION
            </Button>
          </Link>
        </div>
      ) : (
        <div className="cart-layout">
          {/* Items Column */}
          <section className="cart-items-column" aria-label="Cart items">
            {cart.has_unavailable_items && (
              <div className="cart-alert-banner" role="alert">
                <AlertTriangle size={18} />
                <span>
                  Some items in your bag are currently unavailable or have limited inventory. Please update quantities before proceeding.
                </span>
              </div>
            )}

            <div className="cart-items-list">
              {items.map((item) => {
                const itemImg = item.image_url || '/placeholder-product.svg';
                const isItemUnavailable = item.is_available === false;

                return (
                  <article
                    key={item.id || item.variant_id}
                    className={`cart-item-card ${isItemUnavailable ? 'is-unavailable' : ''}`}
                    data-testid={`cart-item-${item.variant_id}`}
                  >
                    {/* Thumbnail */}
                    <Link
                      to={item.product_slug ? `/products/${item.product_slug}` : `/products/id/${item.product_id}`}
                      className="cart-item-image-link"
                      tabIndex="-1"
                      aria-hidden="true"
                    >
                      <img
                        src={itemImg}
                        alt={item.product_name}
                        className="cart-item-image"
                        onError={(e) => {
                          e.currentTarget.style.display = 'none';
                        }}
                      />
                    </Link>

                    {/* Content */}
                    <div className="cart-item-details">
                      <div className="cart-item-header">
                        <div>
                          <span className="cart-item-brand">
                            {item.brand_name || 'DRAPE ATELIER'}
                          </span>
                          <h3 className="cart-item-title">
                            <Link
                              to={item.product_slug ? `/products/${item.product_slug}` : `/products/id/${item.product_id}`}
                            >
                              {item.product_name}
                            </Link>
                          </h3>
                        </div>

                        <button
                          type="button"
                          className="cart-item-remove-btn"
                          onClick={() => removeItem(item.id, item.variant_id)}
                          aria-label={`Remove ${item.product_name} from bag`}
                          disabled={isLoading}
                        >
                          <Trash2 size={16} />
                        </button>
                      </div>

                      {/* Variant Specs */}
                      <div className="cart-item-specs">
                        {item.size && (
                          <span className="cart-spec-pill">SIZE: {item.size}</span>
                        )}
                        {item.color && (
                          <span className="cart-spec-pill">COLOR: {item.color}</span>
                        )}
                        {item.sku && (
                          <span className="cart-spec-sku">SKU: {item.sku}</span>
                        )}
                      </div>

                      {/* Status Warnings */}
                      {isItemUnavailable && (
                        <div className="cart-item-stock-warning">
                          <StatusPill variant="danger" label="Unavailable" />
                          <span>{item.availability_reason || 'Out of stock in atelier'}</span>
                        </div>
                      )}

                      {/* Footer: Quantity & Line Total */}
                      <div className="cart-item-footer">
                        <div className="cart-qty-control" aria-label="Adjust quantity">
                          <button
                            type="button"
                            className="cart-qty-btn"
                            onClick={() =>
                              updateQuantity(
                                item.id,
                                Math.max(1, item.quantity - 1),
                                item.variant_id,
                              )
                            }
                            disabled={item.quantity <= 1 || isLoading}
                            aria-label="Decrease quantity"
                          >
                            <Minus size={12} />
                          </button>
                          <span className="cart-qty-val">{item.quantity}</span>
                          <button
                            type="button"
                            className="cart-qty-btn"
                            onClick={() =>
                              updateQuantity(
                                item.id,
                                item.quantity + 1,
                                item.variant_id,
                              )
                            }
                            disabled={
                              Boolean(item.available_stock && item.quantity >= item.available_stock) ||
                              isLoading
                            }
                            aria-label="Increase quantity"
                          >
                            <Plus size={12} />
                          </button>
                        </div>

                        <div className="cart-item-pricing">
                          <span className="cart-item-unit-price">
                            {formatCurrencyINR(item.unit_price)} each
                          </span>
                          <span className="cart-item-line-total">
                            {formatCurrencyINR(item.line_total || item.unit_price * item.quantity)}
                          </span>
                        </div>
                      </div>
                    </div>
                  </article>
                );
              })}
            </div>

            <div className="cart-actions-bottom">
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  clearCart()
                }
                }
                disabled={isLoading}
              >
                Clear Entire Bag
              </Button>
              <Link to="/catalog" className="cart-continue-link">
                &larr; Continue Discovering Collection
              </Link>
            </div>
          </section>

          {/* Summary Column */}
          <aside className="cart-summary-column" aria-label="Order summary">
            <div className="cart-summary-card">
              <h2 className="cart-summary-title">Summary</h2>

              <div className="cart-summary-rows">
                <div className="cart-summary-row">
                  <span>Available Subtotal</span>
                  <span className="cart-summary-val">
                    {formatCurrencyINR(cart.available_subtotal)}
                  </span>
                </div>
                <div className="cart-summary-row">
                  <span>Estimated Shipping</span>
                  <span className="cart-summary-val cart-free-shipping">
                    Complimentary
                  </span>
                </div>
                <div className="cart-summary-row">
                  <span>Domestic Duties &amp; Taxes</span>
                  <span className="cart-summary-val">Included</span>
                </div>
              </div>

              <div className="cart-summary-divider" />

              <div className="cart-summary-row cart-total-row">
                <span>Total Amount</span>
                <span className="cart-total-amount">
                  {formatCurrencyINR(cart.available_subtotal)}
                </span>
              </div>

              <p className="cart-tax-notice">
                Authoritative totals, coupon discounts, and address selection applied during checkout.
              </p>

              <Button
                type="button"
                variant="primary"
                size="lg"
                fullWidth
                disabled={cart.has_unavailable_items || isEmpty || isLoading}
                onClick={handleProceedToCheckout}
                rightIcon={<ArrowRight size={14} />}
              >
                {!isAuthenticated ? 'SIGN IN TO CHECKOUT' : 'PROCEED TO CHECKOUT'}
              </Button>

              <div className="cart-summary-guarantees">
                {/* <span>• Insured carbon-neutral domestic dispatches</span> */}
                {/* <span>• 256-bit encrypted checkout</span> */}
              </div>
            </div>
          </aside>
        </div>
      )}
    </div>
  );
}
