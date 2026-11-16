import { useEffect, useState, useCallback } from 'react';
import { Link } from 'react-router-dom';
import {
  Search,
  Filter,
  DollarSign,
  RotateCcw,
  CheckCircle,
  Clock,
  AlertCircle,
  RefreshCw,
  ExternalLink,
} from 'lucide-react';
import {
  listAdminOrders,
  listRefundPendingOrders,
  updateCODPaymentStatus,
  processWalletRefund,
} from '../../../api/adminOperationsService';
import { Badge } from '../../../components/common/Badge';
import { Button } from '../../../components/common/Button';
import { DataTable } from '../../../components/common/DataTable';
import { Modal } from '../../../components/common/Modal';
import { useToast } from '../../../context/ToastContext';
import { formatCurrency, formatDate } from '../../../utils/formatters';
import './AdminPaymentsPage.css';

export function AdminPaymentsPage() {
  const { showToast } = useToast();

  const [activeTab, setActiveTab] = useState('transactions'); // 'transactions' | 'refund_pending'
  const [records, setRecords] = useState([]);
  const [isLoading, setIsLoading] = useState(true);

  // Pagination & Filtering
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [totalPages, setTotalPages] = useState(1);
  const [totalItems, setTotalItems] = useState(0);

  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [paymentStatusFilter, setPaymentStatusFilter] = useState('');

  // Action states
  const [selectedOrderForRefund, setSelectedOrderForRefund] = useState(null);
  const [isRefunding, setIsRefunding] = useState(false);
  const [updatingCodId, setUpdatingCodId] = useState(null);

  const fetchPayments = useCallback(async () => {
    setIsLoading(true);
    try {
      if (activeTab === 'transactions') {
        const data = await listAdminOrders({
          page,
          limit,
          search,
          payment_status: paymentStatusFilter || undefined,
        });
        setRecords(data?.orders ?? []);
        setTotalItems(data?.total_count ?? 0);
        setTotalPages(data?.total_pages ?? (Math.ceil((data?.total_count ?? 0) / limit) || 1));
      } else {
        const data = await listRefundPendingOrders({
          page,
          limit,
          search,
        });
        setRecords(data?.orders ?? []);
        setTotalItems(data?.total_count ?? 0);
        setTotalPages(data?.total_pages ?? (Math.ceil((data?.total_count ?? 0) / limit) || 1));
      }
    } catch (err) {
      console.error('Failed to load payment records:', err);
      showToast(err?.response?.data?.message || 'Failed to load payments', 'error');
    } finally {
      setIsLoading(false);
    }
  }, [activeTab, page, limit, search, paymentStatusFilter, showToast]);

  useEffect(() => {
    fetchPayments();
  }, [fetchPayments]);

  const handleSearchSubmit = (e) => {
    e.preventDefault();
    setPage(1);
    setSearch(searchInput.trim());
  };

  const handleClearFilters = () => {
    setSearch('');
    setSearchInput('');
    setPaymentStatusFilter('');
    setPage(1);
  };

  const handleMarkCodPaid = async (orderId) => {
    setUpdatingCodId(orderId);
    try {
      await updateCODPaymentStatus(orderId, 'paid');
      showToast('COD Payment marked as Paid successfully', 'success');
      await fetchPayments();
    } catch (err) {
      console.error('Failed to update COD payment:', err);
      showToast(err?.response?.data?.message || 'Failed to update payment status', 'error');
    } finally {
      setUpdatingCodId(null);
    }
  };

  const handleConfirmRefund = async () => {
    if (!selectedOrderForRefund) return;
    setIsRefunding(true);
    try {
      await processWalletRefund(selectedOrderForRefund.id);
      showToast('Refund processed successfully to customer wallet', 'success');
      setSelectedOrderForRefund(null);
      await fetchPayments();
    } catch (err) {
      console.error('Failed to process wallet refund:', err);
      showToast(err?.response?.data?.message || 'Failed to process wallet refund', 'error');
    } finally {
      setIsRefunding(false);
    }
  };

  const getPaymentBadge = (status) => {
    const s = String(status || '').toLowerCase();
    switch (s) {
      case 'paid':
        return <Badge variant="success">Paid</Badge>;
      case 'pending':
        return <Badge variant="warning">Pending</Badge>;
      case 'refunded':
        return <Badge variant="info">Refunded</Badge>;
      case 'failed':
        return <Badge variant="danger">Failed</Badge>;
      default:
        return <Badge variant="neutral">{status || '—'}</Badge>;
    }
  };

  const columns = [
    {
      key: 'order_number',
      header: 'Order #',
      render: (order) => (
        <div className="drape-order-cell">
          <Link
            to={`/admin/orders/${order.id}`}
            className="drape-order-id-link"
          >
            {order.order_number || `#${order.id}`}
          </Link>
          <span className="drape-order-date">{formatDate(order.created_at)}</span>
        </div>
      ),
    },
    {
      key: 'customer',
      header: 'Customer',
      render: (order) => (
        <div className="drape-customer-cell">
          <span className="drape-customer-name">
            {order.user?.name || order.shipping_address?.name || 'Customer'}
          </span>
          <span className="drape-customer-email">
            {order.user?.email || order.shipping_address?.phone || '—'}
          </span>
        </div>
      ),
    },
    {
      key: 'method',
      header: 'Method',
      render: (order) => (
        <span className="font-semibold text-uppercase">
          {order.payment_method || 'Online'}
        </span>
      ),
    },
    {
      key: 'amount',
      header: 'Amount',
      align: 'right',
      render: (order) => (
        <span className="drape-order-total">{formatCurrency(order.total_amount)}</span>
      ),
    },
    {
      key: 'payment_status',
      header: 'Payment Status',
      align: 'center',
      render: (order) => getPaymentBadge(order.payment_status),
    },
    {
      key: 'order_status',
      header: 'Order Status',
      align: 'center',
      render: (order) => <Badge variant="neutral">{order.order_status}</Badge>,
    },
    {
      key: 'actions',
      header: 'Actions',
      align: 'right',
      render: (order) => {
        const isCodPending =
          String(order.payment_method || '').toLowerCase() === 'cod' &&
          order.payment_status === 'pending';
        const isEligibleForRefund =
          (order.order_status === 'cancelled' || order.order_status === 'returned') &&
          order.payment_status !== 'refunded';

        return (
          <div className="drape-table-actions">
            {isCodPending && (
              <Button
                variant="secondary"
                size="sm"
                icon={<DollarSign size={13} />}
                onClick={() => handleMarkCodPaid(order.id)}
                disabled={updatingCodId === order.id}
              >
                {updatingCodId === order.id ? 'Saving...' : 'Mark Paid'}
              </Button>
            )}

            {isEligibleForRefund && (
              <Button
                variant="outline"
                size="sm"
                icon={<RotateCcw size={13} />}
                onClick={() => setSelectedOrderForRefund(order)}
              >
                Refund
              </Button>
            )}

            <Link to={`/admin/orders/${order.id}`}>
              <Button variant="ghost" size="sm" icon={<ExternalLink size={13} />}>
                Order
              </Button>
            </Link>
          </div>
        );
      },
    },
  ];

  return (
    <div className="admin-payments-page">
      {/* Header */}
      <div className="admin-page-header">
        <div>
          <span className="admin-page-eyebrow">Finance & Settlements</span>
          <h1 className="admin-page-title">Payment Management</h1>
          <p className="admin-page-description">
            Audit customer transactions, settle Cash on Delivery, and process wallet refunds.
          </p>
        </div>
        <div className="admin-header-actions">
          <Button
            variant="outline"
            size="sm"
            icon={<RefreshCw size={14} className={isLoading ? 'spin-icon' : ''} />}
            onClick={fetchPayments}
            disabled={isLoading}
          >
            Refresh
          </Button>
        </div>
      </div>

      {/* Navigation Tabs */}
      <div className="admin-tabs">
        <button
          type="button"
          className={`admin-tab-btn ${activeTab === 'transactions' ? 'active' : ''}`}
          onClick={() => {
            setActiveTab('transactions');
            setPage(1);
          }}
        >
          All Transactions
        </button>
        <button
          type="button"
          className={`admin-tab-btn ${activeTab === 'refund_pending' ? 'active' : ''}`}
          onClick={() => {
            setActiveTab('refund_pending');
            setPage(1);
          }}
        >
          Refund Pending Orders
        </button>
      </div>

      {/* Filter Toolbar */}
      <div className="admin-toolbar">
        <form onSubmit={handleSearchSubmit} className="admin-search-form">
          <div className="admin-search-input-wrapper">
            <Search size={16} className="admin-search-icon" />
            <input
              type="text"
              placeholder="Search by order #, customer, email..."
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              className="admin-search-input"
            />
          </div>
          <Button type="submit" variant="secondary" size="sm">
            Search
          </Button>
        </form>

        {activeTab === 'transactions' && (
          <div className="admin-filter-group">
            <select
              value={paymentStatusFilter}
              onChange={(e) => {
                setPaymentStatusFilter(e.target.value);
                setPage(1);
              }}
              className="admin-select"
            >
              <option value="">All Payment Statuses</option>
              <option value="paid">Paid</option>
              <option value="pending">Pending</option>
              <option value="failed">Failed</option>
              <option value="refunded">Refunded</option>
            </select>

            {(search || paymentStatusFilter) && (
              <Button variant="ghost" size="sm" onClick={handleClearFilters}>
                Clear Filters
              </Button>
            )}
          </div>
        )}
      </div>

      {/* Table */}
      <DataTable
        columns={columns}
        data={records}
        keyField="id"
        isLoading={isLoading}
        emptyMessage={
          activeTab === 'refund_pending'
            ? 'No orders are currently waiting for refund.'
            : 'No transaction records found matching your filters.'
        }
        pagination={{
          page,
          totalPages,
          totalItems,
          limit,
          onPageChange: (newPage) => setPage(newPage),
        }}
      />

      {/* Refund Confirmation Modal */}
      <Modal
        isOpen={Boolean(selectedOrderForRefund)}
        onClose={() => setSelectedOrderForRefund(null)}
        title="Process Wallet Refund"
      >
        {selectedOrderForRefund && (
          <div className="admin-modal-content">
            <p>
              Issue a refund for order{' '}
              <strong>{selectedOrderForRefund.order_number || `#${selectedOrderForRefund.id}`}</strong>?
            </p>
            <p className="text-muted">
              Amount to refund:{' '}
              <strong>{formatCurrency(selectedOrderForRefund.total_amount)}</strong>
            </p>
            <p className="admin-refund-notice">
              The refund amount will be directly deposited into the customer&apos;s DRAPE
              Wallet balance.
            </p>
            <div className="admin-modal-actions">
              <Button
                type="button"
                variant="outline"
                onClick={() => setSelectedOrderForRefund(null)}
              >
                Cancel
              </Button>
              <Button
                type="button"
                variant="primary"
                onClick={handleConfirmRefund}
                disabled={isRefunding}
              >
                {isRefunding ? 'Processing...' : 'Confirm Wallet Refund'}
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
}
