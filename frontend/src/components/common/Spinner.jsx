import './Loading.css';

/**
 * Drape Accessible SVG Spinner
 */
export function Spinner({ size = 20, color = 'currentColor', className = '', label = 'Loading...' }) {
  return (
    <span
      className={`drape-spinner ${className}`}
      style={{ width: size, height: size, color }}
      role="status"
      aria-label={label}
    >
      <svg
        className="drape-spinner__svg"
        viewBox="0 0 24 24"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <circle
          className="drape-spinner__track"
          cx="12"
          cy="12"
          r="9"
          stroke="currentColor"
          strokeWidth="2.5"
          opacity="0.2"
        />
        <path
          className="drape-spinner__head"
          d="M12 3a9 9 0 0 1 9 9"
          stroke="currentColor"
          strokeWidth="2.5"
          strokeLinecap="round"
        />
      </svg>
    </span>
  );
}
