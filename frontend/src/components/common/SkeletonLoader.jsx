import './Loading.css';

/**
 * Drape Accessible Shimmer Skeleton Loader
 */
export function SkeletonLoader({
  width = '100%',
  height = '20px',
  borderRadius = 'var(--radius-sm)',
  className = '',
  style = {},
  count = 1,
}) {
  const items = Array.from({ length: count });

  if (count === 1) {
    return (
      <span
        className={`drape-skeleton ${className}`}
        style={{
          width,
          height,
          borderRadius,
          ...style,
        }}
        aria-hidden="true"
      />
    );
  }

  return (
    <div className="drape-skeleton-group">
      {items.map((_, i) => (
        <span
          key={i}
          className={`drape-skeleton ${className}`}
          style={{
            width,
            height,
            borderRadius,
            ...style,
          }}
          aria-hidden="true"
        />
      ))}
    </div>
  );
}
