import { AlertCircle } from 'lucide-react';
import './FormField.css';

/**
 * Drape FormField Wrapper
 * Encapsulates standard form field typography from Figma:
 * - Uppercase tracking label (11px, weight 600)
 * - Required indicator
 * - Helper text / hint
 * - Inline validation error message
 */
export function FormField({
  id,
  label,
  required = false,
  error = null,
  hint = null,
  children,
  className = '',
}) {
  return (
    <div className={`drape-form-field ${error ? 'drape-form-field--error' : ''} ${className}`}>
      {label && (
        <label htmlFor={id} className="drape-form-label">
          <span>{label}</span>
          {required && <span className="drape-form-label__required" aria-hidden="true">*</span>}
        </label>
      )}

      <div className="drape-form-field__control">{children}</div>

      {error ? (
        <p className="drape-form-error" role="alert" id={id ? `${id}-error` : undefined}>
          <AlertCircle size={14} className="drape-form-error__icon" aria-hidden="true" />
          <span>{error}</span>
        </p>
      ) : hint ? (
        <p className="drape-form-hint" id={id ? `${id}-hint` : undefined}>
          {hint}
        </p>
      ) : null}
    </div>
  );
}
