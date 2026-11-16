import React, { useState, useEffect, useCallback } from 'react';
import {
  Gift,
  Copy,
  Check,
  Share2,
  Users,
  Clock,
  Sparkles,
  ArrowRight,
  ShieldCheck,
  Award,
} from 'lucide-react';
import {
  Button,
  StatusPill,
  Badge,
  SkeletonLoader,
} from '../../components/common';
import { useToast } from '../../context/ToastContext';
import { getReferralDashboard } from '../../api/referralService';
import { formatCurrencyINR, formatDate } from '../../utils/formatters';
import './ReferralPage.css';

export function ReferralPage() {
  const { showToast } = useToast();

  const [dashboardData, setDashboardData] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [hasCopiedCode, setHasCopiedCode] = useState(false);
  const [hasCopiedLink, setHasCopiedLink] = useState(false);

  const fetchDashboard = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await getReferralDashboard({ limit: 20, page: 1 });
      setDashboardData(data);
    } catch (err) {
      showToast({
        title: 'Referral Dashboard Unavailable',
        message: err.message || 'Unable to retrieve referral information.',
        type: 'danger',
      });
    } finally {
      setIsLoading(false);
    }
  }, [showToast]);

  useEffect(() => {
    fetchDashboard();
  }, [fetchDashboard]);

  async function handleCopyCode() {
    const code = dashboardData?.referral_code;
    if (!code) return;
    try {
      await navigator.clipboard.writeText(code);
      setHasCopiedCode(true);
      showToast({
        title: 'Referral Code Copied',
        message: `Code "${code}" copied to clipboard.`,
        type: 'info',
      });
      setTimeout(() => setHasCopiedCode(false), 2500);
    } catch {
      showToast({
        title: 'Copy Failed',
        message: 'Could not write to clipboard.',
        type: 'warning',
      });
    }
  }

  async function handleCopyLink() {
    const link =
      dashboardData?.referral_link ||
      `${window.location.origin}/register?ref=${dashboardData?.referral_code || ''}`;
    try {
      await navigator.clipboard.writeText(link);
      setHasCopiedLink(true);
      showToast({
        title: 'Invitation Link Copied',
        message: 'Your personal invitation link has been copied.',
        type: 'info',
      });
      setTimeout(() => setHasCopiedLink(false), 2500);
    } catch {
      showToast({
        title: 'Copy Failed',
        message: 'Could not write to clipboard.',
        type: 'warning',
      });
    }
  }

  const summary = dashboardData?.summary || {
    total_referrals: 0,
    pending_referrals: 0,
    completed_referrals: 0,
    total_rewards_earned: 0,
  };

  const referrals = dashboardData?.referrals || [];
  const referralCode = dashboardData?.referral_code || '—';

  return (
    <div className="account-referral-page" data-testid="account-referral-page">
      {/* Top Header */}
      <div className="account-referral-header">
        <div>
          <h1 className="account-referral-title">Referral Program</h1>
          <p className="account-referral-subtitle">
            Introduce discerning patrons to Drape Atelier and earn commissioning credits.
          </p>
        </div>
      </div>

      {/* Hero Referral Code Share Box */}
      <div className="referral-hero-card" data-testid="referral-hero-card">
        <div className="referral-hero-content">
          <div className="referral-hero-eyebrow">
            <Gift size={15} />
            <span>EXCLUSIVE ATELIER INITIATIVE</span>
          </div>
          <h2 className="referral-hero-heading">Share Drape with Friends</h2>
          <p className="referral-hero-desc">
            Invite fellow enthusiasts of slow-fashion tailoring. When your referred friend completes
            their first delivered order, rewards credit directly to your Atelier Wallet.
          </p>

          <div className="referral-code-bar">
            <div className="referral-code-display">
              <span className="referral-code-label">YOUR BESPOKE CODE:</span>
              <strong className="referral-code-text" data-testid="referral-code">
                {isLoading ? '...' : referralCode}
              </strong>
            </div>

            <div className="referral-code-actions">
              <Button
                variant="outline"
                size="sm"
                onClick={handleCopyCode}
                leftIcon={hasCopiedCode ? <Check size={14} /> : <Copy size={14} />}
                data-testid="copy-referral-code-btn"
              >
                {hasCopiedCode ? 'COPIED' : 'COPY CODE'}
              </Button>
              <Button
                variant="primary"
                size="sm"
                onClick={handleCopyLink}
                leftIcon={hasCopiedLink ? <Check size={14} /> : <Share2 size={14} />}
                data-testid="copy-referral-link-btn"
              >
                {hasCopiedLink ? 'LINK COPIED' : 'COPY INVITATION LINK'}
              </Button>
            </div>
          </div>
        </div>
      </div>

      {/* 4 Summary Metric Cards */}
      <div className="referral-stats-grid" data-testid="referral-stats-grid">
        <div className="referral-stat-card">
          <span className="referral-stat-label">TOTAL REFERRALS</span>
          <strong className="referral-stat-value" data-testid="stat-total-referrals">
            {String(summary.total_referrals ?? 0).padStart(2, '0')}
          </strong>
        </div>

        <div className="referral-stat-card">
          <span className="referral-stat-label">PENDING FIRST ORDER</span>
          <strong className="referral-stat-value" data-testid="stat-pending-referrals">
            {String(summary.pending_referrals ?? 0).padStart(2, '0')}
          </strong>
        </div>

        <div className="referral-stat-card">
          <span className="referral-stat-label">DELIVERED & QUALIFIED</span>
          <strong className="referral-stat-value" data-testid="stat-completed-referrals">
            {String(summary.completed_referrals ?? 0).padStart(2, '0')}
          </strong>
        </div>

        <div className="referral-stat-card is-reward">
          <span className="referral-stat-label">TOTAL REWARDS EARNED</span>
          <strong className="referral-stat-value" data-testid="stat-total-rewards">
            {formatCurrencyINR(summary.total_rewards_earned ?? 0)}
          </strong>
        </div>
      </div>

      {/* How It Works Guide */}
      <section className="account-card referral-how-card">
        <div className="account-card__header">
          <div className="account-card__title-group">
            <Sparkles size={18} className="account-card__icon" />
            <h2 className="account-card__title">How The Program Works</h2>
          </div>
        </div>

        <div className="account-card__body">
          <div className="referral-steps-grid">
            <div className="referral-step">
              <span className="referral-step-num">01</span>
              <strong className="referral-step-title">Share Your Code</strong>
              <p className="referral-step-desc">
                Send your unique referral code or invitation link to your network.
              </p>
            </div>

            <div className="referral-step">
              <span className="referral-step-num">02</span>
              <strong className="referral-step-title">Friend Registers & Orders</strong>
              <p className="referral-step-desc">
                Your friend registers with your code and places their first atelier commission.
              </p>
            </div>

            <div className="referral-step">
              <span className="referral-step-num">03</span>
              <strong className="referral-step-title">Order Delivered & Verified</strong>
              <p className="referral-step-desc">
                Once their garment is delivered and the standard return window elapses, the referral qualifies.
              </p>
            </div>

            <div className="referral-step">
              <span className="referral-step-num">04</span>
              <strong className="referral-step-title">Wallet Disbursal</strong>
              <p className="referral-step-desc">
                Referral reward points credit automatically to your Atelier Wallet for instant commissioning.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Referral Activity Ledger */}
      <section className="account-card" aria-labelledby="referral-history-heading">
        <div className="account-card__header">
          <div className="account-card__title-group">
            <Users size={18} className="account-card__icon" aria-hidden="true" />
            <h2 id="referral-history-heading" className="account-card__title">
              Referral Activity Ledger
            </h2>
          </div>
        </div>

        <div className="account-card__body">
          {isLoading ? (
            <SkeletonLoader lines={5} />
          ) : referrals.length > 0 ? (
            <div className="referral-ledger-table" data-testid="referral-ledger-table">
              {referrals.map((item) => {
                const status = (item.status || 'pending').toLowerCase();
                const rewardStatus = (item.reward_status || 'pending').toLowerCase();

                return (
                  <div key={item.id} className="referral-ledger-row" data-testid={`referral-item-${item.id}`}>
                    <div className="referral-ledger-identity">
                      <strong className="referral-ledger-name">
                        {item.name || 'Patron Friend'}
                      </strong>
                      <span className="referral-ledger-date">
                        Joined on {formatDate(item.created_at)}
                      </span>
                    </div>

                    <div className="referral-ledger-status-col">
                      <span className="referral-status-label">Order:</span>
                      <StatusPill
                        status={status}
                        variant={status === 'completed' ? 'success' : 'warning'}
                      />
                    </div>

                    <div className="referral-ledger-reward-col">
                      <span className="referral-status-label">Reward:</span>
                      <StatusPill
                        status={rewardStatus}
                        variant={rewardStatus === 'completed' ? 'success' : 'warning'}
                      />
                    </div>

                    <div className="referral-ledger-amount-col">
                      <strong className="referral-ledger-amount">
                        {item.reward_earned > 0
                          ? `+ ${formatCurrencyINR(item.reward_earned)}`
                          : 'Pending Delivery'}
                      </strong>
                    </div>
                  </div>
                );
              })}
            </div>
          ) : (
            <div className="account-empty-block" data-testid="referrals-empty-state">
              <Users size={36} className="account-empty-icon" />
              <h2 className="account-empty-title">No Referrals Yet</h2>
              <p className="account-empty-text">
                Share your invitation link with friends to begin earning rewards on delivered commissions.
              </p>
              <Button
                variant="outline"
                size="sm"
                onClick={handleCopyLink}
                leftIcon={<Share2 size={14} />}
              >
                COPY INVITATION LINK
              </Button>
            </div>
          )}
        </div>
      </section>
    </div>
  );
}
