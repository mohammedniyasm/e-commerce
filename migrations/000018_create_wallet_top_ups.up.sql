CREATE TABLE wallet_top_ups (
    id BIGSERIAL PRIMARY KEY,
    wallet_id BIGINT NOT NULL
        REFERENCES wallets(id) ON DELETE RESTRICT,
    amount NUMERIC(12,2) NOT NULL
        CHECK (amount > 0),
    gateway_order_id VARCHAR(255) NOT NULL UNIQUE,
    gateway_payment_id VARCHAR(255) UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'paid', 'failed')),
    paid_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_wallet_top_ups_wallet_created_at
    ON wallet_top_ups(wallet_id, created_at DESC);