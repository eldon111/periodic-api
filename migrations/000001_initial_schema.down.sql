-- Rollback initial schema migration
DROP INDEX IF EXISTS idx_scheduled_items_next_execution;
DROP TABLE IF EXISTS scheduled_items;
DROP TABLE IF EXISTS todo_items;
DROP TABLE IF EXISTS users;