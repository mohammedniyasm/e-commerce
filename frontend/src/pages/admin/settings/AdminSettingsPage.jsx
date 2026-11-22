import { useEffect, useState, useCallback } from 'react';
import {
  Settings,
  Gift,
  Shield,
  CreditCard,
  Truck,
  CheckCircle,
  Save,
  RefreshCw,
  Server,
  DollarSign,
} from 'lucide-react';
import {
  getReferralSettings,
  updateReferralSettings,
} from '../../../api/adminOperationsService';
import { getAdminProfile } from '../../../api/adminService';
import { Badge } from '../../../components/common/Badge';
import { Button } from '../../../components/common/Button';
import { SkeletonLoader } from '../../../components/common/SkeletonLoader';
import { useToast } from '../../../context/ToastContext';
import { formatCurrency } from '../../../utils/formatters';
import './AdminSettingsPage.css';

export function AdminSettingsPage() {
  const { showToast } = useToast();

  const [adminProfile, setAdminProfile] = useState(null);
  const [isLoading, setIsLoading] = useState(true);

  // Referral Settings Form
  const [referrerBonus, setReferrerBonus] = useState('');
  const [refereeBonus, setRefereeBonus] = useState('');
  const [minOrderToReward, setMinOrderToReward] = useState('');
  const [isReferralActive, setIsReferralActive] = useState(true);
  const [isSavingReferral, setIsSavingReferral] = useState(false);

  const fetchSettings = useCallback(async () => {
    setIsLoading(true);
    try {
      const [profileRes, referralRes] = await Promise.allSettled([
        getAdminProfile(),
        getReferralSettings(),
      ]);

      if (profileRes.status === 'fulfilled') {
        setAdminProfile(profileRes.value);
      }

      if (referralRes.status === 'fulfilled' && referralRes.value) {
        const ref = referralRes.value;
        setReferrerBonus(String(ref.referrer_reward ?? 500));
        setRefereeBonus(String(ref.referred_reward ?? 250));
        setMinOrderToReward(String(ref.minimum_order_amount ?? 2000));
        setIsReferralActive(Boolean(ref.is_active ?? true));
      }
    } catch (err) {
      console.error('Failed to load settings:', err);
      showToast('Failed to load settings', 'error');
    } finally {
      setIsLoading(false);
    }
  }, [showToast]);

  useEffect(() => {
    fetchSettings();
  }, [fetchSettings]);

  const handleSaveReferrals = async (e) => {
    e.preventDefault();
    setIsSavingReferral(true);
    try {
      const payload = {
        referrer_reward: parseFloat(referrerBonus),
        referred_reward: parseFloat(refereeBonus),
        minimum_order_amount: parseFloat(minOrderToReward),
        is_active: isReferralActive,
      };

      await updateReferralSettings(payload);
      showToast('Referral program settings updated successfully', 'success');
    } catch (err) {
      console.error('Failed to update referral settings:', err);
      showToast(err?.response?.data?.message || 'Failed to update referral settings', 'error');
    } finally {
      setIsSavingReferral(false);
    }
  };

  if (isLoading) {
    return (
      <div className="admin-settings-page">
        <SkeletonLoader height={40} width={220} />
        <SkeletonLoader height={240} />
        <SkeletonLoader height={300} />
      </div>
    );
  }

  return (
    <div className="admin-settings-page">
      {/* Header */}
      <div className="admin-page-header">
        <div>
          <span className="admin-page-eyebrow">System & Configuration</span>
          <h1 className="admin-page-title">Store Settings</h1>
          <p className="admin-page-description">
            Configure global store parameters, reward referral programs, and view gateway statuses.
          </p>
        </div>
        <div className="admin-header-actions">
          <Button
            variant="outline"
            size="sm"
            leftIcon={<RefreshCw size={14} />}
            onClick={fetchSettings}
          >
            Refresh
          </Button>
        </div>
      </div>

      <div className="admin-settings-grid">
        {/* Referral Program Configuration */}
        <div className="admin-settings-card">
          <div className="admin-settings-card-header">
            <div className="admin-settings-icon-title">
              <Gift size={20} className="text-gold" />
              <div>
                <h3>Customer Referral Rewards</h3>
                <p>Incentivize customer growth via automatic DRAPE Wallet reward credits.</p>
              </div>
            </div>
            <Badge variant={isReferralActive ? 'success' : 'neutral'}>
              {isReferralActive ? 'Active' : 'Disabled'}
            </Badge>
          </div>

          <form onSubmit={handleSaveReferrals} className="admin-settings-form">
            <div className="admin-form-row">
              <label className="admin-form-label">
                Referrer Reward (₹)
                <input
                  type="number"
                  min="0"
                  value={referrerBonus}
                  onChange={(e) => setReferrerBonus(e.target.value)}
                  className="admin-input-full"
                  required
                />
                <span className="admin-field-hint">
                  Credited to the advocate when their friend completes their first qualifying order.
                </span>
              </label>

              <label className="admin-form-label">
                Referee Welcome Reward (₹)
                <input
                  type="number"
                  min="0"
                  value={refereeBonus}
                  onChange={(e) => setRefereeBonus(e.target.value)}
                  className="admin-input-full"
                  required
                />
                <span className="admin-field-hint">
                  Credited to the newly registered customer upon signing up with referral code.
                </span>
              </label>
            </div>

            <div className="admin-form-row">
              <label className="admin-form-label">
                Minimum Qualifying Order (₹)
                <input
                  type="number"
                  min="0"
                  value={minOrderToReward}
                  onChange={(e) => setMinOrderToReward(e.target.value)}
                  className="admin-input-full"
                  required
                />
                <span className="admin-field-hint">
                  Referee&apos;s first order must meet this subtotal before advocate reward is released.
                </span>
              </label>

              <label className="admin-form-label">
                Program Activation
                <div className="admin-checkbox-row">
                  <input
                    type="checkbox"
                    id="referral-active"
                    checked={isReferralActive}
                    onChange={(e) => setIsReferralActive(e.target.checked)}
                  />
                  <label htmlFor="referral-active">Enable viral referral promotions</label>
                </div>
              </label>
            </div>

            <div className="admin-settings-submit-row">
              <Button
                type="submit"
                variant="primary"
                size="sm"
                leftIcon={<Save size={14} />}
                isLoading={isSavingReferral}
              >
                {isSavingReferral ? 'Saving...' : 'Save Referral Configuration'}
              </Button>
            </div>
          </form>
        </div>

        {/* Store & Payment Overview */}
        <div className="admin-settings-card">
          <div className="admin-settings-card-header">
            <div className="admin-settings-icon-title">
              <CreditCard size={20} className="text-gold" />
              <div>
                <h3>Payment Gateways & Settlement</h3>
                <p>Enabled payment channels and transaction providers.</p>
              </div>
            </div>
          </div>

          <div className="admin-gateways-list">
            <div className="admin-gateway-item">
              <div className="admin-gateway-info">
                <span className="font-semibold">Razorpay Direct Gateway</span>
                <span className="admin-field-hint">UPI, Credit/Debit Cards, NetBanking</span>
              </div>
              <Badge variant="success">Production Ready</Badge>
            </div>

            <div className="admin-gateway-item">
              <div className="admin-gateway-info">
                <span className="font-semibold">DRAPE Customer Wallet</span>
                <span className="admin-field-hint">Instant 1-click checkout and automated refunds</span>
              </div>
              <Badge variant="success">Active</Badge>
            </div>

            <div className="admin-gateway-item">
              <div className="admin-gateway-info">
                <span className="font-semibold">Cash On Delivery (COD)</span>
                <span className="admin-field-hint">Settled manually upon carrier delivery confirmation</span>
              </div>
              <Badge variant="success">Active</Badge>
            </div>
          </div>
        </div>

        {/* Administrator Profile & Environment */}
        <div className="admin-settings-card">
          <div className="admin-settings-card-header">
            <div className="admin-settings-icon-title">
              <Shield size={20} className="text-gold" />
              <div>
                <h3>Admin Session & Role</h3>
                <p>Current authenticated administrative identity.</p>
              </div>
            </div>
          </div>

          <div className="admin-admin-profile-grid">
            <div>
              <span className="admin-field-label">Administrator Name</span>
              <span className="admin-field-value font-semibold">
                {adminProfile?.name || 'Super Administrator'}
              </span>
            </div>
            <div>
              <span className="admin-field-label">Email Account</span>
              <span className="admin-field-value font-mono">
                {adminProfile?.email || 'admin@example.com'}
              </span>
            </div>
            <div>
              <span className="admin-field-label">Assigned Role</span>
              <Badge variant="info">{adminProfile?.role || 'admin'}</Badge>
            </div>
            <div>
              <span className="admin-field-label">Environment</span>
              <span className="admin-field-value">PostgreSQL / Go Clean Architecture</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
