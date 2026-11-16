import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export function RequireAuth({ role, children }) {
  const { isAuthenticated, session, status } = useAuth();
  const location = useLocation();

  if (status === 'initializing') {
    return null;
  }

  if (!isAuthenticated) {
    if (role === 'admin') {
      return <Navigate replace to="/admin/login" state={{ from: location.pathname }} />;
    }
    return <Navigate replace to="/login" state={{ from: location.pathname }} />;
  }

  if (role && session?.role !== role) {
    return <Navigate replace to="/" state={{ from: location.pathname }} />;
  }

  return children ?? <Outlet />;
}
