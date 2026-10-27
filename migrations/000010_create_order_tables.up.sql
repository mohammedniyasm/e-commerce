-- Order and payment status enums

CREATE TYPE order_status AS ENUM (
    'pending',
    'confirmed',
    'processing',
    'shipped',
    'delivered',
    'cancelled',
    'returned'
);

CREATE TYPE payment_status AS ENUM (
    'pending',
    'paid',
    'failed',
    'refunded',
    'partially_refunded'
);

CREATE TYPE shipping_status AS ENUM (
    'pending',
    'packed',
    'ready_to_ship',
    'in_transit',
    'out_for_delivery',
    'delivered',
    'failed',
    'returned',
    'cancelled'
);

-- Orders

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    order_number VARCHAR(255) NOT NULL UNIQUE,
    subtotal NUMERIC(12,2) NOT NULL,
    discount_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    shipping_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    total_amount NUMERIC(12,2) NOT NULL,
    order_status order_status NOT NULL,
    payment_status payment_status NOT NULL,
    payment_deadline TIMESTAMPTZ,
    placed_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_orders_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_orders_user_id
    ON orders(user_id);


-- Order addresses

CREATE TABLE order_addresses (
    id BIGSERIAL PRIMARY KEY,
    order_id UUID NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    address_line1 VARCHAR(255) NOT NULL,
    address_line2 VARCHAR(255),
    city VARCHAR(255) NOT NULL,
    state VARCHAR(255) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    country VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_order_addresses_order
        FOREIGN KEY (order_id)
        REFERENCES orders(id)
        ON DELETE RESTRICT
);


-- Order items

CREATE TABLE order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id UUID NOT NULL,
    variant_id BIGINT NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    image_url VARCHAR(500) NOT NULL,
    variant_info JSONB NOT NULL,
    unit_price NUMERIC(12,2) NOT NULL,
    quantity INTEGER NOT NULL,
    subtotal NUMERIC(12,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_order_items_quantity_positive
    CHECK (quantity > 0),
    
    CONSTRAINT fk_order_items_order
        FOREIGN KEY (order_id)
        REFERENCES orders(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_order_items_variant
        FOREIGN KEY (variant_id)
        REFERENCES product_variants(id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_order_items_order_id
    ON order_items(order_id);


-- Order status history

CREATE TABLE order_status_history (
    id BIGSERIAL PRIMARY KEY,
    order_id UUID NOT NULL,
    status order_status NOT NULL,
    note TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_order_status_history_order
        FOREIGN KEY (order_id)
        REFERENCES orders(id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_order_status_history_order_id
    ON order_status_history(order_id);


-- Payments

CREATE TABLE payments (
    id BIGSERIAL PRIMARY KEY,
    order_id UUID NOT NULL,
    payment_method VARCHAR(30) NOT NULL,
    payment_status payment_status NOT NULL,
    gateway_order_id VARCHAR(255),
    transaction_id VARCHAR(255) UNIQUE,
    gateway_response JSONB,
    failure_reason VARCHAR(255),
    amount NUMERIC(12,2) NOT NULL,
    paid_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_payments_order
        FOREIGN KEY (order_id)
        REFERENCES orders(id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_payments_order_id
    ON payments(order_id);

CREATE INDEX idx_payments_gateway_order_id
    ON payments(gateway_order_id);

-- Shipments

CREATE TABLE shipments (
    id BIGSERIAL PRIMARY KEY,
    order_id UUID NOT NULL UNIQUE,
    courier_name VARCHAR(255),
    tracking_number VARCHAR(255),
    shipping_status shipping_status NOT NULL DEFAULT 'pending',
    shipped_at TIMESTAMP,
    delivered_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_shipments_order
        FOREIGN KEY (order_id)
        REFERENCES orders(id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_shipments_tracking_number
    ON shipments(tracking_number);

-- Shipment tracking history

CREATE TABLE shipment_tracking_history (
    id BIGSERIAL PRIMARY KEY,
    shipment_id BIGINT NOT NULL,
    shipping_status shipping_status NOT NULL,
    description TEXT,
    location VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_shipment_tracking_history_shipment
        FOREIGN KEY (shipment_id)
        REFERENCES shipments(id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_shipment_tracking_history_shipment_id
    ON shipment_tracking_history(shipment_id);