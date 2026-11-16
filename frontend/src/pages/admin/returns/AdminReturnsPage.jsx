import { useEffect, useState, useCallback } from 'react';
import { Link } from 'react-router-dom';
import {
  RotateCcw,
  RefreshCw,
  Search,
  Filter,
  CheckCircle,
  XCircle,
  Truck,
  Repeat,
  AlertCircle,
  ExternalLink,
} from 'lucide-react';
import {
  listReturnRequests,
  updateReturnRequestStatus,
  createExchangeOrder,
  createReturnShipment,
} from '../../../api/adminOperationsService';
import { Badge } from '../../../components/common/Badge';
import { Button } from '../../../components/common/Button';
import { DataTable } from '../../../components/common/DataTable';
import { Modal } from '../../../components/common/Modal';
import { useToast } from '../../../context/ToastContext';
import { formatDate } from '../../../utils/formatters';
import './AdminReturnsPage.css';

const STATUS_OPTIONS = [
  { value: '', label: 'All Statuses' },
  { value: 'pending', label: 'Pending Review' },
  { value: 'approved', label: 'Approved' },
  { value: 'rejected', label: 'Rejected' },
  { value: 'completed', label: 'Completed' },
];

export function AdminReturnsPage() {
  const { showToast } = useToast();

  const [requests, setRequests] = useState([]);
  const [isLoading, setIsLoading] = useState(true);

  // Pagination & Filters
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [totalPages, setTotalPages] = useState(1);
  const [totalItems, setTotalItems] = useState(0);
  const [selectedStatus, setSelectedStatus] = useState('');

  // Status Action Modal State
  const [selectedRequest, setSelectedRequest] = useState(null);
  const [actionStatus, setActionStatus] = useState('approved');
  const [adminNote, setAdminNote] = useState('');
  const [isUpdatingStatus, setIsUpdatingStatus] = useState(false);

  // Exchange & Shipment creation states
  const [isProcessingExchange, setIsProcessingExchange] = useState(false);
  const [isCreatingShipment, setIsCreatingShipment] = useState(false);

  const fetchReturns = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await listReturnRequests({
        page,
        limit,
        status: selectedStatus || undefined,
      });

      const items = data?.items ?? [];
      const total = data?.total ?? 0;
      setRequests(items);
      setTotalItems(total);
      setTotalPages(Math.ceil(total / limit) || 1);
    } catch (err) {
      console.error('Failed to load return requests:', err);
      showToast(err?.response?.data?.message || 'Failed to load return requests', 'error');
    } finally {
      setIsLoading(false);
    }
  }, [page, limit, selectedStatus, showToast]);

  useEffect(() => {
    fetchReturns();
  }, [fetchReturns]);

  const handleStatusSubmit = async (e) => {
    e.preventDefault();
    if (!selectedRequest) return;

    setIsUpdatingStatus(true);
    try {
      await updateReturnRequestStatus(selectedRequest.id, {
        status: actionStatus,
        admin_note: adminNote.trim() || undefined,
      });
      showToast(`Request marked as "${actionStatus}"`, 'success');
      setSelectedRequest(null);
      setAdminNote('');
      await fetchReturns();
    } catch (err) {
      console.error('Failed to update return status:', err);
      showToast(err?.response?.data?.message || 'Failed to update status', 'error');
    } finally {
      setIsUpdatingStatus(false);
    }
  };

  const handleCreateExchange = async (request) => {
    setIsProcessingExchange(true);
    try {
      await createExchangeOrder(request.id);
      showToast('Exchange order created successfully', 'success');
      await fetchReturns();
    } catch (err) {
      console.error('Failed to create exchange order:', err);
      showToast(err?.response?.data?.message || 'Failed to create exchange order', 'error');
    } finally {
      setIsProcessingExchange(false);
    }
  };

  const handleCreateShipment = async (request) => {
    setIsCreatingShipment(true);
    try {
      await createReturnShipment(request.id);
      showToast('Return reverse shipment initiated successfully', 'success');
      await fetchReturns();
    } catch (err) {
      console.error('Failed to create return shipment:', err);
      showToast(err?.response?.data?.message || 'Failed to create return shipment', 'error');
    } finally {
      setIsCreatingShipment(false);
    }
  };

  const getStatusBadge = (status) => {
    const s = String(status || '').toLowerCase();
    switch (s) {
      case 'approved':
        return <Badge variant="info">Approved</Badge>;
      case 'completed':
        return <Badge variant="success">Completed</Badge>;
      case 'rejected':
        return <Badge variant="danger">Rejected</Badge>;
      case 'pending':
      default:
        return <Badge variant="warning">Pending</Badge>;
    }
  };

  const columns = [
    {
      key: 'id',
      header: 'Request #',
      render: (req) => (
        <div className="drape-order-cell">
          <span className="font-semibold">REQ-{req.id}</span>
          <span className="drape-order-date">{formatDate(req.created_at)}</span>
        </div>
      ),
    },
    {
      key: 'order',
      header: 'Order #',
      render: (req) => (
        <Link to={`/admin/orders/${req.order_id}`} className="drape-order-id-link">
          {req.order?.order_number || `#${req.order_id}`}
        </Link>
      ),
    },
    {
      key: 'type',
      header: 'Type',
      align: 'center',
      render: (req) => {
        const isExchange = String(req.type || '').toLowerCase() === 'exchange';
        return isExchange ? (
          <Badge variant="neutral" className="drape-type-badge">
            <Repeat size={12} /> Exchange
          </Badge>
        ) : (
          <Badge variant="neutral" className="drape-type-badge">
            <RotateCcw size={12} /> Return
          </Badge>
        );
      },
    },
    {
      key: 'product',
      header: 'Item',
      render: (req) => (
        <div className="drape-customer-cell">
          <span className="font-semibold">{req.order_item?.product_name || req.product_name || `Item #${req.order_item_id}`}</span>
          <span className="drape-customer-email">Qty: {req.quantity || 1}</span>
        </div>
      ),
    },
    {
      key: 'reason',
      header: 'Reason',
      render: (req) => (
        <div className="drape-reason-cell">
          <span className="drape-reason-title">{req.reason || 'Not specified'}</span>
          {req.comments && <span className="drape-reason-desc">{req.comments}</span>}
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      align: 'center',
      render: (req) => getStatusBadge(req.status),
    },
    {
      key: 'actions',
      header: 'Actions',
      align: 'right',
      render: (req) => {
        const isPending = req.status === 'pending';
        const isApproved = req.status === 'approved';
        const isExchange = String(req.type || '').toLowerCase() === 'exchange';

        return (
          <div className="drape-table-actions">
            {isPending && (
              <Button
                variant="primary"
                size="sm"
                onClick={() => {
                  setSelectedRequest(req);
                  setActionStatus('approved');
                  setAdminNote('');
                }}
              >
                Review
              </Button>
            )}

            {isApproved && isExchange && (
              <Button
                variant="secondary"
                size="sm"
                icon={<Repeat size={12} />}
                onClick={() => handleCreateExchange(req)}
                disabled={isProcessingExchange}
              >
                Create Exchange
              </Button>
            )}

            {isApproved && (
              <Button
                variant="outline"
                size="sm"
                icon={<Truck size={12} />}
                onClick={() => handleCreateShipment(req)}
                disabled={isCreatingShipment}
              >
                Pickup
              </Button>
            )}
          </div>
        );
      },
    },
  ];

  return (
    <div className="admin-returns-page">
      {/* Header */}
      <div className="admin-page-header">
        <div>
          <span className="admin-page-eyebrow">Customer Care & Reverse Logistics</span>
          <h1 className="admin-page-title">Returns & Exchanges</h1>
          <p className="admin-page-description">
            Process item returns, approve exchange replacements, and dispatch reverse pickups.
          </p>
        </div>
        <div className="admin-header-actions">
          <Button
            variant="outline"
            size="sm"
            icon={<RefreshCw size={14} className={isLoading ? 'spin-icon' : ''} />}
            onClick={fetchReturns}
            disabled={isLoading}
          >
            Refresh
          </Button>
        </div>
      </div>

      {/* Toolbar */}
      <div className="admin-toolbar">
        <div className="admin-filter-group">
          <select
            value={selectedStatus}
            onChange={(e) => {
              setSelectedStatus(e.target.value);
              setPage(1);
            }}
            className="admin-select"
          >
            {STATUS_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>

          {selectedStatus && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                setSelectedStatus('');
                setPage(1);
              }}
            >
              Clear Filter
            </Button>
          )}
        </div>
      </div>

      {/* Table */}
      <DataTable
        columns={columns}
        data={requests}
        keyField="id"
        isLoading={isLoading}
        emptyMessage="No return or exchange requests found."
        pagination={{
          page,
          totalPages,
          totalItems,
          limit,
          onPageChange: (newPage) => setPage(newPage),
        }}
      />

      {/* Review Modal */}
      <Modal
        isOpen={Boolean(selectedRequest)}
        onClose={() => setSelectedRequest(null)}
        title={`Review Request REQ-${selectedRequest?.id}`}
      >
        {selectedRequest && (
          <form onSubmit={handleStatusSubmit} className="admin-modal-form">
            <div className="admin-modal-info-block">
              <p>
                <strong>Order:</strong> #{selectedRequest.order_id}
              </p>
              <p>
                <strong>Type:</strong>{' '}
                <span className="text-uppercase">{selectedRequest.type}</span>
              </p>
              <p>
                <strong>Customer Reason:</strong> {selectedRequest.reason || 'None provided'}
              </p>
              {selectedRequest.comments && (
                <p>
                  <strong>Customer Comments:</strong> {selectedRequest.comments}
                </p>
              )}
            </div>

            <label className="admin-form-label">
              Decision *
              <select
                value={actionStatus}
                onChange={(e) => setActionStatus(e.target.value)}
                className="admin-select-full"
              >
                <option value="approved">Approve Request</option>
                <option value="rejected">Reject Request</option>
                <option value="completed">Mark as Completed</option>
              </select>
            </label>

            <label className="admin-form-label">
              Admin Notes / Rejection Reason
              <textarea
                value={adminNote}
                onChange={(e) => setAdminNote(e.target.value)}
                placeholder="Reason or notes communicated to customer..."
                className="admin-input-full"
                rows={3}
              />
            </label>

            <div className="admin-modal-actions">
              <Button
                type="button"
                variant="outline"
                onClick={() => setSelectedRequest(null)}
              >
                Cancel
              </Button>
              <Button type="submit" variant="primary" disabled={isUpdatingStatus}>
                {isUpdatingStatus ? 'Saving...' : 'Submit Decision'}
              </Button>
            </div>
          </form>
        )}
      </Modal>
    </div>
  );
}
