-- Migration: Add auto_reply_text column to session_metadata table
-- Date: 2025-08-02
-- Description: Adds auto_reply_text column to support session-level auto reply functionality

-- For SQLite
-- ALTER TABLE session_metadata ADD COLUMN auto_reply_text TEXT;

-- For MySQL
-- ALTER TABLE session_metadata ADD COLUMN auto_reply_text TEXT AFTER webhook_url;

-- Check current database type and run appropriate command:

-- SQLite version:
SELECT 'Adding auto_reply_text column to session_metadata table (SQLite)' as migration_status;
ALTER TABLE session_metadata ADD COLUMN auto_reply_text TEXT;

-- If you're using MySQL, comment out the SQLite version above and uncomment below:
-- SELECT 'Adding auto_reply_text column to session_metadata table (MySQL)' as migration_status;
-- ALTER TABLE session_metadata ADD COLUMN auto_reply_text TEXT AFTER webhook_url;

-- Verify the column was added successfully
SELECT 'Migration completed successfully - auto_reply_text column added' as result;