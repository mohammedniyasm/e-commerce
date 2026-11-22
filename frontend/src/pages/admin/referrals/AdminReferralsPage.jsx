import { useCallback, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  CalendarDays,
  Clock3,
  Gift,
  RefreshCw,
  Settings2,
  TrendingUp,
  Users,
  Wallet,
} from 'lucide-react';
import {
  getAdminReferralDashboard,
  getTopAdminReferrers,
  listAdminReferrals,
} from '../../../api/adminOperationsService';
import { Badge } from '../../../components/common/Badge';
import { Button } from '../../../components/common/Button';
import { DataTable } from '../../../components/common/DataTable';
import { useToast } from '../../../context/ToastContext';
import { formatCurrency, formatDate } from '../../../utils/formatters';
import './AdminReferralsPage.css';

const PAGE_SIZE = 10;

function statusVariant(status) {
  if (status === 'completed') return 'success';
  if (status === 'pending') return 'warning';
  return 'neutral';
}

export function AdminReferralsPage() {
  const { showToast } = useToast();
  const [dashboard, setDashboard] = useState(null);
  const [topReferrers, setTopReferrers] = useState([]);
  const [referrals, setReferrals] = useState([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [isLoadingOverview, setIsLoadingOverview] = useState(true);
  const [isLoadingReferrals, setIsLoadingReferrals] = useState(true);
  const [filters, setFilters] = useState({
    status: '',
    reward_status: '',
    referrer_id: '',
    start_date: '',
    end_date: '',
  });
  const [draftFilters, setDraftFilters] = useState(filters);

  const fetchOverview = useCallback(async () => {
    setIsLoadingOverview(true);
    try {
      const [summary, leaders] = await Promise.all([
        getAdminReferralDashboard(),
        getTopAdminReferrers(5),
      ]);
      setDashboard(summary);
      setTopReferrers(leaders);
    } catch (err) {
      console.error('Failed to load referral overview:', err);
      showToast(err?.response?.data?.message || 'Failed to load referral overview', 'error');
    } finally {
      setIsLoadingOverview(false);
    }
  }, [showToast]);

  const fetchReferrals = useCallback(async () => {
    setIsLoadingReferrals(true);
    try {
      const data = await listAdminReferrals({
        ...filters,
        page,
        limit: PAGE_SIZE,
      });
      setReferrals(data?.referrals ?? []);
      setTotal(data?.pagination?.total ?? 0);
    } catch (err) {
      console.error('Failed to load referrals:', err);
      showToast(err?.response?.data?.message || 'Failed to load referrals', 'error');
    } finally {
      setIsLoadingReferrals(false);
    }
  }, [filters, page, showToast]);

  useEffect(() => {
    fetchOverview();
  }, [fetchOverview]);

  useEffect(() => {
    fetchReferrals();
  }, [fetchReferrals]);

  const handleApplyFilters = (event) => {
    event.preventDefault();
    if (draftFilters.referrer_id && !/^[1-9]\d*$/.test(draftFilters.referrer_id)) {
      showToast('Referrer ID must be a positive number', 'error');
      return;
    }
    if (draftFilters.start_date && draftFilters.end_date &&
      draftFilters.start_date > draftFilters.end_date) {
      showToast('Start date must be on or before end date', 'error');
      return;
    }
    setPage(1);
    setFilters(draftFilters);
  };

  const handleClearFilters = () => {
    const emptyFilters = {
      status: '',
      reward_status: '',
      referrer_id: '',
      start_date: '',
      end_date: '',
    };
    setDraftFilters(emptyFilters);
    setFilters(emptyFilters);
    setPage(1);
  };

  const refresh = () => {
    fetchOverview();
    fetchReferrals();
  };

  const columns = [
    {
      key: 'referrer_name',
      header: 'Referral Participants',
      render: (value, referral) => (
        <div className="admin-referral-participants">
          <span><strong>From:</strong> {value || 'Unknown'} <small>#{referral.referrer_id}</small></span>
          <span><strong>To:</strong> {referral.referred_user_name || 'Unknown'} <small>#{referral.referred_user_id}</small></span>
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Referral Status',
      render: (value) => (
        <Badge variant={statusVariant(value)}>{value || 'Unknown'}</Badge>
      ),
    },
    {
      key: 'referrer_reward',
      header: 'Referrer Reward',
      render: (value, referral) => (
        <div className="admin-referral-reward">
          <strong>{formatCurrency(value)}</strong>
          <Badge size="sm" variant={statusVariant(referral.referrer_reward_status)}>
            {referral.referrer_reward_status || 'unknown'}
          </Badge>
        </div>
      ),
    },
    {
      key: 'referred_reward',
      header: 'Referred Reward',
      render: (value, referral) => (
        <div className="admin-referral-reward">
          <strong>{formatCurrency(value)}</strong>
          <Badge size="sm" variant={statusVariant(referral.referred_reward_status)}>
            {referral.referred_reward_status || 'unknown'}
          </Badge>
        </div>
      ),
    },
    { key: 'created_at', header: 'Joined', render: (value) => formatDate(value) },
    { key: 'completed_at', header: 'Completed', render: (value) => formatDate(value) },
  ];

  const metricCards = [
    {
      label: 'Total Referrals',
      value: dashboard?.total_referrals ?? '—',
      icon: Users,
      tone: 'blue',
    },
    {
      label: 'Pending Referrals',
      value: dashboard?.pending_referrals ?? '—',
      icon: Clock3,
      tone: 'amber',
    },
    {
      label: 'Completed Referrals',
      value: dashboard?.completed_referrals ?? '—',
      icon: TrendingUp,
      tone: 'green',
    },
    {
      label: 'Rewards Paid',
      value: dashboard ? formatCurrency(dashboard.total_rewards_paid) : '—',
      icon: Wallet,
      tone: 'violet',
    },
  ];

  return (
    <div className="admin-referrals-page">
      <header className="admin-page-header">
        <div>
          <span className="admin-page-eyebrow">Customer Growth</span>
          <h1 className="admin-page-title">Referral Program</h1>
          <p className="admin-page-description">
            Track referral performance, reward payouts, and customer participation.
          </p>
        </div>
        <div className="admin-header-actions">
          <Link className="admin-referral-settings-link" to="/admin/settings">
            <Settings2 size={16} />
            Program Settings
          </Link>
          <Button
            variant="secondary"
            size="sm"
            leftIcon={<RefreshCw size={14} className={isLoadingOverview || isLoadingReferrals ? 'spin-icon' : ''} />}
            onClick={refresh}
            disabled={isLoadingOverview || isLoadingReferrals}
          >
            Refresh
          </Button>
        </div>
      </header>

      <section className="admin-referral-metrics" aria-label="Referral summary">
        {metricCards.map(({ label, value, icon: Icon, tone }) => (
          <article className="admin-referral-metric-card" key={label}>
            <div className={`admin-referral-metric-icon admin-referral-metric-icon--${tone}`}>
              <Icon size={18} aria-hidden="true" />
            </div>
            <span className="admin-referral-metric-label">{label}</span>
            <strong className="admin-referral-metric-value">
              {isLoadingOverview ? '…' : value}
            </strong>
          </article>
        ))}
      </section>

      <div className="admin-referral-secondary-stats">
        <div>
          <span>Referrer rewards paid</span>
          <strong>
            {isLoadingOverview
              ? '…'
              : dashboard
                ? formatCurrency(dashboard.referrer_rewards_paid)
                : '—'}
          </strong>
        </div>
        <div>
          <span>Referred-user rewards paid</span>
          <strong>
            {isLoadingOverview
              ? '…'
              : dashboard
                ? formatCurrency(dashboard.referred_rewards_paid)
                : '—'}
          </strong>
        </div>
        <div>
          <span>Conversion rate</span>
          <strong>
            {isLoadingOverview
              ? '…'
              : dashboard
                ? `${Number(dashboard.conversion_rate ?? 0).toFixed(1)}%`
                : '—'}
          </strong>
        </div>
      </div>

      <section className="admin-referral-section">
        <div className="admin-referral-section-heading">
          <div>
            <span className="admin-page-eyebrow">Leaderboard</span>
            <h2>Top Referrers</h2>
          </div>
          <Gift size={20} aria-hidden="true" />
        </div>
        <DataTable
          columns={[
            { key: 'name', header: 'Customer' },
            { key: 'successful_referrals', header: 'Successful Referrals', align: 'center' },
            {
              key: 'rewards_earned',
              header: 'Rewards Earned',
              align: 'right',
              render: (value) => <strong>{formatCurrency(value)}</strong>,
            },
          ]}
          data={topReferrers}
          keyField="referrer_id"
          isLoading={isLoadingOverview}
          emptyMessage="No completed referrals yet."
        />
      </section>

      <section className="admin-referral-section">
        <div className="admin-referral-section-heading">
          <div>
            <span className="admin-page-eyebrow">Program Activity</span>
            <h2>All Referrals</h2>
          </div>
          <span className="admin-referral-record-count">{total} records</span>
        </div>

        <form className="admin-referral-filters" onSubmit={handleApplyFilters}>
          <label>
            Referral status
            <select
              value={draftFilters.status}
              onChange={(event) => setDraftFilters({ ...draftFilters, status: event.target.value })}
            >
              <option value="">All statuses</option>
              <option value="pending">Pending</option>
              <option value="completed">Completed</option>
            </select>
          </label>
          <label>
            Reward status
            <select
              value={draftFilters.reward_status}
              onChange={(event) => setDraftFilters({ ...draftFilters, reward_status: event.target.value })}
            >
              <option value="">All rewards</option>
              <option value="pending">Pending</option>
              <option value="completed">Completed</option>
            </select>
          </label>
          <label>
            Referrer ID
            <input
              type="number"
              min="1"
              value={draftFilters.referrer_id}
              onChange={(event) => setDraftFilters({ ...draftFilters, referrer_id: event.target.value })}
              placeholder="Any customer"
            />
          </label>
          <label>
            <span><CalendarDays size={13} /> From</span>
            <input
              type="date"
              value={draftFilters.start_date}
              onChange={(event) => setDraftFilters({ ...draftFilters, start_date: event.target.value })}
            />
          </label>
          <label>
            <span><CalendarDays size={13} /> To</span>
            <input
              type="date"
              value={draftFilters.end_date}
              onChange={(event) => setDraftFilters({ ...draftFilters, end_date: event.target.value })}
            />
          </label>
          <div className="admin-referral-filter-actions">
            <Button type="submit" size="sm" disabled={isLoadingReferrals}>Apply</Button>
            <Button type="button" size="sm" variant="outline" onClick={handleClearFilters}>
              Clear
            </Button>
          </div>
        </form>

        <DataTable
          columns={columns}
          data={referrals}
          keyField="id"
          isLoading={isLoadingReferrals}
          emptyMessage="No referrals match the selected filters."
          pagination={{
            page,
            totalPages: Math.ceil(total / PAGE_SIZE) || 1,
            totalItems: total,
            limit: PAGE_SIZE,
            onPageChange: setPage,
          }}
        />
      </section>
    </div>
  );
}
