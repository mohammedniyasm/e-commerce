import { useEffect, useRef, useState } from 'react';
import googleIcon from '../../assets/google-icon.svg';
import { useAuth } from '../../context/AuthContext';
import { useToast } from '../../context/ToastContext';
import './GoogleAuthButton.css';

let googleIdentityServicesPromise = null;
let initializedGoogleClientId = null;
let activeGoogleCredentialHandler = null;

function loadGoogleIdentityServices() {
  if (window.google?.accounts?.id) {
    return Promise.resolve(window.google.accounts.id);
  }

  if (!googleIdentityServicesPromise) {
    googleIdentityServicesPromise = new Promise((resolve, reject) => {
      const script = document.createElement('script');
      script.src = 'https://accounts.google.com/gsi/client';
      script.async = true;
      script.onload = () => {
        const googleIdentity = window.google?.accounts?.id;
        if (googleIdentity) {
          resolve(googleIdentity);
        } else {
          reject(new Error('Google Sign-In did not initialize.'));
        }
      };
      script.onerror = () => reject(new Error('Google Sign-In could not be loaded.'));
      document.head.appendChild(script);
    }).catch((error) => {
      googleIdentityServicesPromise = null;
      throw error;
    });
  }

  return googleIdentityServicesPromise;
}

export function GoogleAuthButton({ referralCode = '', disabled = false, onSuccess }) {
  const { loginGoogle } = useAuth();
  const { showToast } = useToast();
  const [isLoading, setIsLoading] = useState(false);
  const [sdkState, setSdkState] = useState('loading');
  const buttonContainerRef = useRef(null);
  const credentialHandlerRef = useRef(null);

  const googleClientId = import.meta.env.VITE_GOOGLE_CLIENT_ID;

  credentialHandlerRef.current = async (response) => {
    if (!response?.credential) {
      showToast({
        title: 'Google Authentication Failed',
        message: 'Google did not return a sign-in credential. Please try again.',
        type: 'error',
      });
      return;
    }

    try {
      setIsLoading(true);
      await loginGoogle({
        id_token: response.credential,
        referral_code: referralCode || undefined,
      });
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
  };

  useEffect(() => {
    let isMounted = true;
    const handleGoogleCredential = (response) => credentialHandlerRef.current?.(response);

    if (!googleClientId) {
      setSdkState('missing-client-id');
      return () => {
        isMounted = false;
      };
    }

    activeGoogleCredentialHandler = handleGoogleCredential;
    loadGoogleIdentityServices()
      .then((googleIdentity) => {
        if (!isMounted || !buttonContainerRef.current) return;

        if (initializedGoogleClientId !== googleClientId) {
          googleIdentity.initialize({
            client_id: googleClientId,
            callback: (response) => activeGoogleCredentialHandler?.(response),
          });
          initializedGoogleClientId = googleClientId;
        }
        const width = Math.min(
          400,
          Math.max(200, Math.floor(buttonContainerRef.current.getBoundingClientRect().width)),
        );
        googleIdentity.renderButton(buttonContainerRef.current, {
          theme: 'outline',
          size: 'large',
          text: 'continue_with',
          shape: 'rectangular',
          width,
        });
        setSdkState('ready');
      })
      .catch((err) => {
        console.error('Failed to load Google Sign-In:', err);
        if (isMounted) {
          setSdkState('unavailable');
          showToast({
            title: 'Google Sign-In Unavailable',
            message: 'Allow accounts.google.com in Brave Shields and reload the page to try again.',
            type: 'warning',
          });
        }
      });

    return () => {
      isMounted = false;
      if (activeGoogleCredentialHandler === handleGoogleCredential) {
        activeGoogleCredentialHandler = null;
      }
    };
  }, [googleClientId, showToast]);

  function handleFallbackClick() {
    if (sdkState === 'missing-client-id') {
      showToast({
        title: 'Google Sign-In',
        message: 'Google Client ID is not configured. Set VITE_GOOGLE_CLIENT_ID in your environment.',
        type: 'warning',
      });
    } else if (sdkState === 'unavailable') {
      showToast({
        title: 'Google Sign-In Unavailable',
        message: 'Allow accounts.google.com in Brave Shields and reload the page to try again.',
        type: 'warning',
      });
    }
  }

  return (
    <div className="google-auth-control" aria-busy={isLoading}>
      <div ref={buttonContainerRef} className="google-auth-sdk-button" />
      {sdkState !== 'ready' && (
        <button
          type="button"
          className="google-auth-button"
          onClick={handleFallbackClick}
          disabled={disabled || isLoading || sdkState === 'loading'}
          aria-label="Continue with Google"
        >
          <img src={googleIcon} alt="" className="google-auth-icon" aria-hidden="true" />
          <span className="google-auth-text">
            {isLoading ? 'Connecting to Google...' : 'Continue with Google'}
          </span>
        </button>
      )}
      {isLoading && <div className="google-auth-loading">Signing in...</div>}
    </div>
  );
}
