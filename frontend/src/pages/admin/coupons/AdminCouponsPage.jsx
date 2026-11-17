import { useEffect, useState, useCallback } from 'react';
import {
  Tag,
  Plus,
  Search,
  Power,
  Trash2,
  Edit2,
  Users,
  Calendar,
  Percent,
  DollarSign,
  RefreshCw,
  AlertCircle,
} from 'lucide-react';
import {
  listAdminCoupons,
  createAdminCoupon,
  updateAdminCoupon,
  deleteAdminCoupon,
  toggleAdminCouponActive,
  listCouponUsages,
} from '../../../api/adminOperationsService';
import { Badge } from '../../../components/common/Badge';
import { Button } from '../../../components/common/Button';
import { DataTable } from '../../../components/common/DataTable';
import { Modal } from '../../../components/common/Modal';
import { useToast } from '../../../context/ToastContext';
import { formatCurrency, formatDate } from '../../../utils/formatters';
import './AdminCouponsPage.css';

export function AdminCouponsPage() {
  const { showToast } = useToast();

  const [coupons, setCoupons] = useState([]);
  const [isLoading, setIsLoading] = useState(true);

  // Filters & Pagination
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [totalPages, setTotalPages] = useState(1);
  const [totalItems, setTotalItems] = useState(0);
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');

  // Modals
  const [isFormModalOpen, setIsFormModalOpen] = useState(false);
  const [editingCoupon, setEditingCoupon] = useState(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Form Fields
  const [code, setCode] = useState('');
  const [discountType, setDiscountType] = useState('percentage'); // 'percentage' | 'fixed'
  const [discountValue, setDiscountValue] = useState('');
  const [minOrderAmount, setMinOrderAmount] = useState('');
  const [maxDiscountAmount, setMaxDiscountAmount] = useState('');
  const [usageLimit, setUsageLimit] = useState('');
  const [userLimit, setUserLimit] = useState('1');
  const [validFrom, setValidFrom] = useState('');
  const [validUntil, setValidUntil] = useState('');
  const [isActive, setIsActive] = useState(true);

  // Delete modal
  const [couponToDelete, setCouponToDelete] = useState(null);
  const [isDeleting, setIsDeleting] = useState(false);

  // Usages modal
  const [viewingUsagesCoupon, setViewingUsagesCoupon] = useState(null);
  const [usagesList, setUsagesList] = useState([]);
  const [isLoadingUsages, setIsLoadingUsages] = useState(false);

  const fetchCoupons = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await listAdminCoupons({
        page,
        limit,
        search,
      });

      setCoupons(data?.coupons ?? []);
      setTotalItems(data?.total ?? 0);
      setTotalPages(data?.total_pages ?? (Math.ceil((data?.total ?? 0) / limit) || 1));
    } catch (err) {
      console.error('Failed to load coupons:', err);
      showToast(err?.response?.data?.message || 'Failed to load coupons', 'error');
    } finally {
      setIsLoading(false);
    }
  }, [page, limit, search, showToast]);

  useEffect(() => {
    fetchCoupons();
  }, [fetchCoupons]);

  const handleSearchSubmit = (e) => {
    e.preventDefault();
    setPage(1);
    setSearch(searchInput.trim());
  };

  const openCreateModal = () => {
    setEditingCoupon(null);
    setCode('');
    setDiscountType('percentage');
    setDiscountValue('');
    setMinOrderAmount('');
    setMaxDiscountAmount('');
    setUsageLimit('');
    setUserLimit('1');
    setValidFrom(new Date().toISOString().split('T')[0]);
    setValidUntil('');
    setIsActive(true);
    setIsFormModalOpen(true);
  };

  const openEditModal = (coupon) => {
    setEditingCoupon(coupon);
    setCode(coupon.code || '');
    setDiscountType(coupon.discount_type || 'percentage');
    setDiscountValue(String(coupon.discount_value || ''));
    setMinOrderAmount(String(coupon.min_order_amount || ''));
    setMaxDiscountAmount(String(coupon.max_discount_amount || ''));
    setUsageLimit(String(coupon.usage_limit || ''));
    setUserLimit(String(coupon.user_limit || '1'));
    setValidFrom(coupon.valid_from ? coupon.valid_from.split('T')[0] : '');
    setValidUntil(coupon.valid_until ? coupon.valid_until.split('T')[0] : '');
    setIsActive(Boolean(coupon.is_active));
    setIsFormModalOpen(true);
  };

  const handleFormSubmit = async (e) => {
    e.preventDefault();
    if (!code.trim() || !discountValue) {
      showToast('Coupon code and discount value are required', 'error');
      return;
    }

    const payload = {
      code: code.trim().toUpperCase(),
      discount_type: discountType,
      discount_value: parseFloat(discountValue),
      min_order_amount: minOrderAmount ? parseFloat(minOrderAmount) : 0,
      max_discount_amount: maxDiscountAmount ? parseFloat(maxDiscountAmount) : 0,
      usage_limit: usageLimit ? parseInt(usageLimit, 10) : 0,
      user_limit: userLimit ? parseInt(userLimit, 10) : 1,
      valid_from: validFrom ? new Date(validFrom).toISOString() : new Date().toISOString(),
      valid_until: validUntil ? new Date(validUntil).toISOString() : undefined,
      is_active: isActive,
    };

    setIsSubmitting(true);
    try {
      if (editingCoupon) {
        await updateAdminCoupon(editingCoupon.id, payload);
        showToast(`Coupon ${payload.code} updated successfully`, 'success');
      } else {
        await createAdminCoupon(payload);
        showToast(`Coupon ${payload.code} created successfully`, 'success');
      }
      setIsFormModalOpen(false);
      await fetchCoupons();
    } catch (err) {
      console.error('Failed to save coupon:', err);
      showToast(err?.response?.data?.message || 'Failed to save coupon', 'error');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleToggleActive = async (coupon) => {
    try {
      await toggleAdminCouponActive(coupon.id);
      showToast(
        `Coupon ${coupon.code} is now ${coupon.is_active ? 'inactive' : 'active'}`,
        'success'
      );
      await fetchCoupons();
    } catch (err) {
      console.error('Failed to toggle coupon status:', err);
      showToast(err?.response?.data?.message || 'Failed to toggle coupon status', 'error');
    }
  };

  const handleDeleteConfirm = async () => {
    if (!couponToDelete) return;
    setIsDeleting(true);
    try {
      await deleteAdminCoupon(couponToDelete.id);
      showToast(`Coupon ${couponToDelete.code} deleted successfully`, 'success');
      setCouponToDelete(null);
      await fetchCoupons();
    } catch (err) {
      console.error('Failed to delete coupon:', err);
      showToast(err?.response?.data?.message || 'Failed to delete coupon', 'error');
    } finally {
      setIsDeleting(false);
    }
  };

  const handleOpenUsages = async (coupon) => {
    setViewingUsagesCoupon(coupon);
    setIsLoadingUsages(true);
    try {
      const data = await listCouponUsages(coupon.id, { page: 1, limit: 50 });
      setUsagesList(data?.usages ?? []);
    } catch (err) {
      console.error('Failed to load usages:', err);
      showToast('Failed to load redemption history', 'error');
    } finally {
      setIsLoadingUsages(false);
    }
  };

  const columns = [
    {
      key: 'code',
      header: 'Coupon Code',
      render: (val, record) => {
        const c = (typeof val === 'object' && val !== null) ? val : (record || {});
        return (
          <div className="drape-coupon-cell">
            <span className="drape-coupon-code">{c.code || val}</span>
            <span className="drape-coupon-sub">
              {c.discount_type === 'percentage' ? `${c.discount_value}% OFF` : `₹${c.discount_value} OFF`}
            </span>
          </div>
        );
      },
    },
    {
      key: 'conditions',
      header: 'Conditions',
      render: (val, record) => {
        const c = (typeof val === 'object' && val !== null) ? val : (record || {});
        return (
          <div className="drape-conditions-cell">
            <span>Min Order: {formatCurrency(c.min_order_amount || 0)}</span>
            {c.max_discount_amount > 0 && (
              <span className="text-muted">Max Cap: {formatCurrency(c.max_discount_amount)}</span>
            )}
          </div>
        );
      },
    },
    {
      key: 'validity',
      header: 'Validity',
      render: (val, record) => {
        const c = (typeof val === 'object' && val !== null) ? val : (record || {});
        return (
          <div className="drape-conditions-cell">
            <span>Until: {c.valid_until ? formatDate(c.valid_until) : 'No Expiry'}</span>
            <span className="text-muted">From: {formatDate(c.valid_from)}</span>
          </div>
        );
      },
    },
    {
      key: 'usage',
      header: 'Usage',
      align: 'center',
      render: (val, record) => {
        const c = (typeof val === 'object' && val !== null) ? val : (record || {});
        return (
          <button
            type="button"
            onClick={() => handleOpenUsages(c)}
            className="admin-link-btn"
            title="View redemption history"
          >
            {c.usage_count || 0} / {c.usage_limit || '∞'}
          </button>
        );
      },
    },
    {
      key: 'status',
      header: 'Status',
      align: 'center',
      render: (val, record) => {
        const c = (typeof val === 'object' && val !== null) ? val : (record || {});
        const isActive = c.is_active !== undefined ? c.is_active : val;
        return (
          <Badge variant={isActive ? 'success' : 'neutral'}>
            {isActive ? 'Active' : 'Inactive'}
          </Badge>
        );
      },
    },
    {
      key: 'actions',
      header: 'Actions',
      align: 'right',
      render: (val, record) => {
        const c = (typeof val === 'object' && val !== null) ? val : (record || {});
        return (
          <div className="drape-table-actions">
            <Button
              variant="ghost"
              size="sm"
              icon={<Power size={13} />}
              onClick={() => handleToggleActive(c)}
              title={c.is_active ? 'Deactivate' : 'Activate'}
            />
            <Button
              variant="ghost"
              size="sm"
              icon={<Edit2 size={13} />}
              onClick={() => openEditModal(c)}
              title="Edit Coupon"
            />
            <Button
              variant="ghost"
              size="sm"
              icon={<Trash2 size={13} />}
              onClick={() => setCouponToDelete(c)}
              title="Delete Coupon"
            />
          </div>
        );
      },
    },
  ];

  return (
    <div className="admin-coupons-page">
      {/* Header */}
      <div className="admin-page-header">
        <div>
          <span className="admin-page-eyebrow">Promotions & Growth</span>
          <h1 className="admin-page-title">Coupons & Campaigns</h1>
          <p className="admin-page-description">
            Create discount vouchers, enforce order limits, and monitor redemption volume.
          </p>
        </div>
        <div className="admin-header-actions">
          <Button
            variant="outline"
            size="sm"
            icon={<RefreshCw size={14} className={isLoading ? 'spin-icon' : ''} />}
            onClick={fetchCoupons}
            disabled={isLoading}
          >
            Refresh
          </Button>
          <Button
            variant="primary"
            size="sm"
            icon={<Plus size={14} />}
            onClick={openCreateModal}
          >
            Create Coupon
          </Button>
        </div>
      </div>

      {/* Toolbar */}
      <div className="admin-toolbar">
        <form onSubmit={handleSearchSubmit} className="admin-search-form">
          <div className="admin-search-input-wrapper">
            <Search size={16} className="admin-search-icon" />
            <input
              type="text"
              placeholder="Search by coupon code..."
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              className="admin-search-input"
            />
          </div>
          <Button type="submit" variant="secondary" size="sm">
            Search
          </Button>
        </form>

        {search && (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => {
              setSearch('');
              setSearchInput('');
              setPage(1);
            }}
          >
            Clear Search
          </Button>
        )}
      </div>

      {/* Table */}
      <DataTable
        columns={columns}
        data={coupons}
        keyField="id"
        isLoading={isLoading}
        emptyMessage="No promotional coupons found."
        pagination={{
          page,
          totalPages,
          totalItems,
          limit,
          onPageChange: (newPage) => setPage(newPage),
        }}
      />

      {/* Create / Edit Modal */}
      <Modal
        isOpen={isFormModalOpen}
        onClose={() => setIsFormModalOpen(false)}
        title={editingCoupon ? `Edit Coupon ${editingCoupon.code}` : 'Create Promotional Coupon'}
      >
        <form onSubmit={handleFormSubmit} className="admin-modal-form">
          <div className="admin-form-row">
            <label className="admin-form-label">
              Coupon Code *
              <input
                type="text"
                value={code}
                onChange={(e) => setCode(e.target.value.toUpperCase())}
                placeholder="e.g. FESTIVE20"
                required
                className="admin-input-full font-mono"
              />
            </label>

            <label className="admin-form-label">
              Discount Type *
              <select
                value={discountType}
                onChange={(e) => setDiscountType(e.target.value)}
                className="admin-select-full"
              >
                <option value="percentage">Percentage (%)</option>
                <option value="fixed">Fixed Amount (₹)</option>
              </select>
            </label>
          </div>

          <div className="admin-form-row">
            <label className="admin-form-label">
              Discount Value *
              <input
                type="number"
                step="0.01"
                min="1"
                value={discountValue}
                onChange={(e) => setDiscountValue(e.target.value)}
                placeholder={discountType === 'percentage' ? '20' : '500'}
                required
                className="admin-input-full"
              />
            </label>

            <label className="admin-form-label">
              Min Order Amount (₹)
              <input
                type="number"
                min="0"
                value={minOrderAmount}
                onChange={(e) => setMinOrderAmount(e.target.value)}
                placeholder="1000"
                className="admin-input-full"
              />
            </label>
          </div>

          <div className="admin-form-row">
            <label className="admin-form-label">
              Max Discount Cap (₹)
              <input
                type="number"
                min="0"
                value={maxDiscountAmount}
                onChange={(e) => setMaxDiscountAmount(e.target.value)}
                placeholder="2000 (0 for no cap)"
                className="admin-input-full"
              />
            </label>

            <label className="admin-form-label">
              Total Usage Limit
              <input
                type="number"
                min="0"
                value={usageLimit}
                onChange={(e) => setUsageLimit(e.target.value)}
                placeholder="100 (0 for unlimited)"
                className="admin-input-full"
              />
            </label>
          </div>

          <div className="admin-form-row">
            <label className="admin-form-label">
              Usage Limit Per User
              <input
                type="number"
                min="1"
                value={userLimit}
                onChange={(e) => setUserLimit(e.target.value)}
                placeholder="1"
                className="admin-input-full"
              />
            </label>

            <label className="admin-form-label">
              Status
              <div className="admin-checkbox-row">
                <input
                  type="checkbox"
                  id="coupon-active"
                  checked={isActive}
                  onChange={(e) => setIsActive(e.target.checked)}
                />
                <label htmlFor="coupon-active">Active and redeemable</label>
              </div>
            </label>
          </div>

          <div className="admin-form-row">
            <label className="admin-form-label">
              Valid From Date
              <input
                type="date"
                value={validFrom}
                onChange={(e) => setValidFrom(e.target.value)}
                className="admin-input-full"
              />
            </label>

            <label className="admin-form-label">
              Valid Until Date
              <input
                type="date"
                value={validUntil}
                onChange={(e) => setValidUntil(e.target.value)}
                className="admin-input-full"
              />
            </label>
          </div>

          <div className="admin-modal-actions">
            <Button
              type="button"
              variant="outline"
              onClick={() => setIsFormModalOpen(false)}
            >
              Cancel
            </Button>
            <Button type="submit" variant="primary" disabled={isSubmitting}>
              {isSubmitting ? 'Saving...' : editingCoupon ? 'Update Coupon' : 'Create Coupon'}
            </Button>
          </div>
        </form>
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={Boolean(couponToDelete)}
        onClose={() => setCouponToDelete(null)}
        title="Delete Coupon"
      >
        {couponToDelete && (
          <div className="admin-modal-content">
            <p>
              Are you sure you want to delete coupon <strong>{couponToDelete.code}</strong>?
            </p>
            <p className="text-muted">
              Customers will no longer be able to apply this coupon code at checkout.
            </p>
            <div className="admin-modal-actions">
              <Button
                type="button"
                variant="outline"
                onClick={() => setCouponToDelete(null)}
              >
                Cancel
              </Button>
              <Button
                type="button"
                variant="danger"
                onClick={handleDeleteConfirm}
                disabled={isDeleting}
              >
                {isDeleting ? 'Deleting...' : 'Delete Coupon'}
              </Button>
            </div>
          </div>
        )}
      </Modal>

      {/* Usages Modal */}
      <Modal
        isOpen={Boolean(viewingUsagesCoupon)}
        onClose={() => setViewingUsagesCoupon(null)}
        title={`Redemption History: ${viewingUsagesCoupon?.code}`}
      >
        <div className="admin-modal-content">
          {isLoadingUsages ? (
            <p>Loading redemptions...</p>
          ) : usagesList.length === 0 ? (
            <p className="text-muted">This coupon has not been redeemed yet.</p>
          ) : (
            <div className="admin-usages-table-wrapper">
              <table className="admin-items-table">
                <thead>
                  <tr>
                    <th>Order #</th>
                    <th>User ID</th>
                    <th className="text-right">Discount</th>
                    <th>Date</th>
                  </tr>
                </thead>
                <tbody>
                  {usagesList.map((u, i) => (
                    <tr key={u.id || i}>
                      <td>#{u.order_id}</td>
                      <td>{u.user_id || 'User'}</td>
                      <td className="text-right font-semibold">
                        {formatCurrency(u.discount_amount || 0)}
                      </td>
                      <td>{formatDate(u.created_at)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
          <div className="admin-modal-actions">
            <Button
              type="button"
              variant="outline"
              onClick={() => setViewingUsagesCoupon(null)}
            >
              Close
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
