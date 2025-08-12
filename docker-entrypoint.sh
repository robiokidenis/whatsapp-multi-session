#!/bin/sh
set -e

echo "Starting WhatsApp Multi-Session Manager..."

# Create necessary directories if they don't exist
mkdir -p /app/data /app/database /app/logs

# Fix permissions for mounted volumes (running as root so this will work)
echo "Setting permissions for mounted directories (running as root)..."
chmod -R 755 /app/data /app/database /app/logs 2>/dev/null || true
chown -R root:root /app/data /app/database /app/logs 2>/dev/null || true

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

# Running as root, so we can manage permissions properly

echo "Environment configured. Starting application..."

# Execute the main application
exec "$@"