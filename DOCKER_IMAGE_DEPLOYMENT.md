# Docker Image Deployment Guide

This guide explains how to deploy WhatsApp Multi-Session Manager using the pre-built Docker image from Docker Hub.

## Quick Start

### 1. Using Make Commands (Recommended)

```bash
# Deploy with Docker Hub image
make deploy-image

# Or if already pulled, just start
make deploy-image-start
```

### 2. Using docker-compose directly

```bash
# Pull and start
docker-compose -f docker-compose.image.yml up -d

# Check status
docker-compose -f docker-compose.image.yml ps

# View logs
docker-compose -f docker-compose.image.yml logs -f
```

## Available Make Commands

| Command | Description |
|---------|-------------|
| `make deploy-image` | Pull and deploy the Docker Hub image |
| `make deploy-image-start` | Start without pulling (if already downloaded) |
| `make stop-image` | Stop the application |
| `make restart-image` | Restart the application |
| `make logs-image` | View application logs |
| `make status-image` | Check application status |
| `make update-image` | Pull latest image and restart |

## Configuration

### Environment Variables (.env)

Create a `.env` file in the project root:

```env
# Application Port
PORT=8080

# JWT Secret (change this!)
JWT_SECRET=your-super-secret-jwt-key-here

# Environment (release/debug)
GIN_MODE=release

# MySQL Configuration
MYSQL_HOST=mysql
MYSQL_PORT=3306
MYSQL_USER=whatsapp
MYSQL_PASSWORD=whatsapp123
MYSQL_DATABASE=waGo
MYSQL_ROOT_PASSWORD=rootpassword

# Admin User
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123

# Session Configuration
MAX_SESSIONS=10
SESSION_TIMEOUT=24h
QR_TIMEOUT=30s

# Webhook Configuration
WEBHOOK_TIMEOUT=30s
```

### Volume Mounts

The docker-compose.image.yml automatically creates and mounts:

- `./whatsapp` - WhatsApp session databases
- `./config` - Configuration files
- `./logs` - Application logs

## Using Specific Image Versions

### Option 1: Environment Variable

```bash
DOCKER_TAG=v1.5.2 make deploy-image
```

### Option 2: Edit docker-compose.image.yml

```yaml
services:
  whatsapp-multi-session:
    image: rod16/whatsapp-multi-session:v1.5.2
```

Then run:
```bash
docker-compose -f docker-compose.image.yml up -d
```

## Docker Hub Image

- **Repository**: [rod16/whatsapp-multi-session](https://hub.docker.com/r/rod16/whatsapp-multi-session)
- **Tags Available**:
  - `latest` - Most recent stable release
  - `v1.5.2` - Versioned releases

## Pulling the Image Manually

```bash
# Pull latest
docker pull rod16/whatsapp-multi-session:latest

# Pull specific version
docker pull rod16/whatsapp-multi-session:v1.5.2

# Pull all tags
docker pull rod16/whatsapp-multi-session
```

## Running without docker-compose

```bash
docker run -d \
  --name whatsapp-multi-session \
  -p 8080:8080 \
  -v $(pwd)/whatsapp:/app/database \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/logs:/app/logs \
  -e PORT=8080 \
  -e JWT_SECRET=your-secret-here \
  -e MYSQL_HOST=host.docker.internal \
  -e MYSQL_PORT=3306 \
  rod16/whatsapp-multi-session:latest
```

## Troubleshooting

### Check if container is running

```bash
docker ps | grep whatsapp-multi-session
```

### View container logs

```bash
# Using make
make logs-image

# Or directly
docker-compose -f docker-compose.image.yml logs -f

# Or Docker logs
docker logs whatsapp-multi-session
```

### Restart the container

```bash
make restart-image
```

### Stop and remove everything

```bash
make stop-image
docker-compose -f docker-compose.image.yml down -v
```

### Reset and start fresh

```bash
# Stop containers
make stop-image

# Remove volumes (WARNING: This deletes all data!)
docker-compose -f docker-compose.image.yml down -v

# Remove WhatsApp data directory
rm -rf whatsapp config logs

# Deploy fresh
make deploy-image
```

## Updating to Latest Image

```bash
# Pull latest and restart
make update-image

# Or manually
docker-compose -f docker-compose.image.yml pull
docker-compose -f docker-compose.image.yml up -d
```

## Benefits of Using Docker Hub Image

✅ **Faster Deployment** - No build time, just pull and run
✅ **Smaller Size** - Optimized multi-stage build (61MB)
✅ **Version Control** - Use specific tags for reproducibility
✅ **Easy Updates** - Simple pull and restart
✅ **Cross-Platform** - Works on any system with Docker

## Differences from docker-compose.yml

| Feature | docker-compose.yml | docker-compose.image.yml |
|---------|-------------------|-------------------------|
| Build Time | ~5 minutes | Instant (pull only) |
| Image Size | Full build | Optimized 61MB |
| Updates | Rebuild required | Pull latest |
| Customization | Full source code access | Config via env vars only |
| Use Case | Development | Production deployment |

## Accessing the Application

After deployment:

- **Web Interface**: http://localhost:8080
- **Default Credentials**:
  - Username: `admin`
  - Password: `admin123`

⚠️ **IMPORTANT**: Change the default credentials after first login!

## Support

For issues and questions:
- Check the main [README.md](../README.md)
- View logs: `make logs-image`
- Check status: `make status-image`
- Open an issue on GitHub
