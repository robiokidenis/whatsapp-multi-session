# WhatsApp Multi-Session API - Postman Guide

This guide shows how to use the WhatsApp Multi-Session API with Postman for managing multiple WhatsApp sessions and sending messages.

## Prerequisites

1. **Server Running**: Ensure the server is running on `http://localhost:8080`
2. **Postman Installed**: Download from https://www.postman.com/
3. **Basic Auth Setup**: Default credentials - `admin:admin`

## Quick Start Collection

### Import Collection
You can import this collection into Postman by creating a new collection and adding these requests:

## 🔐 Authentication

### 1. Login to Get Token

**POST** `http://localhost:8080/api/login`

```json
{
  "username": "admin",
  "password": "admin"
}
```

**Response:**
```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "admin"
  }
}
```

**💡 Important**: Copy the `token` value and use it as Bearer token for all subsequent requests.

---

## 📱 Session Management

### 2. Get All Sessions

**GET** `http://localhost:8080/api/sessions`

**Headers:**
```
Authorization: Bearer YOUR_TOKEN_HERE
```

**Response:**
```json
[
  {
    "id": "1234567890",
    "phone": "+1234567890",
    "actual_phone": "6282260790996@s.whatsapp.net",
    "name": "My WhatsApp Session",
    "position": 1,
    "webhook_url": "https://myapp.com/webhook",
    "connected": true,
    "logged_in": true
  }
]
```

### 3. Create New Session

**POST** `http://localhost:8080/api/sessions`

**Headers:**
```
Authorization: Bearer YOUR_TOKEN_HERE
Content-Type: application/json
```

**Body (Optional):**
```json
{
  "phone": "+1234567890",
  "name": "My New Session",
  "webhook_url": "https://myapp.com/webhook"
}
```

### 4. Update Session Name

**PUT** `http://localhost:8080/api/sessions/{sessionId}/name`

**Headers:**
```
Authorization: Bearer YOUR_TOKEN_HERE
Content-Type: application/json
```

**Body:**
```json
{
  "name": "Updated Session Name"
}
```

### 5. Update Session Webhook

**PUT** `http://localhost:8080/api/sessions/{sessionId}/webhook`

**Headers:**
```
Authorization: Bearer YOUR_TOKEN_HERE
Content-Type: application/json
```

**Body:**
```json
{
  "webhook_url": "https://myapp.com/new-webhook"
}
```

### 6. Connect Session (Generate QR)

**POST** `http://localhost:8080/api/sessions/{sessionId}/connect`

**Headers:**
```
Authorization: Bearer YOUR_TOKEN_HERE
```

**Response:**
```json
{
  "message": "Connection started",
  "websocket_url": "ws://localhost:8080/api/ws/1234567890?token=YOUR_TOKEN"
}
```

**🔗 Next Step**: Use the WebSocket URL to connect and receive QR code data.

### 7. Disconnect Session

**POST** `http://localhost:8080/api/sessions/{sessionId}/disconnect`

**Headers:**
```
Authorization: Bearer YOUR_TOKEN_HERE
```

### 8. Delete Session

**DELETE** `http://localhost:8080/api/sessions/{sessionId}`

**Headers:**
```
Authorization: Bearer YOUR_TOKEN_HERE
```

---

## 💬 Sending Messages

### 9. Send Message from Specific Session

**POST** `http://localhost:8080/api/sessions/{sessionId}/send`

**Headers:**
```
Authorization: Bearer YOUR_TOKEN_HERE
Content-Type: application/json
```

**Body:**
```json
{
  "recipient": "6281381393739",
  "message": "Hello from WhatsApp API!"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Message sent successfully",
  "message_id": "3EB0BEDDE45EF384377BD7",
  "session_id": "1234567890"
}
```

### 10. Send Message via API (Token + Phone Selection)

**POST** `http://localhost:8080/api/send`

**Headers:**
```
Authorization: Bearer YOUR_TOKEN_HERE
Content-Type: application/json
```

**Body:**
```json
{
  "phone": "6282260790996",
  "recipient": "6281381393739",
  "message": "Hello from selected phone number!"
}
```

---

## 🌐 WebSocket for QR Code Authentication

### Connecting with WebSocket

You cannot test WebSocket directly in Postman, but here's how it works:

1. **Get WebSocket URL** from the connect session response
2. **Connect to WebSocket**: `ws://localhost:8080/api/ws/{sessionId}?token={authToken}`
3. **Receive Messages**:

```javascript
// JavaScript WebSocket example
const ws = new WebSocket('ws://localhost:8080/api/ws/1234567890?token=YOUR_TOKEN');

ws.onmessage = function(event) {
  const data = JSON.parse(event.data);
  
  if (data.type === 'qr') {
    console.log('QR Code Data:', data.data.qr);
    // Display QR code to user
  } else if (data.type === 'success') {
    console.log('Login successful!');
  } else if (data.type === 'error') {
    console.log('Error:', data.data.error);
  }
};
```

**QR Code Message:**
```json
{
  "type": "qr",
  "data": {
    "qr": "2@BAxEFSgFCONIGGPQVR8DBxC/XSU5W8nwGDFCGfj+VJ2G6I3gpTkeZSWXuuMBu2WN2oB4Mgnob01jm4hofuE++9RnmxCH/k53ogo=",
    "timeout": 60000000000
  }
}
```

---

## 🎯 Complete Workflow Examples

### Example 1: Single User, Multiple Phone Numbers

```bash
# 1. Login
POST /api/login
{
  "username": "admin",
  "password": "admin"
}

# 2. Create session for Phone 1
POST /api/sessions
{
  "name": "Business Phone",
  "webhook_url": "https://myapp.com/webhook/business"
}

# 3. Create session for Phone 2  
POST /api/sessions
{
  "name": "Personal Phone",
  "webhook_url": "https://myapp.com/webhook/personal"
}

# 4. Connect and authenticate both sessions via QR codes
POST /api/sessions/{session1}/connect
POST /api/sessions/{session2}/connect

# 5. Send messages from different phones
POST /api/send
{
  "phone": "6282260790996",
  "recipient": "6281381393739",
  "message": "Business message"
}

POST /api/send
{
  "phone": "6285591500390",
  "recipient": "6281381393739", 
  "message": "Personal message"
}
```

### Example 2: Multiple Users with Different Tokens

```bash
# User 1 Login
POST /api/login
{
  "username": "admin",
  "password": "admin"
}
# Use token1 for User 1's requests

# User 2 Login (if multi-user is implemented)
POST /api/login
{
  "username": "user2",
  "password": "password2"
}
# Use token2 for User 2's requests

# Each user manages their own sessions
GET /api/sessions (with respective tokens)
```

---

## 🎣 Webhook Integration

### Setting Up Webhook Endpoint

Your webhook endpoint should handle POST requests:

```javascript
// Express.js example
app.post('/webhook', (req, res) => {
  const { session_id, message } = req.body;
  
  console.log(`Message from session ${session_id}:`);
  console.log(`From: ${message.from}`);
  console.log(`Message: ${message.message}`);
  
  res.status(200).send('OK');
});
```

**Webhook Payload:**
```json
{
  "session_id": "1234567890",
  "message": {
    "id": "3EB0BEDDE45EF384377BD7",
    "from": "6281381393739@s.whatsapp.net",
    "to": "6282260790996@s.whatsapp.net", 
    "message": "Hello!",
    "timestamp": 1643723400
  }
}
```

---

## 🚨 Common Issues & Solutions

### 1. Authentication Errors
```json
{
  "error": "Unauthorized",
  "message": "Invalid or missing authentication token"
}
```
**Solution**: Ensure Bearer token is included in Authorization header.

### 2. Session Not Found
```json
{
  "error": "Session not found"
}
```
**Solution**: Verify session ID exists by calling `GET /api/sessions`.

### 3. Message Send Failure
```json
{
  "error": "Session not logged in"
}
```
**Solution**: Ensure session is connected and authenticated via QR code.

### 4. WebSocket Connection Issues
**Solution**: Check if session exists and token is valid in the WebSocket URL.

---

## 📋 Postman Environment Variables

Create a Postman environment with these variables:

- `base_url`: `http://localhost:8080/api`
- `auth_token`: `{{token_from_login_response}}`
- `session_id`: `{{session_id_from_create_response}}`

This allows you to use `{{base_url}}/sessions` instead of full URLs and `{{auth_token}}` in Authorization headers.

---

## 🔄 Testing Multiple Sessions

1. **Create 3 sessions** for different phone numbers
2. **Connect each session** and scan QR codes with different phones
3. **Send messages** from each session to test isolation
4. **Update session names** for easy identification
5. **Set different webhooks** for each session
6. **Test API sending** using phone number selection

This API supports unlimited sessions, perfect for managing multiple WhatsApp business accounts or different user groups.