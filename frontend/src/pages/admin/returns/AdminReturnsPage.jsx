import { useEffect, useState, useCallback } from 'react';
import { Link } from 'react-router-dom';
import {
  RotateCcw,
  RefreshCw,
  Truck,
  Repeat,
  AlertCircle,
  Eye,
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
import { formatCurrency, formatDate } from '../../../utils/formatters';
import './AdminReturnsPage.css';

const STATUS_OPTIONS = [
  { value: '', label: 'All Statuses' },
  { value: 'requested', label: 'Requested' },
  { value: 'approved', label: 'Approved' },
  { value: 'rejected', label: 'Rejected' },
  { value: 'pickup_scheduled', label: 'Pickup Scheduled' },
  { value: 'in_transit', label: 'In Transit' },
  { value: 'received', label: 'Received' },
  { value: 'inspection_pending', label: 'Inspection Pending' },
  { value: 'inspection_passed', label: 'Inspection Passed' },
  { value: 'inspection_failed', label: 'Inspection Failed' },
  { value: 'closed', label: 'Closed' },
];

const STATUS_TRANSITIONS = {
  requested: ['approved', 'rejected'],
  approved: ['pickup_scheduled', 'closed'],
  pickup_scheduled: ['in_transit', 'closed'],
  in_transit: ['received'],
  received: ['inspection_pending'],
  inspection_pending: ['inspection_passed', 'inspection_failed'],
  inspection_passed: ['closed'],
};

const STATUS_LABELS = {
  requested: 'Requested',
  approved: 'Approved',
  rejected: 'Rejected',
  pickup_scheduled: 'Pickup Scheduled',
  in_transit: 'In Transit',
  received: 'Received',
  inspection_pending: 'Inspection Pending',
  inspection_passed: 'Inspection Passed',
  inspection_failed: 'Inspection Failed',
  closed: 'Closed',
};

const STATUS_ACTION_LABELS = {
  requested: 'Review Request',
  approved: 'Schedule Pickup',
  pickup_scheduled: 'Update Pickup Status',
  in_transit: 'Mark Received',
  received: 'Start Inspection',
  inspection_pending: 'Record Inspection',
  inspection_passed: 'Close Request',
};

function normalizeStatus(status) {
  return String(status || '').trim().toLowerCase();
}

function getRequestType(request) {
  return String(request.request_type || request.type || '').trim().toLowerCase();
}

function getVariantInfoLabel(variantInfo) {
  if (!variantInfo || typeof variantInfo !== 'object') {
    return typeof variantInfo === 'string' ? variantInfo : '';
  }
  return Object.entries(variantInfo)
    .filter(([, value]) => value !== null && value !== undefined && value !== '')
    .map(([key, value]) => `${key.replaceAll('_', ' ')}: ${value}`)
    .join(' · ');
}

function RequestedExchangeVariantDetails({ variant, variantID, title }) {
  if (!variant) {
    return (
      <section className="admin-return-detail-card">
        <h3>{title}</h3>
        <p>Variant #{variantID || '—'} details are unavailable.</p>
      </section>
    );
  }

  const variantLabel = [variant.size, variant.color].filter(Boolean).join(' · ');

  return (
    <section className="admin-return-detail-card">
      <h3>{title}</h3>
      <div className="admin-return-item">
        {variant.image_url && (
          <img
            src={variant.image_url}
            alt={`Requested exchange variant ${variantLabel || variant.sku || variant.id}`}
          />
        )}
        <div>
          <p>{variantLabel || `Variant #${variant.id}`}</p>
          <p>SKU: {variant.sku || '—'}</p>
          <p>
            Price: {formatCurrency(variant.selling_price)} · Available stock: {variant.stock}
          </p>
        </div>
      </div>
    </section>
  );
}

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
  const [modalMode, setModalMode] = useState('status');
  const [actionStatus, setActionStatus] = useState('approved');
  const [adminNote, setAdminNote] = useState('');
  const [courierName, setCourierName] = useState('');
  const [trackingNumber, setTrackingNumber] = useState('');
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

  const closeActionModal = () => {
    setSelectedRequest(null);
    setModalMode('status');
    setAdminNote('');
    setCourierName('');
    setTrackingNumber('');
  };

  const openStatusModal = (request) => {
    const transitions = STATUS_TRANSITIONS[normalizeStatus(request.status)] || [];
    if (transitions.length === 0) return;

    setSelectedRequest(request);
    setModalMode('status');
    setActionStatus(transitions[0]);
    setAdminNote('');
    setCourierName('');
    setTrackingNumber('');
  };

  const openDetailsModal = (request) => {
    setSelectedRequest(request);
    setModalMode('details');
  };

  const openReturnShipmentModal = (request) => {
    setSelectedRequest(request);
    setModalMode('return-shipment');
    setAdminNote('');
    setCourierName('');
    setTrackingNumber('');
  };

  const handleStatusSubmit = async (e) => {
    e.preventDefault();
    if (!selectedRequest) return;

    setIsUpdatingStatus(true);
    try {
      const payload = {
        status: actionStatus,
        admin_note: adminNote.trim() || undefined,
      };
      if (actionStatus === 'pickup_scheduled') {
        payload.courier_name = courierName.trim();
        payload.tracking_number = trackingNumber.trim();
      }

      await updateReturnRequestStatus(selectedRequest.id, payload);
      showToast(`Request moved to "${STATUS_LABELS[actionStatus]}"`, 'success');
      closeActionModal();
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

  const handleCreateShipment = async (e) => {
    e.preventDefault();
    if (!selectedRequest) return;

    setIsCreatingShipment(true);
    try {
      await createReturnShipment(selectedRequest.id, {
        courier_name: courierName.trim(),
        tracking_number: trackingNumber.trim(),
      });
      showToast('Return-to-customer shipment created successfully', 'success');
      closeActionModal();
      await fetchReturns();
    } catch (err) {
      console.error('Failed to create return shipment:', err);
      showToast(err?.response?.data?.message || 'Failed to create return shipment', 'error');
    } finally {
      setIsCreatingShipment(false);
    }
  };

  const getStatusBadge = (status) => {
    const s = normalizeStatus(status);
    const label = STATUS_LABELS[s] || s.replaceAll('_', ' ');
    switch (s) {
      case 'approved':
        return <Badge variant="info">Approved</Badge>;
      case 'inspection_passed':
        return <Badge variant="success">{label}</Badge>;
      case 'rejected':
        return <Badge variant="danger">Rejected</Badge>;
      case 'inspection_failed':
        return <Badge variant="danger">{label}</Badge>;
      case 'requested':
        return <Badge variant="warning">Requested</Badge>;
      case 'inspection_pending':
        return <Badge variant="warning">{label}</Badge>;
      case 'closed':
        return <Badge variant="neutral">Closed</Badge>;
      default:
        return <Badge variant="info">{label}</Badge>;
    }
  };

  const columns = [
    {
      key: 'id',
      header: 'Request #',
      render: (val, record) => {
        const req = (typeof val === 'object' && val !== null) ? val : (record || {});
        return (
          <div className="drape-order-cell">
            <span className="font-semibold">REQ-{req.id}</span>
            <span className="drape-order-date">{formatDate(req.created_at)}</span>
          </div>
        );
      },
    },
    {
      key: 'order',
      header: 'Order #',
      render: (val, record) => {
        const req = record || {};
        const orderId = req.order_id || req.order?.id;
        const orderLabel =
          req.order?.order_number || req.order_number || (orderId ? `#${orderId}` : '—');

        if (!orderId) {
          return <span className="drape-order-id-link">{orderLabel}</span>;
        }

        return (
          <Link to={`/admin/orders/${orderId}`} className="drape-order-id-link">
            {orderLabel}
          </Link>
        );
      },
    },
    {
      key: 'type',
      header: 'Type',
      align: 'center',
      render: (val, record) => {
        const req = (typeof val === 'object' && val !== null) ? val : (record || {});
        const isExchange = getRequestType(req) === 'exchange';
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
      render: (val, record) => {
        const req = (typeof val === 'object' && val !== null) ? val : (record || {});
        return (
          <div className="drape-customer-cell">
            <span className="font-semibold">{req.order_item?.product_name || req.product_name || `Item #${req.order_item_id}`}</span>
            <span className="drape-customer-email">Qty: {req.quantity || 1}</span>
          </div>
        );
      },
    },
    {
      key: 'reason',
      header: 'Reason',
      render: (val, record) => {
        const req = (typeof val === 'object' && val !== null) ? val : (record || {});
        return (
          <div className="drape-reason-cell">
            <span className="drape-reason-title">{req.reason || 'Not specified'}</span>
            {req.comments && <span className="drape-reason-desc">{req.comments}</span>}
          </div>
        );
      },
    },
    {
      key: 'status',
      header: 'Status',
      align: 'center',
      render: (val, record) => {
        const req = (typeof val === 'object' && val !== null) ? val : (record || {});
        return getStatusBadge(req.status || val);
      },
    },
    {
      key: 'actions',
      header: 'Actions',
      align: 'right',
      render: (val, record) => {
        const req = (typeof val === 'object' && val !== null) ? val : (record || {});
        const status = normalizeStatus(req.status);
        const transitions = STATUS_TRANSITIONS[status] || [];
        const isExchange = getRequestType(req) === 'exchange';

        return (
          <div className="drape-table-actions">
            <Button
              variant="ghost"
              size="sm"
              leftIcon={<Eye size={14} />}
              onClick={() => openDetailsModal(req)}
            >
              Details
            </Button>

            {transitions.length > 0 && (
              <Button
                variant="primary"
                size="sm"
                leftIcon={status === 'inspection_pending' ? <AlertCircle size={12} /> : null}
                onClick={() => openStatusModal(req)}
              >
                {STATUS_ACTION_LABELS[status] || 'Update Status'}
              </Button>
            )}

            {status === 'inspection_passed' &&
              isExchange &&
              Number(req.requested_variant_id) > 0 && (
                <Button
                  variant="secondary"
                  size="sm"
                  leftIcon={<Repeat size={12} />}
                  onClick={() => handleCreateExchange(req)}
                  disabled={isProcessingExchange}
                >
                  Create Exchange Order
                </Button>
              )}

            {status === 'inspection_failed' && (
              <Button
                variant="outline"
                size="sm"
                leftIcon={<Truck size={12} />}
                onClick={() => openReturnShipmentModal(req)}
                disabled={isCreatingShipment}
              >
                Return to Customer
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
            leftIcon={<RefreshCw size={14} className={isLoading ? 'spin-icon' : ''} />}
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
        className="admin-returns-table"
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

      <Modal
        isOpen={Boolean(selectedRequest)}
        onClose={closeActionModal}
        title={
          modalMode === 'return-shipment'
            ? `Return Item to Customer REQ-${selectedRequest?.id}`
            : modalMode === 'details'
              ? `Return Request REQ-${selectedRequest?.id}`
            : `${normalizeStatus(selectedRequest?.status) === 'requested' ? 'Review' : 'Update'} Request REQ-${selectedRequest?.id}`
        }
      >
        {selectedRequest && (
          modalMode === 'details' ? (
            <div className="admin-return-details">
              <div className="admin-return-details-grid">
                <div>
                  <span>Request Type</span>
                  <strong>{getRequestType(selectedRequest) || 'Return'}</strong>
                </div>
                <div>
                  <span>Status</span>
                  <strong>{STATUS_LABELS[normalizeStatus(selectedRequest.status)] || selectedRequest.status}</strong>
                </div>
                <div>
                  <span>Order</span>
                  <strong>{selectedRequest.order?.order_number || `#${selectedRequest.order_id}`}</strong>
                </div>
                <div>
                  <span>Customer ID</span>
                  <strong>{selectedRequest.order?.user_id || '—'}</strong>
                </div>
                <div>
                  <span>Request Date</span>
                  <strong>{formatDate(selectedRequest.created_at)}</strong>
                </div>
                <div>
                  <span>Requested Quantity</span>
                  <strong>{selectedRequest.quantity}</strong>
                </div>
              </div>

              <section className="admin-return-detail-card">
                <h3>Requested Item</h3>
                <div className="admin-return-item">
                  {selectedRequest.order_item?.image_url && (
                    <img
                      src={selectedRequest.order_item.image_url}
                      alt={selectedRequest.order_item.product_name || 'Requested product'}
                    />
                  )}
                  <div>
                    <strong>
                      {selectedRequest.order_item?.product_name || `Item #${selectedRequest.order_item_id}`}
                    </strong>
                    <p>
                      {getVariantInfoLabel(selectedRequest.order_item?.variant_info) ||
                        `Variant #${selectedRequest.order_item?.variant_id || '—'}`}
                    </p>
                    <p>
                      Unit price: {formatCurrency(selectedRequest.order_item?.unit_price || 0)} ·
                      {' '}Ordered quantity: {selectedRequest.order_item?.quantity || '—'}
                    </p>
                    <p>
                      Requested quantity: {selectedRequest.quantity} ·
                      {' '}Subtotal: {formatCurrency(selectedRequest.order_item?.subtotal || 0)}
                    </p>
                  </div>
                </div>
              </section>

              {getRequestType(selectedRequest) === 'exchange' && (
                <RequestedExchangeVariantDetails
                  variant={selectedRequest.requested_variant}
                  variantID={selectedRequest.requested_variant_id}
                  title="Requested Exchange Variant"
                />
              )}

              <section className="admin-return-detail-card">
                <h3>Customer Reason</h3>
                <p>{selectedRequest.reason || 'No reason provided.'}</p>
              </section>

              {selectedRequest.admin_note && (
                <section className="admin-return-detail-card">
                  <h3>Admin Notes</h3>
                  <p>{selectedRequest.admin_note}</p>
                </section>
              )}
            </div>
          ) : modalMode === 'status' ? (
            <form onSubmit={handleStatusSubmit} className="admin-modal-form">
              {normalizeStatus(selectedRequest.status) === 'inspection_pending' && (
                <div className="admin-return-inspection-info">
                  <h3>Inspection Information</h3>
                  <div className="admin-return-details-grid">
                    <div>
                      <span>Request Type</span>
                      <strong>{getRequestType(selectedRequest) || 'Return'}</strong>
                    </div>
                    <div>
                      <span>Order Number</span>
                      <strong>{selectedRequest.order?.order_number || `#${selectedRequest.order_id}`}</strong>
                    </div>
                    <div>
                      <span>Customer ID</span>
                      <strong>{selectedRequest.order?.user_id || '—'}</strong>
                    </div>
                    <div>
                      <span>Order Status</span>
                      <strong>{selectedRequest.order?.order_status || '—'}</strong>
                    </div>
                    <div>
                      <span>Payment Status</span>
                      <strong>{selectedRequest.order?.payment_status || '—'}</strong>
                    </div>
                    <div>
                      <span>Request Submitted</span>
                      <strong>{formatDate(selectedRequest.created_at)}</strong>
                    </div>
                    <div>
                      <span>Requested Quantity</span>
                      <strong>{selectedRequest.quantity}</strong>
                    </div>
                  </div>

                  <section className="admin-return-detail-card">
                    <h3>Item Received for Inspection</h3>
                    <div className="admin-return-item">
                      {selectedRequest.order_item?.image_url && (
                        <img
                          src={selectedRequest.order_item.image_url}
                          alt={selectedRequest.order_item.product_name || 'Returned product'}
                        />
                      )}
                      <div>
                        <strong>
                          {selectedRequest.order_item?.product_name || `Item #${selectedRequest.order_item_id}`}
                        </strong>
                        <p>
                          Original variant:{' '}
                          {getVariantInfoLabel(selectedRequest.order_item?.variant_info) ||
                            `Variant #${selectedRequest.order_item?.variant_id || '—'}`}
                        </p>
                        <p>
                          Ordered: {selectedRequest.order_item?.quantity || '—'} ·
                          {' '}Returning: {selectedRequest.quantity}
                        </p>
                        <p>
                          Unit price: {formatCurrency(selectedRequest.order_item?.unit_price || 0)} ·
                          {' '}Order item subtotal: {formatCurrency(selectedRequest.order_item?.subtotal || 0)}
                        </p>
                      </div>
                    </div>
                  </section>

                  {getRequestType(selectedRequest) === 'exchange' && (
                    <RequestedExchangeVariantDetails
                      variant={selectedRequest.requested_variant}
                      variantID={selectedRequest.requested_variant_id}
                      title="Requested Replacement Variant"
                    />
                  )}

                  <section className="admin-return-detail-card">
                    <h3>Customer Reason</h3>
                    <p>{selectedRequest.reason || 'No reason provided.'}</p>
                  </section>
                </div>
              )}

              <div className="admin-modal-info-block">
                <p><strong>Current Status:</strong> {STATUS_LABELS[normalizeStatus(selectedRequest.status)]}</p>
                {normalizeStatus(selectedRequest.status) !== 'inspection_pending' && (
                  <>
                    <p><strong>Order:</strong> #{selectedRequest.order_id}</p>
                    <p><strong>Type:</strong> {getRequestType(selectedRequest)}</p>
                    <p><strong>Customer Reason:</strong> {selectedRequest.reason || 'None provided'}</p>
                    {selectedRequest.comments && (
                      <p>
                        <strong>Customer Comments:</strong> {selectedRequest.comments}
                      </p>
                    )}
                  </>
                )}
              </div>

              <label className="admin-form-label">
                Next Status *
                <select
                  value={actionStatus}
                  onChange={(e) => setActionStatus(e.target.value)}
                  className="admin-select-full"
                  required
                >
                  {(STATUS_TRANSITIONS[normalizeStatus(selectedRequest.status)] || []).map((status) => (
                    <option key={status} value={status}>{STATUS_LABELS[status]}</option>
                  ))}
                </select>
              </label>

              {actionStatus === 'pickup_scheduled' && (
                <>
                  <label className="admin-form-label">
                    Courier Name *
                    <input
                      type="text"
                      value={courierName}
                      onChange={(e) => setCourierName(e.target.value)}
                      className="admin-input-full"
                      pattern=".*\S.*"
                      minLength={2}
                      maxLength={255}
                      required
                    />
                  </label>
                  <label className="admin-form-label">
                    Tracking Number *
                    <input
                      type="text"
                      value={trackingNumber}
                      onChange={(e) => setTrackingNumber(e.target.value)}
                      className="admin-input-full"
                      pattern=".*\S.*"
                      minLength={3}
                      maxLength={255}
                      required
                    />
                  </label>
                </>
              )}

              {actionStatus === 'inspection_passed' && getRequestType(selectedRequest) === 'return' && (
                <div className="admin-modal-notice">
                  <AlertCircle size={16} aria-hidden="true" />
                  <span>
                    Passing inspection restores the item to stock and credits the refund to the
                    customer&apos;s wallet.
                  </span>
                </div>
              )}

              <label className="admin-form-label">
                {normalizeStatus(selectedRequest.status) === 'inspection_pending'
                  ? 'Inspection Findings / Admin Notes'
                  : 'Admin Notes'}
                <textarea
                  value={adminNote}
                  onChange={(e) => setAdminNote(e.target.value)}
                  placeholder={
                    normalizeStatus(selectedRequest.status) === 'inspection_pending'
                      ? 'Record item condition, packaging, defects, and inspection findings...'
                      : 'Add notes for this status update...'
                  }
                  className="admin-input-full"
                  rows={3}
                />
              </label>

              <div className="admin-modal-actions">
                <Button type="button" variant="outline" onClick={closeActionModal}>
                  Cancel
                </Button>
                <Button type="submit" variant="primary" disabled={isUpdatingStatus}>
                  {isUpdatingStatus ? 'Saving...' : 'Update Status'}
                </Button>
              </div>
            </form>
          ) : (
            <form onSubmit={handleCreateShipment} className="admin-modal-form">
              <div className="admin-modal-notice">
                <AlertCircle size={16} aria-hidden="true" />
                <span>
                  The inbound return shipment must be delivered and inspected as failed before an
                  item can be shipped back to the customer.
                </span>
              </div>
              <label className="admin-form-label">
                Courier Name *
                <input
                  type="text"
                  value={courierName}
                  onChange={(e) => setCourierName(e.target.value)}
                  className="admin-input-full"
                  pattern=".*\S.*"
                  minLength={2}
                  maxLength={255}
                  required
                />
              </label>
              <label className="admin-form-label">
                Tracking Number *
                <input
                  type="text"
                  value={trackingNumber}
                  onChange={(e) => setTrackingNumber(e.target.value)}
                  className="admin-input-full"
                  pattern=".*\S.*"
                  minLength={3}
                  maxLength={255}
                  required
                />
              </label>
              <div className="admin-modal-actions">
                <Button type="button" variant="outline" onClick={closeActionModal}>
                  Cancel
                </Button>
                <Button type="submit" variant="primary" disabled={isCreatingShipment}>
                  {isCreatingShipment ? 'Creating...' : 'Create Shipment'}
                </Button>
              </div>
            </form>
          )
        )}
      </Modal>
    </div>
  );
}
