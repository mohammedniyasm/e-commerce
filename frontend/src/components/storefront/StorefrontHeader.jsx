import { useState } from 'react';
import { Link, NavLink, useNavigate } from 'react-router-dom';
import {
  Heart,
  LogOut,
  Menu,
  Search,
  ShoppingBag,
  User,
  X,
} from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { useCart } from '../../context/CartContext';
import './StorefrontHeader.css';

export function StorefrontHeader() {
  const navigate = useNavigate();
  const { isAuthenticated, session, logout } = useAuth();
  const { totalItems } = useCart();
  const [searchQuery, setSearchQuery] = useState('');
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false);

  function handleSearchSubmit(e) {
    e.preventDefault();
    const query = searchQuery.trim();
    if (query) {
      navigate(`/catalog?search=${encodeURIComponent(query)}`);
      setIsMobileMenuOpen(false);
    }
  }

  async function handleLogout() {
    setIsUserMenuOpen(false);
    await logout();
    navigate('/');
  }

  return (
    <header className="storefront-header">
      <div className="storefront-header__inner">
        {/* Mobile Hamburger Toggle */}
        <button
          type="button"
          className="storefront-header__menu-btn"
          onClick={() => setIsMobileMenuOpen((prev) => !prev)}
          aria-label={isMobileMenuOpen ? 'Close menu' : 'Open menu'}
          aria-expanded={isMobileMenuOpen}
        >
          {isMobileMenuOpen ? <X size={22} /> : <Menu size={22} />}
        </button>

        {/* Brand Logo */}
        <Link to="/" className="storefront-header__logo" aria-label="Drape Home">
          DRAPE
        </Link>

        {/* Search Bar */}
        <form className="storefront-header__search-form" onSubmit={handleSearchSubmit} role="search">
          <div className="storefront-header__search-input-wrapper">
            <Search size={15} className="storefront-header__search-icon" aria-hidden="true" />
            <input
              type="search"
              className="storefront-header__search-input"
              placeholder="Search collection, clothing, accessories..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              aria-label="Search collection"
            />
          </div>
        </form>

        {/* Desktop Navigation Links */}
        <nav className="storefront-header__nav" aria-label="Main Storefront">
          <NavLink
            to="/catalog"
            className={({ isActive }) =>
              `storefront-header__nav-link ${isActive ? 'is-active' : ''}`
            }
          >
            COLLECTION
          </NavLink>
        </nav>

        {/* Action Icons */}
        <div className="storefront-header__actions">
          {/* Wishlist Link */}
          <Link
            to="/wishlist"
            className="storefront-header__action-btn"
            aria-label="Wishlist"
          >
            <Heart size={20} />
          </Link>

          {/* Cart Link with Badge */}
          <Link
            to="/cart"
            className="storefront-header__action-btn storefront-header__cart-btn"
            aria-label="Shopping Cart"
          >
            <ShoppingBag size={20} />
            <span className="storefront-header__cart-badge" aria-label="Items in cart">
              {totalItems}
            </span>
          </Link>

          {/* Account / User Menu */}
          {isAuthenticated ? (
            <div className="storefront-header__user-menu-wrapper">
              <button
                type="button"
                className="storefront-header__user-avatar-btn"
                onClick={() => setIsUserMenuOpen((prev) => !prev)}
                aria-label="Customer profile menu"
                aria-expanded={isUserMenuOpen}
              >
                <div className="storefront-header__avatar-circle">
                  <User size={16} />
                </div>
              </button>

              {isUserMenuOpen && (
                <div className="storefront-header__user-dropdown" role="menu">
                  <div className="storefront-header__user-info">
                    <span className="storefront-header__user-name">
                      {session?.name || 'Customer Account'}
                    </span>
                    <span className="storefront-header__user-role">
                      {session?.role === 'admin' ? 'Administrator' : 'Customer'}
                    </span>
                  </div>
                  <div className="storefront-header__dropdown-divider" />
                  <Link
                    to="/account"
                    className="storefront-header__dropdown-item"
                    role="menuitem"
                    onClick={() => setIsUserMenuOpen(false)}
                  >
                    <User size={14} />
                    <span>My Profile</span>
                  </Link>
                  {session?.role === 'admin' && (
                    <Link
                      to="/admin"
                      className="storefront-header__dropdown-item"
                      role="menuitem"
                      onClick={() => setIsUserMenuOpen(false)}
                    >
                      <span>Admin Dashboard</span>
                    </Link>
                  )}
                  <button
                    type="button"
                    className="storefront-header__dropdown-item storefront-header__logout-btn"
                    role="menuitem"
                    onClick={handleLogout}
                  >
                    <LogOut size={14} />
                    <span>Sign Out</span>
                  </button>
                </div>
              )}
            </div>
          ) : (
            <Link
              to="/login"
              className="storefront-header__signin-link"
              aria-label="Sign In"
            >
              <div className="storefront-header__avatar-circle">
                <User size={16} />
              </div>
              <span className="storefront-header__signin-text">SIGN IN</span>
            </Link>
          )}
        </div>
      </div>

      {/* Mobile Navigation Drawer */}
      {isMobileMenuOpen && (
        <div className="storefront-header__mobile-drawer" role="dialog" aria-label="Mobile navigation">
          <form className="storefront-header__mobile-search" onSubmit={handleSearchSubmit}>
            <div className="storefront-header__search-input-wrapper">
              <Search size={15} className="storefront-header__search-icon" aria-hidden="true" />
              <input
                type="search"
                className="storefront-header__search-input"
                placeholder="Search collection..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
            </div>
          </form>

          <nav className="storefront-header__mobile-nav">
            <Link
              to="/catalog"
              className="storefront-header__mobile-nav-link"
              onClick={() => setIsMobileMenuOpen(false)}
            >
              COLLECTION
            </Link>
            <Link
              to="/catalog?sort=price_asc"
              className="storefront-header__mobile-nav-link"
              onClick={() => setIsMobileMenuOpen(false)}
            >
              PRICE: LOW TO HIGH
            </Link>
            <Link
              to="/catalog?sort=price_desc"
              className="storefront-header__mobile-nav-link"
              onClick={() => setIsMobileMenuOpen(false)}
            >
              PRICE: HIGH TO LOW
            </Link>
            {isAuthenticated ? (
              <>
                <Link
                  to="/account"
                  className="storefront-header__mobile-nav-link"
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  MY ACCOUNT
                </Link>
                <button
                  type="button"
                  className="storefront-header__mobile-nav-link storefront-header__mobile-logout"
                  onClick={handleLogout}
                >
                  SIGN OUT
                </button>
              </>
            ) : (
              <Link
                to="/login"
                className="storefront-header__mobile-nav-link"
                onClick={() => setIsMobileMenuOpen(false)}
              >
                SIGN IN / REGISTER
              </Link>
            )}
          </nav>
        </div>
      )}
    </header>
  );
}
