import React, { useState, useEffect } from 'react';
import { ShieldAlert, ArrowRight, X, MailCheck } from 'lucide-react';
import { Button } from '../common/Button';
import './UnverifiedAccountBanner.css';

export function UnverifiedAccountBanner({
  profile,
  onVerifyClick,
  isLoading = false,
  className = '',
}) {
  const [isDismissed, setIsDismissed] = useState(false);

  useEffect(() => {
    try {
      const dismissed = sessionStorage.getItem('drape_verify_banner_dismissed');
      if (dismissed === 'true') {
        setIsDismissed(true);
      }
    } catch {
      // ignore storage access errors
    }
  }, []);

  // Do not display if already verified, loading, or no profile
  if (isLoading || !profile || profile.email_verified_at || isDismissed) {
    return null;
  }

  const handleDismiss = () => {
    setIsDismissed(true);
    try {
      sessionStorage.setItem('drape_verify_banner_dismissed', 'true');
    } catch {
      // ignore storage access errors
    }
  };

  return (
    <div
      className={`unverified-banner ${className}`}
      role="alert"
      aria-live="polite"
      data-testid="unverified-account-banner"
    >
      <div className="unverified-banner__content">
        <div className="unverified-banner__icon-wrap" aria-hidden="true">
          <ShieldAlert size={20} className="unverified-banner__icon" />
        </div>

        <div className="unverified-banner__text-wrap">
          <div className="unverified-banner__title-row">
            <span className="unverified-banner__title">ACTION REQUIRED: EMAIL UNVERIFIED</span>
            <span className="unverified-banner__pill">SECURITY NOTICE</span>
          </div>
          <p className="unverified-banner__desc">
            Your atelier account email <strong>{profile.email}</strong> is not verified. Verify
            now to secure your account, receive real-time order tracking notifications, and unlock
            all customer privileges.
          </p>
        </div>
      </div>

      <div className="unverified-banner__actions">
        <Button
          variant="primary"
          size="sm"
          onClick={onVerifyClick}
          rightIcon={<ArrowRight size={14} />}
          data-testid="banner-verify-email-btn"
        >
          VERIFY EMAIL NOW
        </Button>

        <button
          type="button"
          className="unverified-banner__dismiss-btn"
          onClick={handleDismiss}
          aria-label="Dismiss unverified email banner"
          title="Dismiss notification"
        >
          <X size={16} />
        </button>
      </div>
    </div>
  );
}

/**
 * Slide-in toast-like popup prompt for unverified users
 */
export function UnverifiedAccountPopup({
  profile,
  onVerifyClick,
  isLoading = false,
}) {
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    if (isLoading || !profile || profile.email_verified_at) {
      setIsVisible(false);
      return;
    }

    try {
      const dismissed = sessionStorage.getItem('drape_verify_popup_dismissed');
      if (!dismissed) {
        // Show after a brief 1-second delay for smooth entry
        const timer = setTimeout(() => setIsVisible(true), 800);
        return () => clearTimeout(timer);
      }
    } catch {
      setIsVisible(true);
    }
  }, [profile, isLoading]);

  if (!isVisible || !profile || profile.email_verified_at) {
    return null;
  }

  const handleDismiss = () => {
    setIsVisible(false);
    try {
      sessionStorage.setItem('drape_verify_popup_dismissed', 'true');
    } catch {
      // ignore storage access errors
    }
  };

  return (
    <aside
      className="unverified-popup"
      role="complementary"
      aria-label="Account Verification Prompt"
      data-testid="unverified-account-popup"
    >
      <button
        type="button"
        className="unverified-popup__close"
        onClick={handleDismiss}
        aria-label="Close verification prompt"
      >
        <X size={14} />
      </button>

      <div className="unverified-popup__header">
        <div className="unverified-popup__icon-badge">
          <MailCheck size={16} />
        </div>
        <span className="unverified-popup__eyebrow">ATELIER PRIVILEGES</span>
      </div>

      <h4 className="unverified-popup__title">Verify Your Email Address</h4>
      <p className="unverified-popup__text">
        Confirm <strong>{profile.email}</strong> to activate full order delivery notifications and
        priority atelier services.
      </p>

      <div className="unverified-popup__footer">
        <Button
          variant="primary"
          size="sm"
          onClick={() => {
            handleDismiss();
            onVerifyClick();
          }}
          data-testid="popup-verify-email-btn"
          className="w-full"
        >
          VERIFY EMAIL NOW
        </Button>
      </div>
    </aside>
  );
}
