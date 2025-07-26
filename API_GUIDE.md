# WhatsApp Multi-Session API Guide

Complete guide for using the WhatsApp Multi-Session API with multiple users, phone numbers, and WebSocket integration.

## 🚀 Quick Start

### 1. Server Setup
```bash
# Start the server
./main

# Server runs on: http://localhost:8080
# API base URL: http://localhost:8080/api
# Frontend: http://localhost:8080
```

### 2. Import API Documentation
- **OpenAPI/Swagger**: Use `api-docs.yaml` with Swagger UI or any OpenAPI tool
- **Postman Collection**: Import `postman-collection.json` 
- **Manual Setup**: Follow examples in `POSTMAN.md`

## 📋 API Overview

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/login` | POST | User authentication |
| `/api/sessions` | GET/POST | List/create sessions |
| `/api/sessions/{id}` | GET/DELETE | Get/delete specific session |
| `/api/sessions/{id}/name` | PUT | Update session name |
| `/api/sessions/{id}/webhook` | PUT | Update webhook URL |
| `/api/sessions/{id}/connect` | POST | Start QR authentication |
| `/api/sessions/{id}/disconnect` | POST | Disconnect session |
| `/api/sessions/{id}/send` | POST | Send message from session |
| `/api/sessions/{id}/check-number` | POST | Check if number is on WhatsApp |
| `/api/sessions/{id}/groups` | GET | List all groups for session |
| `/api/send` | POST | Send message via token + phone |
| `/api/ws/{id}?token={token}` | WebSocket | QR code authentication |

## 🔐 Authentication Flow

### Step 1: Login
```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin"
  }'
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

### Step 2: Use Token
All subsequent requests require the Bearer token:
```bash
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## 📱 Multi-Session Scenarios

### Scenario 1: Single User, Multiple Phone Numbers

Perfect for businesses managing multiple WhatsApp accounts:

```bash
# 1. Create Business Session
curl -X POST http://localhost:8080/api/sessions \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Business Support",
    "webhook_url": "https://myapp.com/webhook/business"
  }'

# 2. Create Personal Session  
curl -X POST http://localhost:8080/api/sessions \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Personal Account",
    "webhook_url": "https://myapp.com/webhook/personal"
  }'

# 3. Connect both sessions (scan different QR codes)
curl -X POST http://localhost:8080/api/sessions/SESSION_ID_1/connect \
  -H "Authorization: Bearer YOUR_TOKEN"

curl -X POST http://localhost:8080/api/sessions/SESSION_ID_2/connect \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Scenario 2: Multiple Users with Separate Sessions

Each user has their own token and manages their own sessions:

```bash
# User 1 Login
curl -X POST http://localhost:8080/api/login \
  -d '{"username": "admin", "password": "admin"}'
# Get token1

# User 2 Login (if multiple users are set up)
curl -X POST http://localhost:8080/api/login \
  -d '{"username": "user2", "password": "password2"}'
# Get token2

# Each user creates their own sessions
curl -X POST http://localhost:8080/api/sessions \
  -H "Authorization: Bearer TOKEN1" \
  -d '{"name": "User 1 Session"}'

curl -X POST http://localhost:8080/api/sessions \
  -H "Authorization: Bearer TOKEN2" \
  -d '{"name": "User 2 Session"}'
```

## 🌐 WebSocket Integration

### QR Code Authentication Process

1. **Create/Connect Session**
```bash
curl -X POST http://localhost:8080/api/sessions/1234567890/connect \
  -H "Authorization: Bearer YOUR_TOKEN"
```

2. **Connect to WebSocket**
```javascript
const ws = new WebSocket('ws://localhost:8080/api/ws/1234567890?token=YOUR_TOKEN');

ws.onopen = function() {
  console.log('WebSocket connected');
};

ws.onmessage = function(event) {
  const data = JSON.parse(event.data);
  
  switch(data.type) {
    case 'qr':
      console.log('QR Code:', data.data.qr);
      // Display QR code for user to scan
      displayQRCode(data.data.qr);
      break;
      
    case 'success':
      console.log('Authentication successful!');
      // Session is now connected
      break;
      
    case 'error':
      console.error('Error:', data.data.error);
      break;
  }
};
```

3. **QR Code Display Options**

**Option A: Generate QR Image**
```javascript
function displayQRCode(qrData) {
  const qrUrl = `https://api.qrserver.com/v1/create-qr-code/?size=256x256&data=${encodeURIComponent(qrData)}`;
  document.getElementById('qr-image').src = qrUrl;
}
```

**Option B: Use QR Code Library**
```javascript
import QRCode from 'qrcode';

function displayQRCode(qrData) {
  const canvas = document.getElementById('qr-canvas');
  QRCode.toCanvas(canvas, qrData, { width: 256 });
}
```

## 💬 Message Sending

### Method 1: Session-Specific Sending
```bash
curl -X POST http://localhost:8080/api/sessions/1234567890/send \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "recipient": "6281381393739",
    "message": "Hello from specific session!"
  }'
```

### Method 2: Phone Number Selection
```bash
curl -X POST http://localhost:8080/api/send \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "6282260790996",
    "recipient": "6281381393739",
    "message": "Hello from selected phone!"
  }'
```

## 🔍 Number & Group Management

### Check if Number is on WhatsApp
Verify if a phone number is registered on WhatsApp before sending messages:

```bash
curl -X POST http://localhost:8080/api/sessions/1234567890/check-number \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "number": "6281381393739"
  }'
```

**Response:**
```json
{
  "success": true,
  "number": "6281381393739",
  "on_whatsapp": true,
  "verified": false,
  "verified_name": "Business Name"
}
```

**Use Cases:**
- **Validate recipients** before sending bulk messages
- **Check business accounts** with verified badges
- **Filter contact lists** to only WhatsApp users
- **Verify phone numbers** in registration flows

### List All Groups
Get all WhatsApp groups that the session is part of:

```bash
curl -X GET http://localhost:8080/api/sessions/1234567890/groups \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response:**
```json
{
  "success": true,
  "count": 2,
  "groups": [
    {
      "jid": "120363302171476233@g.us",
      "name": "Family Group",
      "topic": "Family chat group",
      "description": "Family chat group",
      "owner": "6281381393739@s.whatsapp.net",
      "participant_count": 8,
      "is_admin": false,
      "is_super_admin": false
    },
    {
      "jid": "6281381393739-1624931511@g.us",
      "name": "Work Team",
      "topic": "",
      "description": "",
      "owner": "6282260790996@s.whatsapp.net",
      "participant_count": 15,
      "is_admin": true,
      "is_super_admin": false
    }
  ]
}
```

**Use Cases:**
- **Group management** for bot applications
- **Broadcast targeting** specific groups
- **Admin verification** before group operations
- **Group analytics** and reporting

### Integration Examples

**1. Bulk Message with Number Validation:**
```javascript
async function sendBulkMessages(sessionId, recipients, message) {
  const validNumbers = [];
  
  // Check all numbers first
  for (const number of recipients) {
    const response = await fetch(`/api/sessions/${sessionId}/check-number`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ number })
    });
    
    const result = await response.json();
    if (result.on_whatsapp) {
      validNumbers.push(number);
    }
  }
  
  // Send messages to valid numbers only
  for (const number of validNumbers) {
    await sendMessage(sessionId, number, message);
  }
}
```

**2. Group-Based Broadcasting:**
```javascript
async function broadcastToGroups(sessionId, message, adminOnly = false) {
  // Get all groups
  const response = await fetch(`/api/sessions/${sessionId}/groups`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  
  const { groups } = await response.json();
  
  // Filter groups if admin-only broadcasting
  const targetGroups = adminOnly 
    ? groups.filter(g => g.is_admin || g.is_super_admin)
    : groups;
  
  // Send messages to groups
  for (const group of targetGroups) {
    await sendMessage(sessionId, group.jid, message);
  }
}
```

## 🎣 Webhook Integration

### Setting Up Webhooks

1. **Configure Webhook URL**
```bash
curl -X PUT http://localhost:8080/api/sessions/1234567890/webhook \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "webhook_url": "https://myapp.com/webhook"
  }'
```

2. **Handle Incoming Messages**
```javascript
// Express.js webhook handler
app.post('/webhook', express.json(), (req, res) => {
  const { session_id, message } = req.body;
  
  console.log(`New message in session ${session_id}:`);
  console.log(`From: ${message.from}`);
  console.log(`Message: ${message.message}`);
  console.log(`Timestamp: ${message.timestamp}`);
  
  // Process the message
  processIncomingMessage(session_id, message);
  
  res.status(200).send('OK');
});
```

3. **Webhook Payload Structure**
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

## 🏗️ Advanced Use Cases

### Use Case 1: WhatsApp Business API Alternative
```javascript
// Multi-department routing
const departments = {
  'sales': 'session_sales',
  'support': 'session_support',
  'billing': 'session_billing'
};

function routeMessage(department, recipient, message) {
  const sessionId = departments[department];
  
  return fetch(`/api/sessions/${sessionId}/send`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ recipient, message })
  });
}
```

### Use Case 2: Multi-Tenant SaaS Platform
```javascript
// Each tenant has their own sessions
class WhatsAppTenant {
  constructor(tenantId, authToken) {
    this.tenantId = tenantId;
    this.authToken = authToken;
    this.sessions = new Map();
  }
  
  async createSession(name, webhookUrl) {
    const response = await fetch('/api/sessions', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.authToken}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ name, webhook_url: webhookUrl })
    });
    
    const session = await response.json();
    this.sessions.set(session.id, session);
    return session;
  }
  
  async sendMessage(sessionId, recipient, message) {
    return fetch(`/api/sessions/${sessionId}/send`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.authToken}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ recipient, message })
    });
  }
}
```

### Use Case 3: Load Balancing Messages
```javascript
// Distribute messages across multiple sessions
class MessageBalancer {
  constructor(sessions, authToken) {
    this.sessions = sessions;
    this.authToken = authToken;
    this.currentIndex = 0;
  }
  
  async sendBalanced(recipient, message) {
    const session = this.sessions[this.currentIndex];
    this.currentIndex = (this.currentIndex + 1) % this.sessions.length;
    
    return fetch('/api/send', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.authToken}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        phone: session.phone,
        recipient,
        message
      })
    });
  }
}
```

## 🔧 Error Handling

### Common HTTP Status Codes
- `200`: Success
- `201`: Created (new session)
- `400`: Bad Request (invalid parameters)
- `401`: Unauthorized (invalid/missing token)
- `404`: Not Found (session doesn't exist)
- `500`: Internal Server Error

### Error Response Format
```json
{
  "error": "Error Type",
  "message": "Detailed error description"
}
```

### Retry Logic Example
```javascript
async function sendMessageWithRetry(sessionId, recipient, message, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      const response = await fetch(`/api/sessions/${sessionId}/send`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ recipient, message })
      });
      
      if (response.ok) {
        return await response.json();
      }
      
      if (response.status === 401) {
        // Token expired, need to re-authenticate
        await refreshToken();
        continue;
      }
      
      if (response.status === 404) {
        throw new Error('Session not found or not logged in');
      }
      
    } catch (error) {
      if (i === maxRetries - 1) throw error;
      await new Promise(resolve => setTimeout(resolve, 1000 * (i + 1)));
    }
  }
}
```

## 📊 Monitoring & Analytics

### Session Status Monitoring
```javascript
async function monitorSessions() {
  const response = await fetch('/api/sessions', {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  
  const sessions = await response.json();
  
  const stats = {
    total: sessions.length,
    connected: sessions.filter(s => s.connected).length,
    logged_in: sessions.filter(s => s.logged_in).length,
    disconnected: sessions.filter(s => !s.connected).length
  };
  
  console.log('Session Stats:', stats);
  return stats;
}

// Run every minute
setInterval(monitorSessions, 60000);
```

## 🚀 Production Deployment

### Environment Variables
```bash
export PORT=8080
export JWT_SECRET=your-secret-key
export DB_TYPE=mysql
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASS=password
export DB_NAME=whatsapp_sessions
```

### Docker Deployment
```dockerfile
FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/frontend/dist ./frontend/dist
EXPOSE 8080
CMD ["./main"]
```

This comprehensive guide covers all aspects of using the WhatsApp Multi-Session API for various scenarios from simple single-user setups to complex multi-tenant platforms.