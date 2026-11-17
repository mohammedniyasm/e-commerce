import React, { useState, useRef } from 'react';
import { useOutletContext } from 'react-router-dom';
import {
  User,
  Mail,
  Lock,
  Camera,
  Trash2,
  CheckCircle2,
  AlertCircle,
  ShieldCheck,
  KeyRound,
} from 'lucide-react';
import {
  FormField,
  Input,
  Button,
  Badge,
  Modal,
} from '../../components/common';
import { EmailVerificationModal } from '../../components/account/EmailVerificationModal';
import { useToast } from '../../context/ToastContext';
import {
  updateProfile,
  changePassword,
  sendEmailChangeOTP,
  verifyEmailChangeOTP,
  uploadProfileImage,
  deleteProfileImage,
} from '../../api/profileService';
import {
  validateName,
  validatePhone,
  validateEmail,
  checkPasswordRules,
  validateConfirmPassword,
  validateOTP,
} from '../../utils/validators';
import './ProfilePage.css';

export function ProfilePage() {
  const context = useOutletContext() || {};
  const { profile, refreshData = () => {}, openVerifyEmailModal } = context;
  const { showToast } = useToast();
  const [isLocalVerifyModalOpen, setIsLocalVerifyModalOpen] = useState(false);

  const handleOpenVerifyModal = () => {
    if (openVerifyEmailModal) {
      openVerifyEmailModal();
    } else {
      setIsLocalVerifyModalOpen(true);
    }
  };

  // Personal Information state
  const [name, setName] = useState(profile?.name || '');
  const [phone, setPhone] = useState(profile?.phone || '');
  const [personalErrors, setPersonalErrors] = useState({});
  const [isUpdatingPersonal, setIsUpdatingPersonal] = useState(false);

  // Profile image upload state
  const fileInputRef = useRef(null);
  const [isUploadingImage, setIsUploadingImage] = useState(false);
  const [isDeletingImage, setIsDeletingImage] = useState(false);

  // Email Change Modal state
  const [isEmailModalOpen, setIsEmailModalOpen] = useState(false);
  const [emailStep, setEmailStep] = useState('request'); // 'request' | 'verify'
  const [newEmail, setNewEmail] = useState('');
  const [emailOtp, setEmailOtp] = useState('');
  const [emailModalError, setEmailModalError] = useState('');
  const [isEmailSubmitting, setIsEmailSubmitting] = useState(false);

  // Password Security state
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [passwordErrors, setPasswordErrors] = useState({});
  const [isUpdatingPassword, setIsUpdatingPassword] = useState(false);

  // Sync state if profile changes
  React.useEffect(() => {
    if (profile?.name) setName(profile.name);
    if (profile?.phone) setPhone(profile.phone);
  }, [profile]);

  // Handlers for Personal Details
  async function handlePersonalSubmit(e) {
    e.preventDefault();
    const errors = {};
    const nameErr = validateName(name);
    if (nameErr) errors.name = nameErr;
    const phoneErr = validatePhone(phone);
    if (phoneErr) errors.phone = phoneErr;

    if (Object.keys(errors).length > 0) {
      setPersonalErrors(errors);
      return;
    }
    setPersonalErrors({});
    setIsUpdatingPersonal(true);

    try {
      await updateProfile({ name: name.trim(), phone: phone.trim() });
      await refreshData();
      showToast({
        title: 'Profile Updated',
        message: 'Your personal information has been saved successfully.',
        type: 'success',
      });
    } catch (err) {
      showToast({
        title: 'Update Failed',
        message: err.message || 'Unable to update profile. Please try again.',
        type: 'danger',
      });
    } finally {
      setIsUpdatingPersonal(false);
    }
  }

  // Handlers for Profile Image
  async function handleImageFileChange(e) {
    const file = e.target.files?.[0];
    if (!file) return;

    if (!file.type.startsWith('image/')) {
      showToast({
        title: 'Invalid File',
        message: 'Please select an image file (JPEG, PNG, WEBP).',
        type: 'warning',
      });
      return;
    }

    if (file.size > 5 * 1024 * 1024) {
      showToast({
        title: 'File Too Large',
        message: 'Profile image size must not exceed 5MB.',
        type: 'warning',
      });
      return;
    }

    setIsUploadingImage(true);
    try {
      await uploadProfileImage(file);
      await refreshData();
      showToast({
        title: 'Photo Uploaded',
        message: 'Your profile picture has been updated.',
        type: 'success',
      });
    } catch (err) {
      showToast({
        title: 'Upload Failed',
        message: err.message || 'Failed to upload image. Please try again.',
        type: 'danger',
      });
    } finally {
      setIsUploadingImage(false);
      if (fileInputRef.current) fileInputRef.current.value = '';
    }
  }

  async function handleDeleteImage() {
    setIsDeletingImage(true);
    try {
      await deleteProfileImage();
      await refreshData();
      showToast({
        title: 'Photo Removed',
        message: 'Your profile picture has been removed.',
        type: 'info',
      });
    } catch (err) {
      showToast({
        title: 'Action Failed',
        message: err.message || 'Failed to remove image.',
        type: 'danger',
      });
    } finally {
      setIsDeletingImage(false);
    }
  }

  // Handlers for Email Change Flow
  function handleOpenEmailModal() {
    setEmailStep('request');
    setNewEmail('');
    setEmailOtp('');
    setEmailModalError('');
    setIsEmailModalOpen(true);
  }

  async function handleSendEmailOTP(e) {
    e.preventDefault();
    const emailErr = validateEmail(newEmail);
    if (emailErr) {
      setEmailModalError(emailErr);
      return;
    }
    setEmailModalError('');
    setIsEmailSubmitting(true);

    try {
      await sendEmailChangeOTP({ new_email: newEmail.trim() });
      setEmailStep('verify');
      showToast({
        title: 'OTP Dispatched',
        message: `A 6-digit verification code was sent to ${newEmail.trim()}.`,
        type: 'info',
      });
    } catch (err) {
      setEmailModalError(err.message || 'Failed to send verification code.');
    } finally {
      setIsEmailSubmitting(false);
    }
  }

  async function handleVerifyEmailOTP(e) {
    e.preventDefault();
    const otpErr = validateOTP(emailOtp);
    if (otpErr) {
      setEmailModalError(otpErr);
      return;
    }
    setEmailModalError('');
    setIsEmailSubmitting(true);

    try {
      await verifyEmailChangeOTP({ otp: emailOtp.trim() });
      await refreshData();
      setIsEmailModalOpen(false);
      showToast({
        title: 'Email Updated',
        message: 'Your account email address has been changed successfully.',
        type: 'success',
      });
    } catch (err) {
      setEmailModalError(err.message || 'Invalid or expired verification code.');
    } finally {
      setIsEmailSubmitting(false);
    }
  }

  // Handlers for Password Update
  async function handlePasswordSubmit(e) {
    e.preventDefault();
    const errors = {};
    if (!currentPassword) {
      errors.currentPassword = 'Enter your current password';
    }

    const rules = checkPasswordRules(newPassword);
    if (!rules.isValid) {
      errors.newPassword = 'Password must meet all complexity requirements';
    }

    const confirmErr = validateConfirmPassword(newPassword, confirmPassword);
    if (confirmErr) {
      errors.confirmPassword = confirmErr;
    }

    if (Object.keys(errors).length > 0) {
      setPasswordErrors(errors);
      return;
    }

    setPasswordErrors({});
    setIsUpdatingPassword(true);

    try {
      await changePassword({
        current_password: currentPassword,
        new_password: newPassword,
      });
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
      showToast({
        title: 'Password Updated',
        message: 'Your atelier account password has been changed securely.',
        type: 'success',
      });
    } catch (err) {
      showToast({
        title: 'Update Failed',
        message: err.message || 'Current password incorrect or validation error.',
        type: 'danger',
      });
    } finally {
      setIsUpdatingPassword(false);
    }
  }

  const passwordRules = checkPasswordRules(newPassword);

  return (
    <div className="account-profile-page" data-testid="account-profile-page">
      {/* SECTION 1: Profile Photo / Avatar */}
      <section className="account-card profile-photo-card" aria-labelledby="photo-heading">
        <div className="account-card__header">
          <div className="account-card__title-group">
            <Camera size={18} className="account-card__icon" aria-hidden="true" />
            <h2 id="photo-heading" className="account-card__title">
              Profile Portrait
            </h2>
          </div>
        </div>

        <div className="account-card__body profile-photo-body">
          <div className="profile-photo-preview" aria-hidden="true">
            {profile?.profile_image ? (
              <img
                src={profile.profile_image}
                alt={profile.name || 'Profile'}
                className="profile-photo-img"
              />
            ) : (
              <div className="profile-photo-placeholder">
                <User size={32} />
              </div>
            )}
          </div>

          <div className="profile-photo-actions">
            <input
              type="file"
              ref={fileInputRef}
              onChange={handleImageFileChange}
              accept="image/*"
              style={{ display: 'none' }}
              aria-label="Upload profile image"
            />
            <Button
              variant="outline"
              size="sm"
              onClick={() => fileInputRef.current?.click()}
              isLoading={isUploadingImage}
              data-testid="upload-photo-btn"
            >
              UPLOAD NEW PHOTO
            </Button>
            {profile?.profile_image && (
              <Button
                variant="ghost"
                size="sm"
                onClick={handleDeleteImage}
                isLoading={isDeletingImage}
                leftIcon={<Trash2 size={14} />}
                data-testid="remove-photo-btn"
              >
                REMOVE
              </Button>
            )}
            <p className="profile-photo-hint">
              Accepts PNG, JPG or WEBP. Maximum file size: 5 MB.
            </p>
          </div>
        </div>
      </section>

      {/* SECTION 2: Personal Details */}
      <section className="account-card" aria-labelledby="personal-heading">
        <div className="account-card__header">
          <div className="account-card__title-group">
            <User size={18} className="account-card__icon" aria-hidden="true" />
            <h2 id="personal-heading" className="account-card__title">
              Personal Information
            </h2>
          </div>
        </div>

        <div className="account-card__body">
          <form onSubmit={handlePersonalSubmit} className="profile-form" noValidate>
            <div className="profile-form-grid">
              <FormField
                label="FULL NAME"
                htmlFor="profile-name"
                required
                error={personalErrors.name}
              >
                <Input
                  id="profile-name"
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g. Arjun Nair"
                  hasError={Boolean(personalErrors.name)}
                  data-testid="profile-input-name"
                />
              </FormField>

              <FormField
                label="PHONE NUMBER"
                htmlFor="profile-phone"
                required
                error={personalErrors.phone}
              >
                <Input
                  id="profile-phone"
                  type="tel"
                  value={phone}
                  onChange={(e) => setPhone(e.target.value)}
                  placeholder="+91 98470 12345"
                  hasError={Boolean(personalErrors.phone)}
                  data-testid="profile-input-phone"
                />
              </FormField>
            </div>

            <div className="profile-form-actions">
              <Button
                type="submit"
                variant="primary"
                size="md"
                isLoading={isUpdatingPersonal}
                data-testid="save-personal-btn"
              >
                SAVE CHANGES
              </Button>
            </div>
          </form>
        </div>
      </section>

      {/* SECTION 3: Email Address & Verification */}
      <section className="account-card" aria-labelledby="email-heading">
        <div className="account-card__header">
          <div className="account-card__title-group">
            <Mail size={18} className="account-card__icon" aria-hidden="true" />
            <h2 id="email-heading" className="account-card__title">
              Email Address
            </h2>
          </div>
          <div className="account-card__header-actions">
            {!profile?.email_verified_at && (
              <Button
                variant="primary"
                size="sm"
                onClick={handleOpenVerifyModal}
                leftIcon={<ShieldCheck size={14} />}
                data-testid="profile-verify-email-btn"
              >
                VERIFY EMAIL
              </Button>
            )}
            <Button
              variant="outline"
              size="sm"
              onClick={handleOpenEmailModal}
              data-testid="change-email-btn"
            >
              CHANGE EMAIL
            </Button>
          </div>
        </div>

        <div className="account-card__body">
          <div className="profile-email-row">
            <div className="profile-email-info">
              <span className="profile-email-address" data-testid="current-email-text">
                {profile?.email || '—'}
              </span>
              {profile?.email_verified_at ? (
                <Badge variant="success" size="sm">
                  VERIFIED
                </Badge>
              ) : (
                <Badge variant="warning" size="sm">
                  UNVERIFIED
                </Badge>
              )}
            </div>
            <p className="profile-email-desc">
              Your email is used for order confirmations, delivery tracking notifications, and atelier security updates.
            </p>

            {!profile?.email_verified_at && (
              <div className="profile-unverified-alert" data-testid="profile-unverified-alert">
                <div className="profile-unverified-alert__icon">
                  <AlertCircle size={20} />
                </div>
                <div className="profile-unverified-alert__content">
                  <h4 className="profile-unverified-alert__title">Email Not Verified</h4>
                  <p className="profile-unverified-alert__text">
                    Your account email <strong>{profile?.email}</strong> is unverified. Verify now to secure your account and receive order delivery notifications.
                  </p>
                </div>
                <Button
                  variant="primary"
                  size="sm"
                  onClick={handleOpenVerifyModal}
                  data-testid="profile-verify-email-now-btn"
                >
                  VERIFY EMAIL NOW
                </Button>
              </div>
            )}
          </div>
        </div>
      </section>

      {/* SECTION 4: Password Security */}
      <section className="account-card" aria-labelledby="security-heading">
        <div className="account-card__header">
          <div className="account-card__title-group">
            <Lock size={18} className="account-card__icon" aria-hidden="true" />
            <h2 id="security-heading" className="account-card__title">
              Password & Security
            </h2>
          </div>
        </div>

        <div className="account-card__body">
          <form onSubmit={handlePasswordSubmit} className="profile-form" noValidate>
            <FormField
              label="CURRENT PASSWORD"
              htmlFor="current-password"
              required
              error={passwordErrors.currentPassword}
            >
              <Input
                id="current-password"
                type="password"
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                placeholder="Enter current password"
                hasError={Boolean(passwordErrors.currentPassword)}
                data-testid="current-password-input"
              />
            </FormField>

            <div className="profile-form-grid">
              <FormField
                label="NEW PASSWORD"
                htmlFor="new-password"
                required
                error={passwordErrors.newPassword}
              >
                <Input
                  id="new-password"
                  type="password"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  placeholder="Minimum 8 characters"
                  hasError={Boolean(passwordErrors.newPassword)}
                  data-testid="new-password-input"
                />
              </FormField>

              <FormField
                label="CONFIRM NEW PASSWORD"
                htmlFor="confirm-new-password"
                required
                error={passwordErrors.confirmPassword}
              >
                <Input
                  id="confirm-new-password"
                  type="password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  placeholder="Re-enter new password"
                  hasError={Boolean(passwordErrors.confirmPassword)}
                  data-testid="confirm-password-input"
                />
              </FormField>
            </div>

            {/* Password Complexity Checklist */}
            <div className="password-checklist" role="region" aria-label="Password Requirements">
              <span className="password-checklist__title">SECURITY REQUIREMENTS:</span>
              <ul className="password-checklist__list">
                <li className={passwordRules.hasMinLength ? 'is-valid' : ''}>
                  {passwordRules.hasMinLength ? <CheckCircle2 size={12} /> : <span className="bullet">•</span>}
                  At least 8 characters
                </li>
                <li className={passwordRules.hasUpperCase ? 'is-valid' : ''}>
                  {passwordRules.hasUpperCase ? <CheckCircle2 size={12} /> : <span className="bullet">•</span>}
                  One uppercase letter (A-Z)
                </li>
                <li className={passwordRules.hasLowerCase ? 'is-valid' : ''}>
                  {passwordRules.hasLowerCase ? <CheckCircle2 size={12} /> : <span className="bullet">•</span>}
                  One lowercase letter (a-z)
                </li>
                <li className={passwordRules.hasDigit ? 'is-valid' : ''}>
                  {passwordRules.hasDigit ? <CheckCircle2 size={12} /> : <span className="bullet">•</span>}
                  One number (0-9)
                </li>
                <li className={passwordRules.hasSpecialChar ? 'is-valid' : ''}>
                  {passwordRules.hasSpecialChar ? <CheckCircle2 size={12} /> : <span className="bullet">•</span>}
                  One special character (@, #, $, %, etc.)
                </li>
              </ul>
            </div>

            <div className="profile-form-actions">
              <Button
                type="submit"
                variant="primary"
                size="md"
                isLoading={isUpdatingPassword}
                data-testid="update-password-btn"
              >
                UPDATE PASSWORD
              </Button>
            </div>
          </form>
        </div>
      </section>

      {/* Email Change Modal */}
      <Modal
        isOpen={isEmailModalOpen}
        onClose={() => setIsEmailModalOpen(false)}
        title={emailStep === 'request' ? 'Change Account Email' : 'Verify Email Code'}
        eyebrow="SECURITY VERIFICATION"
        size="sm"
      >
        <div className="email-change-modal" data-testid="email-change-modal">
          {emailModalError && (
            <div className="email-change-error" role="alert">
              <AlertCircle size={14} />
              <span>{emailModalError}</span>
            </div>
          )}

          {emailStep === 'request' ? (
            <form onSubmit={handleSendEmailOTP} noValidate>
              <p className="email-change-desc">
                Enter your new email address. We will send a 6-digit one-time password to verify ownership.
              </p>
              <FormField label="NEW EMAIL ADDRESS" htmlFor="modal-new-email" required>
                <Input
                  id="modal-new-email"
                  type="email"
                  value={newEmail}
                  onChange={(e) => setNewEmail(e.target.value)}
                  placeholder="name@example.com"
                  autoFocus
                  data-testid="new-email-input"
                />
              </FormField>
              <div className="email-change-actions">
                <Button
                  type="button"
                  variant="outline"
                  size="md"
                  onClick={() => setIsEmailModalOpen(false)}
                >
                  CANCEL
                </Button>
                <Button
                  type="submit"
                  variant="primary"
                  size="md"
                  isLoading={isEmailSubmitting}
                  data-testid="send-otp-btn"
                >
                  SEND CODE
                </Button>
              </div>
            </form>
          ) : (
            <form onSubmit={handleVerifyEmailOTP} noValidate>
              <p className="email-change-desc">
                Enter the 6-digit code sent to <strong>{newEmail}</strong>.
              </p>
              <FormField label="6-DIGIT VERIFICATION CODE" htmlFor="modal-email-otp" required>
                <Input
                  id="modal-email-otp"
                  type="text"
                  maxLength={6}
                  value={emailOtp}
                  onChange={(e) => setEmailOtp(e.target.value.replace(/\D/g, ''))}
                  placeholder="000000"
                  autoFocus
                  data-testid="email-otp-input"
                />
              </FormField>
              <div className="email-change-actions">
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => setEmailStep('request')}
                >
                  BACK
                </Button>
                <Button
                  type="submit"
                  variant="primary"
                  size="md"
                  isLoading={isEmailSubmitting}
                  data-testid="confirm-email-otp-btn"
                >
                  VERIFY & UPDATE
                </Button>
              </div>
            </form>
          )}
        </div>
      </Modal>

      {/* Local fallback Email Verification Modal */}
      <EmailVerificationModal
        isOpen={isLocalVerifyModalOpen}
        onClose={() => setIsLocalVerifyModalOpen(false)}
        email={profile?.email}
        onVerified={refreshData}
      />
    </div>
  );
}
