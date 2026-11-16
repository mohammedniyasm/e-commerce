import { Link } from 'react-router-dom';

export function NotFoundPage() {
  return (
    <section className="placeholder-page" aria-labelledby="page-title">
      <p className="placeholder-page__eyebrow">404</p>
      <h1 id="page-title">Page not found</h1>
      <p>The requested route does not exist yet.</p>
      <Link to="/">Return to Drape</Link>
    </section>
  );
}
