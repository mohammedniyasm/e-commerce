import { Link, Navigate, Outlet, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import heroCampaignImg from '../assets/hero-campaign.png';
import './AuthLayout.css';

export function AuthLayout() {
  const { isAuthenticated } = useAuth();
  const location = useLocation();

  // Redirect authenticated customers away from auth pages
  if (isAuthenticated) {
    const from = location.state?.from || '/';
    return <Navigate replace to={from} />;
  }

  return (
    <div className="auth-layout">
      {/* Left Column: Editorial Showcase */}
      <aside className="auth-editorial-column" aria-label="Editorial Showcase">
        <div className="auth-editorial-image-wrapper">
          <img
            src={heroCampaignImg}
            alt="Drape clothing collection"
            className="auth-editorial-image"
          />
          <div className="auth-editorial-overlay" />
        </div>

        {/* Top Curatorial Micro-Label */}
        <div className="auth-editorial-top">
          <span className="auth-editorial-edition">NEW SEASON</span>
          <div className="auth-editorial-location">
            <span className="auth-location-dot" />
            <span className="auth-location-text">KOCHI FLAGSHIP</span>
          </div>
        </div>

        {/* Bottom Editorial Content */}
        <div className="auth-editorial-bottom">
          <div className="auth-editorial-collection-label">
            DRAPE — NEW SEASON COLLECTION
          </div>
          <blockquote className="auth-editorial-quote">
            <p>“Comfortable clothing for everyday life.”</p>
          </blockquote>
          <div className="auth-editorial-meta">
            <span>COMFORTABLE EVERYDAY CLOTHING</span>
          </div>
        </div>
      </aside>

      {/* Right Column: Authentication Form Container */}
      <main className="auth-content-column">
        <header className="auth-header">
          <Link to="/" className="auth-brand-logo" aria-label="Drape Home">
            DRAPE
          </Link>
        </header>

        <div className="auth-card-container">
          <Outlet />
        </div>

        <footer className="auth-footer">
          <div className="auth-footer-legal">
            <span>© {new Date().getFullYear()} DRAPE. ALL RIGHTS RESERVED.</span>
          </div>
          <nav className="auth-footer-links" aria-label="Legal links">
            <Link to="/catalog">COLLECTION</Link>
            <span className="auth-footer-dot">•</span>
            <span className="auth-footer-text">SECURE CHECKOUT</span>
          </nav>
        </footer>
      </main>
    </div>
  );
}
