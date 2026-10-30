DROP INDEX IF EXISTS idx_wishlist_user_id;

ALTER TABLE wishlist
DROP COLUMN IF EXISTS name;

ALTER TABLE wishlist
ADD CONSTRAINT wishlist_user_id_key UNIQUE (user_id);

ALTER TABLE wishlist
DROP CONSTRAINT IF EXISTS uq_wishlist_user_id_name;