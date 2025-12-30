#!/bin/bash

# Reset WhatsApp Database Script
# Use this script if you encounter foreign key constraint errors

echo "🔄 Resetting WhatsApp database to fix foreign key constraints..."

# Backup existing database if it exists
if [ -f "./database/sessions.db" ]; then
    echo "📦 Backing up existing WhatsApp database..."
    cp "./database/sessions.db" "./database/sessions.db.backup.$(date +%Y%m%d-%H%M%S)"
fi

# Remove the problematic WhatsApp database
echo "🗑️  Removing problematic WhatsApp database..."
rm -f "./database/sessions.db"
rm -f "./database/sessions.db-shm"
rm -f "./database/sessions.db-wal"

# Clear WhatsApp session files
echo "🧹 Clearing WhatsApp session files..."
rm -rf "./whatsapp/sessions/"*

echo "✅ WhatsApp database reset complete!"
echo ""
echo "ℹ️  Note: All WhatsApp sessions will need to be re-authenticated."
echo "ℹ️  Your session metadata (names, webhooks) is preserved."
echo ""
echo "🚀 You can now restart the application and create new sessions."