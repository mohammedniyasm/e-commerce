import { createContext, useContext, useEffect, useMemo, useState } from 'react';
import {
  loginAdmin,
  loginCustomer,
  loginWithGoogle,
  logout,
  registerCustomer,
  restoreSession,
} from '../api/authService';
import { subscribeToExpiredSession } from '../api/client';
import { clearAccessToken, getSession, subscribeToSession } from '../api/tokenStore';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const initialSession = getSession();
  const [session, setSession] = useState(initialSession);
  const [status, setStatus] = useState(initialSession ? 'authenticated' : 'initializing');

  useEffect(() => {
    let isMounted = true;
    const unsubscribeFromSession = subscribeToSession((nextSession) => {
      if (isMounted) {
        setSession(nextSession);
      }
    });
    const unsubscribeFromExpiry = subscribeToExpiredSession(() => {
      if (isMounted) {
        setStatus('anonymous');
      }
    });

    restoreSession().then((restoredSession) => {
      if (isMounted) {
        setStatus(restoredSession ? 'authenticated' : 'anonymous');
      }
    });

    return () => {
      isMounted = false;
      unsubscribeFromSession();
      unsubscribeFromExpiry();
    };
  }, []);

  async function startCustomerSession(credentials) {
    const nextSession = await loginCustomer(credentials);
    setStatus('authenticated');
    return nextSession;
  }

  async function startGoogleSession(payload) {
    const nextSession = await loginWithGoogle(payload);
    setStatus('authenticated');
    return nextSession;
  }

  async function registerNewCustomer(payload) {
    const response = await registerCustomer(payload);
    return response?.data;
  }

  async function startAdminSession(credentials) {
    const nextSession = await loginAdmin(credentials);
    setStatus('authenticated');
    return nextSession;
  }

  async function endSession() {
    try {
      await logout({ role: session?.role });
    } finally {
      clearAccessToken();
      setStatus('anonymous');
    }
  }

  const value = useMemo(
    () => ({
      session,
      status,
      isAuthenticated: Boolean(session?.accessToken),
      loginCustomer: startCustomerSession,
      loginGoogle: startGoogleSession,
      registerCustomer: registerNewCustomer,
      loginAdmin: startAdminSession,
      logout: endSession,
    }),
    [session, status],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);

  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider.');
  }

  return context;
}
