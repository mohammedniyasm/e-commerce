import { forwardRef } from 'react';
import './FormField.css';

/**
 * Drape Reusable Input Primitive
 * Matches Figma input specifications (height 48px, 15px font, shadow, left/right icon support)
 */
export const Input = forwardRef(function Input(
  {
    type = 'text',
    size = 'md',
    hasError = false,
    leftIcon = null,
    rightIcon = null,
    className = '',
    disabled = false,
    ...rest
  },
  ref,
) {
  const containerClass = [
    'drape-input-wrapper',
    `drape-input-wrapper--${size}`,
    leftIcon ? 'drape-input-wrapper--has-left-icon' : '',
    rightIcon ? 'drape-input-wrapper--has-right-icon' : '',
    hasError ? 'drape-input-wrapper--error' : '',
    disabled ? 'drape-input-wrapper--disabled' : '',
    className,
  ]
    .filter(Boolean)
    .join(' ');

  return (
    <div className={containerClass}>
      {leftIcon && (
        <span className="drape-input__icon drape-input__icon--left" aria-hidden="true">
          {leftIcon}
        </span>
      )}

      <input
        ref={ref}
        type={type}
        className="drape-input"
        disabled={disabled}
        aria-invalid={hasError}
        {...rest}
      />

      {rightIcon && (
        <span className="drape-input__icon drape-input__icon--right">
          {rightIcon}
        </span>
      )}
    </div>
  );
});
