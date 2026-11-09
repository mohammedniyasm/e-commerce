ALTER TABLE wallet_transactions
ADD COLUMN return_request_id BIGINT;

ALTER TABLE wallet_transactions
ADD CONSTRAINT fk_wallet_transactions_return_request
FOREIGN KEY (return_request_id)
REFERENCES return_requests(id)
ON DELETE RESTRICT;

CREATE UNIQUE INDEX idx_wallet_transactions_return_request_id
ON wallet_transactions(return_request_id);

ALTER TABLE return_requests
ADD COLUMN replacement_order_id UUID;

ALTER TABLE return_requests
ADD CONSTRAINT fk_return_requests_replacement_order
FOREIGN KEY (replacement_order_id)
REFERENCES orders(id)
ON DELETE RESTRICT;

CREATE UNIQUE INDEX idx_return_requests_replacement_order_id
ON return_requests(replacement_order_id);