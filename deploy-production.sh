#!/bin/bash

# Production Deployment Script for WhatsApp Multi-Session Manager
# This script handles permission issues in production environments

set -e

echo "🚀 WhatsApp Multi-Session Manager - Production Deployment"
echo "========================================================="

# Detect OS
OS=$(uname -s)
if [ "$OS" = "Linux" ]; then
    echo "✅ Running on Linux (Production Environment)"
else
    echo "⚠️  Running on $OS (Development Environment)"
fi

# Option 1: Run container as root (simplest for production)
use_root_container() {
    echo "📦 Using root container approach (recommended for production)..."
    
    # Ensure directories exist
    mkdir -p ./whatsapp
    mkdir -p ./whatsapp/logs
    mkdir -p ./config
    
    # Set liberal permissions so container (running as root) can write
    chmod -R 777 ./whatsapp
    
    # Create sessions.db if it doesn't exist
    if [ ! -f ./whatsapp/sessions.db ]; then
        touch ./whatsapp/sessions.db
        chmod 666 ./whatsapp/sessions.db
    fi
    
    # Deploy with production override
    echo "🐳 Starting Docker containers with production configuration..."
    docker-compose -f docker-compose.yml -f docker-compose.prod.yml down
    docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
}

# Option 2: Fix host permissions (if you prefer non-root container)
fix_host_permissions() {
    echo "🔧 Fixing host permissions for non-root container..."
    
    # Create directories
    mkdir -p ./whatsapp
    mkdir -p ./whatsapp/logs
    mkdir -p ./config
    
    # Create sessions.db if it doesn't exist
    if [ ! -f ./whatsapp/sessions.db ]; then
        touch ./whatsapp/sessions.db
    fi
    
    # Change ownership to root (UID 0) since container runs as root
    echo "🔐 Setting ownership to root (UID 0 - container user)..."
    sudo chown -R 0:0 ./whatsapp
    sudo chown -R 0:0 ./config
    
    # Ensure write permissions
    sudo chmod -R u+w ./whatsapp
    sudo chmod 664 ./whatsapp/sessions.db
    
    # Deploy normally
    echo "🐳 Starting Docker containers..."
    docker-compose down
    docker-compose up -d
}

# Option 3: Use Docker volumes instead of bind mounts (most reliable)
use_docker_volumes() {
    echo "📂 Using Docker named volumes (most reliable)..."
    
    # Create a modified docker-compose with named volumes
    cat > docker-compose.volumes.yml << 'EOF'
version: '3.8'

services:
  whatsapp-multi-session:
    volumes:
      # Use named volumes instead of bind mounts
      - whatsapp-data:/app/data
      - whatsapp-db:/app/database
      - whatsapp-logs:/app/logs
      - ./config:/app/config:ro

volumes:
  whatsapp-data:
    driver: local
  whatsapp-db:
    driver: local
  whatsapp-logs:
    driver: local
EOF
    
    echo "🐳 Starting with Docker volumes..."
    docker-compose -f docker-compose.yml -f docker-compose.volumes.yml down
    docker-compose -f docker-compose.yml -f docker-compose.volumes.yml up -d
}

# Show menu
echo ""
echo "Choose deployment method:"
echo "1) Run container as root (simplest, recommended for production)"
echo "2) Fix host permissions (keep non-root container)"
echo "3) Use Docker volumes (most reliable, but data in Docker volumes)"
echo ""
read -p "Enter choice [1-3] (default: 1): " choice

case $choice in
    2)
        fix_host_permissions
        ;;
    3)
        use_docker_volumes
        ;;
    *)
        use_root_container
        ;;
esac

# Check deployment status
echo ""
echo "⏳ Waiting for container to be ready..."
sleep 5

# Check if container is running
if docker ps | grep -q whatsapp-multi-session; then
    echo "✅ Container is running!"
    
    # Show logs
    echo ""
    echo "📋 Recent logs:"
    docker logs --tail 20 whatsapp-multi-session
    
    # Health check
    echo ""
    echo "🏥 Health check:"
    curl -s http://localhost:8080/api/health || echo "⚠️  Health check failed (service may still be starting)"
    
    echo ""
    echo "✅ Deployment complete!"
    echo "🌐 Access the application at: http://$(hostname -I | awk '{print $1}'):8080"
    echo "👤 Default login: admin / admin123"
else
    echo "❌ Container failed to start. Check logs with: docker logs whatsapp-multi-session"
    exit 1
fi