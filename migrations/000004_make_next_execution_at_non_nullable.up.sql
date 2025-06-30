-- Make next_execution_at column non-nullable and update index
-- First, drop the existing conditional index
DROP INDEX IF EXISTS idx_scheduled_items_next_execution;

-- Alter the column to be NOT NULL
ALTER TABLE scheduled_items 
ALTER COLUMN next_execution_at SET NOT NULL;

-- Create new unconditional index since column is now always non-null
CREATE INDEX IF NOT EXISTS idx_scheduled_items_next_execution 
ON scheduled_items (next_execution_at);