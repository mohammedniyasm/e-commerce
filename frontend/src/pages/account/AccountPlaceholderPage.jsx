import React from 'react';
import { Link } from 'react-router-dom';
import { ArrowLeft } from 'lucide-react';
import { Button } from '../../components/common/Button';
import './AccountOverviewPage.css';

export function AccountPlaceholderPage({ title, milestone }) {
  return (
    <div className="account-card" data-testid={`account-placeholder-${title.toLowerCase().replace(/[^a-z0-9]+/g, '-')}`}>
      <div className="account-card__header">
        <h2 className="account-card__title">{title}</h2>
        <span className="account-detail-label">{milestone}</span>
      </div>
      <div className="account-card__body">
        <p className="account-card__desc">
          This section is being crafted as part of {milestone}. All backend contracts and Figma
          designs are aligned and ready for implementation.
        </p>
        <Link to="/account">
          <Button variant="outline" size="sm" leftIcon={<ArrowLeft size={14} />}>
            BACK TO OVERVIEW
          </Button>
        </Link>
      </div>
    </div>
  );
}
