import { useEffect, useState, useCallback } from 'react';
import { Link } from 'react-router-dom';
import {
  Search,
  Filter,
  Eye,
  ShoppingBag,
  Clock,
  CheckCircle,
  Truck,
  XCircle,
  RotateCcw,
  RefreshCw,
} from 'lucide-react';
import { listAdminOrders } from '../../../api/adminOperationsService';
import { Badge } from '../../../components/common/Badge';
import { Button } from '../../../components/common/Button';
import { DataTable } from '../../../components/common/DataTable';
import { useToast } from '../../../context/ToastContext';
import { formatCurrency, formatDate } from '../../../utils/formatters';
import './AdminOrdersPage.css';

const ORDER_STATUS_OPTIONS = [
  { value: '', label: 'All Order Statuses' },
  { value: 'pending', label: 'Pending' },
  { value: 'confirmed', label: 'Confirmed' },
  { value: 'processing', label: 'Processing' },
  { value: 'shipped', label: 'Shipped' },
  { value: 'out_for_delivery', label: 'Out for Delivery' },
  { value: 'delivered', label: 'Delivered' },
  { value: 'cancelled', label: 'Cancelled' },
  { value: 'returned', label: 'Returned' },
];

const PAYMENT_STATUS_OPTIONS = [
  { value: '', label: 'All Payment Statuses' },
  { value: 'pending', label: 'Payment Pending' },
  { value: 'paid', label: 'Paid' },
  { value: 'failed', label: 'Payment Failed' },
  { value: 'refunded', label: 'Refunded' },
];

export function AdminOrdersPage() {
  const { showToast } = useToast();

  const [orders, setOrders] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [totalPages, setTotalPages] = useState(1);
  const [totalItems, setTotalItems] = useState(0);

  // Filters
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [selectedOrderStatus, setSelectedOrderStatus] = useState('');
  const [selectedPaymentStatus, setSelectedPaymentStatus] = useState('');

  const fetchOrders = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await listAdminOrders({
        page,
        limit,
        search,
        order_status: selectedOrderStatus || undefined,
        payment_status: selectedPaymentStatus || undefined,
      });

      setOrders(data?.orders ?? []);
      setTotalItems(data?.total_count ?? 0);
      setTotalPages(data?.total_pages ?? (Math.ceil((data?.total_count ?? 0) / limit) || 1));
    } catch (err) {
      console.error('Failed to load admin orders:', err);
      showToast(err?.response?.data?.message || 'Failed to load orders', 'error');
    } finally {
      setIsLoading(false);
    }
  }, [page, limit, search, selectedOrderStatus, selectedPaymentStatus, showToast]);

  useEffect(() => {
    fetchOrders();
  }, [fetchOrders]);

  const handleSearchSubmit = (e) => {
    e.preventDefault();
    setPage(1);
    setSearch(searchInput.trim());
  };

  const handleClearFilters = () => {
    setSearch('');
    setSearchInput('');
    setSelectedOrderStatus('');
    setSelectedPaymentStatus('');
    setPage(1);
  };

  const getOrderStatusBadge = (status) => {
    const s = String(status || '').toLowerCase();
    switch (s) {
      case 'confirmed':
      case 'processing':
        return <Badge variant="info">{status}</Badge>;
      case 'shipped':
      case 'out_for_delivery':
        return <Badge variant="warning">{status}</Badge>;
      case 'delivered':
        return <Badge variant="success">{status}</Badge>;
      case 'cancelled':
        return <Badge variant="danger">{status}</Badge>;
      case 'returned':
        return <Badge variant="neutral">{status}</Badge>;
      case 'pending':
      default:
        return <Badge variant="neutral">{status || 'pending'}</Badge>;
    }
  };

  const getPaymentStatusBadge = (status) => {
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
      render: (val, order) => {
        const item = (typeof val === 'object' && val !== null) ? val : (order || {});
        return (
          <div className="drape-order-cell">
            <Link
              to={`/admin/orders/${item.id}`}
              className="drape-order-id-link"
              title="View order details"
            >
              {item.order_number || `#${item.id}`}
            </Link>
            <span className="drape-order-date">{formatDate(item.created_at)}</span>
          </div>
        );
      },
    },
    {
      key: 'customer',
      header: 'Customer',
      render: (val, order) => {
        const item = (typeof val === 'object' && val !== null) ? val : (order || {});
        return (
          <div className="drape-customer-cell">
            <span className="drape-customer-name">
              {item.user?.name || item.shipping_address?.name || 'Customer'}
            </span>
            <span className="drape-customer-email">
              {item.user?.email || item.shipping_address?.phone || '—'}
            </span>
          </div>
        );
      },
    },
    {
      key: 'items_count',
      header: 'Items',
      align: 'center',
      render: (val, order) => {
        const item = (typeof val === 'object' && val !== null) ? val : (order || {});
        const count = item.items?.length || item.items_count || 1;
        return (
          <span className="drape-items-count">
            {count} {count === 1 ? 'item' : 'items'}
          </span>
        );
      },
    },
    {
      key: 'total_amount',
      header: 'Total',
      align: 'right',
      render: (amount, order) => {
        const item = (typeof amount === 'object' && amount !== null) ? amount : (order || {});
        const total = typeof amount === 'number' ? amount : (item.total_amount || 0);
        return (
          <span className="drape-order-total">{formatCurrency(total)}</span>
        );
      },
    },
    {
      key: 'payment_status',
      header: 'Payment',
      align: 'center',
      render: (status, order) => {
        const item = (typeof status === 'object' && status !== null) ? status : (order || {});
        const val = typeof status === 'string' ? status : item.payment_status;
        return getPaymentStatusBadge(val);
      },
    },
    {
      key: 'order_status',
      header: 'Order Status',
      align: 'center',
      render: (status, order) => {
        const item = (typeof status === 'object' && status !== null) ? status : (order || {});
        const val = typeof status === 'string' ? status : item.order_status;
        return getOrderStatusBadge(val);
      },
    },
    {
      key: 'actions',
      header: 'Actions',
      align: 'right',
      render: (val, order) => {
        const item = (typeof val === 'object' && val !== null) ? val : (order || {});
        return (
          <div className="drape-table-actions">
            <Link to={`/admin/orders/${item.id}`}>
              <Button variant="ghost" size="sm" icon={<Eye size={14} />}>
                Details
              </Button>
            </Link>
          </div>
        );
      },
    },
  ];

  return (
    <div className="admin-orders-page">
      {/* Header */}
      <div className="admin-page-header">
        <div>
          <span className="admin-page-eyebrow">Fulfillment & Operations</span>
          <h1 className="admin-page-title">Orders Management</h1>
          <p className="admin-page-description">
            Monitor, fulfill, and track customer orders across their entire lifecycle.
          </p>
        </div>
        <div className="admin-header-actions">
          <Button
            variant="outline"
            size="sm"
            icon={<RefreshCw size={14} className={isLoading ? 'spin-icon' : ''} />}
            onClick={fetchOrders}
            disabled={isLoading}
          >
            Refresh
          </Button>
        </div>
      </div>

      {/* Filter Toolbar */}
      <div className="admin-toolbar">
        <form onSubmit={handleSearchSubmit} className="admin-search-form">
          <div className="admin-search-input-wrapper">
            <Search size={16} className="admin-search-icon" />
            <input
              type="text"
              placeholder="Search by order #, customer name, email..."
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              className="admin-search-input"
            />
          </div>
          <Button type="submit" variant="secondary" size="sm">
            Search
          </Button>
        </form>

        <div className="admin-filter-group">
          <select
            value={selectedOrderStatus}
            onChange={(e) => {
              setSelectedOrderStatus(e.target.value);
              setPage(1);
            }}
            className="admin-select"
          >
            {ORDER_STATUS_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>

          <select
            value={selectedPaymentStatus}
            onChange={(e) => {
              setSelectedPaymentStatus(e.target.value);
              setPage(1);
            }}
            className="admin-select"
          >
            {PAYMENT_STATUS_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>

          {(search || selectedOrderStatus || selectedPaymentStatus) && (
            <Button variant="ghost" size="sm" onClick={handleClearFilters}>
              Clear Filters
            </Button>
          )}
        </div>
      </div>

      {/* Orders Table */}
      <DataTable
        columns={columns}
        data={orders}
        keyField="id"
        isLoading={isLoading}
        emptyMessage="No customer orders found matching your criteria."
        pagination={{
          page,
          totalPages,
          totalItems,
          limit,
          onPageChange: (newPage) => setPage(newPage),
        }}
      />
    </div>
  );
}
