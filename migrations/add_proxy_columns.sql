-- Migration: Add proxy configuration columns to session_metadata table
-- Description: Adds proxy configuration fields to support per-session proxy settings

-- SQLite Migration
-- For SQLite, add columns one by one

SELECT 'Adding proxy columns to session_metadata table (SQLite)' as migration_status;

-- Add proxy columns (SQLite allows adding columns without checking existence)
ALTER TABLE session_metadata ADD COLUMN proxy_enabled BOOLEAN DEFAULT 0;
ALTER TABLE session_metadata ADD COLUMN proxy_type TEXT DEFAULT '';
ALTER TABLE session_metadata ADD COLUMN proxy_host TEXT DEFAULT '';
ALTER TABLE session_metadata ADD COLUMN proxy_port INTEGER DEFAULT 0;
ALTER TABLE session_metadata ADD COLUMN proxy_username TEXT DEFAULT '';
ALTER TABLE session_metadata ADD COLUMN proxy_password TEXT DEFAULT '';

SELECT 'Migration completed successfully - proxy columns added' as result;