-- Add email field to users table
ALTER TABLE users ADD COLUMN email VARCHAR(255) UNIQUE;
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);