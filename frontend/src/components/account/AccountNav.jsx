import React from 'react';
import { NavLink } from 'react-router-dom';
import {
  LayoutDashboard,
  User,
  Package,
  MapPin,
  Wallet,
  RotateCcw,
  Gift,
  LogOut,
} from 'lucide-react';
import './AccountNav.css';

const NAV_ITEMS = [
  {
    to: '/account',
    label: 'Overview',
    icon: LayoutDashboard,
    end: true,
  },
  {
    to: '/account/profile',
    label: 'Personal Profile',
    icon: User,
  },
  {
    to: '/account/orders',
    label: 'My Orders',
    icon: Package,
  },
  {
    to: '/account/addresses',
    label: 'Address Book',
    icon: MapPin,
  },
  {
    to: '/account/wallet',
    label: 'Wallet',
    icon: Wallet,
  },
  {
    to: '/account/returns',
    label: 'Returns & Exchanges',
    icon: RotateCcw,
  },
  {
    to: '/account/referrals',
    label: 'Referral Program',
    icon: Gift,
  },
];

export function AccountNav({ onSignOutClick }) {
  return (
    <nav className="account-nav" aria-label="Account Sub-Navigation">
      <div className="account-nav__menu" role="menu">
        <span className="account-nav__section-title">ACCOUNT MANAGEMENT</span>
        {NAV_ITEMS.map((item) => {
          const Icon = item.icon;
          return (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              className={({ isActive }) =>
                `account-nav__item ${isActive ? 'is-active' : ''}`
              }
              role="menuitem"
              data-testid={`account-nav-${item.label.toLowerCase().replace(/[^a-z0-9]+/g, '-')}`}
            >
              <Icon size={18} className="account-nav__icon" aria-hidden="true" />
              <span className="account-nav__label">{item.label}</span>
            </NavLink>
          );
        })}

        <div className="account-nav__divider" aria-hidden="true" />

        <button
          type="button"
          className="account-nav__item account-nav__item--logout"
          onClick={onSignOutClick}
          role="menuitem"
          data-testid="account-nav-signout-btn"
        >
          <LogOut size={18} className="account-nav__icon" aria-hidden="true" />
          <span className="account-nav__label">Sign Out</span>
        </button>
      </div>
    </nav>
  );
}
