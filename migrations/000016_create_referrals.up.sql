-- Add referral code to existing users.
ALTER TABLE users
ADD COLUMN referral_code VARCHAR(20);

-- Generate unique referral codes for existing users.
UPDATE users
SET referral_code = UPPER(
    SUBSTRING(REPLACE(gen_random_uuid()::TEXT, '-', ''), 1, 12)
);

ALTER TABLE users
ALTER COLUMN referral_code SET NOT NULL;

ALTER TABLE users
ADD CONSTRAINT uq_users_referral_code UNIQUE (referral_code);


-- Referral settings.
CREATE TABLE referral_settings (
    id BIGINT PRIMARY KEY CHECK (id = 1),

    referrer_reward NUMERIC(12,2) NOT NULL DEFAULT 0
        CHECK (referrer_reward >= 0),

    referred_reward NUMERIC(12,2) NOT NULL DEFAULT 0
        CHECK (referred_reward >= 0),

    minimum_order_amount NUMERIC(12,2) NOT NULL DEFAULT 0
        CHECK (minimum_order_amount >= 0),

    is_active BOOLEAN NOT NULL DEFAULT FALSE,

    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create the default settings row.
INSERT INTO referral_settings (id)
VALUES (1);


-- Referrals.
CREATE TABLE referrals (
    id BIGSERIAL PRIMARY KEY,
    referrer_id BIGINT NOT NULL
        REFERENCES users(id) ON DELETE RESTRICT,
    referred_user_id BIGINT NOT NULL UNIQUE
        REFERENCES users(id) ON DELETE RESTRICT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'completed')),
    referrer_reward NUMERIC(12,2) NOT NULL
        CHECK (referrer_reward >= 0),
    referred_reward NUMERIC(12,2) NOT NULL
        CHECK (referred_reward >= 0),
    minimum_order_amount NUMERIC(12,2) NOT NULL
        CHECK (minimum_order_amount >= 0),
    referrer_reward_status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (referrer_reward_status IN ('pending', 'completed')),
    referred_reward_status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (referred_reward_status IN ('pending', 'completed')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,

    CONSTRAINT chk_referral_no_self_referral
        CHECK (referrer_id <> referred_user_id)
);

CREATE INDEX idx_referrals_referrer_id
ON referrals(referrer_id);

CREATE INDEX idx_referrals_status
ON referrals(status);

ALTER TABLE wallet_transactions
ADD COLUMN referral_id BIGINT
    REFERENCES referrals(id) ON DELETE RESTRICT;

CREATE UNIQUE INDEX uq_wallet_transactions_referral_wallet
ON wallet_transactions (referral_id, wallet_id)
WHERE referral_id IS NOT NULL;