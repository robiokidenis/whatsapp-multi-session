-- Add position column
ALTER TABLE session_metadata ADD COLUMN position INT DEFAULT 0;

-- Add webhook_url column  
ALTER TABLE session_metadata ADD COLUMN webhook_url VARCHAR(500);

-- Add index for position
CREATE INDEX idx_position ON session_metadata(position);

-- Update existing rows to have incremental positions
SET @row_number = 0;
UPDATE session_metadata SET position = (@row_number := @row_number + 1) ORDER BY created_at;