#!/bin/bash

# Fix Docker Volume Permissions for WhatsApp Multi-Session Manager
# This script fixes permission issues for Docker volumes in production

echo "🔧 Fixing Docker volume permissions..."

# Check if running as root (required for chown)
if [ "$EUID" -ne 0 ]; then 
    echo "⚠️  This script must be run with sudo for production fixes"
    echo "   Run: sudo ./fix-docker-permissions.sh"
    exit 1
fi

# Docker container runs as UID 1001 (appuser)
DOCKER_UID=1001
DOCKER_GID=1001

# Create directories if they don't exist
echo "📁 Creating necessary directories..."
mkdir -p ./whatsapp
mkdir -p ./whatsapp/logs
mkdir -p ./config

# Fix ownership for all mounted volumes
echo "🔐 Setting ownership to UID:GID ${DOCKER_UID}:${DOCKER_GID}..."
chown -R ${DOCKER_UID}:${DOCKER_GID} ./whatsapp
chown -R ${DOCKER_UID}:${DOCKER_GID} ./config 2>/dev/null || true

# Ensure write permissions
echo "✏️  Setting write permissions..."
chmod -R u+w ./whatsapp
chmod 664 ./whatsapp/sessions.db 2>/dev/null || true

# Display results
echo "✅ Permissions fixed! Current status:"
ls -la ./whatsapp/
echo ""
echo "📝 sessions.db permissions:"
ls -la ./whatsapp/sessions.db 2>/dev/null || echo "   (will be created on first run)"

echo ""
echo "🚀 You can now restart your Docker container:"
echo "   docker-compose restart"