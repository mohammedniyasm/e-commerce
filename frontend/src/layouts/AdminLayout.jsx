import { useState } from 'react';
import { Outlet, useNavigate } from 'react-router-dom';
import { AdminSidebar } from '../components/admin/AdminSidebar';
import { AdminHeader } from '../components/admin/AdminHeader';
import { Modal } from '../components/common/Modal';
import { useAuth } from '../context/AuthContext';
import { useToast } from '../context/ToastContext';
import './AdminLayout.css';

export function AdminLayout() {
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);
  const [isLogoutModalOpen, setIsLogoutModalOpen] = useState(false);
  const [isLoggingOut, setIsLoggingOut] = useState(false);
  const { logout } = useAuth();
  const { showToast } = useToast();
  const navigate = useNavigate();

  async function handleConfirmLogout() {
    try {
      setIsLoggingOut(true);
      await logout();
      showToast({
        type: 'info',
        title: 'Signed Out',
        message: 'Your administrator session has ended.',
      });
      navigate('/admin/login', { replace: true });
    } catch {
      navigate('/admin/login', { replace: true });
    } finally {
      setIsLoggingOut(false);
      setIsLogoutModalOpen(false);
    }
  }

  return (
    <div className="admin-shell">
      <AdminSidebar
        isOpen={isSidebarOpen}
        onClose={() => setIsSidebarOpen(false)}
        onLogoutClick={() => setIsLogoutModalOpen(true)}
      />

      <div className="admin-main-container">
        <AdminHeader
          onMenuToggle={() => setIsSidebarOpen((prev) => !prev)}
          onLogoutClick={() => setIsLogoutModalOpen(true)}
        />

        <main className="admin-content">
          <Outlet />
        </main>
      </div>

      <Modal
        isOpen={isLogoutModalOpen}
        onClose={() => setIsLogoutModalOpen(false)}
        title="Confirm Sign Out"
        eyebrow="Admin Session"
        description="Are you sure you want to sign out of the Drape Admin Console? Any unsaved changes in active forms will be lost."
        confirmText="Sign Out"
        cancelText="Stay Signed In"
        confirmVariant="danger"
        isConfirmLoading={isLoggingOut}
        onConfirm={handleConfirmLogout}
      />
    </div>
  );
}
