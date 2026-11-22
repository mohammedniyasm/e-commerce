import { useEffect, useState } from 'react';
import { Link, useLocation, useSearchParams } from 'react-router-dom';
import {
  ArrowRight,
  CheckCircle2,
  Clock,
  CreditCard,
  FileText,
  MapPin,
  Package,
  ShoppingBag,
  Truck,
} from 'lucide-react';
import { getMyOrderById } from '../../api/checkoutService';
import { formatCurrencyINR } from '../../utils/formatters';
import { Badge } from '../../components/common/Badge';
import { Button } from '../../components/common/Button';
import { SkeletonLoader } from '../../components/common/SkeletonLoader';
import './OrderConfirmationPage.css';

function formatDeliveryDates(startDateStr, endDateStr) {
  if (!startDateStr && !endDateStr) {
    return 'Standard delivery (confirmed at dispatch)';
  }

  const options = { day: 'numeric', month: 'short' };
  try {
    if (startDateStr && endDateStr) {
      const start = new Date(startDateStr).toLocaleDateString('en-IN', options);
      const end = new Date(endDateStr).toLocaleDateString('en-IN', { ...options, year: 'numeric' });
      return `${start} – ${end}`;
    }
    if (startDateStr) {
      return `From ${new Date(startDateStr).toLocaleDateString('en-IN', { ...options, year: 'numeric' })}`;
    }
    if (endDateStr) {
      return `By ${new Date(endDateStr).toLocaleDateString('en-IN', { ...options, year: 'numeric' })}`;
    }
  } catch {
    return 'Standard delivery (confirmed at dispatch)';
  }
  return 'Standard delivery (confirmed at dispatch)';
}

export function OrderConfirmationPage() {
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const queryOrderId = searchParams.get('order_id');

  const [order, setOrder] = useState(location.state?.order || null);
  const [isLoading, setIsLoading] = useState(!location.state?.order && Boolean(queryOrderId));

  useEffect(() => {
    let isMounted = true;

    async function fetchOrder() {
      if (!order && queryOrderId) {
        setIsLoading(true);
        try {
          const fetched = await getMyOrderById(queryOrderId);
          if (isMounted && fetched) {
            setOrder(fetched);
          }
        } catch {
          // Keep whatever is available
        } finally {
          if (isMounted) setIsLoading(false);
        }
      }
    }

    fetchOrder();

    return () => {
      isMounted = false;
    };
  }, [queryOrderId, order]);

  if (isLoading) {
    return (
      <div className="order-confirm-page order-confirm-loading">
        <SkeletonLoader lines={6} />
      </div>
    );
  }

  const orderNum = order?.order_number || order?.id || 'DRP-CONFIRMED';
  const orderStatus = order?.order_status || 'placed';
  // Safe fallback: never default unknown online status to "paid"
  const paymentStatus = order?.payment_status || (order?.payment_method === 'cod' ? 'cod' : 'pending');
  const totalAmount = order?.total_amount || 0;

  const deliveryStart = order?.estimated_delivery_start || order?.estimated_start;
  const deliveryEnd = order?.estimated_delivery_end || order?.estimated_end;
  const estimatedDeliveryText = formatDeliveryDates(deliveryStart, deliveryEnd);

  return (
    <div className="order-confirm-page" data-testid="order-confirmation-page">
      <div className="order-confirm-card">
        {/* Success Emblem */}
        <div className="order-confirm-emblem" aria-hidden="true">
          <CheckCircle2 size={54} />
        </div>

        <h1 className="order-confirm-title">Order Confirmed</h1>
        <p className="order-confirm-subtitle">
          Thank you for your order. We are preparing it for shipment.
        </p>

        {/* Order Reference Badge */}
        <div className="order-confirm-ref-badge">
          <span>ORDER REFERENCE</span>
          <strong data-testid="order-number">{orderNum}</strong>
        </div>

        {/* Details Grid */}
        <div className="order-confirm-details-grid">
          <div className="order-confirm-detail-box">
            <div className="order-confirm-detail-label">
              <Package size={16} />
              <span>Order Status</span>
            </div>
            <div className="order-confirm-detail-value">
              <Badge variant="success" size="md">
                {orderStatus.toUpperCase()}
              </Badge>
            </div>
          </div>

          <div className="order-confirm-detail-box">
            <div className="order-confirm-detail-label">
              <CreditCard size={16} />
              <span>Payment Status</span>
            </div>
            <div className="order-confirm-detail-value">
              <Badge
                variant={
                  paymentStatus === 'paid'
                    ? 'success'
                    : paymentStatus === 'pending'
                    ? 'warning'
                    : 'neutral'
                }
                size="md"
              >
                {paymentStatus.toUpperCase()}
              </Badge>
            </div>
          </div>

          <div className="order-confirm-detail-box">
            <div className="order-confirm-detail-label">
              <Clock size={16} />
              <span>Estimated Delivery</span>
            </div>
            <div className="order-confirm-detail-value">
              <strong data-testid="order-delivery-date">{estimatedDeliveryText}</strong>
            </div>
          </div>

          <div className="order-confirm-detail-box">
            <div className="order-confirm-detail-label">
              <FileText size={16} />
              <span>Total Paid</span>
            </div>
            <div className="order-confirm-detail-value">
              <strong className="order-confirm-total-amount">
                {totalAmount > 0 ? formatCurrencyINR(totalAmount) : 'Paid'}
              </strong>
            </div>
          </div>
        </div>

        {/* Dispatch Notes */}
        <div className="order-confirm-note">
          <Truck size={18} />
          <p>
            An order confirmation and insured shipment tracking link will be sent to your registered email
            address once your package ships.
          </p>
        </div>

        {/* Actions */}
        <div className="order-confirm-actions">
          <Link to="/account">
            <Button variant="outline" size="lg">
              VIEW MY ORDERS
            </Button>
          </Link>
          <Link to="/catalog">
            <Button variant="primary" size="lg" rightIcon={<ArrowRight size={14} />}>
              CONTINUE SHOPPING
            </Button>
          </Link>
        </div>
      </div>
    </div>
  );
}
