-- MySQL migration to add enabled column to session_metadata table
-- This migration adds an enabled boolean column to control session activation

USE waGo;

-- Check if enabled column exists before adding it
SET @col_exists = 0;
SELECT COUNT(*) INTO @col_exists 
FROM information_schema.columns 
WHERE table_schema = 'waGo' 
AND table_name = 'session_metadata' 
AND column_name = 'enabled';

-- Add enabled column if it doesn't exist
SET @sql = IF(@col_exists = 0,
    'ALTER TABLE session_metadata ADD COLUMN enabled BOOLEAN DEFAULT TRUE',
    'SELECT ''enabled column already exists'' as result');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Update existing sessions to be enabled by default
UPDATE session_metadata SET enabled = TRUE WHERE enabled IS NULL;

-- Show result
SELECT CASE 
    WHEN @col_exists = 0 THEN 'SUCCESS: enabled column added to session_metadata table'
    ELSE 'INFO: enabled column already exists in session_metadata table'
END as result;