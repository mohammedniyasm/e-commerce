DROP INDEX IF EXISTS uq_wallet_transactions_referral_wallet;

ALTER TABLE wallet_transactions
DROP COLUMN IF EXISTS referral_id;

DROP TABLE IF EXISTS referrals;

DROP TABLE IF EXISTS referral_settings;

ALTER TABLE users
DROP CONSTRAINT IF EXISTS uq_users_referral_code;

ALTER TABLE users
DROP COLUMN IF EXISTS referral_code;