# Database Logging Configuration

This application supports flexible logging configuration that allows you to control where logs are stored and displayed.

## Configuration Options

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ENABLE_LOGGING` | `true` | Enable/disable console logging |
| `ENABLE_DATABASE_LOG` | `true` | Enable/disable database logging |
| `LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` |

## Logging Modes

### 1. Full Logging (Default)
```bash
ENABLE_LOGGING=true
ENABLE_DATABASE_LOG=true
```

**Features:**
- ✅ Logs displayed in console
- ✅ Logs stored in database
- ✅ Web interface shows logs with filtering, search, pagination
- ✅ Auto-refresh functionality
- ✅ Delete and cleanup options
- ✅ Export capabilities

### 2. Console Only Logging
```bash
ENABLE_LOGGING=true
ENABLE_DATABASE_LOG=false
```

**Features:**
- ✅ Logs displayed in console
- ❌ No database storage
- ❌ Web interface shows informative message
- ✅ Better performance (no DB writes)
- ✅ Suitable for development/debugging

### 3. No Logging
```bash
ENABLE_LOGGING=false
ENABLE_DATABASE_LOG=false
```

**Features:**
- ❌ No console output
- ❌ No database storage
- ⚠️ Only errors will be shown
- ✅ Maximum performance

## Usage Examples

### Development Mode (Console Only)
```bash
# Set environment variable
export ENABLE_DATABASE_LOG=false

# Or run directly
ENABLE_DATABASE_LOG=false ./server
```

### Production Mode (Full Logging)
```bash
# Set environment variable
export ENABLE_DATABASE_LOG=true

# Or run directly  
ENABLE_DATABASE_LOG=true ./server
```

### Using .env File
Create a `.env.local` file:
```bash
# Disable database logging for development
ENABLE_DATABASE_LOG=false
LOG_LEVEL=debug

# Local database paths
DATABASE_PATH=./database/session_metadata.db
WHATSAPP_DB_PATH=./database/sessions.db
```

## Web Interface Behavior

### When Database Logging is Enabled
- Navigate to `/logs` in the admin interface
- Full log management interface with:
  - Real-time log viewing
  - Filtering by level, component, session, time
  - Pagination with latest-first ordering
  - Delete old logs (7+ days, 30+ days)
  - Clear all logs functionality
  - Auto-refresh toggle
  - Export options

### When Database Logging is Disabled
- Navigate to `/logs` in the admin interface
- Informative screen showing:
  - Current logging configuration
  - Instructions to enable database logging
  - Status of console logging and log level
  - Clear guidance for configuration changes

## Performance Considerations

### Database Logging Enabled
- **Pros**: Full audit trail, web interface, search capabilities
- **Cons**: Additional I/O operations, database storage usage

### Database Logging Disabled  
- **Pros**: Better performance, less disk usage, simpler setup
- **Cons**: No historical log storage, no web interface

## Troubleshooting

### Check Current Configuration
The server shows the current configuration on startup:
```
2025/07/29 07:11:12 [INFO] Configuration loaded - Port: 8080, Log Level: info, Database Logging: false
```

### Verify Database Logging Status
Access the status endpoint (requires admin authentication):
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     http://localhost:8080/api/admin/logs/status
```

Response:
```json
{
  "database_logging_enabled": false,
  "console_logging_enabled": true,
  "log_level": "info"
}
```

### Common Issues

**Error: "mkdir /app: read-only file system"**
- Solution: Set correct database paths for local development:
  ```bash
  DATABASE_PATH=./database/session_metadata.db
  WHATSAPP_DB_PATH=./database/sessions.db
  ```

**Logs not appearing in web interface**
- Check if database logging is enabled
- Verify admin authentication
- Check browser console for errors
- Confirm API endpoints are accessible

## Best Practices

1. **Development**: Use console-only logging for faster iteration
2. **Testing**: Enable database logging to verify log storage
3. **Production**: Enable full logging for audit trails
4. **Performance**: Disable database logging for high-traffic scenarios
5. **Debugging**: Use `debug` log level with console-only mode

## Database Schema

When database logging is enabled, logs are stored in the `logs` table:

```sql
CREATE TABLE logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    level TEXT NOT NULL,           -- debug, info, warn, error
    message TEXT NOT NULL,         -- log message
    component TEXT,                -- application component
    session_id TEXT,               -- WhatsApp session ID
    user_id INTEGER,               -- user ID (if applicable)
    metadata TEXT,                 -- JSON metadata
    created_at INTEGER NOT NULL    -- Unix timestamp
);
```

The table includes indexes for efficient querying by level, component, session, and time.