import { useEffect, useState, useCallback } from 'react';
import {
  Star,
  Eye,
  EyeOff,
  RefreshCw,
  AlertCircle,
  MessageSquare,
  ShieldCheck,
} from 'lucide-react';
import {
  listAdminReviews,
  setReviewVisibility,
} from '../../../api/adminOperationsService';
import { Badge } from '../../../components/common/Badge';
import { Button } from '../../../components/common/Button';
import { DataTable } from '../../../components/common/DataTable';
import { useToast } from '../../../context/ToastContext';
import { formatDate } from '../../../utils/formatters';
import './AdminReviewsPage.css';

export function AdminReviewsPage() {
  const { showToast } = useToast();

  const [reviews, setReviews] = useState([]);
  const [isLoading, setIsLoading] = useState(true);

  // Pagination
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [totalPages, setTotalPages] = useState(1);
  const [totalItems, setTotalItems] = useState(0);

  // Action state
  const [updatingId, setUpdatingId] = useState(null);

  const fetchReviews = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await listAdminReviews({ page, limit });
      setReviews(data?.reviews ?? []);
      setTotalItems(data?.total ?? 0);
      setTotalPages(data?.total_pages ?? (Math.ceil((data?.total ?? 0) / limit) || 1));
    } catch (err) {
      console.error('Failed to load reviews:', err);
      showToast(err?.response?.data?.message || 'Failed to load reviews', 'error');
    } finally {
      setIsLoading(false);
    }
  }, [page, limit, showToast]);

  useEffect(() => {
    fetchReviews();
  }, [fetchReviews]);

  const handleToggleVisibility = async (review) => {
    const nextVisibility = !review.is_visible;
    setUpdatingId(review.id);
    try {
      await setReviewVisibility(review.id, nextVisibility);
      showToast(
        `Review #${review.id} visibility is now ${nextVisibility ? 'Visible' : 'Hidden'}`,
        'success'
      );
      await fetchReviews();
    } catch (err) {
      console.error('Failed to update visibility:', err);
      showToast(err?.response?.data?.message || 'Failed to update review visibility', 'error');
    } finally {
      setUpdatingId(null);
    }
  };

  const renderStars = (rating) => {
    const stars = [];
    const num = Math.max(1, Math.min(5, Math.round(rating || 5)));
    for (let i = 1; i <= 5; i++) {
      stars.push(
        <Star
          key={i}
          size={14}
          className={i <= num ? 'drape-star filled' : 'drape-star empty'}
        />
      );
    }
    return <div className="drape-stars-row">{stars}</div>;
  };

  const columns = [
    {
      key: 'product',
      header: 'Product',
      render: (val, record) => {
        const rev = (typeof val === 'object' && val !== null) ? val : (record || {});
        return (
          <div className="drape-item-info">
            {rev.product?.thumbnail && (
              <img
                src={rev.product.thumbnail}
                alt={rev.product?.name || 'Product'}
                className="admin-item-thumb"
              />
            )}
            <div>
              <span className="admin-item-name">
                {rev.product?.name || `Product #${rev.product_id}`}
              </span>
              <span className="admin-review-date">{formatDate(rev.created_at)}</span>
            </div>
          </div>
        );
      },
    },
    {
      key: 'customer',
      header: 'Customer',
      render: (val, record) => {
        const rev = (typeof val === 'object' && val !== null) ? val : (record || {});
        return (
          <div className="drape-customer-cell">
            <span className="font-semibold">{rev.user?.name || 'Customer'}</span>
            <span className="drape-customer-email">{rev.user?.email || '—'}</span>
          </div>
        );
      },
    },
    {
      key: 'rating',
      header: 'Rating',
      align: 'center',
      render: (val, record) => {
        const rev = (typeof val === 'object' && val !== null) ? val : (record || {});
        const rating = typeof val === 'number' ? val : (rev.rating || 0);
        return (
          <div className="drape-rating-cell">
            {renderStars(rating)}
            <span className="drape-rating-number">{rating} / 5</span>
          </div>
        );
      },
    },
    {
      key: 'comment',
      header: 'Review Content',
      render: (val, record) => {
        const rev = (typeof val === 'object' && val !== null) ? val : (record || {});
        return (
          <div className="drape-comment-cell">
            {rev.title && <span className="drape-comment-title">{rev.title}</span>}
            <p className="drape-comment-body">{rev.comment || val || 'No text provided.'}</p>
          </div>
        );
      },
    },
    {
      key: 'status',
      header: 'Status',
      align: 'center',
      render: (val, record) => {
        const rev = (typeof val === 'object' && val !== null) ? val : (record || {});
        const isVisible = rev.is_visible !== undefined ? rev.is_visible : val;
        return (
          <Badge variant={isVisible ? 'success' : 'neutral'}>
            {isVisible ? 'Visible' : 'Hidden'}
          </Badge>
        );
      },
    },
    {
      key: 'actions',
      header: 'Moderation',
      align: 'right',
      render: (val, record) => {
        const rev = (typeof val === 'object' && val !== null) ? val : (record || {});
        return (
          <div className="drape-table-actions">
            <Button
              variant={rev.is_visible ? 'outline' : 'primary'}
              size="sm"
              icon={rev.is_visible ? <EyeOff size={13} /> : <Eye size={13} />}
              onClick={() => handleToggleVisibility(rev)}
              disabled={updatingId === rev.id}
            >
              {updatingId === rev.id
                ? 'Saving...'
                : rev.is_visible
                ? 'Hide Review'
                : 'Restore'}
            </Button>
          </div>
        );
      },
    },
  ];

  return (
    <div className="admin-reviews-page">
      {/* Header */}
      <div className="admin-page-header">
        <div>
          <span className="admin-page-eyebrow">Trust & Moderation</span>
          <h1 className="admin-page-title">Reviews & Feedback</h1>
          <p className="admin-page-description">
            Audit customer product ratings, verify authenticity, and control storefront visibility.
          </p>
        </div>
        <div className="admin-header-actions">
          <Button
            variant="outline"
            size="sm"
            icon={<RefreshCw size={14} className={isLoading ? 'spin-icon' : ''} />}
            onClick={fetchReviews}
            disabled={isLoading}
          >
            Refresh
          </Button>
        </div>
      </div>

      {/* Domain Rule Moderation Notice */}
      <div className="admin-moderation-notice">
        <ShieldCheck size={18} className="admin-notice-icon" />
        <div className="admin-notice-content">
          <h4>Review Moderation Policy</h4>
          <p>
            Per domain guidelines, administrators can moderate inappropriate content by toggling
            visibility (hiding from the storefront). Deletion of customer feedback is reserved
            strictly for customer authors to preserve review integrity.
          </p>
        </div>
      </div>

      {/* Table */}
      <DataTable
        columns={columns}
        data={reviews}
        keyField="id"
        isLoading={isLoading}
        emptyMessage="No customer product reviews found."
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
