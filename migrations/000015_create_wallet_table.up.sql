CREATE TABLE wallets (
    id BIGINT PRIMARY KEY
        REFERENCES users(id) ON DELETE RESTRICT,
    balance NUMERIC(12,2) NOT NULL DEFAULT 0
        CHECK (balance >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE wallet_transactions (
    id BIGSERIAL PRIMARY KEY,
    wallet_id BIGINT NOT NULL
        REFERENCES wallets(id) ON DELETE RESTRICT,
    transaction_type VARCHAR(20) NOT NULL
        CHECK (transaction_type IN ('credit', 'debit')),
    source VARCHAR(30) NOT NULL
        CHECK (
            source IN (
                'order_payment',
                'refund',
                'referral_reward',
                'top_up',
                'admin_adjustment'
            )
        ),
    amount NUMERIC(12,2) NOT NULL
        CHECK (amount > 0),
    balance_before NUMERIC(12,2) NOT NULL
        CHECK (balance_before >= 0),
    balance_after NUMERIC(12,2) NOT NULL
        CHECK (balance_after >= 0),
    reference_id UUID,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_wallet_transaction_balance
        CHECK (
            (transaction_type = 'credit'
                AND balance_after = balance_before + amount)
            OR
            (transaction_type = 'debit'
                AND balance_after = balance_before - amount)
        )
);
-- Wallet transaction history lookup
CREATE INDEX idx_wallet_transactions_wallet_created_at
    ON wallet_transactions(wallet_id, created_at DESC);
-- Reference lookup
CREATE INDEX idx_wallet_transactions_reference_id
    ON wallet_transactions(reference_id);