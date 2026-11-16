import React, { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import {
  Package,
  ArrowRight,
  ExternalLink,
  ChevronLeft,
  ChevronRight,
  Filter,
  AlertCircle,
  XCircle,
} from 'lucide-react';
import {
  Button,
  StatusPill,
  SkeletonLoader,
  Modal,
} from '../../components/common';
import { useToast } from '../../context/ToastContext';
import { getMyOrders, cancelOrder } from '../../api/orderService';
import { formatCurrencyINR, formatDate } from '../../utils/formatters';
import './OrdersPage.css';

const FILTER_TABS = [
  { key: 'all', label: 'All Orders' },
  { key: 'pending', label: 'Pending' },
  { key: 'confirmed', label: 'Confirmed' },
  { key: 'shipped', label: 'In Transit' },
  { key: 'delivered', label: 'Delivered' },
  { key: 'cancelled', label: 'Cancelled' },
];

export function OrdersPage() {
  const { showToast } = useToast();

  const [orders, setOrders] = useState([]);
  const [totalCount, setTotalCount] = useState(0);
  const [totalPages, setTotalPages] = useState(1);
  const [currentPage, setCurrentPage] = useState(1);
  const [activeFilter, setActiveFilter] = useState('all');
  const [isLoading, setIsLoading] = useState(true);

  // Cancellation modal state
  const [cancelCandidate, setCancelCandidate] = useState(null);
  const [isCancelling, setIsCancelling] = useState(false);

  const fetchOrders = useCallback(async (page = 1) => {
    setIsLoading(true);
    try {
      const data = await getMyOrders({ page, limit: 8 });
      const rawOrders = data?.orders || [];
      setOrders(rawOrders);
      setTotalCount(data?.total_count || rawOrders.length);
      setTotalPages(data?.total_pages || 1);
      setCurrentPage(data?.page || page);
    } catch (err) {
      showToast({
        title: 'Orders Unavailable',
        message: err.message || 'Unable to retrieve your order history.',
        type: 'danger',
      });
    } finally {
      setIsLoading(false);
    }
  }, [showToast]);

  useEffect(() => {
    fetchOrders(currentPage);
  }, [fetchOrders, currentPage]);

  async function handleConfirmCancel() {
    if (!cancelCandidate) return;
    setIsCancelling(true);
    try {
      await cancelOrder(cancelCandidate.id);
      showToast({
        title: 'Order Cancelled',
        message: `Order #${cancelCandidate.order_number || cancelCandidate.id} has been cancelled.`,
        type: 'info',
      });
      setCancelCandidate(null);
      await fetchOrders(currentPage);
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

  // Filter orders client-side for immediate responsive UX
  const filteredOrders = orders.filter((order) => {
    if (activeFilter === 'all') return true;
    const status = (order.order_status || '').toLowerCase();
    if (activeFilter === 'shipped') {
      return status === 'shipped' || status === 'in_transit';
    }
    return status === activeFilter;
  });

  return (
    <div className="account-orders-page" data-testid="account-orders-page">
      {/* Top Header */}
      <div className="account-orders-header">
        <div>
          <h1 className="account-orders-title">My Orders</h1>
          <p className="account-orders-subtitle">
            Review your tailored garments, delivery status, and historical invoices.
          </p>
        </div>
      </div>

      {/* Filter Tabs */}
      <div className="orders-filter-tabs" role="tablist" aria-label="Order status filters">
        {FILTER_TABS.map((tab) => (
          <button
            key={tab.key}
            type="button"
            role="tab"
            aria-selected={activeFilter === tab.key}
            className={`orders-filter-tab ${activeFilter === tab.key ? 'is-active' : ''}`}
            onClick={() => setActiveFilter(tab.key)}
            data-testid={`filter-tab-${tab.key}`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Orders List */}
      {isLoading ? (
        <div className="account-orders-loading" data-testid="orders-loading">
          <SkeletonLoader lines={8} />
        </div>
      ) : filteredOrders.length > 0 ? (
        <div className="orders-cards-list" data-testid="orders-list">
          {filteredOrders.map((order) => {
            const orderId = order.id;
            const status = (order.order_status || 'pending').toLowerCase();
            const paymentStatus = (order.payment_status || 'pending').toLowerCase();
            const isCancellable = status === 'pending' || status === 'confirmed';

            const pillVariant =
              status === 'delivered'
                ? 'success'
                : status === 'cancelled'
                ? 'danger'
                : status === 'shipped' || status === 'in_transit'
                ? 'info'
                : 'warning';

            const paymentPillVariant =
              paymentStatus === 'paid' || paymentStatus === 'success'
                ? 'success'
                : paymentStatus === 'failed'
                ? 'danger'
                : 'warning';

            const itemsCount =
              order.items?.length ||
              order.items_count ||
              1;

            return (
              <article
                key={orderId}
                className="order-card"
                data-testid={`order-card-${orderId}`}
              >
                {/* Order Top Ribbon */}
                <div className="order-card__header">
                  <div className="order-card__meta-group">
                    <span className="order-card__number" data-testid="order-number">
                      #{order.order_number || orderId}
                    </span>
                    <span className="order-card__date">
                      Placed on {formatDate(order.placed_at || order.created_at)}
                    </span>
                  </div>

                  <div className="order-card__status-group">
                    <StatusPill status={status} variant={pillVariant} />
                    <span className="order-card__payment-pill">
                      Payment: <StatusPill status={paymentStatus} variant={paymentPillVariant} />
                    </span>
                  </div>
                </div>

                {/* Order Items Preview */}
                <div className="order-card__body">
                  <div className="order-card__items-preview">
                    {Array.isArray(order.items) && order.items.length > 0 ? (
                      order.items.slice(0, 3).map((item, idx) => (
                        <div key={item.id || idx} className="order-preview-item">
                          {item.image_url ? (
                            <img
                              src={item.image_url}
                              alt={item.product_name || 'Garment'}
                              className="order-preview-thumb"
                            />
                          ) : (
                            <div className="order-preview-thumb-placeholder">
                              <Package size={16} />
                            </div>
                          )}
                          <div className="order-preview-details">
                            <span className="order-preview-name">{item.product_name}</span>
                            <span className="order-preview-variant">
                              {item.size && `Size: ${item.size}`}
                              {item.size && item.color && ' • '}
                              {item.color && `Color: ${item.color}`}
                              {' • '}Qty: {item.quantity}
                            </span>
                          </div>
                        </div>
                      ))
                    ) : (
                      <p className="order-card__summary-text">
                        {itemsCount} {itemsCount === 1 ? 'Garment' : 'Garments'} in collection
                      </p>
                    )}
                  </div>

                  <div className="order-card__financials">
                    <span className="order-card__total-label">Total Amount</span>
                    <span className="order-card__total-val" data-testid="order-total">
                      {formatCurrencyINR(order.total_amount)}
                    </span>
                  </div>
                </div>

                {/* Order Card Footer */}
                <div className="order-card__footer">
                  <div className="order-card__footer-left">
                    {isCancellable && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setCancelCandidate(order)}
                        data-testid={`cancel-order-btn-${orderId}`}
                      >
                        Cancel Order
                      </Button>
                    )}
                  </div>

                  <Link to={`/account/orders/${orderId}`}>
                    <Button
                      variant="outline"
                      size="sm"
                      rightIcon={<ArrowRight size={14} />}
                      data-testid={`view-order-details-${orderId}`}
                    >
                      VIEW ORDER DETAILS
                    </Button>
                  </Link>
                </div>
              </article>
            );
          })}
        </div>
      ) : (
        <div className="account-empty-block" data-testid="orders-empty-state">
          <Package size={36} className="account-empty-icon" />
          <h2 className="account-empty-title">
            {activeFilter === 'all'
              ? 'No Orders Placed Yet'
              : `No ${activeFilter.toUpperCase()} Orders Found`}
          </h2>
          <p className="account-empty-text">
            Explore our curated collections and commission your next tailored essential.
          </p>
          <Link to="/catalog">
            <Button variant="primary" size="sm">
              EXPLORE ATELIER COLLECTION
            </Button>
          </Link>
        </div>
      )}

      {/* Pagination Controls */}
      {totalPages > 1 && (
        <div className="orders-pagination" data-testid="orders-pagination">
          <Button
            variant="outline"
            size="sm"
            disabled={currentPage <= 1}
            onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
            leftIcon={<ChevronLeft size={14} />}
          >
            PREVIOUS
          </Button>
          <span className="orders-pagination-info">
            Page {currentPage} of {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={currentPage >= totalPages}
            onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
            rightIcon={<ChevronRight size={14} />}
          >
            NEXT
          </Button>
        </div>
      )}

      {/* Order Cancellation Modal */}
      <Modal
        isOpen={Boolean(cancelCandidate)}
        onClose={() => setCancelCandidate(null)}
        title="Cancel Order"
        eyebrow="ORDER MANAGEMENT"
        size="sm"
      >
        <div className="order-cancel-modal" data-testid="order-cancel-modal">
          <p className="order-cancel-desc">
            Are you sure you wish to cancel order{' '}
            <strong>#{cancelCandidate?.order_number || cancelCandidate?.id}</strong>?
          </p>
          <p className="order-cancel-note">
            Reserved inventory will be released immediately. If you made an online payment, refund
            processing will initiate to your Atelier Wallet or original method.
          </p>
          <div className="order-cancel-actions">
            <Button
              variant="outline"
              size="md"
              onClick={() => setCancelCandidate(null)}
              disabled={isCancelling}
            >
              KEEP ORDER
            </Button>
            <Button
              variant="primary"
              size="md"
              onClick={handleConfirmCancel}
              isLoading={isCancelling}
              data-testid="confirm-cancel-order-btn"
            >
              CONFIRM CANCELLATION
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
