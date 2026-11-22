import { useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { ArrowRight, Eye, EyeOff, Lock, Mail, Phone, Tag, User } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { useToast } from '../../context/ToastContext';
import { Button } from '../../components/common/Button';
import { FormField } from '../../components/common/FormField';
import { Input } from '../../components/common/Input';
import { GoogleAuthButton } from '../../components/auth/GoogleAuthButton';
import {
  checkPasswordRules,
  cleanPhoneNumber,
  validateEmail,
  validateName,
  validatePassword,
  validatePhone,
} from '../../utils/validators';
import '../../styles/auth.css';

export function RegisterPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { registerCustomer } = useAuth();
  const { showToast } = useToast();
  const referralCodeFromUrl = searchParams.get('ref')?.trim() || '';

  const [formData, setFormData] = useState({
    name: '',
    email: '',
    phone: '',
    password: '',
    referralCode: referralCodeFromUrl,
  });

  const [errors, setErrors] = useState({});
  const [serverError, setServerError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [showReferralInput, setShowReferralInput] = useState(Boolean(referralCodeFromUrl));

  const passwordRules = checkPasswordRules(formData.password);

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

    const nameError = validateName(formData.name);
    if (nameError) nextErrors.name = nameError;

    const emailError = validateEmail(formData.email);
    if (emailError) nextErrors.email = emailError;

    const phoneError = validatePhone(formData.phone);
    if (phoneError) nextErrors.phone = phoneError;

    const passwordError = validatePassword(formData.password);
    if (passwordError) nextErrors.password = passwordError;

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
      const payload = {
        name: formData.name.trim(),
        email: formData.email.trim().toLowerCase(),
        phone: cleanPhoneNumber(formData.phone),
        password: formData.password,
        referral_code: formData.referralCode?.trim() || undefined,
      };

      await registerCustomer(payload);

      showToast({
        title: 'Account Created',
        message: 'Your Drape account was created successfully. Please sign in.',
        type: 'success',
      });

      navigate('/login', {
        replace: true,
        state: { email: formData.email.trim().toLowerCase() },
      });
    } catch (err) {
      const msg = err.message || 'Failed to create your account. Please try again.';
      setServerError(msg);
      showToast({
        title: 'Registration Failed',
        message: msg,
        type: 'error',
      });
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className="auth-card" data-testid="register-card">
      <div className="auth-card-header">
        <h1 className="auth-card-title">Create Account</h1>
        <p className="auth-card-subtitle">Create an account to save products and track your orders.</p>
      </div>

      {serverError && (
        <div className="auth-error-banner" role="alert">
          <span>{serverError}</span>
        </div>
      )}

      <form className="auth-form" onSubmit={handleSubmit} noValidate>
        {/* Full Name Field */}
        <FormField
          id="register-name"
          label="FULL NAME"
          required
          error={errors.name}
        >
          <Input
            id="register-name"
            name="name"
            type="text"
            size="lg"
            placeholder="Your full name"
            value={formData.name}
            onChange={(e) => handleChange('name', e.target.value)}
            hasError={Boolean(errors.name)}
            disabled={isSubmitting}
            leftIcon={<User size={16} />}
            autoComplete="name"
            required
          />
        </FormField>

        {/* Email Address Field */}
        <FormField
          id="register-email"
          label="EMAIL ADDRESS"
          required
          error={errors.email}
        >
          <Input
            id="register-email"
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

        {/* Phone Number Field */}
        <FormField
          id="register-phone"
          label="PHONE NUMBER"
          required
          error={errors.phone}
          hint="10-digit mobile number"
        >
          <div className="auth-phone-field">
            <span className="auth-phone-prefix" aria-label="Country Code +91">
              +91
            </span>
            <Input
              id="register-phone"
              name="phone"
              type="tel"
              size="lg"
              placeholder="98765 43210"
              value={formData.phone}
              onChange={(e) => handleChange('phone', e.target.value)}
              hasError={Boolean(errors.phone)}
              disabled={isSubmitting}
              leftIcon={<Phone size={16} />}
              className="auth-phone-input"
              autoComplete="tel-national"
              required
            />
          </div>
        </FormField>

        {/* Password Field */}
        <FormField
          id="register-password"
          label="PASSWORD"
          required
          error={errors.password}
        >
          <Input
            id="register-password"
            name="password"
            type={showPassword ? 'text' : 'password'}
            size="lg"
            placeholder="Minimum 8 characters"
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
            autoComplete="new-password"
            required
          />
        </FormField>

        {/* Live Password Requirements Checklist */}
        {formData.password.length > 0 && (
          <div className="auth-password-checklist" aria-label="Password requirements">
            <div className={`auth-password-rule ${passwordRules.minLength ? 'is-valid' : ''}`}>
              <span className="auth-password-dot" />
              <span>At least 8 characters</span>
            </div>
            <div className={`auth-password-rule ${passwordRules.hasUppercase ? 'is-valid' : ''}`}>
              <span className="auth-password-dot" />
              <span>Uppercase letter (A-Z)</span>
            </div>
            <div className={`auth-password-rule ${passwordRules.hasLowercase ? 'is-valid' : ''}`}>
              <span className="auth-password-dot" />
              <span>Lowercase letter (a-z)</span>
            </div>
            <div className={`auth-password-rule ${passwordRules.hasNumber ? 'is-valid' : ''}`}>
              <span className="auth-password-dot" />
              <span>At least 1 digit (0-9)</span>
            </div>
            <div className={`auth-password-rule ${passwordRules.hasSpecial ? 'is-valid' : ''}`}>
              <span className="auth-password-dot" />
              <span>Special character (!@#$)</span>
            </div>
          </div>
        )}

        {/* Referral Code (Collapsible) */}
        {!showReferralInput ? (
          <button
            type="button"
            className="auth-resend-button"
            style={{ alignSelf: 'flex-start', margin: '2px 0 6px 0', fontSize: '12px' }}
            onClick={() => setShowReferralInput(true)}
          >
            + Have a referral code?
          </button>
        ) : (
          <FormField
            id="register-referral"
            label="REFERRAL CODE (OPTIONAL)"
          >
            <Input
              id="register-referral"
              name="referralCode"
              type="text"
              size="md"
              placeholder="e.g. FRIEND10"
              value={formData.referralCode}
              onChange={(e) => handleChange('referralCode', e.target.value)}
              disabled={isSubmitting}
              leftIcon={<Tag size={16} />}
            />
          </FormField>
        )}

        {/* Primary Submit Button */}
        <Button
          type="submit"
          variant="primary"
          size="lg"
          fullWidth
          isLoading={isSubmitting}
          rightIcon={<ArrowRight size={14} />}
        >
          CREATE ACCOUNT
        </Button>
      </form>

      {/* Divider */}
      <div className="auth-divider-container" aria-hidden="true">
        <div className="auth-divider-line" />
        <span className="auth-divider-text">OR CONTINUE WITH</span>
      </div>

      {/* Google Sign-On */}
      <GoogleAuthButton
        referralCode={formData.referralCode}
        disabled={isSubmitting}
        onSuccess={() => navigate('/', { replace: true })}
      />

      {/* Redirect to Login */}
      <div className="auth-switch-link-wrapper">
        <span>Already have an account?</span>
        <Link to="/login" className="auth-switch-link">
          Login
        </Link>
      </div>
    </div>
  );
}
