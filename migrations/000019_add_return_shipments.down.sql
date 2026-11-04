DROP INDEX IF EXISTS idx_shipments_return_request_type;

DROP INDEX IF EXISTS idx_shipments_one_delivery_per_order;

DROP INDEX IF EXISTS idx_shipments_return_request_id;

DROP INDEX IF EXISTS idx_shipments_order_id;


ALTER TABLE shipments
DROP CONSTRAINT IF EXISTS chk_shipments_return_request;

ALTER TABLE shipments
DROP CONSTRAINT IF EXISTS chk_shipments_type;

ALTER TABLE shipments
DROP CONSTRAINT IF EXISTS fk_shipments_return_request;


ALTER TABLE shipments
DROP COLUMN IF EXISTS return_request_id;

ALTER TABLE shipments
DROP COLUMN IF EXISTS type;


-- Restore original one-shipment-per-order constraint.
ALTER TABLE shipments
ADD CONSTRAINT shipments_order_id_key UNIQUE (order_id);