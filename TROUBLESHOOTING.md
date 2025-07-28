# Troubleshooting Guide

## WhatsApp Foreign Key Constraint Error

### Problem
```
Failed to pair device: failed to store main device identity: FOREIGN KEY constraint failed
```

### Cause
This error occurs when the WhatsApp library's SQLite database has corrupted foreign key relationships, usually after database migrations or system changes.

### Solution

#### Quick Fix (Recommended)
1. **Stop the application**
2. **Run the reset script:**
   ```bash
   ./reset-whatsapp-db.sh
   ```
3. **Restart the application**
4. **Re-authenticate all WhatsApp sessions**

#### Manual Fix
1. **Stop the application**
2. **Backup existing databases:**
   ```bash
   cp ./database/sessions.db ./database/sessions.db.backup.$(date +%Y%m%d-%H%M%S)
   ```
3. **Remove problematic WhatsApp database:**
   ```bash
   rm -f ./database/sessions.db
   rm -f ./database/sessions.db-shm
   rm -f ./database/sessions.db-wal
   ```
4. **Clear session files:**
   ```bash
   rm -rf ./whatsapp/sessions/*
   ```
5. **Restart the application**

### What Gets Reset
- ✅ **WhatsApp authentication data** (sessions will need re-authentication)
- ✅ **Device identity keys** (new device identity will be generated)
- ❌ **Session metadata** (names, webhooks, etc. are preserved)
- ❌ **User accounts** (login credentials remain intact)

### Prevention
- Avoid force-stopping the application during session creation
- Regular database backups before major updates
- Use the proper shutdown procedure

### MySQL Configuration Issues

#### Problem
Application fails to start with MySQL connection errors.

#### Solution
1. **Check MySQL connection:**
   ```bash
   mysql -h localhost -u root -p -e "SELECT 1"
   ```

2. **Create database if needed:**
   ```sql
   CREATE DATABASE waGo CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   ```

3. **Update environment variables:**
   ```bash
   export DATABASE_TYPE=mysql
   export MYSQL_HOST=localhost
   export MYSQL_USER=root
   export MYSQL_PASSWORD=robioki
   export MYSQL_DATABASE=waGo
   ```

4. **For Docker, ensure MySQL service is running:**
   ```bash
   docker-compose -f docker-compose.yml -f docker-compose.mysql.yml up -d
   ```

## Common Issues

### Session Creation Fails
- **Check disk space** in `./database/` and `./whatsapp/` directories
- **Verify permissions** on database directories
- **Run reset script** if foreign key errors persist

### QR Code Not Displaying
- **Check session status** in the dashboard
- **Verify WebSocket connection** in browser dev tools
- **Restart session** if stuck in connecting state

### Messages Not Sending
- **Verify session is authenticated** (green status)
- **Check phone number format** (include country code)
- **Verify recipient exists** on WhatsApp

## Getting Help

If issues persist:
1. Check application logs for detailed error messages
2. Verify database connectivity
3. Ensure proper environment configuration
4. Consider database reset as last resort

## Database Architecture

- **Application Database**: MySQL/SQLite (configurable) - stores users, session metadata
- **WhatsApp Database**: SQLite only (required by whatsmeow library) - stores device keys, protocol data

Both databases work independently and the foreign key error specifically affects the WhatsApp database.