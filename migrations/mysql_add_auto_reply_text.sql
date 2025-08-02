-- MySQL Migration: Add auto_reply_text column to session_metadata table
-- Date: 2025-08-02
-- Description: Adds auto_reply_text column to support session-level auto reply functionality

-- Use the application database
USE whatsapp_multi_session;

-- Check if the column already exists
SELECT 
    CASE 
        WHEN COUNT(*) > 0 THEN 'Column auto_reply_text already exists'
        ELSE 'Column auto_reply_text does not exist - adding now'
    END as status
FROM information_schema.COLUMNS 
WHERE table_schema = 'whatsapp_multi_session' 
AND table_name = 'session_metadata' 
AND column_name = 'auto_reply_text';

-- Add the auto_reply_text column if it doesn't exist
-- MySQL syntax allows for conditional execution through stored procedures
DELIMITER //

CREATE PROCEDURE AddAutoReplyTextColumn()
BEGIN
    DECLARE column_exists INT DEFAULT 0;
    
    SELECT COUNT(*) INTO column_exists 
    FROM information_schema.COLUMNS 
    WHERE table_schema = DATABASE() 
    AND table_name = 'session_metadata' 
    AND column_name = 'auto_reply_text';
    
    IF column_exists = 0 THEN
        ALTER TABLE session_metadata ADD COLUMN auto_reply_text TEXT AFTER webhook_url;
        SELECT 'SUCCESS: auto_reply_text column added to session_metadata table' as result;
    ELSE
        SELECT 'INFO: auto_reply_text column already exists in session_metadata table' as result;
    END IF;
END //

DELIMITER ;

-- Execute the procedure
CALL AddAutoReplyTextColumn();

-- Drop the procedure as it's no longer needed
DROP PROCEDURE AddAutoReplyTextColumn;

-- Verify the final table structure
DESCRIBE session_metadata;