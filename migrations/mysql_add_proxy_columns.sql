-- MySQL Migration: Add proxy configuration columns to session_metadata table
-- Description: Adds proxy configuration fields to support per-session proxy settings

SELECT 'Checking if proxy columns exist in session_metadata table...' as status;

-- Check if proxy_enabled column exists
SELECT 
    CASE 
        WHEN COUNT(*) > 0 THEN 'Proxy columns already exist'
        ELSE 'Proxy columns do not exist - adding now'
    END as column_status
FROM information_schema.columns 
WHERE table_schema = DATABASE() 
AND table_name = 'session_metadata' 
AND column_name = 'proxy_enabled';

-- Add the proxy columns if they don't exist
SET @sql = (
    SELECT IF(
        (SELECT COUNT(*) FROM information_schema.columns 
         WHERE table_schema = DATABASE() 
         AND table_name = 'session_metadata' 
         AND column_name = 'proxy_enabled') = 0,
        'ALTER TABLE session_metadata ADD COLUMN proxy_enabled BOOLEAN DEFAULT FALSE AFTER auto_reply_text,
         ADD COLUMN proxy_type VARCHAR(10) DEFAULT \'\' AFTER proxy_enabled,
         ADD COLUMN proxy_host VARCHAR(255) DEFAULT \'\' AFTER proxy_type,
         ADD COLUMN proxy_port INT DEFAULT 0 AFTER proxy_host,
         ADD COLUMN proxy_username VARCHAR(255) DEFAULT \'\' AFTER proxy_port,
         ADD COLUMN proxy_password VARCHAR(255) DEFAULT \'\' AFTER proxy_username;',
        'SELECT \'Proxy columns already exist\' as result;'
    )
);

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SELECT 'Migration completed - proxy columns are now available' as final_result;