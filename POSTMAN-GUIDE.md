# 📮 WhatsApp Multi-Session Manager - Postman Collection Guide

## 🚀 Quick Start

### 1. Import Collection
```bash
# Import the collection file into Postman
postman-collection.json
```

### 2. Setup Environment Variables
Create a new environment in Postman with these variables:

| Variable | Initial Value | Description |
|----------|---------------|-------------|
| `base_url` | `http://localhost:8080` | API server URL |
| `auth_token` | _(auto-set)_ | JWT authentication token |
| `session_id` | _(auto-set)_ | Active session ID |
| `user_id` | _(auto-set)_ | Current user ID |
| `expires_valid` | _(auto-set)_ | Valid expiration timestamp |
| `expires_expired` | _(auto-set)_ | Expired timestamp for testing |

### 3. Authentication Flow
1. **Run "Login (Get JWT Token)"** - Token will be automatically saved
2. **Create or manage sessions** using the session endpoints
3. **Start messaging** with the messaging endpoints

---

## 📋 Collection Structure

### 🔐 Authentication
- **Login (Get JWT Token)** - Get authentication token (auto-saves to environment)
- **Register New User** - Create new user account

### 📱 Session Management
- **Get All Sessions** - List all WhatsApp sessions
- **Create New Session** - Create new session (auto-saves session_id)
- **Get Session Details** - Get specific session info
- **Connect/Disconnect Session** - Manage session connections
- **Get QR Code** - Retrieve QR code for authentication
- **Update Session Webhook** - Configure webhook for incoming messages
- **Update Session Name** - Change session display name
- **Delete Session** - Remove session

### 💬 Messaging
- **Send Text Message** - Send plain text messages
- **Send Image (Base64)** - Send images using base64 encoding
- **Send File from URL** - Download and send files from URLs
- **Send Attachment (File Upload)** - Send files using base64
- **Check Number Validity** - Verify WhatsApp number exists
- **Send/Stop Typing Indicator** - Show typing status
- **Get Groups** - List session groups
- **Send Message (General)** - Compatible with original API

### 🖼️ Media Management (SECURED) 🔒
- **Access Temporary Media (Valid)** - Download received media files
- **Access Media - No Auth (Should Fail)** - Security test (401 expected)
- **Access Media - Expired Link (Should Fail)** - Expiration test (410 expected)  
- **Access Media - Invalid Token (Should Fail)** - Token validation test (401 expected)

### 👥 User Management (Admin)
- **Get All Users** - List all users (admin only)
- **Create User** - Add new user (admin only)
- **Get/Update/Delete User** - User CRUD operations (admin only)

### 🔧 Utility
- **Health Check** - API status check (no auth required)

---

## 🔒 Security Features

### Authentication Required
All endpoints except login/register/health require JWT authentication:
```http
Authorization: Bearer <jwt-token>
```

### Media Access Security
Media endpoints now require **both**:
1. **Valid JWT Token** - Authentication required
2. **Valid Expiration** - Timestamps must not be expired

**Example Secure Media URL:**
```http
GET /api/media/temp/filename.jpg?expires=1753708880
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Security Test Cases
The collection includes automated tests for:
- ❌ **No Authentication** → 401 Unauthorized
- ❌ **Invalid Token** → 401 Unauthorized  
- ❌ **Expired Media Links** → 410 Gone
- ✅ **Valid Authentication** → 200 Success

---

## 🎯 Common Use Cases

### 1. Setup New Session
```
1. Login (Get JWT Token)
2. Create New Session
3. Connect Session  
4. Get QR Code (scan with WhatsApp)
5. Start messaging
```

### 2. Send Media Messages
```
# Send Image
POST /api/sessions/{id}/send-image
{
    "to": "6281234567890",
    "image": "base64-encoded-data",
    "caption": "My image"
}

# Send File from URL  
POST /api/sessions/{id}/send-file-url
{
    "to": "6281234567890", 
    "url": "https://example.com/file.pdf",
    "filename": "document.pdf",
    "type": "document"
}
```

### 3. Access Received Media
```
# From webhook, you'll receive:
{
    "message_type": "image",
    "media_url": "/api/media/temp/filename.jpg?expires=1753708880"
}

# Access with authentication:
GET http://localhost:8080/api/media/temp/filename.jpg?expires=1753708880
Authorization: Bearer <your-jwt-token>
```

---

## 🚨 Important Notes

### Media Security Update 🔒
**Breaking Change:** Media access now requires authentication!

**Before:**
- ❌ Anyone could access media URLs
- ❌ Only expiration provided protection

**Now:**
- ✅ JWT authentication required
- ✅ Expiration timestamps validated
- ✅ Secure access control

### Webhook Integration
When receiving webhooks with media URLs, you must:
1. **Store your JWT token** from login
2. **Include Authorization header** when accessing media URLs
3. **Handle 401/410 errors** appropriately

### Environment Variables
The collection automatically manages these variables:
- `auth_token` - Set after successful login
- `session_id` - Set after session creation
- `expires_valid` - Generated for media access tests
- `expires_expired` - Generated for expiration tests

---

## 🔧 Testing Features

### Automated Test Scripts
Many requests include test scripts that:
- ✅ **Validate response codes**
- ✅ **Check response content**
- ✅ **Auto-save important values** (tokens, IDs)
- ✅ **Verify security measures**

### Pre-request Scripts
Some requests include pre-request scripts that:
- 🔄 **Generate timestamps** for media expiration
- 🔄 **Prepare test data**
- 🔄 **Validate prerequisites**

---

## 📞 Support

### Default Credentials
```
Username: admin
Password: admin123
```

### Server URL
```
Base URL: http://localhost:8080
Health Check: http://localhost:8080/api/health
```

### File Locations
```
📁 postman-collection.json (Main collection)
📁 postman-collection-backup.json (Original backup)
📁 POSTMAN-GUIDE.md (This guide)
```

---

## 🎉 Happy Testing!

This collection provides comprehensive testing for all WhatsApp Multi-Session Manager features including the new secure media handling system. All security measures are tested and validated.

For questions or issues, check the API logs at `server.log`.