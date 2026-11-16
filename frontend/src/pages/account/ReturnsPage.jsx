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
} from 'lucide-react';
import {
  Button,
  StatusPill,
  Badge,
  SkeletonLoader,
} from '../../components/common';
import { useToast } from '../../context/ToastContext';
import { getMyReturnRequests } from '../../api/returnService';
import { formatDate } from '../../utils/formatters';
import './ReturnsPage.css';

export function ReturnsPage() {
  const { showToast } = useToast();
  const [returnRequests, setReturnRequests] = useState([]);
  const [isLoading, setIsLoading] = useState(true);

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
            Track return authorizations, sizing exchanges, and atelier wallet credits.
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
                      <span>{req.quantity} Garment(s)</span>
                    </p>

                    <p className="return-card__field">
                      <span className="return-card__label">REASON:</span>
                      <span className="return-card__reason-text">{req.reason}</span>
                    </p>

                    {req.admin_note && (
                      <div className="return-card__admin-note">
                        <span className="return-card__label">ATELIER CONCIERGE NOTE:</span>
                        <p>{req.admin_note}</p>
                      </div>
                    )}
                  </div>
                </div>

                <div className="return-card__footer">
                  <p className="return-card__footer-policy">
                    {isReturn
                      ? 'Approved returns are refunded directly to your Atelier Wallet.'
                      : 'Replacement variant will be dispatched upon return inspection.'}
                  </p>
                </div>
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

      {/* Atelier Returns Guide Card */}
      <section className="account-card returns-guide-card">
        <div className="account-card__header">
          <div className="account-card__title-group">
            <Sparkles size={18} className="account-card__icon" />
            <h2 className="account-card__title">Atelier Returns & Wallet Refunds</h2>
          </div>
        </div>

        <div className="account-card__body">
          <div className="returns-steps-grid">
            <div className="returns-step">
              <span className="returns-step-number">01</span>
              <strong className="returns-step-title">Select Delivered Order</strong>
              <p className="returns-step-desc">
                Locate your delivered commission in My Orders and tap "Return / Exchange".
              </p>
            </div>

            <div className="returns-step">
              <span className="returns-step-number">02</span>
              <strong className="returns-step-title">Provide Reason & Sizing</strong>
              <p className="returns-step-desc">
                Specify sizing preferences or reasons for return to assist our master tailors.
              </p>
            </div>

            <div className="returns-step">
              <span className="returns-step-number">03</span>
              <strong className="returns-step-title">Doorstep Courier Pickup</strong>
              <p className="returns-step-desc">
                Our logistics partner will collect the garment safely in original packaging.
              </p>
            </div>

            <div className="returns-step">
              <span className="returns-step-number">04</span>
              <strong className="returns-step-title">Instant Wallet Credit</strong>
              <p className="returns-step-desc">
                Upon quality inspection, the approved refund is credited to your Atelier Wallet.
              </p>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}
