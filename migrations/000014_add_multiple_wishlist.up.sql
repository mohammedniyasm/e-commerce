ALTER TABLE wishlist 
DROP CONSTRAINT IF EXISTS wishlist_user_id_key;

ALTER TABLE wishlist
ADD COLUMN name VARCHAR(100);

UPDATE wishlist
SET name = 'My Wishlist'
WHERE name IS NULL;

ALTER TABLE wishlist
ALTER COLUMN name SET NOT NULL;

CREATE INDEX idx_wishlist_user_id
ON wishlist(user_id);

ALTER TABLE wishlist
ADD CONSTRAINT uq_wishlist_user_id_name
UNIQUE (user_id, name);