export function PlaceholderPage({ eyebrow, title, description }) {
  return (
    <section className="placeholder-page" aria-labelledby="page-title">
      <p className="placeholder-page__eyebrow">{eyebrow}</p>
      <h1 id="page-title">{title}</h1>
      <p>{description}</p>
    </section>
  );
}
