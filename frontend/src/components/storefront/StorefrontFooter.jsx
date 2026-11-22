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
              Drape offers comfortable, well-made clothing for everyday life.
            </p>
          </div>

          {/* Column 2: Quick Links */}
          <div className="storefront-footer__col">
            <h2 className="storefront-footer__heading">QUICK LINKS</h2>
            <ul className="storefront-footer__list">
              <li><Link to="/">Home</Link></li>
              <li><Link to="/catalog">Collection</Link></li>
              <li><Link to="/catalog?sort=price_asc">Lowest Price First</Link></li>
              <li><Link to="/catalog?sort=price_desc">Highest Price First</Link></li>
            </ul>
          </div>

          {/* Column 3: Customer Service */}
          <div className="storefront-footer__col">
            <h2 className="storefront-footer__heading">CUSTOMER SERVICE</h2>
            <ul className="storefront-footer__list">
              <li><Link to="/account/orders">Orders & Tracking</Link></li>
              <li><Link to="/account/returns">Returns & Exchanges</Link></li>
              <li><Link to="/account/addresses">Delivery Addresses</Link></li>
            </ul>
          </div>

          {/* Column 4: Contact Us */}
          <div className="storefront-footer__col">
            <h2 className="storefront-footer__heading">CONTACT US</h2>
            <address className="storefront-footer__contact">
              <p>
                <a href="mailto:concierge@drapeatelier.com">Email customer support</a>
              </p>
            </address>
          </div>
        </div>

        <div className="storefront-footer__bottom">
          <p className="storefront-footer__copyright">
            © {new Date().getFullYear()} DRAPE. ALL RIGHTS RESERVED.
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
