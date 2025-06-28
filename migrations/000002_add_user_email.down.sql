-- Remove email field from users table
DROP INDEX IF EXISTS idx_users_email;
ALTER TABLE users DROP COLUMN email;