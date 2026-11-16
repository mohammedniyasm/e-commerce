import React, { useState, useEffect, useCallback } from 'react';
import { Outlet, Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { useToast } from '../context/ToastContext';
import { ProfileHeaderRibbon } from '../components/account/ProfileHeaderRibbon';
import { AccountNav } from '../components/account/AccountNav';
import { Modal } from '../components/common/Modal';
import { Button } from '../components/common/Button';
import { getProfile } from '../api/profileService';
import { getMyOrders } from '../api/orderService';
import { getUserAddresses } from '../api/checkoutService';
import { getWishlists } from '../api/wishlistService';
import './AccountLayout.css';

export function AccountLayout() {
  const { user, logout } = useAuth();
  const { showToast } = useToast();
  const navigate = useNavigate();

  const [profile, setProfile] = useState(user || null);
  const [stats, setStats] = useState({
    ordersCount: 0,
    wishlistCount: 0,
    addressesCount: 0,
  });
  const [recentOrders, setRecentOrders] = useState([]);
  const [addresses, setAddresses] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isSignOutModalOpen, setIsSignOutModalOpen] = useState(false);
  const [isLoggingOut, setIsLoggingOut] = useState(false);

  const loadAccountData = useCallback(async () => {
    try {
      setIsLoading(true);
      // Execute parallel requests to populate actual customer data
      const [profileRes, ordersRes, wishlistRes, addressesRes] = await Promise.allSettled([
        getProfile(),
        getMyOrders({ limit: 5, page: 1 }),
        getWishlists(),
        getUserAddresses(),
      ]);

      // Profile
      if (profileRes.status === 'fulfilled' && profileRes.value) {
        setProfile(profileRes.value);
      } else if (user) {
        setProfile(user);
      }

      // Orders count & recent orders
      let ordersCount = 0;
      if (ordersRes.status === 'fulfilled' && ordersRes.value) {
        const orderData = ordersRes.value;
        ordersCount = orderData.total_count || orderData.orders?.length || 0;
        setRecentOrders(orderData.orders?.slice(0, 3) || []);
      }

      // Wishlist items count
      let wishlistCount = 0;
      if (wishlistRes.status === 'fulfilled' && wishlistRes.value) {
        const wishlists = Array.isArray(wishlistRes.value)
          ? wishlistRes.value
          : wishlistRes.value?.wishlists || [];
        wishlistCount = wishlists.reduce(
          (acc, w) => acc + (w.items?.length || w.item_count || 0),
          0,
        );
      }

      // Addresses count & list
      let addressesList = [];
      if (addressesRes.status === 'fulfilled' && addressesRes.value) {
        addressesList = Array.isArray(addressesRes.value)
          ? addressesRes.value
          : addressesRes.value?.addresses || [];
        setAddresses(addressesList);
      }

      setStats({
        ordersCount,
        wishlistCount,
        addressesCount: addressesList.length,
      });
    } catch (err) {
      showToast({
        title: 'Account Information',
        message: 'Unable to load all account metrics. Showing available information.',
        type: 'warning',
      });
    } finally {
      setIsLoading(false);
    }
  }, [user, showToast]);

  useEffect(() => {
    loadAccountData();
  }, [loadAccountData]);

  async function handleConfirmSignOut() {
    setIsLoggingOut(true);
    try {
      await logout();
      setIsSignOutModalOpen(false);
      navigate('/');
      showToast({
        title: 'Signed Out',
        message: 'You have been signed out of your Drape account.',
        type: 'info',
      });
    } catch {
      setIsSignOutModalOpen(false);
      navigate('/');
    } finally {
      setIsLoggingOut(false);
    }
  }

  return (
    <div className="account-layout" data-testid="account-layout">
      {/* Top Breadcrumb Strip */}
      <nav className="account-breadcrumb" aria-label="Breadcrumb">
        <div className="account-breadcrumb__inner">
          <Link to="/" className="account-breadcrumb__link">
            Home
          </Link>
          <span className="account-breadcrumb__sep">/</span>
          <span className="account-breadcrumb__current">My Account</span>
        </div>
      </nav>

      {/* Header Profile Ribbon matching Figma 32:10998 */}
      <ProfileHeaderRibbon
        profile={profile}
        stats={stats}
        isLoading={isLoading}
      />

      {/* Main 2-Column Responsive Grid */}
      <div className="account-layout__container">
        <aside className="account-layout__sidebar">
          <AccountNav onSignOutClick={() => setIsSignOutModalOpen(true)} />
        </aside>

        <main className="account-layout__content" id="account-main-content">
          <Outlet
            context={{
              profile,
              stats,
              recentOrders,
              addresses,
              isLoading,
              refreshData: loadAccountData,
            }}
          />
        </main>
      </div>

      {/* Sign Out Modal matching Figma 33:14649 */}
      <Modal
        isOpen={isSignOutModalOpen}
        onClose={() => setIsSignOutModalOpen(false)}
        title="Sign Out of Atelier"
        eyebrow="SESSION CONFIRMATION"
        size="sm"
      >
        <div className="account-signout-modal" data-testid="signout-modal">
          <p className="account-signout-modal__desc">
            Are you sure you wish to sign out of Drape Atelier? You can sign back in at any time to
            access your tailored orders, saved addresses, and curations.
          </p>
          <div className="account-signout-modal__actions">
            <Button
              variant="outline"
              size="md"
              onClick={() => setIsSignOutModalOpen(false)}
              disabled={isLoggingOut}
            >
              STAY SIGNED IN
            </Button>
            <Button
              variant="primary"
              size="md"
              onClick={handleConfirmSignOut}
              loading={isLoggingOut}
              data-testid="confirm-signout-btn"
            >
              SIGN OUT
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
