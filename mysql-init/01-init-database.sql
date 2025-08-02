-- WhatsApp Multi-Session MySQL Database Initialization
-- This script runs automatically when MySQL container starts for the first time

-- Create the application database if it doesn't exist
CREATE DATABASE IF NOT EXISTS whatsapp_multi_session 
    CHARACTER SET utf8mb4 
    COLLATE utf8mb4_unicode_ci;

-- Create application user with proper permissions
CREATE USER IF NOT EXISTS 'whatsapp_user'@'%' IDENTIFIED BY 'your_secure_mysql_password_here';

-- Grant necessary permissions to the application user
GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, ALTER, INDEX, DROP, REFERENCES 
    ON whatsapp_multi_session.* 
    TO 'whatsapp_user'@'%';

-- Grant permission to create temporary tables (needed for some operations)
GRANT CREATE TEMPORARY TABLES ON whatsapp_multi_session.* TO 'whatsapp_user'@'%';

-- Flush privileges to ensure changes take effect
FLUSH PRIVILEGES;

-- Use the application database
USE whatsapp_multi_session;

-- Optional: Pre-create tables with optimized structure for MySQL
-- Note: The application will create these automatically, but you can pre-create them here

-- Users table (for authentication)
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    session_limit INT NOT NULL DEFAULT 5,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at BIGINT NOT NULL,
    updated_at BIGINT,
    
    INDEX idx_username (username),
    INDEX idx_role (role),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Session metadata table
CREATE TABLE IF NOT EXISTS session_metadata (
    id VARCHAR(255) PRIMARY KEY,
    phone VARCHAR(50) NOT NULL,
    actual_phone VARCHAR(50),
    name VARCHAR(255),
    position INT DEFAULT 0,
    webhook_url TEXT,
    user_id INT NOT NULL,
    created_at BIGINT NOT NULL,
    
    INDEX idx_phone (phone),
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at),
    
    -- Foreign key constraint
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Logs table (for database logging when enabled)
CREATE TABLE IF NOT EXISTS logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    level VARCHAR(10) NOT NULL,
    message TEXT NOT NULL,
    component VARCHAR(100),
    session_id VARCHAR(255),
    user_id INT,
    metadata JSON,
    created_at BIGINT NOT NULL,
    
    INDEX idx_level (level),
    INDEX idx_component (component),
    INDEX idx_session_id (session_id),
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at),
    
    -- Foreign key constraint (optional)
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Optional: Create views for common queries
CREATE OR REPLACE VIEW recent_logs AS
SELECT 
    id,
    level,
    message,
    component,
    session_id,
    user_id,
    created_at,
    FROM_UNIXTIME(created_at) as formatted_time
FROM logs 
WHERE created_at > UNIX_TIMESTAMP(DATE_SUB(NOW(), INTERVAL 24 HOUR))
ORDER BY created_at DESC;

-- Optional: Create stored procedures for maintenance
DELIMITER //

CREATE PROCEDURE IF NOT EXISTS CleanOldLogs(IN days_old INT)
BEGIN
    DECLARE cutoff_time BIGINT;
    SET cutoff_time = UNIX_TIMESTAMP(DATE_SUB(NOW(), INTERVAL days_old DAY));
    
    DELETE FROM logs WHERE created_at < cutoff_time;
    
    SELECT ROW_COUNT() as deleted_rows;
END //

CREATE PROCEDURE IF NOT EXISTS GetLogStats()
BEGIN
    SELECT 
        level,
        COUNT(*) as count,
        MIN(FROM_UNIXTIME(created_at)) as earliest,
        MAX(FROM_UNIXTIME(created_at)) as latest
    FROM logs 
    GROUP BY level
    ORDER BY count DESC;
END //

DELIMITER ;

-- Optional: Create default admin user (password: admin123)
-- Note: The application will create this automatically with the configured credentials
INSERT IGNORE INTO users (username, password_hash, role, session_limit, is_active, created_at)
VALUES (
    'admin',
    '$2a$10$N9qo8uLOickgx2ZMRZoMye1J/QNs1nHFB6w7kp7Jp5kJzfqRqhQzO', -- bcrypt hash of 'admin123'
    'admin',
    50,
    TRUE,
    UNIX_TIMESTAMP(NOW())
);

-- Display initialization completion message
SELECT 'WhatsApp Multi-Session MySQL database initialized successfully!' as message;