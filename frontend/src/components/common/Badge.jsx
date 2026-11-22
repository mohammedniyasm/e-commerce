import './Badge.css';

/**
 * Drape Reusable Badge & Status Pill Component
 * Supports:
 * - Semantic status tags (Orders, Payments, Inventory)
 * - Category filter buttons / pills (All, Men, Women, Outerwear)
 * - Numeric notification counters (Cart item counter)
 */
export function Badge({
  children,
  variant = 'neutral', // 'success', 'warning', 'danger', 'info', 'neutral', 'filter', 'counter'
  size = 'md',        // 'sm', 'md'
  isActive = false,   // For filter pills
  onClick = null,
  onRemove = null,
  className = '',
  ...rest
}) {
  const isClickable = Boolean(onClick);

  const classNames = [
    'drape-badge',
    `drape-badge--${variant}`,
    `drape-badge--${size}`,
    isActive ? 'drape-badge--active' : '',
    isClickable ? 'drape-badge--clickable' : '',
    className,
  ]
    .filter(Boolean)
    .join(' ');

  if (isClickable) {
    return (
      <button
        type="button"
        className={classNames}
        onClick={onClick}
        aria-pressed={isActive}
        {...rest}
      >
        {children}
      </button>
    );
  }

  return (
    <span className={classNames} {...rest}>
      {children}
      {onRemove && (
        <button
          type="button"
          className="drape-badge__remove"
          onClick={onRemove}
          aria-label="Remove filter"
        >
          ×
        </button>
      )}
    </span>
  );
}

/**
 * Convenience wrapper for Status Pills with preconfigured dot indicator
 */
export function StatusPill({ status, label, variant = 'neutral', className = '' }) {
  return (
    <span className={`drape-status-pill drape-status-pill--${variant} ${className}`}>
      <span className="drape-status-pill__dot" aria-hidden="true" />
      <span className="drape-status-pill__label">{label || status}</span>
    </span>
  );
}
