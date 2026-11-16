import { useEffect, useRef, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { ArrowLeft, ArrowRight, KeyRound, Mail, RefreshCw } from 'lucide-react';
import { forgotPassword, resendForgotPasswordOTP, verifyForgotPasswordOTP } from '../../api/authService';
import { useToast } from '../../context/ToastContext';
import { Button } from '../../components/common/Button';
import { FormField } from '../../components/common/FormField';
import { Input } from '../../components/common/Input';
import { validateEmail, validateOTP } from '../../utils/validators';
import '../../styles/auth.css';

export function ForgotPasswordPage() {
  const navigate = useNavigate();
  const { showToast } = useToast();

  const [step, setStep] = useState('request'); // 'request' | 'verify'
  const [email, setEmail] = useState('');
  const [emailError, setEmailError] = useState(null);

  // 6-digit OTP state
  const [otpDigits, setOtpDigits] = useState(['', '', '', '', '', '']);
  const [otpError, setOtpError] = useState(null);
  const otpInputRefs = useRef([]);

  const [serverError, setServerError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [resendCooldown, setResendCooldown] = useState(0);

  // Timer countdown for resending OTP
  useEffect(() => {
    let timer;
    if (resendCooldown > 0) {
      timer = setTimeout(() => {
        setResendCooldown((prev) => prev - 1);
      }, 1000);
    }
    return () => clearTimeout(timer);
  }, [resendCooldown]);

  // Request OTP (Step 1)
  async function handleRequestOTP(e) {
    e.preventDefault();
    if (isSubmitting) return;

    setServerError('');
    const err = validateEmail(email);
    if (err) {
      setEmailError(err);
      return;
    }

    try {
      setIsSubmitting(true);
      await forgotPassword({ email: email.trim().toLowerCase() });
      showToast({
        title: 'Verification Code Sent',
        message: 'A 6-digit verification code was sent to your email address.',
        type: 'info',
      });
      setStep('verify');
      setResendCooldown(60);
      setOtpDigits(['', '', '', '', '', '']);
      // Focus first OTP input
      setTimeout(() => {
        otpInputRefs.current[0]?.focus();
      }, 100);
    } catch (err) {
      const msg = err.message || 'Unable to process your request. Please try again.';
      setServerError(msg);
      showToast({
        title: 'Request Failed',
        message: msg,
        type: 'error',
      });
    } finally {
      setIsSubmitting(false);
    }
  }

  // Handle individual OTP digit change
  function handleOtpChange(index, value) {
    // Only accept numeric input
    const cleanVal = value.replace(/[^0-9]/g, '');
    if (!cleanVal) {
      const next = [...otpDigits];
      next[index] = '';
      setOtpDigits(next);
      return;
    }

    // Single digit entry
    const char = cleanVal.slice(-1);
    const next = [...otpDigits];
    next[index] = char;
    setOtpDigits(next);
    setOtpError(null);
    setServerError('');

    // Advance to next input
    if (index < 5) {
      otpInputRefs.current[index + 1]?.focus();
    }
  }

  // Handle backspace navigation in OTP
  function handleOtpKeyDown(index, e) {
    if (e.key === 'Backspace' && !otpDigits[index] && index > 0) {
      otpInputRefs.current[index - 1]?.focus();
    } else if (e.key === 'ArrowLeft' && index > 0) {
      otpInputRefs.current[index - 1]?.focus();
    } else if (e.key === 'ArrowRight' && index < 5) {
      otpInputRefs.current[index + 1]?.focus();
    }
  }

  // Handle pasting full OTP
  function handleOtpPaste(e) {
    e.preventDefault();
    const pastedData = e.clipboardData.getData('text').trim().replace(/[^0-9]/g, '');
    if (!pastedData) return;

    const digits = pastedData.slice(0, 6).split('');
    const next = [...otpDigits];
    for (let i = 0; i < 6; i++) {
      next[i] = digits[i] || '';
    }
    setOtpDigits(next);
    setOtpError(null);
    setServerError('');

    const targetFocus = Math.min(digits.length, 5);
    otpInputRefs.current[targetFocus]?.focus();
  }

  // Resend OTP
  async function handleResendOTP() {
    if (resendCooldown > 0 || isSubmitting) return;

    try {
      setIsSubmitting(true);
      setServerError('');
      await resendForgotPasswordOTP({ email: email.trim().toLowerCase() });
      showToast({
        title: 'New Code Sent',
        message: 'A fresh verification code was sent to your email.',
        type: 'info',
      });
      setResendCooldown(60);
      setOtpDigits(['', '', '', '', '', '']);
      otpInputRefs.current[0]?.focus();
    } catch (err) {
      const msg = err.message || 'Failed to resend code. Please try again.';
      setServerError(msg);
      showToast({
        title: 'Resend Failed',
        message: msg,
        type: 'error',
      });
    } finally {
      setIsSubmitting(false);
    }
  }

  // Verify OTP (Step 2)
  async function handleVerifyOTP(e) {
    e.preventDefault();
    if (isSubmitting) return;

    const fullOtp = otpDigits.join('');
    const err = validateOTP(fullOtp);
    if (err) {
      setOtpError(err);
      return;
    }

    try {
      setIsSubmitting(true);
      setServerError('');
      // Backend returns 200 and sets HttpOnly 'password_reset' cookie
      await verifyForgotPasswordOTP({
        email: email.trim().toLowerCase(),
        otp: fullOtp,
      });

      showToast({
        title: 'Code Verified',
        message: 'Verification successful. Please enter your new password.',
        type: 'success',
      });

      // Proceed to reset password screen
      navigate('/reset-password', {
        replace: true,
        state: { email: email.trim().toLowerCase() },
      });
    } catch (err) {
      const msg = err.message || 'Invalid or expired verification code.';
      setServerError(msg);
      setOtpError('Invalid code');
      showToast({
        title: 'Verification Failed',
        message: msg,
        type: 'error',
      });
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className="auth-card" data-testid="forgot-password-card">
      {step === 'request' ? (
        <>
          <div className="auth-card-header">
            <h1 className="auth-card-title">Forgot Password</h1>
            <p className="auth-card-subtitle">
              Enter your registered email address to receive a 6-digit verification code
            </p>
          </div>

          {serverError && (
            <div className="auth-error-banner" role="alert">
              <span>{serverError}</span>
            </div>
          )}

          <form className="auth-form" onSubmit={handleRequestOTP} noValidate>
            <FormField
              id="forgot-email"
              label="EMAIL ADDRESS"
              required
              error={emailError}
            >
              <Input
                id="forgot-email"
                name="email"
                type="email"
                size="lg"
                placeholder="name@atelier-drape.com"
                value={email}
                onChange={(e) => {
                  setEmail(e.target.value);
                  setEmailError(null);
                  setServerError('');
                }}
                hasError={Boolean(emailError)}
                disabled={isSubmitting}
                leftIcon={<Mail size={16} />}
                autoComplete="email"
                required
              />
            </FormField>

            <Button
              type="submit"
              variant="primary"
              size="lg"
              fullWidth
              isLoading={isSubmitting}
              rightIcon={<ArrowRight size={14} />}
            >
              SEND OTP
            </Button>
          </form>

          <div className="auth-switch-link-wrapper">
            <Link to="/login" className="auth-switch-link">
              Back to Login
            </Link>
          </div>
        </>
      ) : (
        <>
          <div className="auth-card-header">
            <h1 className="auth-card-title">Enter Verification Code</h1>
            <p className="auth-card-subtitle">
              We sent a 6-digit code to <strong>{email}</strong>
            </p>
          </div>

          {serverError && (
            <div className="auth-error-banner" role="alert">
              <span>{serverError}</span>
            </div>
          )}

          <form className="auth-form" onSubmit={handleVerifyOTP} noValidate>
            <FormField
              id="otp-box-0"
              label="6-DIGIT VERIFICATION CODE"
              required
              error={otpError}
            >
              <div
                className="auth-otp-container"
                onPaste={handleOtpPaste}
                role="group"
                aria-label="Enter 6-digit verification code"
              >
                {otpDigits.map((digit, index) => (
                  <input
                    key={index}
                    ref={(el) => (otpInputRefs.current[index] = el)}
                    id={`otp-box-${index}`}
                    type="text"
                    inputMode="numeric"
                    maxLength={1}
                    value={digit}
                    onChange={(e) => handleOtpChange(index, e.target.value)}
                    onKeyDown={(e) => handleOtpKeyDown(index, e)}
                    className={`auth-otp-input ${otpError ? 'has-error' : ''}`}
                    disabled={isSubmitting}
                    aria-label={`Digit ${index + 1}`}
                  />
                ))}
              </div>
            </FormField>

            <div className="auth-resend-row">
              <span>Didn't receive the code?</span>
              {resendCooldown > 0 ? (
                <span>Resend in {resendCooldown}s</span>
              ) : (
                <button
                  type="button"
                  className="auth-resend-button"
                  onClick={handleResendOTP}
                  disabled={isSubmitting}
                >
                  Resend OTP
                </button>
              )}
            </div>

            <Button
              type="submit"
              variant="primary"
              size="lg"
              fullWidth
              isLoading={isSubmitting}
              rightIcon={<KeyRound size={14} />}
            >
              VERIFY OTP
            </Button>

            <button
              type="button"
              className="auth-resend-button"
              style={{ alignSelf: 'center', marginTop: '12px' }}
              onClick={() => {
                setStep('request');
                setServerError('');
                setOtpError(null);
              }}
              disabled={isSubmitting}
            >
              ← Use a different email
            </button>
          </form>
        </>
      )}
    </div>
  );
}
