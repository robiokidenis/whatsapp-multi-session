# Production Fix Guide

This guide fixes two critical issues in production:

1. **Database Error**: "unable to open database file: is a directory"
2. **Sessions Disappearing**: Sessions are lost after `docker-compose down`

---

## Issue 1: Database "is a directory" Error

### Root Cause
The volume mounts in `docker-compose.yml` are problematic:
```yaml
volumes:
  - ${DATA_DIR:-./whatsapp}:/app/data
  - ${WHATSAPP_DB_DIR:-./whatsapp}:/app/database  # Problem!
  - ${LOGS_DIR:-./whatsapp/logs}:/app/logs
```

Both `/app/data` and `/app/database` map to the SAME host directory (`./whatsapp`), causing confusion.

### Solution A: Fix Directory Structure (Recommended)

```bash
# Stop containers
docker-compose down

# Reorganize directories
mkdir -p whatsapp-data/database
mkdir -p whatsapp-data/logs
mkdir -p whatsapp-data/sessions

# Copy existing data
if [ -f whatsapp/sessions.db ]; then
  mv whatsapp/sessions.db whatsapp-data/database/
fi

if [ -d whatsapp/logs ]; then
  cp -r whatsapp/logs/* whatsapp-data/logs/ 2>/dev/null || true
fi

# Remove problematic directory/file
rm -rf whatsapp/sessions.db  # Remove if it's a directory
```

Then update `.env` file:
```bash
# Add these lines to your .env
DATA_DIR=./whatsapp-data
WHATSAPP_DB_DIR=./whatsapp-data/database
LOGS_DIR=./whatsapp-data/logs
```

### Solution B: Quick Fix (Remove Directory)

```bash
# Stop container
docker-compose down

# Remove if sessions.db is a directory
if [ -d whatsapp/sessions.db ]; then
  echo "Removing sessions.db directory..."
  rm -rf whatsapp/sessions.db
fi

# Start container (will create fresh database)
docker-compose up -d
```

---

## Issue 2: Sessions Disappearing After Restart

### Root Cause
Session metadata is stored in MySQL, but MySQL data is not persisted across restarts.

### Diagnose Your Setup

```bash
# Check which docker-compose files you're using
ps aux | grep docker-compose

# Check if MySQL container is running
docker ps | grep mysql

# Check MySQL data volume
docker volume ls | grep mysql
```

### Solution A: Using MySQL (Production)

If you're using MySQL, ensure data persistence:

```bash
# Check if you have mysql-data volume defined
docker volume inspect whatsapp-multi-session_mysql-data

# Use the MySQL docker-compose configuration
docker-compose -f docker-compose.yml -f docker-compose.mysql.enhanced.yml up -d
```

**Important:** The mysql-data volume MUST be persisted:
```yaml
# In docker-compose.mysql.yml or docker-compose.mysql.enhanced.yml
volumes:
  mysql-data:  # This ensures persistence
    driver: local
```

### Solution B: Backup & Restore Sessions

```bash
#!/bin/bash
# backup-sessions.sh

echo "📦 Backing up sessions..."

# 1. Backup MySQL database
docker exec mysql-container mysqldump \
  -u root -p${MYSQL_PASSWORD} \
  ${MYSQL_DATABASE} \
  > backup_mysql_$(date +%Y%m%d_%H%M%S).sql

# 2. Backup WhatsApp SQLite database
cp whatsapp/sessions.db backup_sessions_$(date +%Y%m%d_%H%M%S).db

# 3. Backup .env file
cp .env backup_env_$(date +%Y%m%d_%H%M%S).txt

echo "✅ Backup complete!"
```

Restore after restart:
```bash
#!/bin/bash
# restore-sessions.sh

BACKUP_DATE=$1  # Pass date as argument

# 1. Restore MySQL
docker exec -i mysql-container mysql \
  -u root -p${MYSQL_PASSWORD} \
  ${MYSQL_DATABASE} \
  < backup_mysql_${BACKUP_DATE}.sql

# 2. Restore SQLite database
cp backup_sessions_${BACKUP_DATE}.db whatsapp/sessions.db

# 3. Restart container
docker-compose restart

echo "✅ Sessions restored!"
```

### Solution C: Complete Production Setup

Create `docker-compose.production.yml`:

```yaml
version: '3.8'

services:
  whatsapp-multi-session:
    extends:
      file: docker-compose.yml
      service: whatsapp-multi-session
    volumes:
      # Fixed directory structure
      - ./data/whatsapp:/app/database
      - ./data/logs:/app/logs
      - ./data/config:/app/config:ro
    environment:
      - WHATSAPP_DB_PATH=/app/database/sessions.db

  mysql:
    image: mysql:8.0
    container_name: whatsapp-mysql
    restart: always
    environment:
      - MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD}
      - MYSQL_DATABASE=${MYSQL_DATABASE}
      - MYSQL_USER=${MYSQL_USER}
      - MYSQL_PASSWORD=${MYSQL_PASSWORD}
    volumes:
      # Persist MySQL data
      - mysql-data:/var/lib/mysql
    ports:
      - "3306:3306"
    networks:
      - whatsapp-network

volumes:
  mysql-data:
    driver: local
    # This volume persists across container restarts!

networks:
  whatsapp-network:
    driver: bridge
```

Usage:
```bash
# Start with persistence
docker-compose -f docker-compose.production.yml up -d

# Stop and restart (data persists!)
docker-compose -f docker-compose.production.yml down
docker-compose -f docker-compose.production.yml up -d

# Sessions are preserved!
```

---

## Complete Production Deployment Script

```bash
#!/bin/bash
# deploy-production-fixed.sh

set -e

echo "🚀 Deploying WhatsApp Multi-Session with persistence..."

# 1. Create proper directory structure
echo "📁 Creating directories..."
mkdir -p data/whatsapp
mkdir -p data/logs
mkdir -p data/config

# 2. Set permissions
echo "🔐 Setting permissions..."
chmod -R 755 data

# 3. Create .env file if not exists
if [ ! -f .env ]; then
  echo "📝 Creating .env file..."
  cat > .env << 'EOF'
# MySQL Configuration
MYSQL_ROOT_PASSWORD=$(openssl rand -base64 32)
MYSQL_DATABASE=whatsapp_multi_session
MYSQL_USER=whatsapp_user
MYSQL_PASSWORD=$(openssl rand -base64 32)

# JWT Secret
JWT_SECRET=$(openssl rand -base64 64)

# Admin User
ADMIN_USERNAME=admin
ADMIN_PASSWORD=$(openssl rand -base64 16)

# Data Directories
DATA_DIR=./data/whatsapp
WHATSAPP_DB_DIR=./data/whatsapp
LOGS_DIR=./data/logs
CONFIG_DIR=./data/config

# WhatsApp Database
WHATSAPP_DB_PATH=/app/database/sessions.db

# Application Settings
PORT=8080
LOG_LEVEL=info
SESSION_TIMEOUT=24h
MAX_SESSIONS=10
EOF
fi

# 4. Source .env
export $(cat .env | grep -v '^#' | xargs)

# 5. Start with MySQL
echo "🐳 Starting containers..."
docker-compose -f docker-compose.yml -f docker-compose.mysql.enhanced.yml up -d

# 6. Wait for MySQL
echo "⏳ Waiting for MySQL..."
sleep 10

# 7. Check status
echo "✅ Checking status..."
docker-compose ps

# 8. Show logs
echo "📋 Recent logs:"
docker-compose logs --tail 20 whatsapp-multi-session

echo ""
echo "✅ Deployment complete!"
echo ""
echo "🔐 Generated credentials:"
echo "Admin: $ADMIN_USERNAME / $ADMIN_PASSWORD"
echo "MySQL: $MYSQL_USER / $MYSQL_PASSWORD"
echo ""
echo "💡 To stop: docker-compose -f docker-compose.yml -f docker-compose.mysql.enhanced.yml down"
echo "💡 To restart: docker-compose -f docker-compose.yml -f docker-compose.mysql.enhanced.yml up -d"
echo "💡 Sessions persist in mysql-data volume!"
```

---

## Verification Commands

```bash
# 1. Check session database persistence
docker exec whatsapp-multi-session ls -la /app/database/sessions.db

# 2. Check MySQL data volume
docker volume ls | grep mysql

# 3. Test session creation and persistence
curl -X POST http://localhost:8080/api/sessions \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"phone":"1234567890","name":"Test"}'

# 4. Stop and restart
docker-compose down
docker-compose up -d

# 5. Verify session still exists
curl http://localhost:8080/api/sessions \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.data.sessions[] | select(.phone == "1234567890")'
```

---

## Emergency Recovery

If you lost sessions after restart:

```bash
# 1. Check if old volume still exists
docker volume ls

# 2. Inspect the volume
docker volume inspect whatsapp-multi-session_mysql-data

# 3. Create new container with old volume to extract data
docker run --rm -v whatsapp-multi-session_mysql-data:/data \
  -v $(pwd):/backup alpine \
  tar czf /backup/mysql-backup.tar.gz -C /data .

# 4. Restore to new setup
docker run --rm -v whatsapp-multi-session_mysql-data:/data \
  -v $(pwd):/backup alpine \
  tar xzf /backup/mysql-backup.tar.gz -C /data
```

---

## Summary

### For Immediate Fix (Issue 1):
```bash
docker-compose down
rm -rf whatsapp/sessions.db
docker-compose up -d
```

### For Persistent Sessions (Issue 2):
```bash
# Use MySQL with volume persistence
docker-compose -f docker-compose.yml -f docker-compose.mysql.enhanced.yml up -d

# Verify mysql-data volume exists
docker volume ls | grep mysql
```

### For Production:
1. Use proper directory structure
2. Use MySQL with named volumes
3. Regular backups
4. Monitor volume persistence
