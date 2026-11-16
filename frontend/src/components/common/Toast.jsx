import { CheckCircle2, AlertCircle, AlertTriangle, Info, X } from 'lucide-react';
import './Toast.css';

const ICON_MAP = {
  success: CheckCircle2,
  error: AlertCircle,
  warning: AlertTriangle,
  info: Info,
};

export function ToastItem({ toast, onDismiss }) {
  const IconComponent = ICON_MAP[toast.type] || Info;

  return (
    <div className={`drape-toast drape-toast--${toast.type}`} role="alert">
      <span className="drape-toast__icon">
        <IconComponent size={18} strokeWidth={2} />
      </span>
      <p className="drape-toast__message">{toast.message}</p>
      <button
        type="button"
        className="drape-toast__close"
        onClick={() => onDismiss(toast.id)}
        aria-label="Dismiss notification"
      >
        <X size={14} />
      </button>
    </div>
  );
}

export function ToastContainer({ toasts, onDismiss }) {
  if (!toasts || toasts.length === 0) return null;

  return (
    <div className="drape-toast-container" aria-live="polite" aria-atomic="true">
      {toasts.map((toast) => (
        <ToastItem key={toast.id} toast={toast} onDismiss={onDismiss} />
      ))}
    </div>
  );
}
