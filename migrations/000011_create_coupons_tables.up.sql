-- 1. Create coupon discount type enum
CREATE TYPE coupon_type AS ENUM (
    'percentage',
    'fixed'
);

-- 2. Create coupons table
CREATE TABLE coupons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    code VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,

    discount_type coupon_type NOT NULL,
    discount_value NUMERIC(12,2) NOT NULL,
    minimum_order_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    maximum_discount NUMERIC(12,2),

    usage_limit INTEGER,
    usage_limit_per_user INTEGER,

    start_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,

    is_active BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_coupons_discount_value
        CHECK (discount_value > 0),

    CONSTRAINT chk_coupons_minimum_order_amount
        CHECK (minimum_order_amount >= 0),

    CONSTRAINT chk_coupons_maximum_discount
        CHECK (maximum_discount IS NULL OR maximum_discount > 0),

    CONSTRAINT chk_coupons_usage_limit
        CHECK (usage_limit IS NULL OR usage_limit > 0),

    CONSTRAINT chk_coupons_usage_limits_consistent
        CHECK (
            (usage_limit IS NULL )
            OR
            (
                usage_limit IS NOT NULL
            )
        ),

    CONSTRAINT chk_coupons_usage_limit_per_user
        CHECK (
            usage_limit_per_user IS NULL
            OR usage_limit_per_user > 0
        ),

    CONSTRAINT chk_coupons_valid_period
        CHECK (expires_at > start_at)
);

-- 3. Create coupon usage table
CREATE TABLE coupon_usage (
    id BIGSERIAL PRIMARY KEY,

    coupon_id UUID NOT NULL,
    user_id BIGINT NOT NULL,
    order_id UUID NOT NULL UNIQUE,

    discount_amount NUMERIC(12,2) NOT NULL,
    used_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT fk_coupon_usage_coupon
        FOREIGN KEY (coupon_id)
        REFERENCES coupons(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_coupon_usage_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_coupon_usage_order
        FOREIGN KEY (order_id)
        REFERENCES orders(id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_coupon_usage_discount
        CHECK (discount_amount >= 0)
);

-- 4. Create indexes
CREATE INDEX coupon_usage_coupon_id_idx
    ON coupon_usage(coupon_id);

CREATE INDEX coupon_usage_user_id_idx
    ON coupon_usage(user_id);
