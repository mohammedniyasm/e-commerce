import React, { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import {
  RotateCcw,
  ArrowRight,
  Package,
  Clock,
  CheckCircle2,
  AlertCircle,
  HelpCircle,
  Sparkles,
  Truck,
  ChevronDown,
  ChevronUp,
  Eye,
} from 'lucide-react';
import {
  Button,
  StatusPill,
  Badge,
  SkeletonLoader,
  Modal,
} from '../../components/common';
import { useToast } from '../../context/ToastContext';
import { getMyReturnRequests, getReturnRequestShipments } from '../../api/returnService';
import { formatCurrency, formatDate, formatDateTime } from '../../utils/formatters';
import './ReturnsPage.css';

function getVariantInfoLabel(variantInfo) {
  if (!variantInfo || typeof variantInfo !== 'object') {
    return typeof variantInfo === 'string' ? variantInfo : '';
  }
  return Object.entries(variantInfo)
    .filter(([, value]) => value !== null && value !== undefined && value !== '')
    .map(([key, value]) => `${key.replaceAll('_', ' ')}: ${value}`)
    .join(' · ');
}

function getShipmentLabel(type) {
  if (type === 'return_to_customer') return 'Shipment back to you';
  if (type === 'return') return 'Return pickup';
  return 'Shipment';
}

function getShipmentStatusVariant(status) {
  switch (String(status || '').toLowerCase()) {
    case 'delivered':
      return 'success';
    case 'failed':
    case 'cancelled':
      return 'danger';
    case 'in_transit':
    case 'out_for_delivery':
    case 'shipped':
      return 'info';
    default:
      return 'warning';
  }
}

export function ReturnsPage() {
  const { showToast } = useToast();
  const [returnRequests, setReturnRequests] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [shipmentDetails, setShipmentDetails] = useState({});
  const [expandedShipmentRequestId, setExpandedShipmentRequestId] = useState(null);
  const [loadingShipmentRequestId, setLoadingShipmentRequestId] = useState(null);
  const [selectedRequest, setSelectedRequest] = useState(null);

  const handleToggleShipmentTracking = async (requestId) => {
    if (expandedShipmentRequestId === requestId) {
      setExpandedShipmentRequestId(null);
      return;
    }

    setExpandedShipmentRequestId(requestId);
    if (Object.prototype.hasOwnProperty.call(shipmentDetails, requestId)) return;

    setLoadingShipmentRequestId(requestId);
    try {
      const shipments = await getReturnRequestShipments(requestId);
      setShipmentDetails((current) => ({
        ...current,
        [requestId]: Array.isArray(shipments) ? shipments : [],
      }));
    } catch (err) {
      setExpandedShipmentRequestId(null);
      showToast({
        title: 'Shipment Tracking Unavailable',
        message: err.message || 'Unable to retrieve shipment tracking details.',
        type: 'danger',
      });
    } finally {
      setLoadingShipmentRequestId(null);
    }
  };

  const fetchReturns = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await getMyReturnRequests({ page: 1, limit: 20 });
      const items = Array.isArray(data) ? data : data?.items || [];
      setReturnRequests(items);
    } catch (err) {
      showToast({
        title: 'Returns Unavailable',
        message: err.message || 'Unable to retrieve returns history.',
        type: 'danger',
      });
    } finally {
      setIsLoading(false);
    }
  }, [showToast]);

  useEffect(() => {
    fetchReturns();
  }, [fetchReturns]);

  return (
    <div className="account-returns-page" data-testid="account-returns-page">
      {/* Header */}
      <div className="account-returns-header">
        <div>
          <h1 className="account-returns-title">Returns & Exchanges</h1>
          <p className="account-returns-subtitle">
            Track return requests, exchanges, and wallet refunds.
          </p>
        </div>
        <Link to="/account/orders">
          <Button variant="outline" size="sm" rightIcon={<ArrowRight size={14} />}>
            VIEW ORDERS FOR RETURNS
          </Button>
        </Link>
      </div>

      {/* Return Requests List */}
      {isLoading ? (
        <div className="account-returns-loading" data-testid="returns-loading">
          <SkeletonLoader lines={6} />
        </div>
      ) : returnRequests.length > 0 ? (
        <div className="returns-list" data-testid="returns-list">
          {returnRequests.map((req) => {
            const status = (req.status || 'pending').toLowerCase();
            const isReturn = (req.request_type || 'return').toLowerCase() === 'return';

            const pillVariant =
              status === 'approved' || status === 'inspection_passed' || status === 'closed'
                ? 'success'
                : status === 'rejected' || status === 'inspection_failed'
                ? 'danger'
                : status === 'in_transit' || status === 'pickup_scheduled'
                ? 'info'
                : 'warning';

            return (
              <article
                key={req.id}
                className="return-request-card"
                data-testid={`return-request-${req.id}`}
              >
                <div className="return-card__header">
                  <div className="return-card__title-group">
                    <Badge variant={isReturn ? 'neutral' : 'info'} size="sm">
                      {isReturn ? 'RETURN' : 'EXCHANGE'}
                    </Badge>
                    <strong className="return-card__id">Request #{req.id}</strong>
                    <span className="return-card__date">
                      Submitted on {formatDate(req.created_at)}
                    </span>
                  </div>
                  <StatusPill status={status.replace(/_/g, ' ')} variant={pillVariant} />
                </div>

                <div className="return-card__body">
                  <div className="return-card__details">
                    <p className="return-card__field">
                      <span className="return-card__label">ORDER REFERENCE:</span>
                      <Link
                        to={`/account/orders/${req.order_id}`}
                        className="return-card__order-link"
                      >
                        #{req.order_id}
                      </Link>
                    </p>

                    <p className="return-card__field">
                      <span className="return-card__label">QUANTITY:</span>
                      <span>{req.quantity} item(s)</span>
                    </p>

                    <p className="return-card__field">
                      <span className="return-card__label">REASON:</span>
                      <span className="return-card__reason-text">{req.reason}</span>
                    </p>

                    {req.admin_note && (
                      <div className="return-card__admin-note">
                        <span className="return-card__label">ADMIN NOTE:</span>
                        <p>{req.admin_note}</p>
                      </div>
                    )}
                  </div>
                </div>

                <div className="return-card__footer">
                  <p className="return-card__footer-policy">
                    {isReturn
                      ? 'Approved refunds are credited to your wallet.'
                      : 'Replacement variant will be dispatched upon return inspection.'}
                  </p>
                  <div className="return-card__actions">
                    <Button
                      variant="ghost"
                      size="sm"
                      leftIcon={<Eye size={14} />}
                      onClick={() => setSelectedRequest(req)}
                      data-testid={`view-return-details-${req.id}`}
                    >
                      View Details
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      leftIcon={<Truck size={14} />}
                      rightIcon={
                        expandedShipmentRequestId === req.id
                          ? <ChevronUp size={14} />
                          : <ChevronDown size={14} />
                      }
                      onClick={() => handleToggleShipmentTracking(req.id)}
                      disabled={loadingShipmentRequestId !== null}
                      isLoading={loadingShipmentRequestId === req.id}
                      data-testid={`track-return-shipment-${req.id}`}
                    >
                      {expandedShipmentRequestId === req.id ? 'Hide Tracking' : 'Track Shipment'}
                    </Button>
                  </div>
                </div>

                {expandedShipmentRequestId === req.id && (
                  <div
                    className="return-card__tracking"
                    data-testid={`return-shipment-tracking-${req.id}`}
                  >
                    {loadingShipmentRequestId === req.id ? (
                      <p className="return-card__tracking-empty">Loading shipment tracking...</p>
                    ) : (shipmentDetails[req.id] || []).length > 0 ? (
                      shipmentDetails[req.id].map((shipment) => (
                        <section className="return-shipment-card" key={shipment.id}>
                          <div className="return-shipment-card__header">
                            <strong>{getShipmentLabel(shipment.type)}</strong>
                            <StatusPill
                              status={(shipment.shipping_status || 'pending').replace(/_/g, ' ')}
                              variant={getShipmentStatusVariant(shipment.shipping_status)}
                            />
                          </div>
                          <div className="return-shipment-card__details">
                            <span>
                              Courier: <strong>{shipment.courier_name || 'Not assigned yet'}</strong>
                            </span>
                            <span>
                              Tracking number: <strong>{shipment.tracking_number || 'Not assigned yet'}</strong>
                            </span>
                            {shipment.delivered_at && (
                              <span>Delivered on {formatDateTime(shipment.delivered_at)}</span>
                            )}
                          </div>
                          {shipment.history?.length > 0 && (
                            <ol className="return-shipment-timeline">
                              {shipment.history.map((update) => (
                                <li key={update.id}>
                                  <div>
                                    <strong>
                                      {(update.shipping_status || '').replace(/_/g, ' ').toUpperCase()}
                                    </strong>
                                    <span>{formatDateTime(update.created_at)}</span>
                                  </div>
                                  {update.location && <span>{update.location}</span>}
                                  {update.description && <p>{update.description}</p>}
                                </li>
                              ))}
                            </ol>
                          )}
                        </section>
                      ))
                    ) : (
                      <p className="return-card__tracking-empty">
                        No shipment tracking is available for this request yet.
                      </p>
                    )}
                  </div>
                )}
              </article>
            );
          })}
        </div>
      ) : (
        <div className="account-empty-block" data-testid="returns-empty-state">
          <RotateCcw size={36} className="account-empty-icon" />
          <h2 className="account-empty-title">No Active Return or Exchange Requests</h2>
          <p className="account-empty-text">
            Need an exchange or return? You can initiate a request on any delivered order from your orders dashboard.
          </p>
          <Link to="/account/orders">
            <Button variant="outline" size="sm" rightIcon={<ArrowRight size={14} />}>
              VIEW DELIVERED ORDERS
            </Button>
          </Link>
        </div>
      )}

      <Modal
        isOpen={Boolean(selectedRequest)}
        onClose={() => setSelectedRequest(null)}
        title={`Return Request #${selectedRequest?.id || ''}`}
      >
        {selectedRequest && (
          <div className="customer-return-details">
            <div className="customer-return-details__summary">
              <div>
                <span>Request Type</span>
                <strong>
                  {(selectedRequest.request_type || 'return').toUpperCase()}
                </strong>
              </div>
              <div>
                <span>Status</span>
                <StatusPill
                  status={(selectedRequest.status || 'pending').replace(/_/g, ' ')}
                  variant={
                    ['approved', 'inspection_passed', 'closed'].includes(
                      (selectedRequest.status || '').toLowerCase(),
                    )
                      ? 'success'
                      : ['rejected', 'inspection_failed'].includes(
                          (selectedRequest.status || '').toLowerCase(),
                        )
                        ? 'danger'
                        : 'warning'
                  }
                />
              </div>
              <div>
                <span>Order</span>
                <strong>
                  {selectedRequest.order?.order_number || `#${selectedRequest.order_id}`}
                </strong>
              </div>
              <div>
                <span>Submitted</span>
                <strong>{formatDateTime(selectedRequest.created_at)}</strong>
              </div>
              <div>
                <span>Requested Quantity</span>
                <strong>{selectedRequest.quantity} item(s)</strong>
              </div>
            </div>

            <section className="customer-return-details__section">
              <h3>Requested Item</h3>
              <div className="customer-return-details__item">
                {selectedRequest.order_item?.image_url && (
                  <img
                    src={selectedRequest.order_item.image_url}
                    alt={selectedRequest.order_item.product_name || 'Requested item'}
                  />
                )}
                <div>
                  <strong>
                    {selectedRequest.order_item?.product_name ||
                      `Item #${selectedRequest.order_item_id}`}
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
                    {' '}Item subtotal: {formatCurrency(selectedRequest.order_item?.subtotal || 0)}
                  </p>
                </div>
              </div>
            </section>

            {(selectedRequest.request_type || '').toLowerCase() === 'exchange' && (
              <section className="customer-return-details__section">
                <h3>Requested Exchange Variant</h3>
                {selectedRequest.requested_variant ? (
                  <div className="customer-return-details__item">
                    {selectedRequest.requested_variant.image_url && (
                      <img
                        src={selectedRequest.requested_variant.image_url}
                        alt="Requested exchange variant"
                      />
                    )}
                    <div>
                      <strong>
                        {[selectedRequest.requested_variant.size, selectedRequest.requested_variant.color]
                          .filter(Boolean)
                          .join(' · ') || `Variant #${selectedRequest.requested_variant.id}`}
                      </strong>
                      <p>SKU: {selectedRequest.requested_variant.sku || '—'}</p>
                      <p>
                        Price: {formatCurrency(selectedRequest.requested_variant.selling_price)}
                      </p>
                    </div>
                  </div>
                ) : (
                  <p>
                    Variant #{selectedRequest.requested_variant_id || '—'} details are unavailable.
                  </p>
                )}
              </section>
            )}

            <section className="customer-return-details__section">
              <h3>Reason</h3>
              <p>{selectedRequest.reason || 'No reason provided.'}</p>
            </section>

            {selectedRequest.admin_note && (
              <section className="customer-return-details__section">
                <h3>Admin Note</h3>
                <p>{selectedRequest.admin_note}</p>
              </section>
            )}
          </div>
        )}
      </Modal>

      {/* Returns Guide */}
      <section className="account-card returns-guide-card">
        <div className="account-card__header">
          <div className="account-card__title-group">
            <Sparkles size={18} className="account-card__icon" />
            <h2 className="account-card__title">How Returns Work</h2>
          </div>
        </div>

        <div className="account-card__body">
          <div className="returns-steps-grid">
            <div className="returns-step">
              <span className="returns-step-number">01</span>
              <strong className="returns-step-title">Select Delivered Order</strong>
              <p className="returns-step-desc">
                Find the delivered order in My Orders and select "Return / Exchange".
              </p>
            </div>

            <div className="returns-step">
              <span className="returns-step-number">02</span>
              <strong className="returns-step-title">Provide Reason & Sizing</strong>
              <p className="returns-step-desc">
                Choose a reason and add any details that will help us review your request.
              </p>
            </div>

            <div className="returns-step">
              <span className="returns-step-number">03</span>
              <strong className="returns-step-title">Doorstep Courier Pickup</strong>
              <p className="returns-step-desc">
                Our delivery partner will collect the item in its original packaging.
              </p>
            </div>

            <div className="returns-step">
              <span className="returns-step-number">04</span>
              <strong className="returns-step-title">Instant Wallet Credit</strong>
              <p className="returns-step-desc">
                After inspection, approved refunds are credited to your wallet.
              </p>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}
