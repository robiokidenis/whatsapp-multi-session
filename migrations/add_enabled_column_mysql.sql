-- Simple MySQL migration to add enabled column
-- Run this manually in your MySQL database

ALTER TABLE session_metadata ADD COLUMN enabled BOOLEAN DEFAULT TRUE;

-- Update existing sessions to be enabled by default  
UPDATE session_metadata SET enabled = TRUE WHERE enabled IS NULL;