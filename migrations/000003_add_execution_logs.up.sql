-- Add execution_logs table for tracking scheduled item processing
CREATE TABLE IF NOT EXISTS execution_logs (
    id SERIAL PRIMARY KEY,
    scheduled_item_id INTEGER NOT NULL,
    executed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'success',
    error_message TEXT,
    todo_item_id INTEGER
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_execution_logs_scheduled_item_id ON execution_logs (scheduled_item_id);
CREATE INDEX IF NOT EXISTS idx_execution_logs_executed_at ON execution_logs (executed_at);
CREATE INDEX IF NOT EXISTS idx_execution_logs_status ON execution_logs (status);

-- Add constraint to ensure status is valid
ALTER TABLE execution_logs ADD CONSTRAINT chk_execution_logs_status 
CHECK (status IN ('success', 'error', 'skipped'));