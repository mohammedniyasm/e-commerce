import { createContext, useCallback, useContext, useMemo, useState } from 'react';
import { ToastContainer } from '../components/common/Toast';

const ToastContext = createContext(null);

let toastCount = 0;

export function ToastProvider({ children }) {
  const [toasts, setToasts] = useState([]);

  const removeToast = useCallback((id) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const addToast = useCallback(
    ({ type = 'info', message = '', duration = 4000 }) => {
      const id = ++toastCount;
      const newToast = { id, type, message };

      setToasts((prev) => [...prev, newToast]);

      if (duration > 0) {
        setTimeout(() => {
          removeToast(id);
        }, duration);
      }

      return id;
    },
    [removeToast],
  );

  const toast = useMemo(
    () => ({
      success: (msg, duration) => addToast({ type: 'success', message: msg, duration }),
      error: (msg, duration) => addToast({ type: 'error', message: msg, duration }),
      warning: (msg, duration) => addToast({ type: 'warning', message: msg, duration }),
      info: (msg, duration) => addToast({ type: 'info', message: msg, duration }),
      dismiss: removeToast,
    }),
    [addToast, removeToast],
  );

  return (
    <ToastContext.Provider value={toast}>
      {children}
      <ToastContainer toasts={toasts} onDismiss={removeToast} />
    </ToastContext.Provider>
  );
}

export function useToast() {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider.');
  }
  return context;
}
