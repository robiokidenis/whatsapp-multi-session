#!/bin/bash

echo "Initializing Docker environment for WhatsApp Multi-Session Manager..."

# Create necessary directories
echo "Creating required directories..."
mkdir -p whatsapp/sessions
mkdir -p whatsapp/logs
mkdir -p whatsapp/backups
mkdir -p config
mkdir -p database

# Copy .env.example to .env if .env doesn't exist
if [ ! -f .env ]; then
    echo "Creating .env file from .env.example..."
    cp .env.example .env
    echo "⚠️  Please edit .env file to set your configuration!"
fi

# Set proper permissions
echo "Setting directory permissions..."
chmod -R 755 whatsapp
chmod -R 755 config
chmod -R 755 database

# Set proper ownership for Docker container (user ID 1001)
echo "Setting directory ownership for Docker container..."
if command -v sudo >/dev/null 2>&1; then
    sudo chown -R 1001:1001 ./whatsapp ./config
    echo "✅ Directory ownership set to 1001:1001"
else
    echo "⚠️  sudo not available. You may need to run manually:"
    echo "   sudo chown -R 1001:1001 ./whatsapp ./config"
fi

# Check if frontend needs to be built
if [ ! -d "frontend/dist" ]; then
    echo "Frontend dist not found. Building frontend..."
    if [ -d "frontend" ] && [ -f "frontend/package.json" ]; then
        cd frontend
        echo "Installing frontend dependencies..."
        npm install
        echo "Building frontend..."
        npm run build
        cd ..
    else
        echo "⚠️  Frontend directory not found. Frontend will be built during Docker build."
    fi
fi

echo "✅ Docker environment initialized successfully!"
echo ""
echo "Next steps:"
echo "1. Edit .env file with your configuration"
echo "2. Run: docker-compose up -d"
echo "3. Access the application at http://localhost:${PORT:-8080}"
echo ""
echo "Default credentials:"
echo "Username: admin"
echo "Password: admin123"
