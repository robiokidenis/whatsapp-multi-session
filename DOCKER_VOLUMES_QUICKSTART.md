# Docker Volumes - Quick Start

## 🚀 Start with Docker Volumes (Production)

```bash
# 1. Start with volumes (includes MySQL)
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml up -d

# 2. Verify volumes created
docker volume ls | grep whatsapp

# Expected output:
# whatsapp-database       # WhatsApp sessions
# whatsapp-logs           # Application logs  
# whatsapp-data           # Session data
# whatsapp-mysql-data     # MySQL database
```

## 📦 Backup All Volumes

```bash
# Interactive menu
./volumes-manager.sh

# Or backup specific volume
./volumes-manager.sh backup whatsapp-database my-backup.tar.gz

# Or export all volumes
./volumes-manager.sh export ./backups/
```

## ✅ Test Persistence

```bash
# 1. Create session
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  | jq -r '.data.token')

curl -X POST http://localhost:8080/api/sessions \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"phone":"9999999999","name":"Test"}'

# 2. Stop everything
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml down

# 3. Start again
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml up -d

# 4. Verify session still exists
curl http://localhost:8080/api/sessions \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.data.sessions[] | select(.phone == "9999999999")'

# ✅ Data persists!
```

## 🛠️ Volume Management

```bash
# List volumes
docker volume ls | grep whatsapp

# Inspect volume
docker volume inspect whatsapp-database

# Check usage
docker system df -v | grep whatsapp

# Interactive volume manager
./volumes-manager.sh
```

## 💾 Backup & Restore

```bash
# Backup single volume
docker run --rm -v whatsapp-database:/data:ro -v $(pwd):/backup \
  alpine tar czf /backup/db-$(date +%Y%m%d).tar.gz -C /data .

# Restore volume
docker run --rm -v whatsapp-database:/data -v $(pwd):/backup \
  alpine tar xzf /backup/db-20251231.tar.gz -C /data
```

## 📁 What's Stored Where

| Volume | Content |
|--------|---------|
| `whatsapp-database` | WhatsApp sessions.db (SQLite) |
| `whatsapp-logs` | Application logs |
| `whatsapp-data` | Session temp data |
| `whatsapp-mysql-data` | MySQL database (users, logs) |

## ⚠️ Important Commands

```bash
# Stop (keeps volumes!)
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml down

# Start (uses existing volumes)
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml up -d

# Remove EVERYTHING (⚠️ deletes data!)
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml down -v
```

## 🔄 Migration from Bind Mounts

```bash
# 1. Backup current data
cp -r whatsapp/* ./backup/

# 2. Switch to volumes
docker-compose down
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml up -d

# 3. Copy data to volumes
docker run --rm -v whatsapp-database:/data -v $(pwd)/backup:/source \
  alpine sh -c "cp /source/sessions.db /data/ 2>/dev/null || true"

# 4. Restart
docker-compose -f docker-compose.yml -f docker-compose.volumes.yml restart
```

## 📚 Documentation

- Full guide: [DOCKER_VOLUMES.md](DOCKER_VOLUMES.md)
- Volume manager: `./volumes-manager.sh`
- Production fix: [PRODUCTION_FIX.md](PRODUCTION_FIX.md)

## 🎯 Key Benefits

✅ Data persists across `docker-compose down`
✅ Better performance than bind mounts
✅ No permission issues
✅ Easy backup/restore
✅ Production-ready
