import { forwardRef } from 'react';
import { Spinner } from './Spinner';
import './Button.css';

/**
 * Drape Reusable Button Component
 * Follows Drape design system from Figma (height 48px, uppercase, 11px bold, letter-spacing)
 */
export const Button = forwardRef(function Button(
  {
    children,
    type = 'button',
    variant = 'primary',
    size = 'md',
    fullWidth = false,
    isLoading = false,
    disabled = false,
    leftIcon = null,
    rightIcon = null,
    className = '',
    onClick,
    ...rest
  },
  ref,
) {
  const isDisabled = disabled || isLoading;

  const classNames = [
    'drape-btn',
    `drape-btn--${variant}`,
    `drape-btn--${size}`,
    fullWidth ? 'drape-btn--full-width' : '',
    isLoading ? 'drape-btn--loading' : '',
    className,
  ]
    .filter(Boolean)
    .join(' ');

  return (
    <button
      ref={ref}
      type={type}
      className={classNames}
      disabled={isDisabled}
      aria-busy={isLoading}
      aria-disabled={isDisabled}
      onClick={onClick}
      {...rest}
    >
      {isLoading ? (
        <span className="drape-btn__spinner">
          <Spinner size={size === 'sm' ? 14 : 16} color="currentColor" />
        </span>
      ) : leftIcon ? (
        <span className="drape-btn__icon drape-btn__icon--left" aria-hidden="true">
          {leftIcon}
        </span>
      ) : null}

      <span className="drape-btn__content">{children}</span>

      {!isLoading && rightIcon ? (
        <span className="drape-btn__icon drape-btn__icon--right" aria-hidden="true">
          {rightIcon}
        </span>
      ) : null}
    </button>
  );
});
