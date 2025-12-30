# 🐳 Docker Deployment Guide

Complete guide for deploying WhatsApp Multi-Session Manager using Docker and Docker Compose.

## 📋 Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+
- At least 512MB RAM available
- 1GB free disk space

## 🚀 Quick Start

### 1. Clone and Setup
```bash
git clone <your-repo>
cd whatsapp-multi-session

# Create required directories
mkdir -p whatsapp/{sessions,logs,backups}
mkdir -p config
```

### 2. Configure Environment
```bash
# Copy environment template
cp .env.example .env

# Edit configuration (IMPORTANT: Change JWT_SECRET and ADMIN_PASSWORD)
nano .env
```

### 3. Deploy
```bash
# Build and start the container
docker-compose up -d

# Check logs
docker-compose logs -f

# Access the application
open http://localhost:8080
```

## 🔧 Configuration

### Environment Variables (.env)

**Required Settings:**
```env
# CHANGE THESE IN PRODUCTION!
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production-12345
ADMIN_PASSWORD=your-secure-admin-password

# Application
PORT=8080
DATABASE_PATH=/app/data/session_metadata.db

# Directories
DATA_DIR=./whatsapp
WHATSAPP_DIR=./whatsapp/sessions
LOGS_DIR=./whatsapp/logs
```

**Optional Settings:**
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
```

### Directory Structure
```
whatsapp-multi-session/
├── docker-compose.yml
├── Dockerfile
├── .env
├── whatsapp/               # Volume mount point
│   ├── session_metadata.db # SQLite database
│   ├── sessions.db         # WhatsApp sessions
│   ├── sessions/           # Session files
│   ├── logs/              # Application logs
│   └── backups/           # Database backups
└── config/                # Optional config files
```

## 🐳 Docker Commands

### Basic Operations
```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# Restart services
docker-compose restart

# View logs
docker-compose logs -f whatsapp-multi-session

# Follow logs (live)
docker-compose logs -f --tail=100

# Check status
docker-compose ps
```

### Management Commands
```bash
# Update and rebuild
docker-compose down
docker-compose build --no-cache
docker-compose up -d

# Scale (if needed)
docker-compose up -d --scale whatsapp-multi-session=2

# Execute commands in container
docker-compose exec whatsapp-multi-session sh

# Backup database
docker-compose exec whatsapp-multi-session cp /app/data/session_metadata.db /app/data/backup_$(date +%Y%m%d_%H%M%S).db
```

### Troubleshooting Commands
```bash
# Check container health
docker-compose exec whatsapp-multi-session wget -qO- http://localhost:8080/api/health

# View container resources
docker stats

# Inspect container
docker-compose exec whatsapp-multi-session ps aux
docker-compose exec whatsapp-multi-session df -h
```

## 📊 Monitoring

### Health Checks
The container includes built-in health monitoring:
```bash
# Check health status
docker inspect whatsapp-multi-session --format='{{.State.Health.Status}}'

# View health logs
docker inspect whatsapp-multi-session --format='{{range .State.Health.Log}}{{.Output}}{{end}}'
```

### Logs
```bash
# Application logs
docker-compose logs whatsapp-multi-session

# System logs (inside container)
docker-compose exec whatsapp-multi-session tail -f /app/logs/app.log

# Access logs
docker-compose exec whatsapp-multi-session tail -f /app/logs/access.log
```

### Resource Monitoring
```bash
# Memory and CPU usage
docker stats whatsapp-multi-session

# Disk usage
docker-compose exec whatsapp-multi-session du -sh /app/data
```

## 🔒 Security

### Production Security Checklist
- [ ] Change `JWT_SECRET` to a strong random value
- [ ] Change `ADMIN_PASSWORD` to a secure password  
- [ ] Set `CORS_ALLOWED_ORIGINS` to specific domains
- [ ] Enable firewall rules for port 8080
- [ ] Use HTTPS with reverse proxy (nginx/traefik)
- [ ] Regular database backups
- [ ] Monitor logs for suspicious activity

### Reverse Proxy Example (Nginx)
```nginx
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
    
    # WebSocket support for QR codes
    location /api/ws/ {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 💾 Backup & Recovery

### Automated Backups
The container supports automatic backups when `ENABLE_BACKUP=true`:
```bash
# Backups are stored in /app/data/backups/
# Format: session_metadata_YYYYMMDD_HHMMSS.db

# Manual backup
docker-compose exec whatsapp-multi-session \
  cp /app/data/session_metadata.db \
  /app/data/backups/manual_backup_$(date +%Y%m%d_%H%M%S).db
```

### Restore from Backup
```bash
# Stop the application
docker-compose down

# Restore database
cp whatsapp/backups/session_metadata_20240126_120000.db whatsapp/session_metadata.db

# Restart application
docker-compose up -d
```

### Migration
```bash
# Export data
docker-compose exec whatsapp-multi-session sqlite3 /app/data/session_metadata.db .dump > backup.sql

# Import to new instance
docker-compose exec whatsapp-multi-session sqlite3 /app/data/session_metadata.db < backup.sql
```

## 🚨 Troubleshooting

### Common Issues

**Port Already in Use:**
```bash
# Check what's using port 8080
lsof -i :8080

# Change port in .env file
PORT=8081
```

**Permission Denied:**
```bash
# Fix directory permissions
sudo chown -R 1001:1001 whatsapp/
chmod -R 755 whatsapp/
```

**Database Locked:**
```bash
# Stop container and check for lock files
docker-compose down
rm -f whatsapp/session_metadata.db-wal whatsapp/session_metadata.db-shm
docker-compose up -d
```

**Out of Memory:**
```bash
# Increase memory limit in docker-compose.yml
deploy:
  resources:
    limits:
      memory: 1G
```

### Debug Mode
```bash
# Enable debug logging
echo "DEBUG=true" >> .env
echo "LOG_LEVEL=debug" >> .env
docker-compose restart

# View debug logs
docker-compose logs -f
```

## 📈 Performance Tuning

### Resource Optimization
```yaml
# docker-compose.yml - adjust based on your needs
deploy:
  resources:
    limits:
      memory: 1G        # Increase for more sessions
      cpus: '2.0'       # Increase for better performance
    reservations:
      memory: 512M
      cpus: '1.0'
```

### Environment Tuning
```env
# .env - Performance settings
MAX_SESSIONS=20           # Increase session limit
SESSION_TIMEOUT=48h      # Longer session timeout
RATE_LIMIT=200          # Higher rate limit
AUTO_DISCONNECT_HOURS=48 # Longer auto-disconnect
```

## 🔄 Updates

### Update Container
```bash
# Pull latest changes
git pull

# Rebuild and restart
docker-compose down
docker-compose build --no-cache
docker-compose up -d

# Check health
docker-compose ps
docker-compose logs -f
```

### Zero-Downtime Updates (Advanced)
```bash
# Use blue-green deployment
cp docker-compose.yml docker-compose.blue.yml
# Update docker-compose.yml with new image
docker-compose -f docker-compose.yml up -d
# Wait for health check to pass
# Stop old version
docker-compose -f docker-compose.blue.yml down
```

## 📞 Support

### Getting Help
1. Check logs: `docker-compose logs -f`
2. Verify configuration: `docker-compose config`
3. Test connectivity: `curl http://localhost:8080/api/health`
4. Check resources: `docker stats`

### Useful Commands
```bash
# Full system information
docker system info
docker-compose version

# Container inspection
docker-compose exec whatsapp-multi-session env
docker-compose exec whatsapp-multi-session ps aux
docker-compose exec whatsapp-multi-session netstat -tlnp
```

---

## 🎯 Quick Reference

| Command | Description |
|---------|-------------|
| `docker-compose up -d` | Start in background |
| `docker-compose down` | Stop and remove |
| `docker-compose logs -f` | View logs |
| `docker-compose restart` | Restart services |
| `docker-compose ps` | Check status |
| `docker-compose exec whatsapp-multi-session sh` | Access container |

**Default Access:**
- **URL:** http://localhost:8080
- **Username:** admin
- **Password:** admin123 (change in .env!)
- **API:** http://localhost:8080/api
- **Health:** http://localhost:8080/api/health