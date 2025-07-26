# 🐳 WhatsApp Multi-Session Manager - Docker Setup

Complete Docker deployment with SQLite database and local volume storage.

## ✅ What's Included

### 📦 Docker Files
- **Dockerfile** - Multi-stage build with Go 1.23 and Alpine Linux
- **docker-compose.yml** - Complete orchestration with environment variables
- **.env** - Comprehensive configuration template
- **.dockerignore** - Optimized build context
- **DOCKER_DEPLOYMENT.md** - Detailed deployment guide

### 🗄️ Database & Storage
- **SQLite-only** configuration (MySQL removed)
- **Local volume mapping** to `./whatsapp/` directory
- **Persistent storage** for databases, sessions, and logs
- **Automatic database initialization** with admin user

### 🚀 Quick Start

```bash
# 1. Clone repository
git clone <your-repo>
cd whatsapp-multi-session

# 2. Create directories
mkdir -p whatsapp/{sessions,logs,backups}

# 3. Configure environment
cp .env .env.local
# Edit .env.local - CHANGE JWT_SECRET and ADMIN_PASSWORD!

# 4. Build and run
docker-compose up -d

# 5. Access application
open http://localhost:8080
```

### 🔐 Default Credentials
- **Username:** admin
- **Password:** admin123 (⚠️ CHANGE THIS!)

## 📁 Directory Structure

```
whatsapp-multi-session/
├── 🐳 Docker Files
│   ├── Dockerfile              # Multi-stage build
│   ├── docker-compose.yml      # Service orchestration
│   ├── .dockerignore           # Build optimization
│   └── .env                    # Environment template
│
├── 📚 Documentation
│   ├── DOCKER_DEPLOYMENT.md    # Complete deployment guide
│   ├── API_GUIDE.md            # API documentation
│   ├── POSTMAN.md              # Postman usage guide
│   └── README_DOCKER.md        # This file
│
├── 🗄️ Data Storage (Local Volumes)
│   └── whatsapp/               # Volume mount point
│       ├── session_metadata.db # SQLite database
│       ├── sessions.db         # WhatsApp sessions DB
│       ├── sessions/           # Session files
│       ├── logs/              # Application logs
│       └── backups/           # Database backups
│
├── 🎯 Application Files
│   ├── main.go                # Go application
│   ├── go.mod, go.sum         # Go dependencies
│   └── frontend/              # React frontend
│       └── dist/              # Built frontend assets
│
└── 📋 API Documentation
    ├── api-docs.yaml          # OpenAPI specification
    └── postman-collection.json # Postman collection
```

## 🔧 Configuration Options (.env)

### 🔑 Required Settings
```env
# SECURITY: Change these in production!
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production-12345
ADMIN_PASSWORD=your-secure-admin-password

# Application
PORT=8080
DATABASE_PATH=/app/data/session_metadata.db

# Directories (local paths)
DATA_DIR=./whatsapp
WHATSAPP_DIR=./whatsapp/sessions
LOGS_DIR=./whatsapp/logs
```

### 🎛️ Optional Settings
```env
# Performance
MAX_SESSIONS=10
SESSION_TIMEOUT=24h
MEMORY_LIMIT=512M

# Security
CORS_ALLOWED_ORIGINS=*
RATE_LIMIT=100

# Features
AUTO_CONNECT=true
ENABLE_BACKUP=true
DEBUG=false
```

## 🚦 Docker Commands Reference

### Basic Operations
```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f

# Check status
docker-compose ps

# Restart services
docker-compose restart
```

### Management
```bash
# Rebuild container
docker-compose build --no-cache
docker-compose up -d

# Access container shell
docker-compose exec whatsapp-multi-session sh

# Backup database
docker-compose exec whatsapp-multi-session \
  cp /app/data/session_metadata.db \
  /app/data/backups/backup_$(date +%Y%m%d_%H%M%S).db
```

### Monitoring
```bash
# Health check
docker-compose ps

# Resource usage
docker stats whatsapp-multi-session

# Container logs
docker-compose logs -f --tail=100

# Check database files
ls -la whatsapp/
```

## 🌐 API Access

### Endpoints
- **Frontend:** http://localhost:8080
- **API Base:** http://localhost:8080/api
- **Login:** POST /api/login
- **Sessions:** GET /api/sessions
- **WebSocket:** ws://localhost:8080/api/ws/{sessionId}?token={jwt}

### Authentication
```bash
# Login
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'

# Use token in subsequent requests
curl -X GET http://localhost:8080/api/sessions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 🔐 Security Considerations

### ⚠️ Production Checklist
- [ ] Change `JWT_SECRET` to a strong random value
- [ ] Change `ADMIN_PASSWORD` to a secure password
- [ ] Set `CORS_ALLOWED_ORIGINS` to specific domains
- [ ] Use HTTPS with reverse proxy (nginx/traefik)
- [ ] Set up firewall rules for port 8080
- [ ] Enable regular database backups
- [ ] Monitor logs for suspicious activity

### 🛡️ Security Features
- **Non-root user** in container (uid: 1001)
- **JWT authentication** for API access
- **CORS protection** (configurable)
- **Rate limiting** (configurable)
- **Health checks** built-in
- **Resource limits** in Docker Compose

## 💾 Data Persistence

### Volume Mapping
All data is stored locally in the `whatsapp/` directory:
- **Database files:** Persistent across container restarts
- **Session data:** WhatsApp authentication persists
- **Logs:** Application logs are preserved
- **Backups:** Optional automatic database backups

### Backup Strategy
```bash
# Manual backup
docker-compose exec whatsapp-multi-session \
  sqlite3 /app/data/session_metadata.db .dump > backup.sql

# Restore from backup
docker-compose exec whatsapp-multi-session \
  sqlite3 /app/data/session_metadata.db < backup.sql
```

## 🚨 Troubleshooting

### Common Issues

**Port already in use:**
```bash
# Change port in .env
PORT=8081
docker-compose up -d
```

**Permission denied:**
```bash
# Fix directory permissions
sudo chown -R 1001:1001 whatsapp/
```

**Container won't start:**
```bash
# Check logs
docker-compose logs whatsapp-multi-session

# Check configuration
docker-compose config
```

**Database issues:**
```bash
# Remove lock files
rm -f whatsapp/*.db-wal whatsapp/*.db-shm
docker-compose restart
```

## 📊 Features

### ✅ Implemented
- [x] **SQLite-only database** (MySQL removed)
- [x] **Docker containerization** with multi-stage build
- [x] **Environment-based configuration** (.env support)
- [x] **Local volume persistence** (whatsapp/ directory)
- [x] **Health checks** and monitoring
- [x] **Security hardening** (non-root user, resource limits)
- [x] **Comprehensive documentation** (API, deployment, troubleshooting)
- [x] **Multi-session support** with persistent storage
- [x] **WebSocket QR code authentication**
- [x] **REST API** for message sending and management
- [x] **Webhook support** for incoming messages

### 🎯 Key Benefits
- **Simplified deployment** - Single Docker command
- **No external dependencies** - SQLite database included
- **Production ready** - Security best practices
- **Scalable configuration** - Environment variables
- **Data persistence** - Local volume mapping
- **Complete documentation** - Deployment and API guides

## 📞 Support

### Getting Help
1. **Check logs:** `docker-compose logs -f`
2. **Verify config:** `docker-compose config`
3. **Test connectivity:** `curl http://localhost:8080/api/sessions`
4. **Check resources:** `docker stats`
5. **Review documentation:** `DOCKER_DEPLOYMENT.md`

### Quick Commands
```bash
# Full restart
docker-compose down && docker-compose up -d

# View all logs
docker-compose logs -f --tail=50

# Check container health
docker inspect whatsapp-multi-session --format='{{.State.Health.Status}}'

# Access application
open http://localhost:8080
```

---

## 🎉 Success!

Your WhatsApp Multi-Session Manager is now fully containerized with:
- ✅ Complete Docker deployment
- ✅ SQLite database with local persistence
- ✅ Environment-based configuration
- ✅ Production-ready security
- ✅ Comprehensive documentation

**Ready to deploy!** 🚀