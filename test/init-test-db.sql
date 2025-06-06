-- Initialize test database with required tables

-- Create scheduled_items table
CREATE TABLE IF NOT EXISTS scheduled_items (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    starts_at TIMESTAMP NOT NULL,
    repeats BOOLEAN NOT NULL,
    cron_expression TEXT,
    expiration TIMESTAMP
);

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

-- Grant permissions to test user
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO testuser;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO testuser;