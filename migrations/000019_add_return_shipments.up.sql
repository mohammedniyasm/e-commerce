-- Allow return_to_customer as a valid shipment type.
ALTER TABLE shipments
DROP CONSTRAINT IF EXISTS chk_shipments_type;

ALTER TABLE shipments
ADD CONSTRAINT chk_shipments_type
CHECK (
    type IN (
        'delivery',
        'return',
        'return_to_customer'
    )
);

ALTER TABLE shipments
DROP CONSTRAINT IF EXISTS chk_shipments_return_request;

ALTER TABLE shipments
ADD CONSTRAINT chk_shipments_return_request
CHECK (
    (type = 'delivery' AND return_request_id IS NULL)
    OR
    (
        type IN ('return', 'return_to_customer')
        AND return_request_id IS NOT NULL
    )
);

CREATE UNIQUE INDEX idx_shipments_return_request_type
ON shipments (return_request_id, type)
WHERE return_request_id IS NOT NULL;