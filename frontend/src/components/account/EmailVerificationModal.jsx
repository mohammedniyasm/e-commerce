import React, { useState, useEffect, useRef } from 'react';
import { Mail, CheckCircle2, AlertCircle, RefreshCw, ShieldCheck } from 'lucide-react';
import { Modal } from '../common/Modal';
import { Button } from '../common/Button';
import { Input } from '../common/Input';
import { FormField } from '../common/FormField';
import { useToast } from '../../context/ToastContext';
import {
  sendVerificationOTP,
  verifyEmailOTP,
  resendVerificationOTP,
} from '../../api/authService';
import { validateOTP } from '../../utils/validators';
import './EmailVerificationModal.css';

export function EmailVerificationModal({
  isOpen,
  onClose,
  email = '',
  onVerified = () => {},
}) {
  const { showToast } = useToast();
  const [otp, setOtp] = useState('');
  const [otpError, setOtpError] = useState('');
  const [isSending, setIsSending] = useState(false);
  const [isVerifying, setIsVerifying] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);
  const [countdown, setCountdown] = useState(0);
  const [statusMessage, setStatusMessage] = useState('');
  const [hasSentInitially, setHasSentInitially] = useState(false);
  const timerRef = useRef(null);

  // Countdown timer effect
  useEffect(() => {
    if (countdown > 0) {
      timerRef.current = setTimeout(() => {
        setCountdown((prev) => prev - 1);
      }, 1000);
    }
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, [countdown]);

  // When modal opens, initialize and trigger initial OTP send
  useEffect(() => {
    if (isOpen) {
      setOtp('');
      setOtpError('');
      setIsSuccess(false);
      if (email && !hasSentInitially) {
        handleSendInitialOTP(email);
      }
    } else {
      setHasSentInitially(false);
      if (timerRef.current) clearTimeout(timerRef.current);
    }
  }, [isOpen, email]);

  async function handleSendInitialOTP(targetEmail) {
    if (!targetEmail) return;
    setIsSending(true);
    setStatusMessage('Sending 6-digit verification code...');
    setOtpError('');

    try {
      await sendVerificationOTP({ email: targetEmail });
      setHasSentInitially(true);
      setCountdown(60);
      setStatusMessage(`Verification code sent to ${targetEmail}. Please check your inbox.`);
      showToast({
        title: 'Verification Code Sent',
        message: `A 6-digit verification code was sent to ${targetEmail}.`,
        type: 'info',
      });
    } catch (err) {
      const msg = err.data?.message || err.message || 'Failed to send verification code.';
      if (msg.toLowerCase().includes('already verified')) {
        setIsSuccess(true);
        setStatusMessage('Your email is already verified.');
        onVerified();
        setTimeout(() => onClose(), 1500);
        return;
      }
      setStatusMessage('');
      setOtpError(msg);
    } finally {
      setIsSending(false);
    }
  }

  async function handleResendCode() {
    if (countdown > 0 || isSending || !email) return;
    setIsSending(true);
    setOtpError('');
    setStatusMessage('Resending verification code...');

    try {
      await resendVerificationOTP({ email });
      setCountdown(60);
      setStatusMessage(`A new 6-digit code has been sent to ${email}.`);
      showToast({
        title: 'Code Resent',
        message: 'A fresh verification code has been dispatched.',
        type: 'success',
      });
    } catch (err) {
      const msg = err.data?.message || err.message || 'Failed to resend code.';
      if (msg.toLowerCase().includes('already verified')) {
        setIsSuccess(true);
        setStatusMessage('Your email is already verified.');
        onVerified();
        setTimeout(() => onClose(), 1500);
        return;
      }
      setOtpError(msg);
      setStatusMessage('');
    } finally {
      setIsSending(false);
    }
  }

  async function handleVerifySubmit(e) {
    e.preventDefault();
    const cleanOtp = otp.trim();
    const validationError = validateOTP(cleanOtp);
    if (validationError) {
      setOtpError(validationError);
      return;
    }

    setOtpError('');
    setIsVerifying(true);

    try {
      await verifyEmailOTP({ otp: cleanOtp });
      setIsSuccess(true);
      setStatusMessage('Email verified successfully!');
      showToast({
        title: 'Email Verified',
        message: 'Your Drape account email has been verified successfully.',
        type: 'success',
      });
      onVerified();
      setTimeout(() => {
        onClose();
      }, 1500);
    } catch (err) {
      const msg = err.data?.message || err.message || 'Verification failed. Please check the code.';
      setOtpError(msg);
    } finally {
      setIsVerifying(false);
    }
  }

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={isSuccess ? 'Email Verified' : 'Verify Your Email'}
      eyebrow="ACCOUNT SECURITY"
      size="sm"
    >
      <div className="email-verify-modal" data-testid="email-verification-modal">
        {isSuccess ? (
          <div className="email-verify-modal__success" data-testid="email-verify-success">
            <div className="email-verify-modal__icon-wrap email-verify-modal__icon-wrap--success">
              <CheckCircle2 size={44} className="email-verify-modal__check-icon" />
            </div>
            <h3 className="email-verify-modal__title">Account Verified</h3>
            <p className="email-verify-modal__desc">
              Your email <strong>{email}</strong> is now verified. You now have full access to
              order updates and important account notifications.
            </p>
            <Button variant="primary" size="md" onClick={onClose} className="w-full">
              CONTINUE TO ACCOUNT
            </Button>
          </div>
        ) : (
          <form onSubmit={handleVerifySubmit} className="email-verify-modal__form">
            <div className="email-verify-modal__header-icon">
              <div className="email-verify-modal__icon-wrap">
                <ShieldCheck size={28} />
              </div>
            </div>

            <p className="email-verify-modal__desc">
              To complete your account security, enter the 6-digit code sent to:
            </p>

            <div className="email-verify-modal__email-pill">
              <Mail size={14} className="text-muted" />
              <span className="email-verify-modal__email-text">{email || 'your email address'}</span>
            </div>

            {statusMessage && !otpError && (
              <div className="email-verify-modal__notice" role="status">
                <span>{statusMessage}</span>
              </div>
            )}

            <div className="email-verify-modal__field-wrap">
              <FormField
                label="6-DIGIT VERIFICATION CODE"
                htmlFor="verify-email-otp"
                required
                error={otpError}
              >
                <Input
                  id="verify-email-otp"
                  type="text"
                  inputMode="numeric"
                  autoComplete="one-time-code"
                  maxLength={6}
                  placeholder="000000"
                  value={otp}
                  onChange={(e) => {
                    const val = e.target.value.replace(/\D/g, '').slice(0, 6);
                    setOtp(val);
                    if (otpError) setOtpError('');
                  }}
                  disabled={isVerifying || isSending}
                  className="email-verify-modal__otp-input"
                  data-testid="email-verify-otp-input"
                  autoFocus
                />
              </FormField>
            </div>

            <div className="email-verify-modal__resend-row">
              <span className="email-verify-modal__resend-text">Didn't receive the code?</span>
              <button
                type="button"
                className="email-verify-modal__resend-btn"
                onClick={handleResendCode}
                disabled={countdown > 0 || isSending}
                data-testid="email-verify-resend-btn"
              >
                {isSending ? (
                  <>
                    <RefreshCw size={13} className="spin" /> Sending...
                  </>
                ) : countdown > 0 ? (
                  `Resend code in ${countdown}s`
                ) : (
                  'Resend Code'
                )}
              </button>
            </div>

            <div className="email-verify-modal__actions">
              <Button
                type="button"
                variant="outline"
                size="md"
                onClick={onClose}
                disabled={isVerifying}
              >
                CANCEL
              </Button>
              <Button
                type="submit"
                variant="primary"
                size="md"
                loading={isVerifying}
                disabled={otp.trim().length !== 6 || isSending}
                data-testid="email-verify-submit-btn"
              >
                VERIFY EMAIL
              </Button>
            </div>
          </form>
        )}
      </div>
    </Modal>
  );
}
