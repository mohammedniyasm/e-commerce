import { CheckCircle2 } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { Button } from '../../components/common/Button';
import '../../styles/auth.css';

export function PasswordResetSuccessPage() {
  const navigate = useNavigate();

  return (
    <div className="auth-card" data-testid="password-reset-success-card">
      <div className="auth-success-card">
        <div className="auth-success-icon-wrapper" aria-hidden="true">
          <CheckCircle2 size={36} color="#166534" />
        </div>

        <div className="auth-card-header" style={{ marginBottom: '24px' }}>
          <h1 className="auth-card-title">Password Reset Successful</h1>
          <p className="auth-card-subtitle">
            Your password has been updated securely. You can now sign in to your Drape account with your new credentials.
          </p>
        </div>

        <Button
          type="button"
          variant="primary"
          size="lg"
          fullWidth
          onClick={() => navigate('/login', { replace: true })}
        >
          BACK TO LOGIN
        </Button>
      </div>
    </div>
  );
}
