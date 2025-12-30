# Docker + MySQL Setup Guide

This guide explains how to run WhatsApp Multi-Session Manager with Docker and MySQL database.

## Quick Start

### 1. Copy Environment Configuration
```bash
# Copy the MySQL environment template
cp .env.docker.mysql .env

# Edit the configuration (IMPORTANT!)
nano .env
```

### 2. Start with Docker Compose
```bash
# Basic setup with MySQL
docker-compose -f docker-compose.yml -f docker-compose.mysql.enhanced.yml up -d

# Or use the original MySQL compose file
docker-compose -f docker-compose.yml -f docker-compose.mysql.yml up -d
```

### 3. Access the Application
- **Main Application**: http://localhost:18080
- **Admin Login**: admin / SuperSecureAdminPassword123! *(change this!)*
- **Database Logs**: http://localhost:18080/logs (admin only)

## Configuration Files

### Environment Files
| File | Purpose |
|------|---------|
| `.env.docker.mysql` | Template for Docker + MySQL setup |
| `.env.local` | Local development configuration |
| `.env.example` | General configuration template |

### Docker Compose Files
| File | Purpose |
|------|---------|
| `docker-compose.yml` | Base configuration |
| `docker-compose.mysql.yml` | Basic MySQL setup |
| `docker-compose.mysql.enhanced.yml` | Enhanced MySQL with extras |

## Database Configuration

### Required Environment Variables
```bash
# Database type
DATABASE_TYPE=mysql

# MySQL connection
MYSQL_HOST=mysql
MYSQL_PORT=3306
MYSQL_USER=whatsapp_user
MYSQL_PASSWORD=your_secure_mysql_password_here
MYSQL_DATABASE=whatsapp_multi_session

# Logging configuration
ENABLE_DATABASE_LOG=true  # Enable database logging
LOG_LEVEL=info
```

### MySQL Settings
```bash
# Root password for MySQL container
MYSQL_ROOT_PASSWORD=very_secure_root_password_123

# Performance tuning
MYSQL_INNODB_BUFFER_POOL_SIZE=256M
MYSQL_MAX_CONNECTIONS=200
```

## Advanced Features

### With Admin Tools (phpMyAdmin)
```bash
# Start with phpMyAdmin for database management
docker-compose -f docker-compose.yml -f docker-compose.mysql.enhanced.yml --profile admin up -d

# Access phpMyAdmin: http://localhost:8081
```

### With Redis Session Storage
```bash
# Start with Redis for improved session handling
docker-compose -f docker-compose.yml -f docker-compose.mysql.enhanced.yml --profile redis up -d
```

### With Monitoring (Prometheus + Grafana)
```bash
# Start with monitoring stack
docker-compose -f docker-compose.yml -f docker-compose.mysql.enhanced.yml --profile monitoring up -d

# Access Grafana: http://localhost:3000 (admin/admin123)
# Access Prometheus: http://localhost:9090
```

### All Services
```bash
# Start everything
docker-compose -f docker-compose.yml -f docker-compose.mysql.enhanced.yml --profile admin --profile redis --profile monitoring up -d
```

## Database Logging Features

When `ENABLE_DATABASE_LOG=true`, you get:

### ✅ Full Database Logging
- All application logs stored in MySQL
- Web interface for log management
- Advanced filtering and search
- Auto-refresh capabilities
- Cleanup and maintenance tools

### 📊 Log Management Interface
Access via: http://localhost:18080/logs (admin required)

**Features:**
- Filter by level (debug, info, warn, error)
- Filter by component, session, user
- Time range filtering
- Pagination with latest-first ordering
- Auto-refresh every 5 seconds
- Delete logs older than X days
- Clear all logs functionality

### 🔧 Log Configuration
```bash
# Console + Database logging (default)
ENABLE_LOGGING=true
ENABLE_DATABASE_LOG=true

# Console only (better performance)
ENABLE_LOGGING=true  
ENABLE_DATABASE_LOG=false

# Minimal logging
ENABLE_LOGGING=false
ENABLE_DATABASE_LOG=false
```

## Database Schema

The application automatically creates these MySQL tables:

### Users Table
```sql
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    session_limit INT NOT NULL DEFAULT 5,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at BIGINT NOT NULL,
    updated_at BIGINT
);
```

### Logs Table (when database logging enabled)
```sql
CREATE TABLE logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    level VARCHAR(10) NOT NULL,
    message TEXT NOT NULL,
    component VARCHAR(100),
    session_id VARCHAR(255),
    user_id INT,
    metadata JSON,
    created_at BIGINT NOT NULL
);
```

### Session Metadata Table
```sql
CREATE TABLE session_metadata (
    id VARCHAR(255) PRIMARY KEY,
    phone VARCHAR(50) NOT NULL,
    actual_phone VARCHAR(50),
    name VARCHAR(255),
    position INT DEFAULT 0,
    webhook_url TEXT,
    created_at BIGINT NOT NULL
);
```

## Volume Mounts

### Data Persistence
```yaml
volumes:
  - ./whatsapp:/app/data          # Session data
  - ./whatsapp/sessions:/app/sessions  # WhatsApp sessions
  - ./media:/app/media            # Media files
  - mysql-data:/var/lib/mysql     # MySQL data
```

### Configuration
```yaml
volumes:
  - ./mysql-init:/docker-entrypoint-initdb.d:ro  # MySQL initialization
  - ./mysql-config:/etc/mysql/conf.d:ro          # MySQL configuration
```

## Security Considerations

### 🔒 Change Default Passwords
```bash
# In .env file:
ADMIN_PASSWORD=YourSecurePassword123!
MYSQL_PASSWORD=YourSecureMySQLPassword456!
MYSQL_ROOT_PASSWORD=YourSecureRootPassword789!
JWT_SECRET=YourSuperSecretJWTKey2024!
```

### 🌐 Configure CORS
```bash
# For production, set specific domains:
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://app.yourdomain.com

# For development:
CORS_ALLOWED_ORIGINS=http://localhost:18080,http://127.0.0.1:18080
```

### 🛡️ Firewall Configuration
- Application: Port 18080
- MySQL: Port 3306 (restrict access)
- phpMyAdmin: Port 8081 (development only)

## Performance Tuning

### MySQL Optimization
```bash
# In .env file:
MYSQL_INNODB_BUFFER_POOL_SIZE=512M  # Increase for more RAM
MYSQL_MAX_CONNECTIONS=400           # Increase for high load
DB_MAX_OPEN_CONNS=50               # Application connection pool
```

### Resource Limits
```bash
# Container limits
MEMORY_LIMIT=2048M  # Application memory
CPU_LIMIT=4.0       # CPU cores
```

### Database Logging Performance
```bash
# Disable database logging for maximum performance
ENABLE_DATABASE_LOG=false

# Keep console logging for debugging
ENABLE_LOGGING=true
LOG_LEVEL=warn  # Reduce log verbosity
```

## Maintenance

### Database Backups
```bash
# Manual backup
docker exec whatsapp-mysql mysqldump -u root -p whatsapp_multi_session > backup.sql

# Restore backup
docker exec -i whatsapp-mysql mysql -u root -p whatsapp_multi_session < backup.sql
```

### Log Cleanup
```bash
# Via web interface: http://localhost:18080/logs
# Click "Delete 30+ days" or "Clear All Logs"

# Via MySQL directly:
docker exec -it whatsapp-mysql mysql -u root -p
USE whatsapp_multi_session;
CALL CleanOldLogs(30);  # Delete logs older than 30 days
```

### Container Management
```bash
# View logs
docker-compose logs -f whatsapp-multi-session
docker-compose logs -f mysql

# Restart services
docker-compose restart whatsapp-multi-session
docker-compose restart mysql

# Update containers
docker-compose pull
docker-compose up -d --force-recreate
```

## Troubleshooting

### Database Connection Issues
```bash
# Check MySQL is running
docker-compose ps mysql

# Check MySQL logs
docker-compose logs mysql

# Test connection
docker exec -it whatsapp-mysql mysql -u whatsapp_user -p
```

### Application Issues
```bash
# Check application logs
docker-compose logs whatsapp-multi-session

# Check configuration
docker exec whatsapp-multi-session env | grep -E "(DATABASE|MYSQL|LOG)"
```

### Common Errors

**"Database logging disabled"**
- Check: `ENABLE_DATABASE_LOG=true` in .env
- Restart: `docker-compose restart whatsapp-multi-session`

**"Connection refused to MySQL"**
- Wait for MySQL to fully start (30-60 seconds)
- Check MySQL health: `docker-compose ps mysql`

**"Access denied for user"**
- Verify MySQL credentials in .env
- Check mysql-init scripts ran correctly

## Production Deployment

### Environment Security
```bash
# Use secrets management
MYSQL_PASSWORD_FILE=/run/secrets/mysql_password
JWT_SECRET_FILE=/run/secrets/jwt_secret

# Enable HTTPS
ENABLE_HTTPS=true
SSL_CERT_PATH=/app/certs/cert.pem
SSL_KEY_PATH=/app/certs/key.pem
```

### Monitoring Setup
```bash
# Enable metrics
ENABLE_METRICS=true
ENABLE_PROMETHEUS=true

# Start monitoring stack
docker-compose --profile monitoring up -d
```

### Resource Planning
- **Minimum**: 2GB RAM, 2 CPU cores, 20GB storage
- **Recommended**: 4GB RAM, 4 CPU cores, 100GB storage
- **High Load**: 8GB RAM, 8 CPU cores, 500GB storage