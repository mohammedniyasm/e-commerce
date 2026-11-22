import { useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { AlertCircle, ArrowRight, Eye, EyeOff, Lock, Mail, Shield } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { useToast } from '../../context/ToastContext';
import { validateEmail } from '../../utils/validators';
import './AdminLoginPage.css';

export function AdminLoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { loginAdmin, logout } = useAuth();
  const { showToast } = useToast();

  const [formData, setFormData] = useState({
    email: '',
    password: '',
  });

  const [errors, setErrors] = useState({});
  const [serverError, setServerError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  function handleChange(field, value) {
    setFormData((prev) => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: null }));
    }
    if (serverError) {
      setServerError('');
    }
  }

  function validateForm() {
    const nextErrors = {};

    const emailError = validateEmail(formData.email);
    if (emailError) nextErrors.email = emailError;

    if (!formData.password) {
      nextErrors.password = 'Password is required.';
    }

    setErrors(nextErrors);
    return Object.keys(nextErrors).length === 0;
  }

  async function handleSubmit(e) {
    e.preventDefault();
    if (isSubmitting) return;

    setServerError('');

    if (!validateForm()) {
      return;
    }

    try {
      setIsSubmitting(true);
      const session = await loginAdmin({
        email: formData.email.trim().toLowerCase(),
        password: formData.password,
      });

      if (session?.role !== 'admin') {
        await logout({ role: session?.role });
        setServerError('Access restricted: This account does not possess administrator privileges.');
        return;
      }

      showToast({
        type: 'success',
        title: 'Authentication Successful',
        message: 'Welcome to DRAPE Admin Console.',
      });

      const destination = location.state?.from || '/admin';
      navigate(destination, { replace: true });
    } catch (err) {
      const message =
        err?.response?.data?.message ||
        err?.message ||
        'Unable to authenticate administrator. Please check your credentials.';
      setServerError(message);
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className="admin-login-page">
      <div className="admin-login-card">
        <div className="admin-login-header">
          <div className="admin-login-brand">
            <span className="admin-login-logo">DRAPE</span>
            <span className="admin-login-badge">Admin</span>
          </div>
          <p className="admin-login-subtitle">
            Sign in with administrator credentials to manage store operations and analytics.
          </p>
        </div>

        {serverError && (
          <div className="admin-login-alert" role="alert">
            <AlertCircle size={18} style={{ flexShrink: 0, marginTop: 2 }} />
            <span>{serverError}</span>
          </div>
        )}

        <form className="admin-login-form" onSubmit={handleSubmit} noValidate>
          <div className="form-field">
            <label className="form-label" htmlFor="admin-email">
              Admin Email
            </label>
            <div className="admin-input-wrapper">
              <Mail className="admin-input-icon" />
              <input
                id="admin-email"
                type="email"
                name="email"
                autoComplete="email"
                placeholder="admin@example.com"
                value={formData.email}
                onChange={(e) => handleChange('email', e.target.value)}
                className={`admin-login-input ${errors.email ? 'has-error' : ''}`}
                disabled={isSubmitting}
                required
              />
            </div>
            {errors.email && <span className="admin-field-error">{errors.email}</span>}
          </div>

          <div className="form-field">
            <label className="form-label" htmlFor="admin-password">
              Password
            </label>
            <div className="admin-input-wrapper">
              <Lock className="admin-input-icon" />
              <input
                id="admin-password"
                type={showPassword ? 'text' : 'password'}
                name="password"
                autoComplete="current-password"
                placeholder="••••••••••••"
                value={formData.password}
                onChange={(e) => handleChange('password', e.target.value)}
                className={`admin-login-input ${errors.password ? 'has-error' : ''}`}
                disabled={isSubmitting}
                required
              />
              <button
                type="button"
                className="admin-password-toggle"
                onClick={() => setShowPassword((prev) => !prev)}
                aria-label={showPassword ? 'Hide password' : 'Show password'}
                tabIndex={-1}
              >
                {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
              </button>
            </div>
            {errors.password && <span className="admin-field-error">{errors.password}</span>}
          </div>

          <button
            type="submit"
            className="admin-submit-btn"
            disabled={isSubmitting}
          >
            {isSubmitting ? (
              <span>Authenticating...</span>
            ) : (
              <>
                <Shield size={16} />
                <span>Enter Admin Console</span>
                <ArrowRight size={16} />
              </>
            )}
          </button>
        </form>

        <div className="admin-login-footer">
          <Link to="/" className="admin-store-link">
            &larr; Return to Storefront
          </Link>
        </div>
      </div>
    </div>
  );
}
