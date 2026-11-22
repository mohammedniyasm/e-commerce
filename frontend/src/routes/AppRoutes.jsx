import { Navigate, Route, Routes } from 'react-router-dom';
import { AccountLayout } from '../layouts/AccountLayout';
import { AdminLayout } from '../layouts/AdminLayout';
import { AuthLayout } from '../layouts/AuthLayout';
import { StorefrontLayout } from '../layouts/StorefrontLayout';
import { AccountOverviewPage } from '../pages/account/AccountOverviewPage';
import { ProfilePage } from '../pages/account/ProfilePage';
import { AddressesPage } from '../pages/account/AddressesPage';
import { OrdersPage } from '../pages/account/OrdersPage';
import { OrderDetailsPage } from '../pages/account/OrderDetailsPage';
import { ReturnsPage } from '../pages/account/ReturnsPage';
import { WalletPage } from '../pages/account/WalletPage';
import { ReferralPage } from '../pages/account/ReferralPage';
import { AdminDashboardPage } from '../pages/admin/AdminDashboardPage';
import { AdminLoginPage } from '../pages/admin/AdminLoginPage';
import { AdminProductsPage } from '../pages/admin/products/AdminProductsPage';
import { AdminProductFormPage } from '../pages/admin/products/AdminProductFormPage';
import { AdminProductDetailsPage } from '../pages/admin/products/AdminProductDetailsPage';
import { AdminCategoriesPage } from '../pages/admin/categories/AdminCategoriesPage';
import { AdminBrandsPage } from '../pages/admin/brands/AdminBrandsPage';
import { AdminOrdersPage } from '../pages/admin/orders/AdminOrdersPage';
import { AdminOrderDetailsPage } from '../pages/admin/orders/AdminOrderDetailsPage';
import { AdminShipmentsPage } from '../pages/admin/shipments/AdminShipmentsPage';
import { AdminPaymentsPage } from '../pages/admin/payments/AdminPaymentsPage';
import { AdminReturnsPage } from '../pages/admin/returns/AdminReturnsPage';
import { AdminCouponsPage } from '../pages/admin/coupons/AdminCouponsPage';
import { AdminReviewsPage } from '../pages/admin/reviews/AdminReviewsPage';
import { AdminCustomersPage } from '../pages/admin/customers/AdminCustomersPage';
import { AdminReportsPage } from '../pages/admin/reports/AdminReportsPage';
import { AdminReferralsPage } from '../pages/admin/referrals/AdminReferralsPage';
import { AdminSettingsPage } from '../pages/admin/settings/AdminSettingsPage';
import { ForgotPasswordPage } from '../pages/auth/ForgotPasswordPage';
import { LoginPage } from '../pages/auth/LoginPage';
import { PasswordResetSuccessPage } from '../pages/auth/PasswordResetSuccessPage';
import { RegisterPage } from '../pages/auth/RegisterPage';
import { ResetPasswordPage } from '../pages/auth/ResetPasswordPage';
import { CartPage } from '../pages/cart/CartPage';
import { CheckoutPage } from '../pages/checkout/CheckoutPage';
import { OrderConfirmationPage } from '../pages/checkout/OrderConfirmationPage';
import { NotFoundPage } from '../pages/NotFoundPage';
import { CatalogPage } from '../pages/storefront/CatalogPage';
import { HomePage } from '../pages/storefront/HomePage';
import { ProductDetailPage } from '../pages/storefront/ProductDetailPage';
import { WishlistPage } from '../pages/wishlist/WishlistPage';
import { RequireAuth } from './RequireAuth';

export function AppRoutes() {
  return (
    <Routes>
      {/* Customer Authentication Flow */}
      <Route element={<AuthLayout />}>
        <Route path="login" element={<LoginPage />} />
        <Route path="register" element={<RegisterPage />} />
        <Route path="forgot-password" element={<ForgotPasswordPage />} />
        <Route path="reset-password" element={<ResetPasswordPage />} />
        <Route path="reset-password/success" element={<PasswordResetSuccessPage />} />
        {/* Support /auth/* aliases */}
        <Route path="auth/login" element={<Navigate replace to="/login" />} />
        <Route path="auth/register" element={<Navigate replace to="/register" />} />
        <Route path="auth/forgot-password" element={<Navigate replace to="/forgot-password" />} />
        <Route path="auth/reset-password" element={<Navigate replace to="/reset-password" />} />
      </Route>

      {/* Storefront Routes */}
      <Route element={<StorefrontLayout />}>
        <Route index element={<HomePage />} />
        <Route path="catalog" element={<CatalogPage />} />
        <Route path="products/:slug" element={<ProductDetailPage />} />
        <Route path="products/id/:id" element={<ProductDetailPage />} />
        <Route path="cart" element={<CartPage />} />
        <Route path="wishlist" element={<WishlistPage />} />

        {/* Customer Protected Checkout & Account */}
        <Route element={<RequireAuth />}>
          <Route path="account" element={<AccountLayout />}>
            <Route index element={<AccountOverviewPage />} />
            <Route path="profile" element={<ProfilePage />} />
            <Route path="orders" element={<OrdersPage />} />
            <Route path="orders/:orderId" element={<OrderDetailsPage />} />
            <Route path="addresses" element={<AddressesPage />} />
            <Route path="wallet" element={<WalletPage />} />
            <Route path="returns" element={<ReturnsPage />} />
            <Route path="referrals" element={<ReferralPage />} />
          </Route>
          <Route path="checkout" element={<CheckoutPage />} />
          <Route path="checkout/success" element={<OrderConfirmationPage />} />
        </Route>
      </Route>

      {/* Admin Authentication */}
      <Route path="admin/login" element={<AdminLoginPage />} />

      {/* Admin Dashboard & Management Routes */}
      <Route element={<RequireAuth role="admin" />}>
        <Route path="admin" element={<AdminLayout />}>
          <Route index element={<AdminDashboardPage />} />
          <Route path="products" element={<AdminProductsPage />} />
          <Route path="products/new" element={<AdminProductFormPage />} />
          <Route path="products/:id" element={<AdminProductDetailsPage />} />
          <Route path="products/:id/edit" element={<AdminProductFormPage />} />
          <Route path="categories" element={<AdminCategoriesPage />} />
          <Route path="brands" element={<AdminBrandsPage />} />
          <Route path="orders" element={<AdminOrdersPage />} />
          <Route path="orders/:orderId" element={<AdminOrderDetailsPage />} />
          <Route path="shipments" element={<AdminShipmentsPage />} />
          <Route path="payments" element={<AdminPaymentsPage />} />
          <Route path="returns" element={<AdminReturnsPage />} />
          <Route path="coupons" element={<AdminCouponsPage />} />
          <Route path="reviews" element={<AdminReviewsPage />} />
          <Route path="customers" element={<AdminCustomersPage />} />
          <Route path="reports" element={<AdminReportsPage />} />
          <Route path="referrals" element={<AdminReferralsPage />} />
          <Route path="settings" element={<AdminSettingsPage />} />
        </Route>
      </Route>

      {/* 404 Route */}
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  );
}
