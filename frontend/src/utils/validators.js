/**
 * Customer Authentication & Form Validators
 * Exactly matches the backend validation in internal/delivery/http/validator/user.go
 */

export function validateName(name) {
  if (!name || typeof name !== 'string') {
    return 'Full name is required.';
  }
  const trimmed = name.trim();
  if (trimmed.length < 2 || trimmed.length > 30) {
    return 'Name must be between 2 and 30 characters.';
  }
  // Letters, spaces, hyphens, and apostrophes only
  const validNameRegex = /^[A-Za-z\s'-]+$/;
  if (!validNameRegex.test(trimmed)) {
    return 'Name can only contain letters, spaces, hyphens, and apostrophes.';
  }
  return null;
}

export function cleanPhoneNumber(phone) {
  if (!phone || typeof phone !== 'string') return '';
  let cleaned = phone.trim().replace(/\s+/g, '');
  if (cleaned.startsWith('+91')) {
    cleaned = cleaned.slice(3);
  }
  return cleaned;
}

export function validatePhone(phone) {
  if (!phone || typeof phone !== 'string') {
    return 'Phone number is required.';
  }
  const cleaned = cleanPhoneNumber(phone);
  if (!/^[0-9]{10}$/.test(cleaned)) {
    return 'Phone number must be exactly 10 digits.';
  }
  return null;
}

export function validateEmail(email) {
  if (!email || typeof email !== 'string') {
    return 'Email address is required.';
  }
  const trimmed = email.trim();
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (!emailRegex.test(trimmed)) {
    return 'Please enter a valid email address.';
  }
  return null;
}

export function checkPasswordRules(password = '') {
  const minLength = password.length >= 8;
  const hasUppercase = /[A-Z]/.test(password);
  const hasLowercase = /[a-z]/.test(password);
  const hasNumber = /[0-9]/.test(password);
  const hasSpecial = /[^A-Za-z0-9]/.test(password);

  const isValid = minLength && hasUppercase && hasLowercase && hasNumber && hasSpecial;

  return {
    minLength,
    hasUppercase,
    hasLowercase,
    hasNumber,
    hasSpecial,
    isValid,
  };
}

export function validatePassword(password) {
  if (!password || typeof password !== 'string') {
    return 'Password is required.';
  }
  const rules = checkPasswordRules(password);
  if (!rules.isValid) {
    return 'Password must be at least 8 characters and include uppercase, lowercase, number, and special character.';
  }
  return null;
}

export function validateConfirmPassword(password, confirmPassword) {
  if (!confirmPassword) {
    return 'Please confirm your password.';
  }
  if (password !== confirmPassword) {
    return 'Passwords do not match.';
  }
  return null;
}

export function validateOTP(otp) {
  if (!otp || typeof otp !== 'string') {
    return 'Verification code is required.';
  }
  const trimmed = otp.trim();
  if (!/^[0-9]{6}$/.test(trimmed)) {
    return 'Verification code must be exactly 6 digits.';
  }
  return null;
}
