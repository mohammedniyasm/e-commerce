import { useEffect, useState, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import {
  ArrowLeft,
  Package,
  Truck,
  CreditCard,
  MapPin,
  User,
  Clock,
  CheckCircle,
  AlertCircle,
  RotateCcw,
  RefreshCw,
  Send,
  DollarSign,
} from 'lucide-react';
import {
  getAdminOrderById,
  updateOrderStatus,
  updateCODPaymentStatus,
  processWalletRefund,
  getShipmentByOrderId,
  createShipment,
  updateShippingStatus,
} from '../../../api/adminOperationsService';
import { Badge } from '../../../components/common/Badge';
import { Button } from '../../../components/common/Button';
import { Modal } from '../../../components/common/Modal';
import { SkeletonLoader } from '../../../components/common/SkeletonLoader';
import { useToast } from '../../../context/ToastContext';
import { formatCurrency, formatDate } from '../../../utils/formatters';
import {
  getShipmentStatusTransitions,
  SHIPMENT_STATUS_LABELS,
} from '../../../utils/shipmentStatus';
import {
  canMarkCodPaymentPaid,
  canProcessWalletRefund,
} from '../../../utils/orderEligibility';
import './AdminOrderDetailsPage.css';

// Exact valid transitions from backend internal/usecase/order/order_usecase.go: isValidOrderTransition
const VALID_TRANSITIONS = {
  pending: ['confirmed', 'cancelled'],
  confirmed: ['processing', 'cancelled'],
  processing: ['shipped', 'cancelled'],
  shipped: ['delivered', 'returned'],
  out_for_delivery: ['delivered', 'cancelled'],
  delivered: [],
  cancelled: [],
  returned: [],
};

export function AdminOrderDetailsPage() {
  const { orderId } = useParams();
  const { showToast } = useToast();

  const [order, setOrder] = useState(null);
  const [shipment, setShipment] = useState(null);
  const [isLoading, setIsLoading] = useState(true);

  // Status transition state
  const [targetStatus, setTargetStatus] = useState('');
  const [statusNote, setStatusNote] = useState('');
  const [isUpdatingStatus, setIsUpdatingStatus] = useState(false);

  // Shipment creation state
  const [isCreateShipmentOpen, setIsCreateShipmentOpen] = useState(false);
  const [courierName, setCourierName] = useState('BlueDart');
  const [trackingNumber, setTrackingNumber] = useState('');
  const [isCreatingShipment, setIsCreatingShipment] = useState(false);

  // Shipment status update state
  const [isUpdateShipmentOpen, setIsUpdateShipmentOpen] = useState(false);
  const [newShipmentStatus, setNewShipmentStatus] = useState('');
  const [shipmentLocation, setShipmentLocation] = useState('');
  const [shipmentUpdateNotes, setShipmentUpdateNotes] = useState('');
  const [isUpdatingShipment, setIsUpdatingShipment] = useState(false);

  // Refund state
  const [isRefundModalOpen, setIsRefundModalOpen] = useState(false);
  const [isRefunding, setIsRefunding] = useState(false);
  const availableShipmentTransitions = getShipmentStatusTransitions(shipment?.shipping_status);

  // COD Payment update
  const [isMarkingPaid, setIsMarkingPaid] = useState(false);

  const fetchOrderData = useCallback(async () => {
    setIsLoading(true);
    setOrder(null);
    setShipment(null);
    try {
      const orderData = await getAdminOrderById(orderId);
      setOrder(orderData);

      // Try fetching shipment details
      try {
        const shipmentData = await getShipmentByOrderId(orderId);
        setShipment(shipmentData);
      } catch (shipErr) {
        if (shipErr?.status === 404) {
          setShipment(null);
        } else {
          console.error('Failed to load shipment details:', shipErr);
          showToast(shipErr?.message || 'Failed to load shipment details', 'error');
        }
      }
    } catch (err) {
      console.error('Failed to load order details:', err);
      showToast(err?.response?.data?.message || 'Failed to load order details', 'error');
    } finally {
      setIsLoading(false);
    }
  }, [orderId, showToast]);

  useEffect(() => {
    fetchOrderData();
  }, [fetchOrderData]);

  // Handle Order Status Update
  const handleStatusUpdate = async (e) => {
    e.preventDefault();
    if (!targetStatus) return;

    setIsUpdatingStatus(true);
    try {
      await updateOrderStatus(orderId, targetStatus, statusNote.trim() || undefined);
      showToast(`Order status updated to "${targetStatus}"`, 'success');
      setStatusNote('');
      setTargetStatus('');
      await fetchOrderData();
    } catch (err) {
      console.error('Failed to update status:', err);
      showToast(err?.response?.data?.message || 'Failed to update order status', 'error');
    } finally {
      setIsUpdatingStatus(false);
    }
  };

  // Handle COD Payment Update
  const handleMarkCODPaid = async () => {
    setIsMarkingPaid(true);
    try {
      await updateCODPaymentStatus(orderId, 'paid');
      showToast('COD Payment marked as Paid', 'success');
      await fetchOrderData();
    } catch (err) {
      console.error('Failed to update COD payment:', err);
      showToast(err?.response?.data?.message || 'Failed to update payment status', 'error');
    } finally {
      setIsMarkingPaid(false);
    }
  };

  // Handle Wallet Refund
  const handleProcessRefund = async () => {
    setIsRefunding(true);
    try {
      await processWalletRefund(orderId);
      showToast('Refund processed successfully to customer wallet', 'success');
      setIsRefundModalOpen(false);
      await fetchOrderData();
    } catch (err) {
      console.error('Failed to process wallet refund:', err);
      showToast(err?.response?.data?.message || 'Failed to process wallet refund', 'error');
    } finally {
      setIsRefunding(false);
    }
  };

  // Handle Create Shipment
  const handleCreateShipment = async (e) => {
    e.preventDefault();
    if (!courierName.trim() || !trackingNumber.trim()) {
      showToast('Courier name and tracking number are required', 'error');
      return;
    }

    setIsCreatingShipment(true);
    try {
      await createShipment({
        order_id: orderId,
        courier_name: courierName.trim(),
        tracking_number: trackingNumber.trim(),
      });
      showToast('Shipment created successfully', 'success');
      setIsCreateShipmentOpen(false);
      await fetchOrderData();
    } catch (err) {
      console.error('Failed to create shipment:', err);
      showToast(err?.message || 'Failed to create shipment', 'error');
    } finally {
      setIsCreatingShipment(false);
    }
  };

  // Handle Update Shipment Status
  const handleUpdateShipmentStatus = async (e) => {
    e.preventDefault();
    if (!shipment?.id) return;
    if (!getShipmentStatusTransitions(shipment.shipping_status).includes(newShipmentStatus)) {
      showToast('That shipment status transition is not allowed.', 'error');
      return;
    }

    setIsUpdatingShipment(true);
    try {
      await updateShippingStatus(shipment.id, {
        status: newShipmentStatus,
        location: shipmentLocation.trim() || undefined,
        description: shipmentUpdateNotes.trim() || undefined,
      });
      showToast('Shipment status updated', 'success');
      setIsUpdateShipmentOpen(false);
      await fetchOrderData();
    } catch (err) {
      console.error('Failed to update shipment status:', err);
      showToast(err?.message || 'Failed to update shipment status', 'error');
    } finally {
      setIsUpdatingShipment(false);
    }
  };

  if (isLoading) {
    return (
      <div className="admin-order-details-page">
        <SkeletonLoader height={40} width={200} />
        <SkeletonLoader height={180} />
        <SkeletonLoader height={300} />
      </div>
    );
  }

  if (!order) {
    return (
      <div className="admin-order-details-page">
        <div className="admin-page-header">
          <Link to="/admin/orders" className="admin-back-link">
            <ArrowLeft size={16} /> Back to Orders
          </Link>
        </div>
        <div className="admin-empty-card">
          <AlertCircle size={36} color="var(--color-danger, #ef4444)" />
          <h3>Order Not Found</h3>
          <p>The requested order #{orderId} does not exist or has been removed.</p>
        </div>
      </div>
    );
  }

  const currentStatus = String(order.order_status || '').toLowerCase();
  const allowedTransitions = VALID_TRANSITIONS[currentStatus] || [];
  const canRefund = canProcessWalletRefund(order);
  const canMarkCODPaid = canMarkCodPaymentPaid(order);
  const shippingAddress = order.address || order.shipping_address;

  return (
    <div className="admin-order-details-page">
      {/* Header & Back Link */}
      <div className="admin-page-header">
        <div>
          <Link to="/admin/orders" className="admin-back-link">
            <ArrowLeft size={16} /> Back to Orders
          </Link>
          <div className="admin-order-title-group">
            <h1 className="admin-page-title">
              Order {order.order_number || `#${order.id}`}
            </h1>
            <Badge
              variant={
                currentStatus === 'delivered'
                  ? 'success'
                  : currentStatus === 'cancelled'
                  ? 'danger'
                  : currentStatus === 'shipped'
                  ? 'warning'
                  : 'info'
              }
            >
              {order.order_status}
            </Badge>
            <Badge
              variant={
                order.payment_status === 'paid'
                  ? 'success'
                  : order.payment_status === 'failed'
                  ? 'danger'
                  : 'warning'
              }
            >
              Payment: {order.payment_status}
            </Badge>
          </div>
          <p className="admin-page-description">Placed on {formatDate(order.created_at)}</p>
        </div>

        <div className="admin-header-actions">
          <Button
            variant="outline"
            size="sm"
            icon={<RefreshCw size={14} />}
            onClick={fetchOrderData}
          >
            Refresh
          </Button>
          {canRefund && (
            <Button
              variant="secondary"
              size="sm"
              icon={<RotateCcw size={14} />}
              onClick={() => setIsRefundModalOpen(true)}
            >
              Process Wallet Refund
            </Button>
          )}
          {canMarkCODPaid && (
            <Button
              variant="primary"
              size="sm"
              leftIcon={<DollarSign size={14} />}
              onClick={handleMarkCODPaid}
              isLoading={isMarkingPaid}
              disabled={isMarkingPaid}
            >
              {isMarkingPaid ? 'Marking...' : 'Mark COD as Paid'}
            </Button>
          )}
        </div>
      </div>

      <div className="admin-order-layout">
        {/* Left Column: Items & Details */}
        <div className="admin-order-main">
          {/* Order Items Table Card */}
          <div className="admin-card">
            <div className="admin-card-header">
              <h2 className="admin-card-title">Order Items</h2>
              <span className="admin-badge-count">
                {order.items?.length || 0} {order.items?.length === 1 ? 'item' : 'items'}
              </span>
            </div>
            <div className="admin-items-table-wrapper">
              <table className="admin-items-table">
                <thead>
                  <tr>
                    <th>Item</th>
                    <th>Variant / SKU</th>
                    <th className="text-right">Price</th>
                    <th className="text-center">Qty</th>
                    <th className="text-right">Total</th>
                  </tr>
                </thead>
                <tbody>
                  {(order.items || []).map((item, idx) => {
                    const variantInfo =
                      item.variant_info && typeof item.variant_info === 'object'
                        ? item.variant_info
                        : {};
                    const variantLabel = [
                      variantInfo.size && `Size: ${variantInfo.size}`,
                      variantInfo.color && `Color: ${variantInfo.color}`,
                      (variantInfo.sku || item.variant?.sku || item.sku) &&
                        `SKU: ${variantInfo.sku || item.variant?.sku || item.sku}`,
                    ]
                      .filter(Boolean)
                      .join(' · ');
                    const unitPrice = Number(item.unit_price ?? item.price ?? 0);
                    const itemTotal = Number(item.subtotal ?? unitPrice * Number(item.quantity || 0));
                    const productName = item.product_name || item.product?.name || `Item #${item.id}`;
                    const imageUrl = item.image_url || item.product?.thumbnail;

                    return (
                      <tr key={item.id || idx}>
                        <td>
                          <div className="admin-item-info">
                            {imageUrl && (
                              <img
                                src={imageUrl}
                                alt={productName}
                                className="admin-item-thumb"
                              />
                            )}
                            <div>
                              <span className="admin-item-name">{productName}</span>
                              {item.product?.brand?.name && (
                                <span className="admin-item-brand">{item.product.brand.name}</span>
                              )}
                            </div>
                          </div>
                        </td>
                        <td>
                          <span className="admin-item-variant">
                            {variantLabel || `Variant #${item.variant_id || item.variant?.id || '—'}`}
                          </span>
                        </td>
                        <td className="text-right">{formatCurrency(unitPrice)}</td>
                        <td className="text-center">{item.quantity}</td>
                        <td className="text-right font-semibold">{formatCurrency(itemTotal)}</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>

            {/* Price Breakdown */}
            <div className="admin-order-breakdown">
              <div className="admin-breakdown-row">
                <span>Subtotal</span>
                <span>{formatCurrency(order.subtotal || order.total_amount)}</span>
              </div>
              {Number(order.discount_amount) > 0 && (
                <div className="admin-breakdown-row text-success">
                  <span>Coupon Discount {order.coupon_code ? `(${order.coupon_code})` : ''}</span>
                  <span>- {formatCurrency(order.discount_amount)}</span>
                </div>
              )}
              {Number(order.delivery_charge) > 0 && (
                <div className="admin-breakdown-row">
                  <span>Shipping Fee</span>
                  <span>{formatCurrency(order.delivery_charge)}</span>
                </div>
              )}
              <div className="admin-breakdown-row admin-breakdown-total">
                <span>Grand Total</span>
                <span>{formatCurrency(order.total_amount)}</span>
              </div>
            </div>
          </div>

          {/* Shipment Management Card */}
          <div className="admin-card">
            <div className="admin-card-header">
              <div className="admin-card-header-left">
                <Truck size={18} />
                <h2 className="admin-card-title">Fulfillment & Shipment</h2>
              </div>
              <div className="admin-shipment-card-actions">
                {shipment ? (
                  availableShipmentTransitions.length > 0 && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        setNewShipmentStatus(availableShipmentTransitions[0]);
                        setIsUpdateShipmentOpen(true);
                      }}
                    >
                      Update Status
                    </Button>
                  )
                ) : (
                  <Button
                    variant="primary"
                    size="sm"
                    onClick={() => setIsCreateShipmentOpen(true)}
                  >
                    Create Shipment
                  </Button>
                )}
                <Link
                  to={`/admin/shipments?orderId=${encodeURIComponent(orderId)}`}
                  className="admin-shipment-manager-link"
                >
                  Shipment Manager
                </Link>
              </div>
            </div>

            {shipment ? (
              <div className="admin-shipment-details">
                <div className="admin-shipment-grid">
                  <div>
                    <span className="admin-field-label">Courier</span>
                    <span className="admin-field-value font-semibold">{shipment.courier_name || 'Not assigned'}</span>
                  </div>
                  <div>
                    <span className="admin-field-label">Tracking Number</span>
                    <span className="admin-field-value font-mono">
                      {shipment.tracking_number}
                    </span>
                  </div>
                  <div>
                    <span className="admin-field-label">Status</span>
                    <Badge variant="info">
                      {SHIPMENT_STATUS_LABELS[shipment.shipping_status] || shipment.shipping_status || 'Unknown'}
                    </Badge>
                  </div>
                  <div>
                    <span className="admin-field-label">Shipped At</span>
                    <span className="admin-field-value">
                      {shipment.shipped_at ? formatDate(shipment.shipped_at) : '—'}
                    </span>
                  </div>
                  <div>
                    <span className="admin-field-label">Delivered At</span>
                    <span className="admin-field-value">
                      {shipment.delivered_at ? formatDate(shipment.delivered_at) : '—'}
                    </span>
                  </div>
                </div>

                {shipment.history && shipment.history.length > 0 && (
                  <div className="admin-tracking-timeline">
                    <h4>Tracking Timeline</h4>
                    <div className="admin-timeline">
                      {shipment.history.map((event, i) => (
                        <div key={i} className="admin-timeline-item">
                          <div className="admin-timeline-dot" />
                          <div className="admin-timeline-content">
                            <span className="admin-timeline-status">
                              {SHIPMENT_STATUS_LABELS[event.shipping_status] || event.shipping_status}
                            </span>
                            {event.location && (
                              <span className="admin-timeline-location">{event.location}</span>
                            )}
                            {event.description && (
                              <span className="admin-timeline-notes">{event.description}</span>
                            )}
                            <span className="admin-timeline-time">{formatDate(event.created_at)}</span>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            ) : (
              <div className="admin-empty-shipment">
                <Package size={28} className="text-muted" />
                <p>No shipment registered for this order yet.</p>
              </div>
            )}
          </div>
        </div>

        {/* Right Column: Customer Info & Status Transition */}
        <div className="admin-order-sidebar">
          {/* Status Transition Control */}
          <div className="admin-card">
            <h3 className="admin-card-subtitle">Manage Status</h3>
            <p className="admin-status-desc">
              Current state: <strong>{order.order_status}</strong>
            </p>

            {allowedTransitions.length > 0 ? (
              <form onSubmit={handleStatusUpdate} className="admin-status-form">
                <label className="admin-form-label">
                  Allowed Next State:
                  <select
                    value={targetStatus}
                    onChange={(e) => setTargetStatus(e.target.value)}
                    className="admin-select-full"
                    required
                  >
                    <option value="">Select Next Status...</option>
                    {allowedTransitions.map((t) => (
                      <option key={t} value={t}>
                        {t.toUpperCase()}
                      </option>
                    ))}
                  </select>
                </label>

                <label className="admin-form-label">
                  Note / Reason (optional):
                  <input
                    type="text"
                    value={statusNote}
                    onChange={(e) => setStatusNote(e.target.value)}
                    placeholder="e.g. Dispatched via express courier"
                    className="admin-input-full"
                  />
                </label>

                <Button
                  type="submit"
                  variant="primary"
                  size="sm"
                  disabled={!targetStatus || isUpdatingStatus}
                  className="w-full"
                >
                  {isUpdatingStatus ? 'Updating...' : 'Update Order Status'}
                </Button>
              </form>
            ) : (
              <div className="admin-terminal-status">
                <CheckCircle size={16} />
                <span>Order is in a terminal status ({order.order_status}).</span>
              </div>
            )}
          </div>

          {/* Customer & Shipping Details */}
          <div className="admin-card">
            <div className="admin-card-section">
              <div className="admin-section-header">
                <User size={16} />
                <h4>Customer Details</h4>
              </div>
              <div className="admin-details-block">
                <span className="font-semibold">
                  {order.user?.name || shippingAddress?.name || 'Guest Customer'}
                </span>
                <span>{order.user?.email || 'No email provided'}</span>
                <span>{order.user?.phone || shippingAddress?.phone || 'No phone provided'}</span>
              </div>
            </div>

            <div className="admin-card-section">
              <div className="admin-section-header">
                <MapPin size={16} />
                <h4>Shipping Address</h4>
              </div>
              <div className="admin-details-block">
                {shippingAddress ? (
                  <>
                    <span className="font-semibold">{shippingAddress.name}</span>
                    <span>
                      {shippingAddress.address_line1 || shippingAddress.street || '—'}
                    </span>
                    {(shippingAddress.address_line2 || shippingAddress.address_line_2) && (
                      <span>{shippingAddress.address_line2 || shippingAddress.address_line_2}</span>
                    )}
                    <span>
                      {[
                        shippingAddress.city,
                        shippingAddress.state,
                        shippingAddress.pincode || shippingAddress.postal_code,
                        shippingAddress.country,
                      ]
                        .filter(Boolean)
                        .join(', ') || '—'}
                    </span>
                    <span>Phone: {shippingAddress.phone || '—'}</span>
                  </>
                ) : (
                  <span>No shipping address on file</span>
                )}
              </div>
            </div>

            <div className="admin-card-section">
              <div className="admin-section-header">
                <CreditCard size={16} />
                <h4>Payment Information</h4>
              </div>
              <div className="admin-details-block">
                <span>Method: <strong>{order.payment_method || 'Online'}</strong></span>
                <span>Status: <strong>{order.payment_status}</strong></span>
                {order.transaction_id && <span>Txn ID: {order.transaction_id}</span>}
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Modal: Create Shipment */}
      <Modal
        isOpen={isCreateShipmentOpen}
        onClose={() => setIsCreateShipmentOpen(false)}
        title="Create Order Shipment"
      >
        <form onSubmit={handleCreateShipment} className="admin-modal-form">
          <label className="admin-form-label">
            Courier Name *
            <input
              type="text"
              value={courierName}
              onChange={(e) => setCourierName(e.target.value)}
              placeholder="e.g. BlueDart, Delhivery, FedEx"
              required
              className="admin-input-full"
            />
          </label>
          <label className="admin-form-label">
            Tracking Number *
            <input
              type="text"
              value={trackingNumber}
              onChange={(e) => setTrackingNumber(e.target.value)}
              placeholder="e.g. BD9823487123IN"
              required
              className="admin-input-full"
            />
          </label>
          <div className="admin-modal-actions">
            <Button
              type="button"
              variant="outline"
              onClick={() => setIsCreateShipmentOpen(false)}
            >
              Cancel
            </Button>
            <Button type="submit" variant="primary" disabled={isCreatingShipment}>
              {isCreatingShipment ? 'Creating...' : 'Create Shipment'}
            </Button>
          </div>
        </form>
      </Modal>

      {/* Modal: Update Shipment Status */}
      <Modal
        isOpen={isUpdateShipmentOpen}
        onClose={() => setIsUpdateShipmentOpen(false)}
        title="Update Shipment Status"
      >
        <form onSubmit={handleUpdateShipmentStatus} className="admin-modal-form">
          <label className="admin-form-label">
            Status *
            <select
              value={newShipmentStatus}
              onChange={(e) => setNewShipmentStatus(e.target.value)}
              className="admin-select-full"
              required
            >
              {availableShipmentTransitions.map((status) => (
                <option key={status} value={status}>
                  {SHIPMENT_STATUS_LABELS[status] || status}
                </option>
              ))}
            </select>
          </label>
          <label className="admin-form-label">
            Current Location
            <input
              type="text"
              value={shipmentLocation}
              onChange={(e) => setShipmentLocation(e.target.value)}
              placeholder="e.g. Regional Hub - Mumbai"
              className="admin-input-full"
            />
          </label>
          <label className="admin-form-label">
            Description
            <textarea
              value={shipmentUpdateNotes}
              onChange={(e) => setShipmentUpdateNotes(e.target.value)}
              placeholder="Package processed at sorting facility."
              className="admin-input-full"
              rows={2}
            />
          </label>

          <div className="admin-modal-actions">
            <Button
              type="button"
              variant="outline"
              onClick={() => setIsUpdateShipmentOpen(false)}
            >
              Cancel
            </Button>
            <Button type="submit" variant="primary" disabled={isUpdatingShipment}>
              {isUpdatingShipment ? 'Updating...' : 'Update Status'}
            </Button>
          </div>
        </form>
      </Modal>

      {/* Modal: Process Wallet Refund */}
      <Modal
        isOpen={isRefundModalOpen}
        onClose={() => setIsRefundModalOpen(false)}
        title="Confirm Wallet Refund"
      >
        <div className="admin-modal-content">
          <p>
            Are you sure you want to process a refund for order{' '}
            <strong>{order.order_number || `#${order.id}`}</strong>?
          </p>
          <p className="text-muted">
            The full refundable amount of{' '}
            <strong>{formatCurrency(order.total_amount)}</strong> will be credited directly
            to the customer&apos;s DRAPE Wallet balance.
          </p>
          <div className="admin-modal-actions">
            <Button
              type="button"
              variant="outline"
              onClick={() => setIsRefundModalOpen(false)}
            >
              Cancel
            </Button>
            <Button
              type="button"
              variant="secondary"
              onClick={handleProcessRefund}
              disabled={isRefunding}
            >
              {isRefunding ? 'Refunding...' : 'Confirm Wallet Refund'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
