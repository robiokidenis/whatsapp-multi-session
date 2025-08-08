-- Add enabled column to session_metadata table
-- This migration adds an enabled boolean column to control session activation

-- For SQLite
ALTER TABLE session_metadata ADD COLUMN enabled BOOLEAN DEFAULT 1;

-- For MySQL (uncomment if using MySQL)
-- ALTER TABLE session_metadata ADD COLUMN enabled BOOLEAN DEFAULT TRUE;

-- Update existing sessions to be enabled by default
UPDATE session_metadata SET enabled = 1 WHERE enabled IS NULL;