import { Link, NavLink, useLocation } from 'react-router-dom';
import { ExternalLink, LogOut, Menu } from 'lucide-react';
import './AdminHeader.css';

const headerTabs = [
  { path: '/admin', label: 'Dashboard', end: true },
  { path: '/admin/orders', label: 'Orders' },
  { path: '/admin/shipments', label: 'Shipments' },
  { path: '/admin/payments', label: 'Payments' },
  { path: '/admin/customers', label: 'Customers' },
  { path: '/admin/products', label: 'Products' },
  { path: '/admin/returns', label: 'Returns' },
  { path: '/admin/coupons', label: 'Coupons' },
  { path: '/admin/reports', label: 'Reports' },
  { path: '/admin/settings', label: 'Settings' },
];

export function AdminHeader({ onMenuToggle, onLogoutClick }) {
  const location = useLocation();

  // Determine current section from path
  const pathParts = location.pathname.split('/').filter(Boolean);
  const currentSection =
    pathParts.length > 1
      ? pathParts[1].replace(/-/g, ' ')
      : 'Dashboard';

  return (
    <header className="admin-header">
      <div className="admin-header-left">
        <button
          type="button"
          className="admin-menu-toggle"
          onClick={onMenuToggle}
          aria-label="Toggle navigation menu"
        >
          <Menu size={20} />
        </button>

        <nav aria-label="Breadcrumb" className="admin-breadcrumb">
          <span className="admin-breadcrumb-root">Admin</span>
          <span className="admin-breadcrumb-sep">/</span>
          <span className="admin-breadcrumb-current">{currentSection}</span>
        </nav>

        {/* Quick Access Top Tabs */}
        {/* <nav className="admin-header-tabs" aria-label="Quick Section Tabs">
          {headerTabs.map((tab) => (
            <NavLink
              key={tab.path}
              to={tab.path}
              end={tab.end}
              className={({ isActive }) =>
                `admin-header-tab ${isActive ? 'active' : ''}`
              }
            >
              {tab.label}
            </NavLink>
          ))}
        </nav> */}
      </div>

      <div className="admin-header-right">
        <Link to="/" target="_blank" rel="noopener noreferrer" className="admin-view-store-link">
          <span>View Store</span>
          <ExternalLink size={13} />
        </Link>

        <div className="admin-header-divider" />

        <div className="admin-header-user">
          <div className="admin-header-avatar" aria-hidden="true">
            AD
          </div>
          <div className="admin-header-user-meta">
            <span className="admin-header-user-name">Admin Console</span>
            <span className="admin-header-user-role">Superuser</span>
          </div>
        </div>

        <button
          type="button"
          className="admin-header-logout-btn"
          onClick={onLogoutClick}
          aria-label="Sign out"
          title="Sign out of Admin Console"
        >
          <LogOut size={16} />
        </button>
      </div>
    </header>
  );
}
