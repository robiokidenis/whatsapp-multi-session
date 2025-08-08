# Session Proxy Configuration Guide

This guide explains how to configure individual proxy settings for each WhatsApp session.

## Database Migration

First, run the proxy columns migration:

### For SQLite:
```bash
sqlite3 database/session_metadata.db < migrations/add_proxy_columns.sql
```

### For MySQL:
```bash
mysql -u your_username -p your_database < migrations/mysql_add_proxy_columns.sql
```

## API Usage

### 1. Create Session with Proxy

```bash
curl -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "name": "My Proxy Session",
    "proxy_config": {
      "enabled": true,
      "type": "socks5",
      "host": "127.0.0.1",
      "port": 1080,
      "username": "proxy_user",
      "password": "proxy_pass"
    }
  }'
```

### 2. Update Session Proxy

```bash
curl -X PUT http://localhost:8080/api/sessions/SESSION_ID \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "proxy_config": {
      "enabled": true,
      "type": "http",
      "host": "proxy.example.com",
      "port": 8080,
      "username": "",
      "password": ""
    }
  }'
```

### 3. Disable Proxy

```bash
curl -X PUT http://localhost:8080/api/sessions/SESSION_ID \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "proxy_config": {
      "enabled": false
    }
  }'
```

## Proxy Types Supported

- `http` - HTTP proxy
- `https` - HTTPS proxy  
- `socks5` - SOCKS5 proxy

## Configuration Example

```json
{
  "proxy_config": {
    "enabled": true,
    "type": "socks5",
    "host": "127.0.0.1",
    "port": 1080,
    "username": "optional_username",
    "password": "optional_password"
  }
}
```

## Implementation Status

✅ **Completed:**
- Added ProxyConfig model
- Updated database schema (both SQLite and MySQL)
- Created migration scripts
- Updated session models to include proxy config

🚧 **To Complete:**
- Update all repository methods for proxy fields
- Implement WhatsApp client proxy configuration
- Add proxy validation
- Test proxy connectivity

## Next Steps

1. Run the migration to add proxy columns
2. The proxy configuration will be stored in the database
3. Each session can have its own unique proxy settings
4. Proxy settings persist across server restarts

## Security Note

Proxy passwords are stored in the database. Consider encrypting sensitive proxy credentials in production environments.