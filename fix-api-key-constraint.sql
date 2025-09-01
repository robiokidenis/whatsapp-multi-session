-- Fix for duplicate empty API key constraint error
-- This script updates all empty API keys to NULL to resolve the unique constraint issue

-- First, let's check if there are any empty API keys
SELECT id, username, api_key 
FROM users 
WHERE api_key = '';

-- Update all empty API keys to NULL
UPDATE users 
SET api_key = NULL 
WHERE api_key = '';

-- Verify the fix
SELECT COUNT(*) as empty_api_keys 
FROM users 
WHERE api_key = '';

-- Show users with NULL API keys (this is valid)
SELECT id, username, 
       CASE 
           WHEN api_key IS NULL THEN 'NULL (valid)'
           WHEN api_key = '' THEN 'EMPTY (invalid)'
           ELSE 'SET'
       END as api_key_status
FROM users
ORDER BY id;