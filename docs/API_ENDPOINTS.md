# WhatsApp Multi-Session API Endpoints

This document lists all available API endpoints after the restructure.

## Base URL
All endpoints are prefixed with `/api`

## Authentication
Most endpoints require a Bearer token in the Authorization header:
```
Authorization: Bearer <token>
```

## Public Endpoints (No Authentication Required)

### POST /api/auth/login
Login with username and password
```json
{
  "username": "admin",
  "password": "admin123"
}
```

### GET /api/health
Health check endpoint

## Session Management (Authentication Required)

### GET /api/sessions
Get all sessions

### POST /api/sessions
Create a new session
```json
{
  "name": "Alex Johnson",
  "phone": "628123456789",
  "webhook_url": "https://example.com/webhook"
}
```

### GET /api/sessions/{sessionId}
Get specific session details

### PUT /api/sessions/{sessionId}
Update session
```json
{
  "name": "Updated Session Name",
  "webhook_url": "https://example.com/new-webhook"
}
```

### DELETE /api/sessions/{sessionId}
Delete a session

### POST /api/sessions/{sessionId}/connect
Connect a session to WhatsApp

### POST /api/sessions/{sessionId}/disconnect
Disconnect a session

### GET /api/sessions/{sessionId}/qr
Get QR code for session login

### GET /api/sessions/{sessionId}/ws
WebSocket endpoint for real-time updates

## Message Endpoints (Authentication Required)

### POST /api/sessions/{sessionId}/send
Send text message
```json
{
  "to": "628987654321@s.whatsapp.net",
  "message": "Hello World!"
}
```

### POST /api/send
General send endpoint (for compatibility)
```json
{
  "session_id": "your_session_id",
  "to": "628987654321@s.whatsapp.net",
  "message": "Hello World!"
}
```

### POST /api/sessions/{sessionId}/send-attachment
Send file attachment
```json
{
  "to": "628987654321@s.whatsapp.net",
  "file": "base64_encoded_file_data",
  "filename": "document.pdf",
  "caption": "Document caption"
}
```

### POST /api/sessions/{sessionId}/check-number
Check if number is on WhatsApp
```json
{
  "number": "628987654321"
}
```

### POST /api/sessions/{sessionId}/typing
Send typing indicator
```json
{
  "to": "628987654321@s.whatsapp.net"
}
```

### POST /api/sessions/{sessionId}/stop-typing
Stop typing indicator
```json
{
  "to": "628987654321@s.whatsapp.net"
}
```

### POST /api/sessions/{sessionId}/presence
Set session presence status (online/offline)
```json
{
  "status": "available"
}
```
Valid status values: `available`, `online`, `unavailable`, `offline`

### GET /api/sessions/{sessionId}/groups
Get all groups for a session

### GET /api/sessions/{sessionId}/conversations
Get all conversations/chats for a session (contacts and groups)
Response:
```json
{
  "success": true,
  "message": "Conversations retrieved successfully",
  "data": {
    "conversations": [
      {
        "jid": "628123456789@s.whatsapp.net",
        "name": "John Doe",
        "is_group": false,
        "last_message_id": "",
        "last_message": "",
        "last_message_time": null,
        "unread_count": 0,
        "is_pinned": false,
        "is_muted": false,
        "is_archived": false,
        "avatar": ""
      },
      {
        "jid": "123456789-1234567890@g.us",
        "name": "Group Chat",
        "is_group": true,
        "last_message_id": "",
        "last_message": "",
        "last_message_time": null,
        "unread_count": 0,
        "is_pinned": false,
        "is_muted": false,
        "is_archived": false,
        "avatar": ""
      }
    ],
    "count": 2
  }
}
```

## Admin User Management (Admin Role Required)

### POST /api/auth/register
Register a new user account (Admin only)
```json
{
  "username": "newuser",
  "password": "password123"
}
```

### GET /api/admin/users
Get all users

### POST /api/admin/users
Create a new user
```json
{
  "username": "newuser",
  "password": "password123",
  "role": "user",
  "session_limit": 5
}
```

### GET /api/admin/users/{id}
Get specific user

### PUT /api/admin/users/{id}
Update user
```json
{
  "username": "updated_username",
  "role": "admin",
  "session_limit": 10,
  "is_active": true
}
```

### DELETE /api/admin/users/{id}
Delete a user

## Webhook Format

When webhook_url is configured for a session, incoming messages will be sent to that URL with this format:

```json
{
  "session_id": "session_123",
  "from": "628987654321@s.whatsapp.net",
  "to": "628123456789@s.whatsapp.net",
  "message": "Message content",
  "message_type": "text",
  "timestamp": "2024-01-01T12:00:00Z",
  "id": "message_id_123",
  "is_group": false,
  "group_id": "",
  "media_url": ""
}
```

## Environment Variables

- `PORT`: Server port (default: 8080)
- `DATABASE_PATH`: SQLite database path (default: ./database/session_metadata.db)
- `WHATSAPP_DB_PATH`: WhatsApp sessions database path (default: ./database/sessions.db)
- `JWT_SECRET`: JWT signing secret
- `ADMIN_USERNAME`: Default admin username (default: admin)
- `ADMIN_PASSWORD`: Default admin password (default: admin123)
- `ENABLE_LOGGING`: Enable logging (default: true)
- `LOG_LEVEL`: Log level (default: info)

## Default Admin Account

- Username: `admin`
- Password: `admin123`

Change these credentials after first login for security.