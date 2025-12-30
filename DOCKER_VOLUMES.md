# Docker Volumes Setup Guide

This guide explains how to use Docker volumes instead of bind mounts for better performance and data persistence.

---

## 🎯 Why Use Docker Volumes?

### Advantages over Bind Mounts:

✅ **Better Performance** - Volumes are optimized for Docker I/O
✅ **Managed by Docker** - Automatic creation, cleanup, and backup
✅ **Better Permissions** - No UID/GID issues
✅ **Portable** - Works across different systems and OS
✅ **Persistent** - Data survives container deletion and `docker-compose down`
✅ **Safer** - Isolated from host filesystem
✅ **Easier Backup** - Built-in Docker volume commands

### When to Use Each:

| Scenario | Use |
|----------|-----|
| **Production** | Docker Volumes (recommended) |
| **Development** | Bind Mounts (easier file editing) |
| **Testing** | Docker Volumes (clean isolation) |
| **Multi-host** | Docker Volumes (portable) |

---

## 🚀 Quick Start

### 1. Using Docker Volumes (Recommended)

```bash
# Start with Docker volumes
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml up -d

# That's it! All data is stored in Docker volumes
```

### 2. Check Volumes Created

```bash
# List all volumes
docker volume ls | grep whatsapp

# Expected output:
# whatsapp-database    # WhatsApp session database (SQLite)
# whatsapp-logs        # Application logs
# whatsapp-data        # Session data
# whatsapp-mysql-data  # MySQL database
```

### 3. Verify Data Persistence

```bash
# Create a test session
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  | jq -r '.data.token')

curl -X POST http://localhost:8080/api/sessions \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"phone":"9999999999","name":"Test Session"}'

# Stop ALL containers (including MySQL)
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml down

# Start again
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml up -d

# Check if session still exists
curl http://localhost:8080/api/sessions \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.data.sessions[] | select(.phone == "9999999999")'

# ✅ Session persists! Data is safe in Docker volumes
```

---

## 📁 Volume Structure

### What's Stored Where:

| Volume | Container Path | Content | Backup Frequency |
|--------|---------------|---------|------------------|
| `whatsapp-database` | `/app/database` | WhatsApp sessions.db (SQLite) | Daily |
| `whatsapp-logs` | `/app/logs` | Application logs | Weekly |
| `whatsapp-data` | `/app/data` | Session data & temp files | As needed |
| `whatsapp-config` | `/app/config` | Configuration files (ro) | On change |
| `whatsapp-mysql-data` | `/var/lib/mysql` | MySQL database (users, logs) | Daily |

### Volume Locations:

**Docker stores volumes in:**
- Linux: `/var/lib/docker/volumes/`
- macOS: `~/Library/Containers/com.docker.docker/Data/volumes/`
- Windows: `C:\ProgramData\Docker\volumes\`

You don't need to access these directly - Docker manages everything!

---

## 🛠️ Volume Management

### Using the Volume Manager Script

```bash
# Make script executable (already done)
chmod +x volumes-manager.sh

# Interactive mode
./volumes-manager.sh

# Command-line mode
./volumes-manager.sh list
./volumes-manager.sh backup whatsapp-database
./volumes-manager.sh restore backup.tar.gz whatsapp-database
./volumes-manager.sh details whatsapp-mysql-data
```

### Manual Volume Management

#### List Volumes
```bash
# All volumes
docker volume ls

# Filtered
docker volume ls | grep whatsapp
```

#### Inspect Volume
```bash
docker volume inspect whatsapp-database

# Show:
# - Creation date
# - Driver
# - Mount point (on host)
# - Labels
```

#### Show Volume Usage
```bash
# Space usage
docker system df -v | grep whatsapp

# Detailed
docker system df -v
```

---

## 💾 Backup & Restore

### Automatic Backups (Recommended)

Create `backup-volumes.sh`:

```bash
#!/bin/bash

# Backup all WhatsApp volumes
BACKUP_DIR="./backups/$(date +%Y%m%d_%H%M%S)"
mkdir -p $BACKUP_DIR

echo "📦 Backing up volumes to: $BACKUP_DIR"

# Backup each volume
for volume in whatsapp-database whatsapp-logs whatsapp-data whatsapp-mysql-data; do
    echo "Backing up $volume..."
    docker run --rm \
        -v $volume:/data:ro \
        -v $(pwd)/$BACKUP_DIR:/backup \
        alpine tar czf /backup/${volume}.tar.gz -C /data .
done

echo "✅ Backup complete: $BACKUP_DIR"
echo "💾 Total size: $(du -sh $BACKUP_DIR | cut -f1)"

# Cleanup old backups (keep last 7 days)
find ./backups -type d -mtime +7 -exec rm -rf {} + 2>/dev/null || true
```

Make it executable and add to crontab:
```bash
chmod +x backup-volumes.sh

# Daily backup at 2 AM
crontab -e

# Add this line:
0 2 * * * /path/to/whatsapp-multi-session/backup-volumes.sh >> /var/log/whatsapp-backup.log 2>&1
```

### Manual Backup

```bash
# Backup single volume
docker run --rm \
  -v whatsapp-database:/data:ro \
  -v $(pwd):/backup \
  alpine tar czf /backup/whatsapp-database-$(date +%Y%m%d).tar.gz -C /data .

# Backup all volumes
for volume in whatsapp-database whatsapp-logs whatsapp-data whatsapp-mysql-data; do
    docker run --rm \
        -v $volume:/data:ro \
        -v $(pwd):/backup \
        alpine tar czf /backup/${volume}-$(date +%Y%m%d).tar.gz -C /data .
done
```

### Restore from Backup

```bash
# Stop containers first
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml down

# Restore volume
docker run --rm \
  -v whatsapp-database:/data \
  -v $(pwd):/backup \
  alpine tar xzf /backup/whatsapp-database-20251231.tar.gz -C /data

# Start containers
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml up -d
```

---

## 🔄 Migration from Bind Mounts to Volumes

### Option 1: Automatic Migration (Recommended)

```bash
# 1. Stop current containers
docker-compose down

# 2. Backup existing data
mkdir -p ./migration-backup
cp -r whatsapp/* ./migration-backup/ 2>/dev/null || true

# 3. Start with volumes
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml up -d

# 4. Copy data from old location to volumes
# Copy WhatsApp database
if [ -f ./migration-backup/sessions.db ]; then
    docker run --rm \
        -v whatsapp-database:/data \
        -v $(pwd)/migration-backup:/source \
        alpine cp /source/sessions.db /data/
fi

# Copy logs
if [ -d ./migration-backup/logs ]; then
    docker run --rm \
        -v whatsapp-logs:/data \
        -v $(pwd)/migration-backup:/source \
        alpine sh -c "cp -r /source/logs/* /data/"
fi

# 5. Restart to load data
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml restart

# 6. Verify
curl http://localhost:8080/api/sessions -H "Authorization: Bearer $TOKEN"
```

### Option 2: Manual Migration

```bash
# 1. Export from bind mounts
docker run --rm \
  -v $(pwd)/whatsapp:/source:ro \
  -v $(pwd):/backup \
  alpine tar czf /backup/whatsapp-bind-mounts.tar.gz -C /source .

# 2. Switch to volumes
docker-compose down
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml up -d

# 3. Import to volumes
docker run --rm \
  -v whatsapp-database:/data \
  -v $(pwd):/backup \
  alpine tar xzf /backup/whatsapp-bind-mounts.tar.gz -C /data
```

---

## 🧪 Testing Volume Persistence

### Test Script

```bash
#!/bin/bash
# test-volumes-persistence.sh

echo "🧪 Testing Docker volume persistence..."

# 1. Create test session
echo "1️⃣ Creating test session..."
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  | jq -r '.data.token')

curl -s -X POST http://localhost:8080/api/sessions \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"phone":"1111111111","name":"Persistence Test"}' > /dev/null

echo "   ✅ Test session created"

# 2. Verify session exists
echo "2️⃣ Verifying session exists..."
SESSION_EXISTS=$(curl -s http://localhost:8080/api/sessions \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.data.sessions[] | select(.phone == "1111111111") | .phone')

if [ "$SESSION_EXISTS" == "1111111111" ]; then
    echo "   ✅ Session found: $SESSION_EXISTS"
else
    echo "   ❌ Session not found!"
    exit 1
fi

# 3. Stop all containers
echo "3️⃣ Stopping all containers..."
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml down
echo "   ✅ Containers stopped"

# 4. Start containers
echo "4️⃣ Starting containers..."
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml up -d
sleep 5
echo "   ✅ Containers started"

# 5. Verify session still exists
echo "5️⃣ Verifying session still exists after restart..."
SESSION_EXISTS=$(curl -s http://localhost:8080/api/sessions \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.data.sessions[] | select(.phone == "1111111111") | .phone')

if [ "$SESSION_EXISTS" == "1111111111" ]; then
    echo "   ✅ Session still exists: $SESSION_EXISTS"
    echo ""
    echo "🎉 SUCCESS! Data persists in Docker volumes!"
else
    echo "   ❌ Session lost after restart!"
    echo ""
    echo "❌ FAILURE! Data not persisted"
    exit 1
fi
```

Run the test:
```bash
chmod +x test-volumes-persistence.sh
./test-volumes-persistence.sh
```

---

## 🐛 Troubleshooting

### Issue: Volumes Not Created

**Symptom:** `docker volume ls` doesn't show whatsapp volumes

**Solution:**
```bash
# Start containers with volume configuration
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml up -d

# Verify volumes created
docker volume ls | grep whatsapp
```

### Issue: Permission Errors

**Symptom:** "Permission denied" accessing volumes

**Solution:**
```bash
# Volumes handle permissions automatically
# Just ensure container runs as correct user (1001:1001)
# Already configured in docker-compose.volumes.yml
```

### Issue: Data Lost After Restart

**Symptom:** Sessions disappear after `docker-compose down`

**Diagnosis:**
```bash
# Check if you're using volume config
docker-compose ps | grep whatsapp-multi-session

# Check if volumes exist
docker volume ls | grep whatsapp

# If no volumes found, you're using bind mounts (old config)
```

**Solution:**
```bash
# Always use the volume config
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml up -d
```

### Issue: Can't Access Volume Data

**Symptom:** Need to access files inside volume

**Solution:**
```bash
# Temporary container to browse volume
docker run --rm -it \
  -v whatsapp-database:/data \
  alpine sh

# Inside container:
cd /data
ls -la
cat sessions.db | head
exit

# Or copy files out
docker run --rm \
  -v whatsapp-database:/data:ro \
  -v $(pwd):/backup \
  alpine cp /data/sessions.db /backup/
```

---

## 📊 Monitoring

### Check Volume Size

```bash
# All volumes
docker system df -v

# Specific volume
docker volume inspect whatsapp-database | jq '.[0].UsageSize'

# Human readable
du -sh /var/lib/docker/volumes/whatsapp-*
```

### Monitor Volume Growth

```bash
# Daily check script
#!/bin/bash
echo "Volume usage:"
docker system df -v | grep -E "VOLUME NAME|whatsapp|local"
```

---

## 🎯 Best Practices

### 1. Regular Backups
```bash
# Daily automated backups
0 2 * * * /path/to/backup-volumes.sh
```

### 2. Monitor Volume Growth
```bash
# Weekly check
docker system df -v | grep whatsapp
```

### 3. Test Restores
```bash
# Monthly restore test
# Test backup integrity in staging environment
```

### 4. Document Backup Locations
```bash
# Keep record of:
# - Backup schedule
# - Backup locations
# - Restore procedures
# - Retention policy
```

### 5. Use Volume Labels
```bash
# Add labels for organization
docker volume create --label env=prod --label app=whatsapp whatsapp-data
```

---

## 📚 Reference

### Volume Commands Cheat Sheet

```bash
# List volumes
docker volume ls

# Inspect volume
docker volume inspect <volume-name>

# Create volume
docker volume create <volume-name>

# Remove volume
docker volume rm <volume-name>

# Remove all unused volumes
docker volume prune

# Backup volume
docker run --rm -v <volume>:/data -v $(pwd):/backup \
  alpine tar czf /backup/backup.tar.gz -C /data .

# Restore volume
docker run --rm -v <volume>:/data -v $(pwd):/backup \
  alpine tar xzf /backup/backup.tar.gz -C /data
```

### Docker Compose Commands

```bash
# Start with volumes
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml up -d

# Stop (preserves volumes)
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml stop

# Down (preserves volumes!)
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml down

# Down with volume removal (⚠️ deletes data!)
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml down -v
```

---

## 🔗 Quick Links

- [Docker Volumes Documentation](https://docs.docker.com/storage/volumes/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Production Deployment Guide](PRODUCTION_FIX.md)
- [Troubleshooting Guide](TROUBLESHOOTING.md)

---

## ✅ Summary

**✅ Use Docker Volumes when:**
- Running in production
- Need better performance
- Want automatic data persistence
- Deploying across multiple environments
- Don't need direct file access

**❌ Use Bind Mounts when:**
- Developing and editing files frequently
- Need direct access to files on host
- Running in local development environment

**🎯 For Production:**
```bash
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml up -d
```

**💾 Don't forget:**
- Regular backups
- Monitor volume growth
- Test restore procedures
- Document your setup
