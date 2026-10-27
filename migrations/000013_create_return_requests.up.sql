CREATE TABLE return_requests (
    id BIGSERIAL PRIMARY KEY,

    order_id UUID NOT NULL
        REFERENCES orders(id) ON DELETE RESTRICT,

    order_item_id BIGINT NOT NULL
        REFERENCES order_items(id) ON DELETE RESTRICT,

    request_type VARCHAR(20) NOT NULL
        CHECK (request_type IN ('return', 'exchange')),

    requested_variant_id BIGINT
        REFERENCES product_variants(id) ON DELETE RESTRICT,

    quantity INT NOT NULL DEFAULT 1
        CHECK (quantity > 0),

    reason TEXT NOT NULL,

    status VARCHAR(30) NOT NULL DEFAULT 'requested'
        CHECK (
            status IN (
                'requested',
                'approved',
                'rejected',
                'pickup_scheduled',
                'in_transit',
                'received',
                'inspection_pending',
                'inspection_passed',
                'inspection_failed',
                'closed'
            )
        ),

    admin_note TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_return_request_variant
        CHECK (
            (request_type = 'return' AND requested_variant_id IS NULL)
            OR
            (request_type = 'exchange' AND requested_variant_id IS NOT NULL)
        )
);

CREATE INDEX idx_return_requests_order_id
    ON return_requests(order_id);

CREATE INDEX idx_return_requests_order_item_id
    ON return_requests(order_item_id);

CREATE INDEX idx_return_requests_status
    ON return_requests(status);
