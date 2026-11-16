import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { ArrowRight, Eye, EyeOff, Lock, ShieldCheck } from 'lucide-react';
import { resetPassword } from '../../api/authService';
import { useToast } from '../../context/ToastContext';
import { Button } from '../../components/common/Button';
import { FormField } from '../../components/common/FormField';
import { Input } from '../../components/common/Input';
import {
  checkPasswordRules,
  validateConfirmPassword,
  validatePassword,
} from '../../utils/validators';
import '../../styles/auth.css';

export function ResetPasswordPage() {
  const navigate = useNavigate();
  const { showToast } = useToast();

  const [formData, setFormData] = useState({
    newPassword: '',
    confirmPassword: '',
  });

  const [errors, setErrors] = useState({});
  const [serverError, setServerError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  const passwordRules = checkPasswordRules(formData.newPassword);

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

    const passwordError = validatePassword(formData.newPassword);
    if (passwordError) nextErrors.newPassword = passwordError;

    const confirmError = validateConfirmPassword(
      formData.newPassword,
      formData.confirmPassword,
    );
    if (confirmError) nextErrors.confirmPassword = confirmError;

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
      // The browser automatically attaches the HttpOnly 'password_reset' cookie
      await resetPassword({
        new_password: formData.newPassword,
      });

      showToast({
        title: 'Password Updated',
        message: 'Your password has been reset successfully.',
        type: 'success',
      });

      navigate('/reset-password/success', { replace: true });
    } catch (err) {
      const msg = err.message || 'Failed to reset password. Your session may have expired.';
      setServerError(msg);
      showToast({
        title: 'Reset Failed',
        message: msg,
        type: 'error',
      });
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className="auth-card" data-testid="reset-password-card">
      <div className="auth-card-header">
        <h1 className="auth-card-title">Reset Password</h1>
        <p className="auth-card-subtitle">
          Create a new strong password for your Drape account
        </p>
      </div>

      {serverError && (
        <div className="auth-error-banner" role="alert">
          <span>{serverError}</span>
        </div>
      )}

      <form className="auth-form" onSubmit={handleSubmit} noValidate>
        {/* New Password Field */}
        <FormField
          id="reset-new-password"
          label="NEW PASSWORD"
          required
          error={errors.newPassword}
        >
          <Input
            id="reset-new-password"
            name="newPassword"
            type={showNewPassword ? 'text' : 'password'}
            size="lg"
            placeholder="Minimum 8 characters"
            value={formData.newPassword}
            onChange={(e) => handleChange('newPassword', e.target.value)}
            hasError={Boolean(errors.newPassword)}
            disabled={isSubmitting}
            leftIcon={<Lock size={16} />}
            rightIcon={
              <button
                type="button"
                onClick={() => setShowNewPassword((prev) => !prev)}
                className="drape-input__icon-button"
                aria-label={showNewPassword ? 'Hide password' : 'Show password'}
              >
                {showNewPassword ? <EyeOff size={16} /> : <Eye size={16} />}
              </button>
            }
            autoComplete="new-password"
            required
          />
        </FormField>

        {/* Password Strength Checklist */}
        {formData.newPassword.length > 0 && (
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

        {/* Confirm Password Field */}
        <FormField
          id="reset-confirm-password"
          label="CONFIRM NEW PASSWORD"
          required
          error={errors.confirmPassword}
        >
          <Input
            id="reset-confirm-password"
            name="confirmPassword"
            type={showConfirmPassword ? 'text' : 'password'}
            size="lg"
            placeholder="Re-enter new password"
            value={formData.confirmPassword}
            onChange={(e) => handleChange('confirmPassword', e.target.value)}
            hasError={Boolean(errors.confirmPassword)}
            disabled={isSubmitting}
            leftIcon={<Lock size={16} />}
            rightIcon={
              <button
                type="button"
                onClick={() => setShowConfirmPassword((prev) => !prev)}
                className="drape-input__icon-button"
                aria-label={showConfirmPassword ? 'Hide password' : 'Show password'}
              >
                {showConfirmPassword ? <EyeOff size={16} /> : <Eye size={16} />}
              </button>
            }
            autoComplete="new-password"
            required
          />
        </FormField>

        {/* Submit Button */}
        <Button
          type="submit"
          variant="primary"
          size="lg"
          fullWidth
          isLoading={isSubmitting}
          rightIcon={<ShieldCheck size={16} />}
        >
          RESET PASSWORD
        </Button>
      </form>

      {/* Back to Login link */}
      <div className="auth-switch-link-wrapper">
        <Link to="/login" className="auth-switch-link">
          Back to Login
        </Link>
      </div>
    </div>
  );
}
