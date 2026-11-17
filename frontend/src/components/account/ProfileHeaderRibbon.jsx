import React from 'react';
import { Link } from 'react-router-dom';
import { User, Edit3 } from 'lucide-react';
import { Button } from '../common/Button';
import './ProfileHeaderRibbon.css';

export function ProfileHeaderRibbon({
  profile,
  stats = {},
  isLoading = false,
  onVerifyClick = null,
}) {
  const name = profile?.name || 'Customer Account';
  const email = profile?.email || '';
  const phone = profile?.phone || '';

  // Generate initials (e.g. "Arjun Nair" -> "AN")
  const initials = name
    .trim()
    .split(/\s+/)
    .map((word) => word[0])
    .filter(Boolean)
    .slice(0, 2)
    .join('')
    .toUpperCase() || 'DA';

  const ordersCount = String(stats.ordersCount ?? 0).padStart(2, '0');
  const wishlistCount = String(stats.wishlistCount ?? 0).padStart(2, '0');
  const addressesCount = String(stats.addressesCount ?? 0).padStart(2, '0');

  return (
    <section className="profile-header-ribbon" aria-label="Customer Profile Summary">
      <div className="profile-header-ribbon__inner">
        {/* Left: Avatar & Identity */}
        <div className="profile-header-ribbon__identity">
          <div className="profile-header-ribbon__avatar" aria-hidden="true">
            {profile?.profile_image ? (
              <img
                src={profile.profile_image}
                alt={name}
                className="profile-header-ribbon__avatar-img"
              />
            ) : (
              <span className="profile-header-ribbon__initials">{initials}</span>
            )}
          </div>

          <div className="profile-header-ribbon__details">
            <h1 className="profile-header-ribbon__name" data-testid="profile-name">
              {isLoading ? 'Loading profile...' : name}
            </h1>
            <div className="profile-header-ribbon__contact">
              {email && (
                <span className="profile-header-ribbon__email" data-testid="profile-email">
                  {email}
                </span>
              )}
              {profile?.email_verified_at ? (
                <span
                  className="profile-ribbon-badge profile-ribbon-badge--verified"
                  data-testid="profile-verified-badge"
                >
                  ✓ Verified
                </span>
              ) : email ? (
                <button
                  type="button"
                  className="profile-ribbon-badge profile-ribbon-badge--unverified"
                  onClick={onVerifyClick}
                  data-testid="profile-unverified-badge"
                  title="Click to verify your email address"
                >
                  ⚠️ Unverified • Verify Now
                </button>
              ) : null}
              {phone && <span className="profile-header-ribbon__dot">•</span>}
              {phone && (
                <span className="profile-header-ribbon__phone" data-testid="profile-phone">
                  {phone}
                </span>
              )}
            </div>
          </div>
        </div>

        {/* Center/Right: 3 Stat Counters matching Figma 32:10998 */}
        <div className="profile-header-ribbon__stats" role="region" aria-label="Account Statistics">
          <Link to="/account/orders" className="profile-stat-box" data-testid="stat-orders">
            <span className="profile-stat-box__value">{ordersCount}</span>
            <span className="profile-stat-box__label">ORDERS</span>
          </Link>

          <div className="profile-stat-box__divider" aria-hidden="true" />

          <Link to="/wishlist" className="profile-stat-box" data-testid="stat-wishlist">
            <span className="profile-stat-box__value">{wishlistCount}</span>
            <span className="profile-stat-box__label">WISHLIST</span>
          </Link>

          <div className="profile-stat-box__divider" aria-hidden="true" />

          <Link to="/account/addresses" className="profile-stat-box" data-testid="stat-addresses">
            <span className="profile-stat-box__value">{addressesCount}</span>
            <span className="profile-stat-box__label">SAVED ADDRESSES</span>
          </Link>
        </div>

        {/* Action: Edit Profile */}
        <div className="profile-header-ribbon__actions">
          <Link to="/account/profile">
            <Button
              variant="outline"
              size="sm"
              leftIcon={<Edit3 size={14} />}
              data-testid="edit-profile-btn"
            >
              EDIT PROFILE
            </Button>
          </Link>
        </div>
      </div>
    </section>
  );
}
