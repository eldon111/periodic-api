-- Initial schema migration
-- Create scheduled_items table
CREATE TABLE IF NOT EXISTS scheduled_items (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    starts_at TIMESTAMP NOT NULL,
    repeats BOOLEAN NOT NULL,
    cron_expression TEXT,
    expiration TIMESTAMP,
    next_execution_at TIMESTAMP
);

-- Create index on next_execution_at for efficient querying
CREATE INDEX IF NOT EXISTS idx_scheduled_items_next_execution 
ON scheduled_items (next_execution_at) 
WHERE next_execution_at IS NOT NULL;

-- Create todo_items table  
CREATE TABLE IF NOT EXISTS todo_items (
    id SERIAL PRIMARY KEY,
    text TEXT NOT NULL,
    checked BOOLEAN NOT NULL DEFAULT FALSE
);

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    password_hash BYTEA NOT NULL
);

-- Grant permissions (for test environments)
DO $$ 
BEGIN
    IF EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'testuser') THEN
        GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO testuser;
        GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO testuser;
    END IF;
END $$;