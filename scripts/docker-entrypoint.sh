#!/bin/sh
set -e

echo "Starting WhatsApp Multi-Session Manager..."

# Create necessary directories if they don't exist
mkdir -p /app/data /app/database /app/logs

# Check if we can write to mounted volumes (running as www user)
echo "Checking write permissions for mounted directories (running as www user UID 1001)..."
if [ ! -w /app/data ] || [ ! -w /app/database ] || [ ! -w /app/logs ]; then
    echo "⚠️  Warning: Limited write permissions detected."
    echo "   For production, ensure host directories are owned by UID:GID 1001:1001 (www user)"
    echo "   Run: sudo chown -R 1001:1001 ./whatsapp ./config"
fi

# Initialize database if it doesn't exist
if [ ! -f "/app/data/session_metadata.db" ]; then
    echo "Database not found. Application will create it on first run."
fi

# Check if .env file was provided
if [ -f "/app/config/.env" ]; then
    echo "Loading environment from .env file..."
    export $(cat /app/config/.env | grep -v '^#' | xargs)
fi

# Set default values if not provided
export DATABASE_PATH="${DATABASE_PATH:-/app/data/session_metadata.db}"
export WHATSAPP_DB_PATH="${WHATSAPP_DB_PATH:-/app/database/sessions.db}"
export JWT_SECRET="${JWT_SECRET:-default-jwt-secret-change-in-production}"
export ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
export ADMIN_PASSWORD="${ADMIN_PASSWORD:-admin123}"

# Running as www user (UID 1001) for security

echo "Environment configured. Starting application..."

# Execute the main application
exec "$@"