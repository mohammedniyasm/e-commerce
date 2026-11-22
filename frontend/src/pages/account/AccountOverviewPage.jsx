import React from 'react';
import { useOutletContext, Link } from 'react-router-dom';
import {
  User,
  Package,
  MapPin,
  Shield,
  HelpCircle,
  ArrowRight,
  ExternalLink,
  CheckCircle2,
  AlertCircle,
} from 'lucide-react';
import { Button } from '../../components/common/Button';
import { Badge, StatusPill } from '../../components/common/Badge';
import { SkeletonLoader } from '../../components/common/SkeletonLoader';
import { formatCurrencyINR, formatDate } from '../../utils/formatters';
import './AccountOverviewPage.css';

export function AccountOverviewPage() {
  const context = useOutletContext() || {};
  const {
    profile,
    recentOrders = [],
    addresses = [],
    isLoading = false,
    openVerifyEmailModal,
  } = context;

  const defaultAddress =
    addresses.find((addr) => addr.is_default) || addresses[0] || null;

  if (isLoading) {
    return (
      <div className="account-overview-loading" data-testid="account-overview-loading">
        <SkeletonLoader lines={10} />
      </div>
    );
  }

  return (
    <div className="account-overview-page" data-testid="account-overview-page">
      {/* Overview Unverified Notice Banner */}
      {!profile?.email_verified_at && (
        <div className="account-overview-unverified-alert" data-testid="overview-unverified-alert">
          <div className="account-overview-unverified-alert__icon">
            <AlertCircle size={20} />
          </div>
          <div className="account-overview-unverified-alert__text">
            <strong>Account Email Unverified</strong>
            <span>Your email address ({profile?.email}) is not yet verified. Verify it to receive account updates.</span>
          </div>
          <Button
            variant="primary"
            size="sm"
            onClick={openVerifyEmailModal}
            data-testid="overview-alert-verify-btn"
          >
            VERIFY EMAIL
          </Button>
        </div>
      )}

      {/* SECTION 1: Personal Details (Figma 32:10998 Section 1) */}
      <section className="account-card" aria-labelledby="personal-details-heading">
        <div className="account-card__header">
          <div className="account-card__title-group">
            <User size={18} className="account-card__icon" aria-hidden="true" />
            <h2 id="personal-details-heading" className="account-card__title">
              Personal Details
            </h2>
          </div>
          <Link
            to="/account/profile"
            className="account-card__header-link"
            data-testid="edit-personal-details-link"
          >
            <span>Edit</span>
            <ArrowRight size={14} />
          </Link>
        </div>

        <div className="account-card__body">
          <div className="account-details-grid">
            <div className="account-detail-item">
              <span className="account-detail-label">FULL NAME</span>
              <span className="account-detail-value" data-testid="detail-fullname">
                {profile?.name || '—'}
              </span>
            </div>

            <div className="account-detail-item">
              <span className="account-detail-label">EMAIL ADDRESS</span>
              <div className="account-detail-value-with-badge">
                <span data-testid="detail-email">{profile?.email || '—'}</span>
                {profile?.email_verified_at ? (
                  <Badge variant="success" size="sm">
                    VERIFIED
                  </Badge>
                ) : (
                  <div className="account-detail-unverified-group">
                    <Badge variant="warning" size="sm">
                      UNVERIFIED
                    </Badge>
                    <button
                      type="button"
                      className="account-overview-verify-btn"
                      onClick={openVerifyEmailModal}
                      data-testid="overview-verify-email-btn"
                    >
                      Verify Now
                    </button>
                  </div>
                )}
              </div>
            </div>

            <div className="account-detail-item">
              <span className="account-detail-label">PHONE NUMBER</span>
              <span className="account-detail-value" data-testid="detail-phone">
                {profile?.phone || 'Not provided'}
              </span>
            </div>

            <div className="account-detail-item">
              <span className="account-detail-label">ACCOUNT</span>
              <span className="account-detail-value">
                Active Customer
              </span>
            </div>
          </div>
        </div>
      </section>

      {/* SECTION 2: Recent Orders (Figma 32:10998 Section 2) */}
      <section className="account-card" aria-labelledby="recent-orders-heading">
        <div className="account-card__header">
          <div className="account-card__title-group">
            <Package size={18} className="account-card__icon" aria-hidden="true" />
            <h2 id="recent-orders-heading" className="account-card__title">
              Recent Orders
            </h2>
          </div>
          {recentOrders.length > 0 && (
            <Link
              to="/account/orders"
              className="account-card__header-link"
              data-testid="view-all-orders-link"
            >
              <span>View All Orders</span>
              <ArrowRight size={14} />
            </Link>
          )}
        </div>

        <div className="account-card__body">
          {recentOrders.length > 0 ? (
            <div className="account-recent-orders-list">
              {recentOrders.map((order) => {
                const orderId = order.id || order.order_id;
                const status = (order.order_status || 'placed').toLowerCase();
                const pillVariant =
                  status === 'delivered'
                    ? 'success'
                    : status === 'cancelled'
                    ? 'danger'
                    : status === 'shipped' || status === 'in_transit'
                    ? 'info'
                    : 'warning';

                const itemsCount =
                  order.items?.length ||
                  order.items_count ||
                  (Array.isArray(order.items) ? order.items.length : 1);

                return (
                  <div key={orderId} className="account-order-snippet" data-testid="order-snippet">
                    <div className="account-order-snippet__meta">
                      <span className="account-order-snippet__number">
                        #{order.order_number || orderId}
                      </span>
                      <span className="account-order-snippet__date">
                        Placed on {formatDate(order.placed_at || order.created_at)}
                      </span>
                    </div>

                    <div className="account-order-snippet__status">
                      <StatusPill status={status} variant={pillVariant} />
                    </div>

                    <div className="account-order-snippet__summary">
                      <span className="account-order-snippet__items">
                        {itemsCount} {itemsCount === 1 ? 'item' : 'items'}
                      </span>
                      <span className="account-order-snippet__total">
                        {formatCurrencyINR(order.total_amount)}
                      </span>
                    </div>

                    <div className="account-order-snippet__action">
                      <Link
                        to={`/account/orders/${orderId}`}
                        className="account-order-snippet__link"
                        data-testid={`view-order-${orderId}`}
                      >
                        <span>View Details</span>
                        <ArrowRight size={14} />
                      </Link>
                    </div>
                  </div>
                );
              })}
            </div>
          ) : (
            <div className="account-empty-block" data-testid="orders-empty-state">
              <Package size={32} className="account-empty-icon" />
              <p className="account-empty-text">
                You have not placed an order yet. Browse our products to get started.
              </p>
              <Link to="/catalog">
                <Button variant="outline" size="sm">
                  EXPLORE COLLECTION
                </Button>
              </Link>
            </div>
          )}
        </div>
      </section>

      {/* SECTION 3: Saved Addresses (Figma 32:10998 Section 3) */}
      <section className="account-card" aria-labelledby="saved-addresses-heading">
        <div className="account-card__header">
          <div className="account-card__title-group">
            <MapPin size={18} className="account-card__icon" aria-hidden="true" />
            <h2 id="saved-addresses-heading" className="account-card__title">
              Saved Addresses
            </h2>
          </div>
          <Link
            to="/account/addresses"
            className="account-card__header-link"
            data-testid="manage-addresses-link"
          >
            <span>Manage Addresses</span>
            <ArrowRight size={14} />
          </Link>
        </div>

        <div className="account-card__body">
          {defaultAddress ? (
            <div className="account-default-address-card" data-testid="default-address-card">
              <div className="account-default-address-header">
                <strong className="account-default-address-name">
                  {defaultAddress.name}
                </strong>
                {defaultAddress.is_default && (
                  <Badge variant="neutral" size="sm">
                    DEFAULT DELIVERY
                  </Badge>
                )}
              </div>
              <p className="account-default-address-lines">
                {defaultAddress.address_line1}
                {defaultAddress.address_line2 ? `, ${defaultAddress.address_line2}` : ''}
                <br />
                {defaultAddress.city}, {defaultAddress.state} {defaultAddress.postal_code}
              </p>
              <p className="account-default-address-phone">
                Tel: {defaultAddress.phone}
              </p>
            </div>
          ) : (
            <div className="account-empty-block" data-testid="addresses-empty-state">
              <MapPin size={32} className="account-empty-icon" />
              <p className="account-empty-text">No delivery addresses saved yet.</p>
              <Link to="/account/addresses">
                <Button variant="outline" size="sm">
                  + ADD NEW ADDRESS
                </Button>
              </Link>
            </div>
          )}
        </div>
      </section>

      {/* SECTION 4: Account Security & Customer Support */}
      <div className="account-support-grid">
        {/* Account Security */}
        <div className="account-card account-card--compact">
          <div className="account-card__header">
            <div className="account-card__title-group">
              <Shield size={18} className="account-card__icon" aria-hidden="true" />
              <h2 className="account-card__title">Account Security</h2>
            </div>
          </div>
          <div className="account-card__body">
            <p className="account-card__desc">
              Keep your account secure and manage your password.
            </p>
            <div className="account-card__action-row">
              <Link to="/account/profile">
                <Button variant="outline" size="sm" data-testid="security-manage-btn">
                  MANAGE SECURITY
                </Button>
              </Link>
            </div>
          </div>
        </div>

        {/* Customer support */}
        <div className="account-card account-card--compact">
          <div className="account-card__header">
            <div className="account-card__title-group">
              <HelpCircle size={18} className="account-card__icon" aria-hidden="true" />
              <h2 className="account-card__title">Need Help?</h2>
            </div>
          </div>
          <div className="account-card__body">
            <p className="account-card__desc">
              Contact our support team for help with orders, sizing, or returns.
            </p>
            <div className="account-card__action-row">
              <a
                href="mailto:concierge@drapeatelier.com"
                className="account-concierge-link"
              >
                <span>Email Customer Support</span>
                <ExternalLink size={14} />
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
