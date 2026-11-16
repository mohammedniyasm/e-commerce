import { Outlet } from 'react-router-dom';
import { StorefrontHeader } from '../components/storefront/StorefrontHeader';
import { StorefrontFooter } from '../components/storefront/StorefrontFooter';
import './StorefrontLayout.css';

export function StorefrontLayout() {
  return (
    <div className="storefront-layout">
      <StorefrontHeader />
      <main className="storefront-main" id="main-content">
        <Outlet />
      </main>
      <StorefrontFooter />
    </div>
  );
}
