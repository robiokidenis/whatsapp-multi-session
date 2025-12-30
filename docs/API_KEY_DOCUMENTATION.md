# API Key Authentication

The WhatsApp Multi-Session API now supports API key authentication as an alternative to JWT tokens. This provides a more convenient way for developers to authenticate API requests without managing token expiration.

## Overview

- **API Keys**: Long-lived authentication tokens with `wams_` prefix
- **Backward Compatibility**: JWT authentication continues to work alongside API keys
- **User-specific**: Each user can have one API key at a time
- **Secure**: API keys are generated using cryptographically secure random bytes

## API Key Format

API keys follow this format:
```
wams_<base64-encoded-random-data>
```

Example: `wams_xYz123AbC456DeFgHiJkLmNoPqRsTuVwXyZ789`

## Authentication Methods

The API supports two authentication methods:

### 1. JWT Token (Bearer Token)
```bash
curl -H "Authorization: Bearer <jwt-token>" \
  https://api.example.com/api/sessions
```

### 2. API Key (Bearer Token)
```bash
curl -H "Authorization: Bearer wams_xYz123AbC456DeFgHiJkLmNoPqRsTuVwXyZ789" \
  https://api.example.com/api/sessions
```

The system automatically detects the authentication method based on the token format.

## API Key Management Endpoints

### Generate API Key
**POST** `/api/auth/api-key`

Generates a new API key for the authenticated user. If an API key already exists, it will be replaced.

**Headers:**
```
Authorization: Bearer <jwt-token-or-existing-api-key>
```

**Response:**
```json
{
  "success": true,
  "message": "API key generated successfully",
  "data": {
    "success": true,
    "message": "API key generated successfully",
    "api_key": "wams_xYz123AbC456DeFgHiJkLmNoPqRsTuVwXyZ789"
  }
}
```

### Get API Key Information
**GET** `/api/auth/api-key`

Returns information about the user's API key (without exposing the key itself).

**Headers:**
```
Authorization: Bearer <jwt-token-or-api-key>
```

**Response:**
```json
{
  "success": true,
  "message": "API key info retrieved successfully",
  "data": {
    "has_key": true,
    "created_at": "2024-01-01T00:00:00Z",
    "last_used": "2024-01-02T00:00:00Z"
  }
}
```

### Revoke API Key
**DELETE** `/api/auth/api-key`

Revokes the current API key for the authenticated user.

**Headers:**
```
Authorization: Bearer <jwt-token-or-api-key>
```

**Response:**
```json
{
  "success": true,
  "message": "API key revoked successfully"
}
```

## Admin API Key Management

Administrators can manage API keys for any user:

### Generate API Key for User (Admin)
**POST** `/api/admin/users/{userId}/api-key`

**Headers:**
```
Authorization: Bearer <admin-jwt-token-or-api-key>
```

### Revoke API Key for User (Admin)
**DELETE** `/api/admin/users/{userId}/api-key`

**Headers:**
```
Authorization: Bearer <admin-jwt-token-or-api-key>
```

## Usage Examples

### 1. Login and Generate API Key
```bash
# 1. Login to get JWT token
curl -X POST https://api.example.com/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "your_username",
    "password": "your_password"
  }'

# Response will include JWT token
# {
#   "success": true,
#   "data": {
#     "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
#     "user": {...}
#   }
# }

# 2. Generate API key using JWT token
curl -X POST https://api.example.com/api/auth/api-key \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Response will include API key
# {
#   "success": true,
#   "data": {
#     "api_key": "wams_xYz123AbC456DeFgHiJkLmNoPqRsTuVwXyZ789"
#   }
# }
```

### 2. Use API Key for All Subsequent Requests
```bash
# List sessions using API key
curl -H "Authorization: Bearer wams_xYz123AbC456DeFgHiJkLmNoPqRsTuVwXyZ789" \
  https://api.example.com/api/sessions

# Send message using API key
curl -X POST https://api.example.com/api/sessions/session1/send \
  -H "Authorization: Bearer wams_xYz123AbC456DeFgHiJkLmNoPqRsTuVwXyZ789" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "1234567890",
    "message": "Hello from API key!"
  }'

# Get conversations using API key
curl -H "Authorization: Bearer wams_xYz123AbC456DeFgHiJkLmNoPqRsTuVwXyZ789" \
  https://api.example.com/api/sessions/session1/conversations
```

### 3. Environment Variables and Scripts
```bash
# Set API key as environment variable
export WHATSAPP_API_KEY="wams_xYz123AbC456DeFgHiJkLmNoPqRsTuVwXyZ789"

# Use in scripts
curl -H "Authorization: Bearer $WHATSAPP_API_KEY" \
  https://api.example.com/api/sessions
```

## Security Considerations

1. **Store Securely**: Treat API keys like passwords - store them securely and never expose them in client-side code
2. **Rotate Regularly**: Generate new API keys periodically for enhanced security
3. **Revoke When Compromised**: Immediately revoke API keys if they are compromised
4. **Use HTTPS**: Always use HTTPS when transmitting API keys
5. **Environment Variables**: Store API keys in environment variables, not in code

## Migration from JWT to API Keys

For existing applications using JWT tokens:

1. **No Breaking Changes**: JWT authentication continues to work
2. **Gradual Migration**: Generate API keys for users and update client applications gradually
3. **Dual Support**: Both authentication methods work simultaneously

## Error Handling

### Invalid API Key
```json
{
  "success": false,
  "error": "Invalid API key",
  "code": "UNAUTHORIZED"
}
```

### Missing Authorization Header
```json
{
  "success": false,
  "error": "Missing authorization header",
  "code": "UNAUTHORIZED"
}
```

### User Account Disabled
```json
{
  "success": false,
  "error": "Account is disabled",
  "code": "UNAUTHORIZED"
}
```

## Postman Configuration

To use API keys in Postman:

1. **Collection Variables**: Add `api_key` variable to your collection
2. **Authorization**: Set to "Bearer Token" and use `{{api_key}}`
3. **Pre-request Script**: Not needed (unlike JWT tokens, API keys don't expire)

Example collection variable:
```
api_key: wams_xYz123AbC456DeFgHiJkLmNoPqRsTuVwXyZ789
```

## Database Schema

The API key feature adds an `api_key` column to the users table:

```sql
-- MySQL
ALTER TABLE users ADD COLUMN api_key VARCHAR(64) UNIQUE NULL;

-- SQLite
ALTER TABLE users ADD COLUMN api_key TEXT UNIQUE NULL;
```

The migration is handled automatically when the application starts.