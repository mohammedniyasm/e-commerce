import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { ArrowRight, RefreshCw, ShieldCheck, Sparkles, Truck } from 'lucide-react';
import { getStoreCategories, getStoreProducts } from '../../api/productService';
import heroCampaignImg from '../../assets/hero-campaign.png';
import { Button } from '../../components/common/Button';
import { SkeletonLoader } from '../../components/common/SkeletonLoader';
import { ProductCard } from '../../components/storefront/ProductCard';
import './HomePage.css';

export function HomePage() {
  const [featuredProducts, setFeaturedProducts] = useState([]);
  const [categories, setCategories] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    let isMounted = true;

    async function loadHomeData() {
      try {
        setIsLoading(true);
        setError(null);
        const [productsData, categoriesData] = await Promise.allSettled([
          getStoreProducts({ limit: 4, sort: 'name_asc' }),
          getStoreCategories(),
        ]);

        if (isMounted) {
          if (productsData.status === 'fulfilled' && productsData.value?.products) {
            setFeaturedProducts(productsData.value.products);
          }
          if (categoriesData.status === 'fulfilled' && Array.isArray(categoriesData.value)) {
            setCategories(categoriesData.value);
          }
        }
      } catch (err) {
        if (isMounted) {
          setError(err.message || 'Unable to load products. Please check connection.');
        }
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    }

    loadHomeData();

    return () => {
      isMounted = false;
    };
  }, []);

  return (
    <div className="home-page" data-testid="home-page">
      {/* 1. Hero Campaign Banner */}
      <section className="home-hero" aria-label="New season collection">
        <div className="home-hero__image-wrapper">
          <img
            src={heroCampaignImg}
            alt="Drape clothing collection"
            className="home-hero__image"
          />
          <div className="home-hero__overlay" />
        </div>

        <div className="home-hero__content">
          <div className="home-hero__pill">
            <span className="home-hero__pill-dot" />
            <span>NEW SEASON • KOCHI</span>
          </div>

          <h1 className="home-hero__title">
            Clothes for Every Day
          </h1>

          <p className="home-hero__subtitle">
            Comfortable, well-made clothing for work, weekends, and everything in between.
          </p>

          <div className="home-hero__actions">
            <Link to="/catalog">
              <Button variant="primary" size="lg" rightIcon={<ArrowRight size={14} />}>
                SHOP COLLECTION
              </Button>
            </Link>
            <Link to="/catalog">
              <Button variant="outline" size="lg" className="home-hero__btn-secondary">
                VIEW ALL PRODUCTS
              </Button>
            </Link>
          </div>
        </div>
      </section>

      {/* 2. Value Pillars */}
      <section className="home-features" aria-label="Store benefits">
        <div className="home-features__inner">
          <div className="home-feature-item">
            <Truck size={20} className="home-feature-icon" />
            <div>
              <h3 className="home-feature-title">INSURED SHIPPING</h3>
              <p className="home-feature-desc">Track your order from dispatch to delivery.</p>
            </div>
          </div>
          <div className="home-feature-item">
            <Sparkles size={20} className="home-feature-icon" />
            <div>
              <h3 className="home-feature-title">QUALITY MATERIALS</h3>
              <p className="home-feature-desc">Everyday clothing made with quality fabrics.</p>
            </div>
          </div>
          <div className="home-feature-item">
            <ShieldCheck size={20} className="home-feature-icon" />
            <div>
              <h3 className="home-feature-title">EASY RETURNS</h3>
              <p className="home-feature-desc">Request a return or exchange for eligible orders.</p>
            </div>
          </div>
        </div>
      </section>

      {/* 3. Featured Categories Grid */}
      {categories.length > 0 && (
        <section className="home-section" aria-labelledby="shop-categories-title">
          <div className="home-section__header">
            <div>
              <span className="home-section__eyebrow">BROWSE CATEGORIES</span>
              <h2 id="shop-categories-title" className="home-section__title">
                Shop By Category
              </h2>
            </div>
            <Link to="/catalog" className="home-section__link">
              <span>View All Categories</span>
              <ArrowRight size={14} />
            </Link>
          </div>

          <div className="home-categories-grid">
            {categories.slice(0, 4).map((cat) => (
              <Link
                key={cat.id}
                to={`/catalog?category_id=${cat.id}`}
                className="home-category-card"
              >
                <div className="home-category-card__content">
                  <span className="home-category-card__eyebrow">CATEGORY</span>
                  <h3 className="home-category-card__title">{cat.name}</h3>
                  <span className="home-category-card__action">
                    Explore &rarr;
                  </span>
                </div>
              </Link>
            ))}
          </div>
        </section>
      )}

      {/* 4. Featured Latest Arrivals */}
      <section className="home-section" aria-labelledby="featured-arrivals-title">
        <div className="home-section__header">
          <div>
            <span className="home-section__eyebrow">SHOP FAVORITES</span>
            <h2 id="featured-arrivals-title" className="home-section__title">
              Popular Products
            </h2>
          </div>
          <Link to="/catalog" className="home-section__link">
            <span>Explore Full Collection</span>
            <ArrowRight size={14} />
          </Link>
        </div>

        {isLoading ? (
          <div className="home-products-grid">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="home-product-skeleton">
                <SkeletonLoader height="320px" />
                <div style={{ marginTop: '12px' }}>
                  <SkeletonLoader height="14px" width="40%" />
                  <SkeletonLoader height="18px" width="80%" style={{ marginTop: '6px' }} />
                  <SkeletonLoader height="16px" width="30%" style={{ marginTop: '6px' }} />
                </div>
              </div>
            ))}
          </div>
        ) : error ? (
          <div className="home-state-banner">
            <p>{error}</p>
            <Link to="/catalog">
              <Button variant="secondary" size="sm">
                Browse Catalog
              </Button>
            </Link>
          </div>
        ) : featuredProducts.length > 0 ? (
          <div className="home-products-grid">
            {featuredProducts.map((product) => (
              <ProductCard key={product.id} product={product} />
            ))}
          </div>
        ) : (
          <div className="home-state-banner">
            <p>No products are available right now. Please check back soon.</p>
            <Link to="/catalog">
              <Button variant="primary" size="md">
                Browse Catalog
              </Button>
            </Link>
          </div>
        )}
      </section>

      {/* 5. Editorial Philosophy Section */}
      <section className="home-manifesto">
        <div className="home-manifesto__inner">
          <blockquote className="home-manifesto__quote">
            “Drape brings together comfortable clothing and quality you can rely on.”
          </blockquote>
          <p className="home-manifesto__author">— DRAPE, KOCHI</p>
        </div>
      </section>
    </div>
  );
}
