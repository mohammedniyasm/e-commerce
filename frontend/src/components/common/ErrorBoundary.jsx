import React from 'react';
import { AlertCircle, RefreshCw, Home } from 'lucide-react';
import './ErrorBoundary.css';

export class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error) {
    return { hasError: true, error };
  }

  componentDidCatch(error, errorInfo) {
    console.error('Unhandled React render error in component tree:', error, errorInfo);
  }

  handleReload = () => {
    window.location.reload();
  };

  handleGoHome = () => {
    window.location.href = '/';
  };

  render() {
    if (this.state.hasError) {
      return (
        <div className="drape-error-boundary">
          <div className="drape-error-card">
            <div className="drape-error-icon">
              <AlertCircle size={40} />
            </div>
            <p className="drape-error-eyebrow">APPLICATION ERROR</p>
            <h1 className="drape-error-title">Something went wrong</h1>
            <p className="drape-error-desc">
              An unexpected error occurred while rendering this page. You can try refreshing the page or return to the home page.
            </p>
            <div className="drape-error-actions">
              <button
                type="button"
                className="drape-btn-primary"
                onClick={this.handleReload}
              >
                <RefreshCw size={16} />
                Refresh Page
              </button>
              <button
                type="button"
                className="drape-btn-outline"
                onClick={this.handleGoHome}
              >
                <Home size={16} />
                Go to Home
              </button>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
