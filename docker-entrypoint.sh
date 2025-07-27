#!/bin/sh
set -e

echo "Starting WhatsApp Multi-Session Manager..."

# Create necessary directories if they don't exist
mkdir -p /app/data /app/sessions /app/logs /app/whatsapp/sessions /app/whatsapp/logs

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
export WHATSAPP_DB_PATH="${WHATSAPP_DB_PATH:-/app/data/sessions.db}"
export JWT_SECRET="${JWT_SECRET:-default-jwt-secret-change-in-production}"
export ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
export ADMIN_PASSWORD="${ADMIN_PASSWORD:-admin123}"

# Note: Cannot change ownership as we're running as non-root user
# The directories should already have correct permissions from Dockerfile

echo "Environment configured. Starting application..."

# Execute the main application
exec "$@"