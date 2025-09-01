#!/bin/bash

# Production Database Fix Script for API Key Constraint Error
# This script fixes the duplicate empty API key issue in the production database

set -e

echo "🔧 WhatsApp Multi-Session - Fix API Key Constraint Issue"
echo "========================================================"
echo ""

# Check if running in Docker or directly
if [ -f /.dockerenv ]; then
    echo "📦 Running inside Docker container"
    IS_DOCKER=true
else
    echo "💻 Running on host system"
    IS_DOCKER=false
fi

# Function to execute MySQL query
execute_mysql() {
    local query="$1"
    
    if [ "$IS_DOCKER" = true ]; then
        # Inside Docker, use environment variables
        mysql -h"$MYSQL_HOST" -P"$MYSQL_PORT" -u"$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DATABASE" -e "$query"
    else
        # On host, check for docker-compose MySQL service
        if docker ps | grep -q mysql; then
            echo "Using Docker MySQL service..."
            docker exec -i $(docker ps | grep mysql | awk '{print $1}') mysql -u"$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DATABASE" -e "$query"
        else
            # Direct MySQL connection
            echo "Using direct MySQL connection..."
            mysql -h"${MYSQL_HOST:-localhost}" -P"${MYSQL_PORT:-3306}" -u"$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DATABASE" -e "$query"
        fi
    fi
}

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
    echo "✅ Loaded environment variables from .env"
else
    echo "⚠️  No .env file found, using existing environment variables"
fi

# Verify required variables
if [ -z "$MYSQL_DATABASE" ]; then
    echo "❌ Error: MYSQL_DATABASE not set"
    exit 1
fi

echo ""
echo "📊 Database: $MYSQL_DATABASE"
echo "🔗 Host: ${MYSQL_HOST:-localhost}"
echo ""

# Step 1: Check current status
echo "1️⃣ Checking for empty API keys..."
execute_mysql "SELECT COUNT(*) as count FROM users WHERE api_key = '';" || true

echo ""
echo "2️⃣ Listing affected users..."
execute_mysql "SELECT id, username, CASE WHEN api_key = '' THEN 'EMPTY' WHEN api_key IS NULL THEN 'NULL' ELSE 'SET' END as api_key_status FROM users WHERE api_key = '' OR api_key IS NULL;" || true

# Step 2: Fix the issue
echo ""
echo "3️⃣ Fixing empty API keys (setting them to NULL)..."
execute_mysql "UPDATE users SET api_key = NULL WHERE api_key = '';"

# Step 3: Verify the fix
echo ""
echo "4️⃣ Verifying the fix..."
execute_mysql "SELECT COUNT(*) as empty_count FROM users WHERE api_key = '';"

echo ""
echo "5️⃣ Final status check..."
execute_mysql "SELECT CASE WHEN api_key IS NULL THEN 'NULL (valid)' WHEN api_key = '' THEN 'EMPTY (invalid)' ELSE 'SET' END as api_key_status, COUNT(*) as user_count FROM users GROUP BY api_key_status;"

echo ""
echo "✅ Database fix completed!"
echo "🎉 You should now be able to create new users without the duplicate key error."
echo ""

# Optional: Test creating a new user
read -p "Would you like to test creating a new user? (y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    echo "Creating test user..."
    TIMESTAMP=$(date +%s)
    TEST_USERNAME="testuser_$TIMESTAMP"
    
    execute_mysql "INSERT INTO users (username, password_hash, api_key, role, session_limit, is_active, created_at) VALUES ('$TEST_USERNAME', 'test_hash', NULL, 'user', 5, 1, $TIMESTAMP);"
    
    echo "✅ Test user created successfully: $TEST_USERNAME"
    
    # Clean up test user
    read -p "Delete test user? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        execute_mysql "DELETE FROM users WHERE username = '$TEST_USERNAME';"
        echo "✅ Test user deleted"
    fi
fi

echo ""
echo "🏁 Script completed successfully!"