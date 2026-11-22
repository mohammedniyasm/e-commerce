import { useEffect, useRef } from 'react';
import { X } from 'lucide-react';
import { Button } from './Button';
import './Modal.css';

const FOCUSABLE_SELECTOR =
  'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

/**
 * Drape Reusable Accessible Modal Component
 * Directly reproduces Drape Figma modal specifications:
 * - 440px / 672px width
 * - 32px padding
 * - Header with optional icon circle (48px)
 * - Uppercase eyebrow (11px, tracking 1.54px)
 * - Heading 2 (24px)
 * - Dual action buttons container with flex: 1
 *
 * Accessibility:
 * - Traps keyboard focus inside modal (Tab & Shift+Tab cycle)
 * - Restores focus to the triggering element upon close
 * - Dismisses on Escape key press
 * - Locks background document body scrolling while open
 */
export function Modal({
  isOpen,
  onClose,
  title,
  eyebrow = null,
  description = null,
  icon = null,
  iconVariant = 'neutral',
  size = 'md', // 'sm' (400px), 'md' (440px), 'lg' (672px), 'xl' (800px)
  children,
  footer = null,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  onConfirm = null,
  isConfirmLoading = false,
  confirmVariant = 'primary',
  showCloseButton = true,
  closeOnBackdropClick = true,
}) {
  const modalRef = useRef(null);
  const previousActiveElementRef = useRef(null);
  const onCloseRef = useRef(onClose);
  onCloseRef.current = onClose;

  // Focus trap, initial focus, and focus restoration
  useEffect(() => {
    if (!isOpen) return;

    // 1. Save currently focused element to restore when modal closes
    previousActiveElementRef.current = document.activeElement;

    // 2. Lock background scrolling
    const originalOverflow = document.body.style.overflow;
    document.body.style.overflow = 'hidden';

    // 3. Move initial focus into modal
    const focusTimer = setTimeout(() => {
      if (modalRef.current) {
        const focusableElements = modalRef.current.querySelectorAll(FOCUSABLE_SELECTOR);
        if (focusableElements.length > 0) {
          // Focus first interactive element or modal container
          focusableElements[0].focus();
        } else {
          modalRef.current.focus();
        }
      }
    }, 50);

    // 4. Keyboard handler for Escape and Tab trapping
    function handleKeyDown(e) {
      if (e.key === 'Escape') {
        e.preventDefault();
        onCloseRef.current();
        return;
      }

      if (e.key === 'Tab' && modalRef.current) {
        const focusables = Array.from(
          modalRef.current.querySelectorAll(FOCUSABLE_SELECTOR),
        );
        if (focusables.length === 0) {
          e.preventDefault();
          return;
        }

        const firstElement = focusables[0];
        const lastElement = focusables[focusables.length - 1];

        if (e.shiftKey) {
          // Shift + Tab
          if (document.activeElement === firstElement || !modalRef.current.contains(document.activeElement)) {
            e.preventDefault();
            lastElement.focus();
          }
        } else {
          // Tab
          if (document.activeElement === lastElement || !modalRef.current.contains(document.activeElement)) {
            e.preventDefault();
            firstElement.focus();
          }
        }
      }
    }

    document.addEventListener('keydown', handleKeyDown);

    return () => {
      clearTimeout(focusTimer);
      document.removeEventListener('keydown', handleKeyDown);
      document.body.style.overflow = originalOverflow;

      // Restore focus to triggering element
      if (
        previousActiveElementRef.current &&
        typeof previousActiveElementRef.current.focus === 'function'
      ) {
        previousActiveElementRef.current.focus();
      }
    };
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div
      className="drape-modal-overlay"
      role="presentation"
      onClick={closeOnBackdropClick ? onClose : undefined}
    >
      <div
        ref={modalRef}
        className={`drape-modal drape-modal--${size}`}
        role="dialog"
        tabIndex={-1}
        aria-modal="true"
        aria-labelledby={title ? 'drape-modal-title' : undefined}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header Section */}
        <div className="drape-modal__header">
          {icon ? (
            <div className={`drape-modal__icon-circle drape-modal__icon-circle--${iconVariant}`}>
              {icon}
            </div>
          ) : (
            <div />
          )}

          {showCloseButton && (
            <button
              type="button"
              className="drape-modal__close-btn"
              onClick={onClose}
              aria-label="Close modal"
            >
              <X size={20} />
            </button>
          )}
        </div>

        {/* Content Section */}
        <div className="drape-modal__body">
          {eyebrow && <p className="drape-modal__eyebrow">{eyebrow}</p>}
          {title && <h2 id="drape-modal-title" className="drape-modal__title">{title}</h2>}
          {description && <div className="drape-modal__description">{description}</div>}

          {children && <div className="drape-modal__custom-content">{children}</div>}
        </div>

        {/* Actions / Footer */}
        {footer !== null ? (
          footer
        ) : onConfirm ? (
          <div className="drape-modal__actions">
            <Button
              variant="secondary"
              size="md"
              onClick={onClose}
              disabled={isConfirmLoading}
            >
              {cancelText}
            </Button>
            <Button
              variant={confirmVariant}
              size="md"
              onClick={onConfirm}
              isLoading={isConfirmLoading}
            >
              {confirmText}
            </Button>
          </div>
        ) : null}
      </div>
    </div>
  );
}
