import { useState } from 'react';
import googleIcon from '../../assets/google-icon.svg';
import { useAuth } from '../../context/AuthContext';
import { useToast } from '../../context/ToastContext';
import './GoogleAuthButton.css';

export function GoogleAuthButton({ referralCode = '', disabled = false, onSuccess }) {
  const { loginGoogle } = useAuth();
  const { showToast } = useToast();
  const [isLoading, setIsLoading] = useState(false);

  const googleClientId = import.meta.env.VITE_GOOGLE_CLIENT_ID;

  async function handleGoogleClick() {
    if (disabled || isLoading) return;

    if (!googleClientId) {
      showToast({
        title: 'Google Sign-In',
        message: 'Google Client ID is not configured. Set VITE_GOOGLE_CLIENT_ID in your environment.',
        type: 'warning',
      });
      return;
    }

    try {
      setIsLoading(true);
      // If Google Identity Services library is loaded on window
      if (window.google?.accounts?.id) {
        window.google.accounts.id.initialize({
          client_id: googleClientId,
          callback: async (response) => {
            try {
              const idToken = response.credential;
              await loginGoogle({ id_token: idToken, referral_code: referralCode || undefined });
              showToast({
                title: 'Welcome',
                message: 'Successfully signed in with Google.',
                type: 'success',
              });
              onSuccess?.();
            } catch (err) {
              showToast({
                title: 'Google Authentication Failed',
                message: err.message || 'Unable to sign in with Google.',
                type: 'error',
              });
            } finally {
              setIsLoading(false);
            }
          },
        });
        window.google.accounts.id.prompt();
      } else {
        showToast({
          title: 'Google Services Loading',
          message: 'Google Sign-In SDK is still initializing. Please try again.',
          type: 'info',
        });
        setIsLoading(false);
      }
    } catch (err) {
      showToast({
        title: 'Google Authentication Error',
        message: err.message || 'Failed to initiate Google sign in.',
        type: 'error',
      });
      setIsLoading(false);
    }
  }

  return (
    <button
      type="button"
      className="google-auth-button"
      onClick={handleGoogleClick}
      disabled={disabled || isLoading}
      aria-label="Continue with Google"
    >
      <img src={googleIcon} alt="" className="google-auth-icon" aria-hidden="true" />
      <span className="google-auth-text">
        {isLoading ? 'Connecting to Google...' : 'Continue with Google'}
      </span>
    </button>
  );
}
