import { Link } from 'react-router-dom';
import './StorefrontFooter.css';

export function StorefrontFooter() {
  return (
    <footer className="storefront-footer" role="contentinfo">
      <div className="storefront-footer__inner">
        <div className="storefront-footer__grid">
          {/* Column 1: About Us */}
          <div className="storefront-footer__col">
            <h2 className="storefront-footer__heading">ABOUT US</h2>
            <p className="storefront-footer__text">
              DRAPE embodies minimalist slow-fashion, sustainable tailoring, and timeless wardrobe essentials crafted for the intentional silhouette.
            </p>
          </div>

          {/* Column 2: Quick Links */}
          <div className="storefront-footer__col">
            <h2 className="storefront-footer__heading">QUICK LINKS</h2>
            <ul className="storefront-footer__list">
              <li><Link to="/">Home</Link></li>
              <li><Link to="/catalog">Collection</Link></li>
              <li><Link to="/catalog?sort=price_asc">Curated Essentials</Link></li>
              <li><Link to="/catalog?sort=price_desc">New Arrivals</Link></li>
            </ul>
          </div>

          {/* Column 3: Customer Service */}
          <div className="storefront-footer__col">
            <h2 className="storefront-footer__heading">CUSTOMER SERVICE</h2>
            <ul className="storefront-footer__list">
              <li><Link to="/account">Orders & Tracking</Link></li>
              <li><Link to="/catalog">Returns & Exchanges</Link></li>
              <li><Link to="/catalog">Shipping Information</Link></li>
              <li><Link to="/catalog">Frequently Asked Questions</Link></li>
            </ul>
          </div>

          {/* Column 4: Contact Us */}
          <div className="storefront-footer__col">
            <h2 className="storefront-footer__heading">CONTACT US</h2>
            <address className="storefront-footer__contact">
              <p>concierge@drapeatelier.com</p>
              <p>+1 (800) 294-8190</p>
              <p>482 Mercer Street, Soho</p>
              <p>New York, NY 10013</p>
            </address>
          </div>
        </div>

        <div className="storefront-footer__bottom">
          <p className="storefront-footer__copyright">
            © {new Date().getFullYear()} DRAPE ATELIER. ALL RIGHTS RESERVED.
          </p>
          <div className="storefront-footer__badges">
            <span>PRICES IN INDIAN RUPEES (INR ₹)</span>
            <span className="storefront-footer__dot">•</span>
            <span>SECURE 256-BIT ENCRYPTION</span>
          </div>
        </div>
      </div>
    </footer>
  );
}
