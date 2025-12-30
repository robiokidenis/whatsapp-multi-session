# WhatsApp Multi-Session - Troubleshooting Guide

This guide helps you diagnose and fix common issues with the WhatsApp Multi-Session Manager.

## Table of Contents

1. [Database Issues](#database-issues)
2. [Connection Issues](#connection-issues)
3. [Authentication Issues](#authentication-issues)
4. [Message Sending Issues](#message-sending-issues)
5. [Performance Issues](#performance-issues)
6. [Docker Issues](#docker-issues)
7. [Library Updates](#library-updates)

---

## Database Issues

### Symptom: "failed to create WhatsApp store: failed to upgrade database"

**Cause:** SQLite database is corrupted or has permission issues.

**Solutions:**
1. **Quick Fix (Recommended):**
   ```bash
   ./scripts/reset-whatsapp-db.sh
   ```

2. **Manual Fix:**
   ```bash
   rm -f whatsapp/sessions.db
   docker-compose restart
   ```

3. **Permission Fix:**
   ```bash
   chmod 666 whatsapp/sessions.db
   docker-compose restart
   ```

4. **Check File Ownership:**
   ```bash
   ls -la whatsapp/
   # Files should be readable by container user (UID 1001)
   ```

### Symptom: "attempt to write a readonly database"

**Cause:** Database file permissions don't allow container user (UID 1001) to write.

**Solutions:**
1. **Fix Permissions:**
   ```bash
   chmod 666 whatsapp/sessions.db
   ```

2. **Fix Ownership:**
   ```bash
   sudo chown 1001:1001 whatsapp/sessions.db
   ```

3. **Delete and Recreate:**
   ```bash
   rm whatsapp/sessions.db
   docker-compose restart
   ```

### Symptom: "failed to check if foreign keys are enabled"

**Cause:** Database file doesn't exist or is corrupted.

**Solutions:**
1. **Delete Database:**
   ```bash
   rm whatsapp/sessions.db
   ```

2. **Restart Container:**
   ```bash
   docker-compose restart
   ```

3. **Container will create fresh database automatically**

### Symptom: "Duplicate entry '' for key 'messages.message_id'"

**Cause:** Attempting to log messages with empty message IDs (fixed in current version).

**Solution:** This is fixed in the latest code. Update your application:
```bash
git pull
docker-compose up -d --build
```

---

## Connection Issues

### Symptom: "received Connected event but client is not actually connected"

**Cause:** whatsmeow library fired Connected event but websocket failed to establish.

**Solutions:**
1. **Check Network Connectivity:**
   - Verify container has internet access
   - Check firewall/proxy settings
   - Ensure WhatsApp servers are reachable

2. **Update whatsmeow Library:**
   ```bash
   go get go.mau.fi/whatsmeow@latest
   go mod tidy
   docker-compose up -d --build
   ```

3. **Check whatsmeow Version:**
   ```bash
   go list -m go.mau.fi/whatsmeow
   ```

4. **Delete and Recreate Session:**
   - Delete the problematic session
   - Create a new session
   - Scan QR code again

5. **Check Logs:**
   ```bash
   docker logs -f whatsapp-multi-session
   ```

### Symptom: "Stream error, disconnecting"

**Cause:** WhatsApp server closed the connection due to protocol error or session issue.

**Solutions:**
1. **Update whatsmeow Library:**
   ```bash
   go get go.mau.fi/whatsmeow@latest
   go mod tidy
   docker-compose up -d --build
   ```

2. **Re-authenticate Session:**
   - Disconnect session
   - Scan QR code again

3. **Check for Logged Out Elsewhere:**
   - Session might be logged out from another device
   - Re-authenticate by scanning QR code

4. **Delete and Recreate Session:**
   - Clear device store: delete whatsapp/sessions.db
   - Recreate session

### Symptom: Connection succeeds but messages fail with "websocket not connected"

**Cause:** Race condition between connection event and actual socket ready state.

**Solutions:**
1. **This should be fixed in current code** - verify you have latest version

2. **Add Delay Before Sending:**
   - Wait 5-10 seconds after connection before sending messages

3. **Check Connection Status:**
   ```bash
   # Check session status
   curl http://localhost:8080/api/sessions
   ```

4. **Reconnect Session:**
   - Disconnect
   - Wait 5 seconds
   - Connect again

---

## Authentication Issues

### Symptom: "Failed to set online presence: can't send presence without PushName set"

**Cause:** Device store missing PushName required by WhatsApp.

**Solutions:**
1. **This is auto-fixed in current code** by generating random PushName

2. **If Issue Persists, Manual Fix:**
   - Check code in `internal/services/whatsapp_service.go`
   - Look for `generateRandomName()` function
   - Ensure it's being called during connection

### Symptom: QR Code Not Loading

**Cause:** Session not properly initialized or connection timeout.

**Solutions:**
1. **Check Session Status:**
   ```bash
   curl http://localhost:8080/api/sessions
   ```

2. **Recreate Session:**
   - Delete session
   - Create new session
   - Try again

3. **Check Logs:**
   ```bash
   docker logs -f whatsapp-multi-session
   ```

---

## Message Sending Issues

### Symptom: 500 Internal Server Error when sending messages

**Cause:** Session not connected or not authenticated.

**Solutions:**
1. **Check Session Status:**
   ```bash
   curl http://localhost:8080/api/sessions
   ```

2. **Verify Session is Connected and Logged In:**
   - Status should show: `"connected": true, "loggedIn": true`

3. **Re-authenticate if Needed:**
   - Scan QR code again

4. **Check Logs for Specific Error:**
   ```bash
   docker logs -f whatsapp-multi-session
   ```

### Symptom: "session is not connected" error (503)

**Cause:** Session disconnected or never connected.

**Solutions:**
1. **Connect Session First:**
   ```bash
   curl -X POST http://localhost:8080/api/sessions/{session_id}/connect \
     -H "Authorization: Bearer YOUR_TOKEN"
   ```

2. **Authenticate Session:**
   - Scan QR code

3. **Wait for Connection:**
   - Wait 5-10 seconds after connecting before sending messages

---

## Performance Issues

### Symptom: Slow message delivery

**Cause:** Too many concurrent sessions or network latency.

**Solutions:**
1. **Check Session Count:**
   ```bash
   curl http://localhost:8080/api/sessions
   ```

2. **Limit Concurrent Sessions:**
   - Reduce number of active sessions
   - Disable unused sessions

3. **Check Network Speed:**
   - Verify container has sufficient bandwidth

4. **Monitor Resources:**
   ```bash
   docker stats whatsapp-multi-session
   ```

### Symptom: High memory usage

**Cause:** Memory leak or too many sessions.

**Solutions:**
1. **Restart Container:**
   ```bash
   docker-compose restart
   ```

2. **Limit Sessions:**
   - Reduce number of concurrent sessions

3. **Monitor Memory:**
   ```bash
   docker stats whatsapp-multi-session
   ```

---

## Docker Issues

### Symptom: Container won't start

**Cause:** Port conflict or volume mount issues.

**Solutions:**
1. **Check Port Availability:**
   ```bash
   lsof -i :8080
   ```

2. **Check Volume Mounts:**
   ```bash
   docker-compose config
   ```

3. **Check Logs:**
   ```bash
   docker logs whatsapp-multi-session
   ```

### Symptom: "Permission denied" errors

**Cause:** File permissions on mounted volumes.

**Solutions:**
1. **Fix Permissions:**
   ```bash
   ./scripts/fix-docker-permissions.sh
   ```

2. **Or Manual Fix:**
   ```bash
   sudo chown -R 1001:1001 ./whatsapp
   sudo chown -R 1001:1001 ./config
   ```

---

## Library Updates

### Updating whatsmeow Library

**When to Update:**
- Connection issues not resolved by other fixes
- Protocol errors with WhatsApp
- New features needed

**How to Update:**
```bash
# Update the library
go get go.mau.fi/whatsmeow@latest

# Update dependencies
go mod tidy

# Rebuild and restart
docker-compose down
docker-compose up -d --build
```

**Check Current Version:**
```bash
go list -m go.mau.fi/whatsmeow
```

**Check for Updates:**
```bash
# List available updates
go list -m -u all

# Check whatsmeow GitHub for latest releases
# https://github.com/mautic/whatsmeow/releases
```

---

## Getting Help

### Check Logs First

Always check logs for detailed error messages:
```bash
docker logs -f whatsapp-multi-session
```

### Useful Commands

```bash
# Check container status
docker ps

# Check container logs
docker logs whatsapp-multi-session

# Check container resources
docker stats whatsapp-multi-session

# Restart container
docker-compose restart

# Rebuild container
docker-compose up -d --build

# Enter container shell
docker exec -it whatsapp-multi-session sh
```

### Common Script Locations

- Database reset: `./scripts/reset-whatsapp-db.sh`
- Permission fix: `./scripts/fix-docker-permissions.sh`
- Production deploy: `./scripts/deploy-production.sh`

### Report Issues

When reporting issues, include:
1. Full error message from logs
2. whatsmeow version: `go list -m go.mau.fi/whatsmeow`
3. Docker version: `docker --version`
4. System information
5. Steps to reproduce

### Useful Resources

- whatsmeow GitHub: https://github.com/mautic/whatsmeow
- whatsmeow Issues: https://github.com/mautic/whatsmeow/issues
- WhatsApp Web Protocol: https://github.com/FKLC/WhatsApp-Protocol

---

## Quick Reference

### Most Common Issues and Quick Fixes

1. **Database Error:**
   ```bash
   rm -f whatsapp/sessions.db && docker-compose restart
   ```

2. **Permission Error:**
   ```bash
   chmod 666 whatsapp/sessions.db && docker-compose restart
   ```

3. **Connection Issues:**
   ```bash
   ./scripts/reset-whatsapp-db.sh
   ```

4. **Update whatsmeow:**
   ```bash
   go get go.mau.fi/whatsmeow@latest && docker-compose up -d --build
   ```

5. **Complete Reset:**
   ```bash
   docker-compose down
   rm -rf whatsapp/*
   docker-compose up -d
   ```
