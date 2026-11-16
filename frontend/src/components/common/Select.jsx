import { forwardRef } from 'react';
import { ChevronDown } from 'lucide-react';
import './Select.css';

/**
 * Drape Reusable Select Dropdown Component
 */
export const Select = forwardRef(function Select(
  {
    options = [],
    value,
    defaultValue,
    onChange,
    placeholder,
    size = 'md',
    hasError = false,
    disabled = false,
    className = '',
    children,
    ...rest
  },
  ref,
) {
  const wrapperClass = [
    'drape-select-wrapper',
    `drape-select-wrapper--${size}`,
    hasError ? 'drape-select-wrapper--error' : '',
    disabled ? 'drape-select-wrapper--disabled' : '',
    className,
  ]
    .filter(Boolean)
    .join(' ');

  return (
    <div className={wrapperClass}>
      <select
        ref={ref}
        className="drape-select"
        value={value}
        defaultValue={defaultValue}
        onChange={onChange}
        disabled={disabled}
        aria-invalid={hasError}
        {...rest}
      >
        {placeholder && (
          <option value="" disabled>
            {placeholder}
          </option>
        )}
        {options.length > 0
          ? options.map((opt) => (
              <option key={opt.value} value={opt.value} disabled={opt.disabled}>
                {opt.label}
              </option>
            ))
          : children}
      </select>

      <span className="drape-select__chevron" aria-hidden="true">
        <ChevronDown size={16} />
      </span>
    </div>
  );
});
