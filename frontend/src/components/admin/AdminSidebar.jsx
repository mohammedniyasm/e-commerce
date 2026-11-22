import { NavLink } from 'react-router-dom';
import {
  Award,
  BarChart3,
  CreditCard,
  Gift,
  Layers,
  LayoutDashboard,
  LogOut,
  Package,
  RotateCcw,
  Settings,
  ShoppingBag,
  Star,
  TicketPercent,
  Truck,
  Users,
} from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import './AdminSidebar.css';

const navSections = [
  {
    title: 'Operations',
    items: [
      { path: '/admin', label: 'Dashboard', icon: LayoutDashboard, end: true },
      { path: '/admin/orders', label: 'Orders', icon: ShoppingBag },
      { path: '/admin/shipments', label: 'Shipments', icon: Truck },
      { path: '/admin/payments', label: 'Payments', icon: CreditCard },
      { path: '/admin/customers', label: 'Customers', icon: Users },
      { path: '/admin/returns', label: 'Returns', icon: RotateCcw },
    ],
  },
  {
    title: 'Catalog',
    items: [
      { path: '/admin/products', label: 'Products', icon: Package },
      { path: '/admin/categories', label: 'Categories', icon: Layers },
      { path: '/admin/brands', label: 'Brands', icon: Award },
    ],
  },
  {
    title: 'Growth & Settings',
    items: [
      { path: '/admin/coupons', label: 'Coupons', icon: TicketPercent },
      { path: '/admin/reviews', label: 'Reviews', icon: Star },
      { path: '/admin/referrals', label: 'Referrals', icon: Gift },
      { path: '/admin/reports', label: 'Reports', icon: BarChart3 },
      { path: '/admin/settings', label: 'Settings', icon: Settings },
    ],
  },
];

export function AdminSidebar({ isOpen, onClose, onLogoutClick }) {
  const { session } = useAuth();

  return (
    <>
      {isOpen && (
        <div
          className="admin-sidebar-backdrop"
          onClick={onClose}
          aria-hidden="true"
        />
      )}

      <aside className={`admin-sidebar ${isOpen ? 'open' : ''}`} aria-label="Admin Navigation">
        <div className="admin-sidebar-brand">
          <div className="admin-brand-link">
            <span className="admin-brand-logo">DRAPE</span>
            <span className="admin-brand-badge">Admin</span>
          </div>
          <span className="admin-brand-sub">Operations Console</span>
        </div>

        <nav className="admin-sidebar-nav">
          {navSections.map((section) => (
            <div key={section.title} className="admin-nav-group">
              <span className="admin-nav-section-title">{section.title}</span>
              {section.items.map((item) => {
                const Icon = item.icon;
                return (
                  <NavLink
                    key={item.path}
                    to={item.path}
                    end={item.end}
                    className={({ isActive }) =>
                      `admin-nav-item ${isActive ? 'active' : ''}`
                    }
                    onClick={() => {
                      if (isOpen && onClose) onClose();
                    }}
                  >
                    <Icon className="admin-nav-icon" />
                    <span>{item.label}</span>
                  </NavLink>
                );
              })}
            </div>
          ))}
        </nav>

        <div className="admin-sidebar-footer">
          <div className="admin-user-card">
            <div className="admin-user-avatar" aria-hidden="true">
              AD
            </div>
            <div className="admin-user-info">
              <span className="admin-user-name">Administrator</span>
              <span className="admin-user-role">
                <span className="admin-user-status-dot" />
                Active Session
              </span>
            </div>
          </div>

          <button
            type="button"
            className="admin-sidebar-logout-btn"
            onClick={onLogoutClick}
          >
            <LogOut size={14} />
            <span>Sign Out</span>
          </button>
        </div>
      </aside>
    </>
  );
}
