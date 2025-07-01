-- Rollback: make next_execution_at nullable again and restore conditional index
-- Drop the unconditional index
DROP INDEX IF EXISTS idx_scheduled_items_next_execution;

-- Alter the column back to nullable
ALTER TABLE scheduled_items 
ALTER COLUMN next_execution_at DROP NOT NULL;

-- Recreate the conditional index
CREATE INDEX IF NOT EXISTS idx_scheduled_items_next_execution 
ON scheduled_items (next_execution_at) 
WHERE next_execution_at IS NOT NULL;