/**
 * Drape Utility Formatters
 * Currency, Date, and Status display helpers
 */

/**
 * Format a number or numeric string into Indian Rupees (INR - ₹)
 * Matches the Drape Figma visual standard (e.g. ₹1,25,000 or ₹2,499.00)
 *
 * @param {number|string} amount
 * @param {object} options
 * @param {boolean} [options.showDecimals=false]
 * @returns {string}
 */
export function formatCurrency(amount, { showDecimals = false } = {}) {
  const numericAmount = Number(amount);

  if (isNaN(numericAmount)) {
    return '₹0';
  }

  return new Intl.NumberFormat('en-IN', {
    style: 'currency',
    currency: 'INR',
    maximumFractionDigits: showDecimals ? 2 : 0,
    minimumFractionDigits: showDecimals ? 2 : 0,
  }).format(numericAmount);
}

export const formatCurrencyINR = formatCurrency;

/**
 * Format an ISO date string into standard human-readable display (e.g. "14 Nov 2026")
 *
 * @param {string|Date} dateInput
 * @param {object} options
 * @returns {string}
 */
export function formatDate(dateInput, options = {}) {
  if (!dateInput) return '—';

  const date = new Date(dateInput);
  if (isNaN(date.getTime())) return '—';

  const defaultOptions = {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    ...options,
  };

  return new Intl.DateTimeFormat('en-IN', defaultOptions).format(date);
}

/**
 * Format an ISO date string into date with time (e.g. "14 Nov 2026, 01:30 PM")
 *
 * @param {string|Date} dateInput
 * @returns {string}
 */
export function formatDateTime(dateInput) {
  if (!dateInput) return '—';

  const date = new Date(dateInput);
  if (isNaN(date.getTime())) return '—';

  return new Intl.DateTimeFormat('en-IN', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: true,
  }).format(date);
}

/**
 * Order status dictionary mapping API status to label and semantic theme variant
 */
const ORDER_STATUS_MAP = {
  pending: { label: 'Pending', variant: 'warning' },
  processing: { label: 'Processing', variant: 'info' },
  shipped: { label: 'Shipped', variant: 'info' },
  delivered: { label: 'Delivered', variant: 'success' },
  cancelled: { label: 'Cancelled', variant: 'danger' },
  return_requested: { label: 'Return Requested', variant: 'warning' },
  returned: { label: 'Returned', variant: 'danger' },
};

export function formatOrderStatus(status) {
  const normalized = (status || '').toLowerCase();
  return ORDER_STATUS_MAP[normalized] || { label: status || 'Unknown', variant: 'neutral' };
}

/**
 * Payment status dictionary
 */
const PAYMENT_STATUS_MAP = {
  pending: { label: 'Pending', variant: 'warning' },
  paid: { label: 'Paid', variant: 'success' },
  failed: { label: 'Failed', variant: 'danger' },
  refunded: { label: 'Refunded', variant: 'info' },
  partially_refunded: { label: 'Partially Refunded', variant: 'info' },
};

export function formatPaymentStatus(status) {
  const normalized = (status || '').toLowerCase();
  return PAYMENT_STATUS_MAP[normalized] || { label: status || 'Unknown', variant: 'neutral' };
}

/**
 * Return request status dictionary
 */
const RETURN_STATUS_MAP = {
  requested: { label: 'Requested', variant: 'warning' },
  approved: { label: 'Approved', variant: 'info' },
  rejected: { label: 'Rejected', variant: 'danger' },
  pickup_scheduled: { label: 'Pickup Scheduled', variant: 'info' },
  item_received: { label: 'Item Received', variant: 'info' },
  refund_completed: { label: 'Refunded to Wallet', variant: 'success' },
  exchange_dispatched: { label: 'Exchange Dispatched', variant: 'success' },
};

export function formatReturnStatus(status) {
  const normalized = (status || '').toLowerCase();
  return RETURN_STATUS_MAP[normalized] || { label: status || 'Unknown', variant: 'neutral' };
}
