import React, { useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import {
  Package,
  ArrowLeft,
  Download,
  Truck,
  MapPin,
  CreditCard,
  RotateCcw,
  AlertCircle,
  CheckCircle2,
  Clock,
  ChevronRight,
} from 'lucide-react';
import {
  Button,
  StatusPill,
  Badge,
  SkeletonLoader,
  Modal,
  FormField,
  Input,
} from '../../components/common';
import { useToast } from '../../context/ToastContext';
import {
  getMyOrderById,
  cancelOrder,
  downloadInvoice,
  getCustomerShipment,
  getCustomerTrackingHistory,
} from '../../api/orderService';
import { createReturnRequest } from '../../api/returnService';
import { getStoreProductById } from '../../api/productService';
import { formatCurrencyINR, formatDate } from '../../utils/formatters';
import './OrderDetailsPage.css';

export function OrderDetailsPage() {
  const { orderId } = useParams();
  const { showToast } = useToast();

  const [order, setOrder] = useState(null);
  const [shipment, setShipment] = useState(null);
  const [trackingHistory, setTrackingHistory] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isDownloadingInvoice, setIsDownloadingInvoice] = useState(false);

  // Cancellation State
  const [isCancelModalOpen, setIsCancelModalOpen] = useState(false);
  const [isCancelling, setIsCancelling] = useState(false);

  // Return / Exchange Modal State
  const [returnItem, setReturnItem] = useState(null);
  const [requestType, setRequestType] = useState('return'); // 'return' | 'exchange'
  const [exchangeVariants, setExchangeVariants] = useState([]);
  const [requestedVariantID, setRequestedVariantID] = useState('');
  const [isLoadingExchangeVariants, setIsLoadingExchangeVariants] = useState(false);
  const [returnQuantity, setReturnQuantity] = useState(1);
  const [returnReason, setReturnReason] = useState('');
  const [returnError, setReturnError] = useState('');
  const [isSubmittingReturn, setIsSubmittingReturn] = useState(false);
  const availableExchangeVariants = exchangeVariants.filter(
    (variant) => Number(variant.stock) >= returnQuantity,
  );

  const loadOrderData = useCallback(async () => {
    if (!orderId) return;
    setIsLoading(true);
    try {
      const orderData = await getMyOrderById(orderId);
      setOrder(orderData);

      // Attempt shipment & tracking history parallel fetch
      try {
        const [shipmentData, historyData] = await Promise.allSettled([
          getCustomerShipment(orderId),
          getCustomerTrackingHistory(orderId),
        ]);

        if (shipmentData.status === 'fulfilled' && shipmentData.value) {
          setShipment(shipmentData.value);
        }
        if (historyData.status === 'fulfilled' && historyData.value) {
          setTrackingHistory(historyData.value.history || []);
        }
      } catch {
        // Shipment data is optional / may not be dispatched yet
      }
    } catch (err) {
      showToast({
        title: 'Order Not Found',
        message: err.message || 'Unable to retrieve order details.',
        type: 'danger',
      });
    } finally {
      setIsLoading(false);
    }
  }, [orderId, showToast]);

  useEffect(() => {
    loadOrderData();
  }, [loadOrderData]);

  useEffect(() => {
    if (requestType !== 'exchange' || !returnItem) {
      setExchangeVariants([]);
      setRequestedVariantID('');
      setIsLoadingExchangeVariants(false);
      setReturnError('');
      return undefined;
    }

    let isCurrent = true;
    setExchangeVariants([]);
    setRequestedVariantID('');
    setReturnError('');

    if (!returnItem.product_id) {
      setReturnError('Exchange options are unavailable for this order item. Please refresh the order and try again.');
      return undefined;
    }

    setIsLoadingExchangeVariants(true);
    getStoreProductById(returnItem.product_id)
      .then((product) => {
        if (!isCurrent) return;
        const variants = (product?.variants || []).filter(
          (variant) =>
            Number(variant.id) !== Number(returnItem.variant_id) &&
            variant.in_stock &&
            Number(variant.stock) > 0,
        );
        setExchangeVariants(variants);
        if (variants.length === 0) {
          setReturnError('No alternative in-stock variants are available for exchange.');
        }
      })
      .catch((err) => {
        if (!isCurrent) return;
        console.error('Failed to load exchange variants:', err);
        setReturnError(err?.response?.data?.message || 'Failed to load exchange options.');
      })
      .finally(() => {
        if (isCurrent) setIsLoadingExchangeVariants(false);
      });

    return () => {
      isCurrent = false;
    };
  }, [requestType, returnItem?.product_id, returnItem?.variant_id]);

  // Invoice Download Handler
  async function handleDownloadInvoice() {
    setIsDownloadingInvoice(true);
    try {
      const blob = await downloadInvoice(orderId);
      const url = window.URL.createObjectURL(new Blob([blob], { type: 'application/pdf' }));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `drape-invoice-${order?.order_number || orderId}.pdf`);
      document.body.appendChild(link);
      link.click();
      link.parentNode.removeChild(link);
      window.URL.revokeObjectURL(url);

      showToast({
        title: 'Invoice Downloaded',
        message: 'Your official PDF tax invoice has been saved.',
        type: 'success',
      });
    } catch (err) {
      showToast({
        title: 'Download Failed',
        message: err.message || 'Unable to generate invoice at this time.',
        type: 'danger',
      });
    } finally {
      setIsDownloadingInvoice(false);
    }
  }

  // Cancellation Handler
  async function handleCancelOrder() {
    setIsCancelling(true);
    try {
      await cancelOrder(orderId);
      showToast({
        title: 'Order Cancelled',
        message: 'Your order has been cancelled and inventory released.',
        type: 'info',
      });
      setIsCancelModalOpen(false);
      await loadOrderData();
    } catch (err) {
      showToast({
        title: 'Cancellation Failed',
        message: err.message || 'Unable to cancel this order.',
        type: 'danger',
      });
    } finally {
      setIsCancelling(false);
    }
  }

  // Return / Exchange Submission Handler
  function openReturnModal(item) {
    setReturnItem(item);
    setRequestType('return');
    setExchangeVariants([]);
    setRequestedVariantID('');
    setReturnQuantity(1);
    setReturnReason('');
    setReturnError('');
  }

  async function handleReturnSubmit(e) {
    e.preventDefault();
    if (!returnReason.trim() || returnReason.trim().length < 5) {
      setReturnError('Please enter a detailed reason (minimum 5 characters).');
      return;
    }
    if (requestType === 'exchange') {
      const selectedVariant = exchangeVariants.find(
        (variant) => String(variant.id) === String(requestedVariantID),
      );
      if (!selectedVariant || Number(selectedVariant.stock) < returnQuantity) {
        setReturnError('Please select an available exchange variant with enough stock.');
        return;
      }
    }
    setReturnError('');
    setIsSubmittingReturn(true);

    try {
      await createReturnRequest({
        order_id: orderId,
        order_item_id: returnItem.id,
        request_type: requestType,
        requested_variant_id: requestType === 'exchange' ? requestedVariantID : undefined,
        quantity: returnQuantity,
        reason: returnReason.trim(),
      });

      showToast({
        title: 'Request Submitted',
        message: `Your ${requestType} request has been submitted. Approved refunds are credited to your wallet.`,
        type: 'success',
      });
      setReturnItem(null);
    } catch (err) {
      setReturnError(err.message || 'Failed to submit return request.');
    } finally {
      setIsSubmittingReturn(false);
    }
  }

  if (isLoading) {
    return (
      <div className="account-order-details-loading" data-testid="order-details-loading">
        <SkeletonLoader lines={10} />
      </div>
    );
  }

  if (!order) {
    return (
      <div className="account-empty-block">
        <Package size={36} className="account-empty-icon" />
        <h2 className="account-empty-title">Order Not Found</h2>
        <p className="account-empty-text">The requested order could not be located in your history.</p>
        <Link to="/account/orders">
          <Button variant="outline" size="sm" leftIcon={<ArrowLeft size={14} />}>
            BACK TO ORDERS
          </Button>
        </Link>
      </div>
    );
  }

  const status = (order.order_status || 'pending').toLowerCase();
  const paymentStatus = (order.payment_status || 'pending').toLowerCase();
  const isCancellable = status === 'pending' || status === 'confirmed';
  const isDelivered = status === 'delivered';

  const pillVariant =
    status === 'delivered'
      ? 'success'
      : status === 'cancelled'
      ? 'danger'
      : status === 'shipped' || status === 'in_transit'
      ? 'info'
      : 'warning';

  return (
    <div className="account-order-details-page" data-testid="account-order-details-page">
      {/* Top Nav Back Link & Actions */}
      <div className="order-details-top-bar">
        <Link to="/account/orders" className="order-details-back-link">
          <ArrowLeft size={16} />
          <span>Back to All Orders</span>
        </Link>

        <div className="order-details-top-actions">
          <Button
            variant="outline"
            size="sm"
            leftIcon={<Download size={14} />}
            onClick={handleDownloadInvoice}
            isLoading={isDownloadingInvoice}
            data-testid="download-invoice-btn"
          >
            DOWNLOAD INVOICE
          </Button>

          {isCancellable && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setIsCancelModalOpen(true)}
              data-testid="cancel-order-top-btn"
            >
              Cancel Order
            </Button>
          )}
        </div>
      </div>

      {/* Main Order Header Card */}
      <section className="account-card order-details-summary-card">
        <div className="order-details-summary-grid">
          <div className="order-summary-item">
            <span className="order-summary-label">ORDER NUMBER</span>
            <strong className="order-summary-value" data-testid="detail-order-number">
              #{order.order_number || order.id}
            </strong>
          </div>

          <div className="order-summary-item">
            <span className="order-summary-label">DATE PLACED</span>
            <span className="order-summary-value">
              {formatDate(order.placed_at || order.created_at)}
            </span>
          </div>

          <div className="order-summary-item">
            <span className="order-summary-label">ORDER STATUS</span>
            <div>
              <StatusPill status={status} variant={pillVariant} />
            </div>
          </div>

          <div className="order-summary-item">
            <span className="order-summary-label">PAYMENT</span>
            <div className="order-summary-payment-wrap">
              <StatusPill status={paymentStatus} />
              <span className="order-summary-method">
                via {(order.payment_method || 'Online').toUpperCase()}
              </span>
            </div>
          </div>
        </div>
      </section>

      {/* Shipment Tracking */}
      <section className="account-card" aria-labelledby="tracking-heading">
        <div className="account-card__header">
          <div className="account-card__title-group">
            <Truck size={18} className="account-card__icon" aria-hidden="true" />
            <h2 id="tracking-heading" className="account-card__title">
              Shipment Tracking
            </h2>
          </div>
          {shipment?.tracking_number && (
            <span className="order-tracking-badge" data-testid="tracking-number">
              Tracking: <strong>{shipment.tracking_number}</strong> ({shipment.courier_name})
            </span>
          )}
        </div>

        <div className="account-card__body">
          {trackingHistory.length > 0 ? (
            <div className="tracking-timeline" data-testid="tracking-timeline">
              {trackingHistory.map((step, idx) => (
                <div key={step.id || idx} className="tracking-timeline-item">
                  <div className="tracking-timeline-dot">
                    <CheckCircle2 size={16} />
                  </div>
                  <div className="tracking-timeline-content">
                    <div className="tracking-timeline-header">
                      <strong className="tracking-timeline-status">
                        {step.shipping_status?.replace(/_/g, ' ').toUpperCase()}
                      </strong>
                      <span className="tracking-timeline-date">{formatDate(step.created_at)}</span>
                    </div>
                    <p className="tracking-timeline-desc">{step.description}</p>
                    {step.location && (
                      <span className="tracking-timeline-location">
                        <MapPin size={12} /> {step.location}
                      </span>
                    )}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="tracking-pending-notice" data-testid="tracking-pending-notice">
              <Clock size={20} className="tracking-pending-icon" />
              <div>
                <strong>
                  {status === 'delivered'
                    ? 'Delivered to Destination'
                    : status === 'cancelled'
                    ? 'Order Cancelled'
                    : 'Crafting & Quality Assurance'}
                </strong>
                <p className="tracking-pending-text">
                  {status === 'delivered'
                    ? 'Your order was delivered successfully.'
                    : status === 'cancelled'
                    ? 'This order has been cancelled and will not be dispatched.'
                    : 'Your order is being prepared. Tracking updates will appear here after it ships.'}
                </p>
              </div>
            </div>
          )}
        </div>
      </section>

      {/* Order Items */}
      <section className="account-card" aria-labelledby="items-heading">
        <div className="account-card__header">
          <div className="account-card__title-group">
            <Package size={18} className="account-card__icon" aria-hidden="true" />
            <h2 id="items-heading" className="account-card__title">
              Order Items ({order.items?.length || 0})
            </h2>
          </div>
        </div>

        <div className="account-card__body">
          <div className="order-items-table">
            {order.items?.map((item) => (
              <div key={item.id} className="order-item-row" data-testid={`order-item-${item.id}`}>
                <div className="order-item-media">
                  {item.image_url ? (
                    <img src={item.image_url} alt={item.product_name} className="order-item-img" />
                  ) : (
                    <div className="order-item-placeholder">
                      <Package size={20} />
                    </div>
                  )}
                </div>

                <div className="order-item-details">
                  <h3 className="order-item-name">{item.product_name}</h3>
                  <div className="order-item-specs">
                    {item.size && <span>Size: {item.size}</span>}
                    {item.size && item.color && <span>•</span>}
                    {item.color && <span>Color: {item.color}</span>}
                    <span>•</span>
                    <span>Qty: {item.quantity}</span>
                  </div>
                </div>

                <div className="order-item-pricing">
                  <span className="order-item-unit-price">
                    {formatCurrencyINR(item.unit_price)} each
                  </span>
                  <strong className="order-item-total-price">
                    {formatCurrencyINR(item.total_price || item.unit_price * item.quantity)}
                  </strong>
                </div>

                {isDelivered && (
                  <div className="order-item-return-col">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => openReturnModal(item)}
                      leftIcon={<RotateCcw size={13} />}
                      data-testid={`return-item-btn-${item.id}`}
                    >
                      Return / Exchange
                    </Button>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* 2-Column: Delivery Address & Financial Summary */}
      <div className="order-details-bottom-grid">
        {/* Shipping Address */}
        <section className="account-card">
          <div className="account-card__header">
            <div className="account-card__title-group">
              <MapPin size={18} className="account-card__icon" />
              <h2 className="account-card__title">Delivery Destination</h2>
            </div>
          </div>
          <div className="account-card__body">
            {order.address ? (
              <div className="order-address-box">
                <strong className="order-address-name">{order.address.name}</strong>
                <p className="order-address-lines">
                  {order.address.address_line1}
                  {order.address.address_line2 ? `, ${order.address.address_line2}` : ''}
                  <br />
                  {order.address.city}, {order.address.state} {order.address.postal_code}
                </p>
                <p className="order-address-phone">Tel: {order.address.phone}</p>
              </div>
            ) : (
              <p className="account-card__desc">No address details available for this order.</p>
            )}
          </div>
        </section>

        {/* Financial Summary */}
        <section className="account-card">
          <div className="account-card__header">
            <div className="account-card__title-group">
              <CreditCard size={18} className="account-card__icon" />
              <h2 className="account-card__title">Payment Summary</h2>
            </div>
          </div>
          <div className="account-card__body">
            <div className="order-financial-rows">
              <div className="order-financial-row">
                <span>Subtotal</span>
                <span>{formatCurrencyINR(order.subtotal)}</span>
              </div>
              {order.discount_amount > 0 && (
                <div className="order-financial-row is-discount">
                  <span>Discount</span>
                  <span>- {formatCurrencyINR(order.discount_amount)}</span>
                </div>
              )}
              <div className="order-financial-row">
                <span>Shipping</span>
                <span>{order.shipping_amount === 0 ? 'Complimentary' : formatCurrencyINR(order.shipping_amount)}</span>
              </div>
              <div className="order-financial-divider" />
              <div className="order-financial-row is-total">
                <strong>Total Amount</strong>
                <strong className="order-final-amount" data-testid="order-grand-total">
                  {formatCurrencyINR(order.total_amount)}
                </strong>
              </div>
            </div>
          </div>
        </section>
      </div>

      {/* Cancellation Modal */}
      <Modal
        isOpen={isCancelModalOpen}
        onClose={() => setIsCancelModalOpen(false)}
        title="Cancel Order"
        eyebrow="CONFIRMATION"
        size="sm"
      >
        <div className="order-cancel-modal">
          <p className="order-cancel-desc">
            Are you sure you wish to cancel order #{order.order_number || order.id}?
          </p>
          <p className="order-cancel-note">
            Inventory reserved for your order will be released immediately. If you have already
            completed payment, your refund will be processed to your wallet or original
            method upon cancellation.
          </p>
          <div className="order-cancel-actions">
            <Button
              variant="outline"
              size="md"
              onClick={() => setIsCancelModalOpen(false)}
              disabled={isCancelling}
            >
              KEEP ORDER
            </Button>
            <Button
              variant="primary"
              size="md"
              onClick={handleCancelOrder}
              isLoading={isCancelling}
              data-testid="confirm-cancel-details-btn"
            >
              CONFIRM CANCELLATION
            </Button>
          </div>
        </div>
      </Modal>

      {/* Return / Exchange Submission Modal */}
      <Modal
        isOpen={Boolean(returnItem)}
        onClose={() => setReturnItem(null)}
        title={requestType === 'return' ? 'Request Return' : 'Request Exchange'}
        eyebrow="RETURN REQUEST"
        size="md"
      >
        {returnItem && (
          <form onSubmit={handleReturnSubmit} className="return-modal-form" noValidate>
            <div className="return-item-preview">
              <strong className="return-item-preview-name">{returnItem.product_name}</strong>
              <span className="return-item-preview-meta">
                Purchased Qty: {returnItem.quantity} • Unit Price: {formatCurrencyINR(returnItem.unit_price)}
              </span>
            </div>

            {returnError && (
              <div className="return-modal-error" role="alert">
                <AlertCircle size={14} />
                <span>{returnError}</span>
              </div>
            )}

            <div className="return-type-selector">
              <label className="return-type-option">
                <input
                  type="radio"
                  name="requestType"
                  value="return"
                  checked={requestType === 'return'}
                  onChange={() => setRequestType('return')}
                  data-testid="return-radio-return"
                />
                <span>Return Item (Refund to Wallet)</span>
              </label>

              <label className="return-type-option">
                <input
                  type="radio"
                  name="requestType"
                  value="exchange"
                  checked={requestType === 'exchange'}
                  onChange={() => setRequestType('exchange')}
                  data-testid="return-radio-exchange"
                />
                <span>Exchange Sizing / Variant</span>
              </label>
            </div>

            {requestType === 'exchange' && (
              <FormField label="EXCHANGE TO" htmlFor="requested-variant" required>
                {isLoadingExchangeVariants ? (
                  <p className="return-variant-status">Loading available variants...</p>
                ) : availableExchangeVariants.length === 0 ? (
                  <p className="return-variant-status">
                    No available variants have enough stock for this quantity.
                  </p>
                ) : (
                  <select
                    id="requested-variant"
                    className="return-variant-select"
                    value={
                      availableExchangeVariants.some(
                        (variant) => String(variant.id) === String(requestedVariantID),
                      )
                        ? requestedVariantID
                        : ''
                    }
                    onChange={(e) => setRequestedVariantID(e.target.value)}
                    required
                    data-testid="exchange-variant-select"
                  >
                    <option value="">Select an available variant...</option>
                    {availableExchangeVariants.map((variant) => {
                      const details = [variant.size, variant.color, variant.sku]
                        .filter(Boolean)
                        .join(' · ');
                      return (
                        <option key={variant.id} value={variant.id}>
                          {details || 'Standard variant'} — {formatCurrencyINR(variant.selling_price)} ({variant.stock} available)
                        </option>
                      );
                    })}
                  </select>
                )}
              </FormField>
            )}

            <FormField label="RETURN QUANTITY" htmlFor="return-quantity" required>
              <Input
                id="return-quantity"
                type="number"
                min={1}
                max={returnItem.quantity}
                value={returnQuantity}
                onChange={(e) => setReturnQuantity(Math.max(1, Number(e.target.value)))}
                data-testid="return-quantity-input"
              />
            </FormField>

            <FormField label="REASON FOR RETURN / EXCHANGE" htmlFor="return-reason" required>
              <textarea
                id="return-reason"
                className="return-reason-textarea"
                rows={4}
                value={returnReason}
                onChange={(e) => setReturnReason(e.target.value)}
                placeholder="Tell us why you want to return or exchange this item."
                data-testid="return-reason-textarea"
              />
            </FormField>

            <div className="return-policy-notice">
              <p>
                <strong>Refund Policy:</strong> Once we receive and inspect your return, approved
                refunds are credited to your wallet.
              </p>
            </div>

            <div className="order-cancel-actions">
              <Button
                type="button"
                variant="outline"
                size="md"
                onClick={() => setReturnItem(null)}
              >
                CANCEL
              </Button>
              <Button
                type="submit"
                variant="primary"
                size="md"
                isLoading={isSubmittingReturn}
                data-testid="submit-return-request-btn"
              >
                SUBMIT REQUEST
              </Button>
            </div>
          </form>
        )}
      </Modal>
    </div>
  );
}
