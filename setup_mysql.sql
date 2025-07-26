-- Drop existing tables if they exist
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS session_metadata;

-- Create users table
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at BIGINT NOT NULL,
    INDEX idx_username (username),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create session_metadata table
CREATE TABLE session_metadata (
    id VARCHAR(255) PRIMARY KEY,
    phone VARCHAR(255) NOT NULL,
    actual_phone VARCHAR(255),
    name VARCHAR(255),
    created_at BIGINT NOT NULL,
    INDEX idx_phone (phone),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default admin user
INSERT INTO users (username, password, created_at) 
VALUES ('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMye.R89qBMIILTGMNlbLZLKRIkzFWNUsLq', UNIX_TIMESTAMP());

-- Show tables
SHOW TABLES;
DESCRIBE users;
DESCRIBE session_metadata;