import { useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { ArrowRight, Eye, EyeOff, Lock, Mail } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { useToast } from '../../context/ToastContext';
import { Button } from '../../components/common/Button';
import { FormField } from '../../components/common/FormField';
import { Input } from '../../components/common/Input';
import { GoogleAuthButton } from '../../components/auth/GoogleAuthButton';
import { validateEmail } from '../../utils/validators';
import '../../styles/auth.css';

export function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { loginCustomer } = useAuth();
  const { showToast } = useToast();

  const [formData, setFormData] = useState({
    email: location.state?.email || '',
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
      await loginCustomer({
        email: formData.email.trim().toLowerCase(),
        password: formData.password,
      });

      showToast({
        title: 'Welcome Back',
        message: 'Successfully signed in to your Drape account.',
        type: 'success',
      });

      const destination = location.state?.from || '/';
      navigate(destination, { replace: true });
    } catch (err) {
      const msg = err.message || 'Invalid email or password. Please try again.';
      setServerError(msg);
      showToast({
        title: 'Login Failed',
        message: msg,
        type: 'error',
      });
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className="auth-card" data-testid="login-card">
      <div className="auth-card-header">
        <h1 className="auth-card-title">Welcome Back</h1>
        <p className="auth-card-subtitle">Sign in to view your account and orders.</p>
      </div>

      {serverError && (
        <div className="auth-error-banner" role="alert">
          <span>{serverError}</span>
        </div>
      )}

      <form className="auth-form" onSubmit={handleSubmit} noValidate>
        {/* Email Address Field */}
        <FormField
          id="login-email"
          label="EMAIL ADDRESS"
          required
          error={errors.email}
        >
          <Input
            id="login-email"
            name="email"
            type="email"
            size="lg"
            placeholder="you@example.com"
            value={formData.email}
            onChange={(e) => handleChange('email', e.target.value)}
            hasError={Boolean(errors.email)}
            disabled={isSubmitting}
            leftIcon={<Mail size={16} />}
            autoComplete="email"
            required
          />
        </FormField>

        {/* Password Field */}
        <FormField
          id="login-password"
          label="PASSWORD"
          required
          error={errors.password}
        >
          <Input
            id="login-password"
            name="password"
            type={showPassword ? 'text' : 'password'}
            size="lg"
            placeholder="Enter your password"
            value={formData.password}
            onChange={(e) => handleChange('password', e.target.value)}
            hasError={Boolean(errors.password)}
            disabled={isSubmitting}
            leftIcon={<Lock size={16} />}
            rightIcon={
              <button
                type="button"
                onClick={() => setShowPassword((prev) => !prev)}
                className="drape-input__icon-button"
                aria-label={showPassword ? 'Hide password' : 'Show password'}
              >
                {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
              </button>
            }
            autoComplete="current-password"
            required
          />
        </FormField>

        {/* Forgot Password Link */}
        <Link to="/forgot-password" className="auth-forgot-link">
          Forgot Password?
        </Link>

        {/* Submit Button */}
        <Button
          type="submit"
          variant="primary"
          size="lg"
          fullWidth
          isLoading={isSubmitting}
          rightIcon={<ArrowRight size={14} />}
        >
          LOGIN
        </Button>
      </form>

      {/* Divider */}
      <div className="auth-divider-container" aria-hidden="true">
        <div className="auth-divider-line" />
        <span className="auth-divider-text">OR CONTINUE WITH</span>
      </div>

      {/* Google Sign-On */}
      <GoogleAuthButton
        disabled={isSubmitting}
        onSuccess={() => {
          const destination = location.state?.from || '/';
          navigate(destination, { replace: true });
        }}
      />

      {/* Redirect to Register */}
      <div className="auth-switch-link-wrapper">
        <span>Don't have an account?</span>
        <Link to="/register" className="auth-switch-link">
          Create Account
        </Link>
      </div>
    </div>
  );
}
