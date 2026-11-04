-- Remove the original constraint that allows only one shipment per order.
ALTER TABLE shipments
DROP CONSTRAINT IF EXISTS shipments_order_id_key;


-- Add shipment type.
-- Existing shipments will automatically receive 'delivery'.
ALTER TABLE shipments
ADD COLUMN type VARCHAR(30) NOT NULL DEFAULT 'delivery';


-- Link return shipments to their return request.
ALTER TABLE shipments
ADD COLUMN return_request_id BIGINT;


-- Foreign key for return requests.
ALTER TABLE shipments
ADD CONSTRAINT fk_shipments_return_request
FOREIGN KEY (return_request_id)
REFERENCES return_requests(id)
ON DELETE RESTRICT;


-- Allow only valid shipment types.
ALTER TABLE shipments
ADD CONSTRAINT chk_shipments_type
CHECK (
    type IN (
        'delivery',
        'return',
        'return_to_customer'
    )
);


-- Enforce the relationship between shipment type and return request.
ALTER TABLE shipments
ADD CONSTRAINT chk_shipments_return_request
CHECK (
    (
        type = 'delivery'
        AND return_request_id IS NULL
    )
    OR
    (
        type IN ('return', 'return_to_customer')
        AND return_request_id IS NOT NULL
    )
);


-- Normal index for order lookups.
CREATE INDEX idx_shipments_order_id
ON shipments(order_id);


-- Index for return request lookups.
CREATE INDEX idx_shipments_return_request_id
ON shipments(return_request_id);


-- Only one delivery shipment per order.
CREATE UNIQUE INDEX idx_shipments_one_delivery_per_order
ON shipments(order_id)
WHERE type = 'delivery';


-- Only one shipment of each return type per return request.
CREATE UNIQUE INDEX idx_shipments_return_request_type
ON shipments(return_request_id, type)
WHERE return_request_id IS NOT NULL;