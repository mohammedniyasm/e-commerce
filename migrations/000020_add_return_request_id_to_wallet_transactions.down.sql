DROP INDEX IF EXISTS idx_wallet_transactions_return_request_id;

ALTER TABLE wallet_transactions
DROP CONSTRAINT IF EXISTS fk_wallet_transactions_return_request;

ALTER TABLE wallet_transactions
DROP COLUMN IF EXISTS return_request_id;

DROP INDEX IF EXISTS idx_return_requests_replacement_order_id;

ALTER TABLE return_requests
DROP CONSTRAINT IF EXISTS fk_return_requests_replacement_order;

ALTER TABLE return_requests
DROP COLUMN IF EXISTS replacement_order_id;