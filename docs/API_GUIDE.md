# WhatsApp Multi-Session API Guide - Session Creation & QR Code

This guide shows you how to create a new WhatsApp session and get the QR code via API.

## Quick Start

```bash
# 1. Login and get token
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  | jq -r '.data.token')

# 2. Create new session
SESSION_ID=$(curl -s -X POST http://localhost:8080/api/sessions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"phone":"6281234567890","name":"My Session"}' \
  | jq -r '.data.session.id')

# 3. Connect session
curl -X POST http://localhost:8080/api/sessions/$SESSION_ID/connect \
  -H "Authorization: Bearer $TOKEN"

# 4. Login to get QR
curl -X POST http://localhost:8080/api/sessions/$SESSION_ID/login \
  -H "Authorization: Bearer $TOKEN"

# 5. Get QR code and save to file
curl -X GET http://localhost:8080/api/sessions/$SESSION_ID/qr \
  -H "Authorization: Bearer $TOKEN" \
  | jq -r '.data.qr_code' \
  | sed 's/data:image\/png;base64,//' \
  | base64 -d > qr-code.png

echo "QR Code saved to qr-code.png - scan it now!"
open qr-code.png  # macOS
# xdg-open qr-code.png  # Linux
```

## API Endpoints

### 1. POST /api/sessions - Create Session

```json
{
  "phone": "6281234567890",
  "name": "My WhatsApp Business",
  "webhook_url": "https://your-server.com/webhook",
  "enabled": true
}
```

### 2. POST /api/sessions/{id}/connect - Connect Session

### 3. POST /api/sessions/{id}/login - Initiate Login

### 4. GET /api/sessions/{id}/qr - Get QR Code

Returns:
```json
{
  "success": true,
  "data": {
    "qr_code": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA..."
  }
}
```

## Display QR Code in HTML

```html
<!DOCTYPE html>
<html>
<head>
  <title>WhatsApp QR Code</title>
</head>
<body>
  <h1>Scan QR Code</h1>
  <img id="qr-code" alt="Loading..." />
  
  <script>
    const sessionId = '6281234567890';
    const token = 'YOUR_JWT_TOKEN';
    
    fetch(`http://localhost:8080/api/sessions/${sessionId}/qr`, {
      headers: { 'Authorization': `Bearer ${token}` }
    })
    .then(res => res.json())
    .then(data => {
      document.getElementById('qr-code').src = data.data.qr_code;
    });
  </script>
</body>
</html>
```

## Python Example

```python
import requests
import base64
from PIL import Image
from io import BytesIO

# Login
response = requests.post('http://localhost:8080/api/auth/login', json={
    'username': 'admin',
    'password': 'admin123'
})
token = response.json()['data']['token']
headers = {'Authorization': f'Bearer {token}'}

# Create session
response = requests.post('http://localhost:8080/api/sessions', 
                        headers=headers, 
                        json={'phone': '6281234567890', 'name': 'Python Bot'})
session_id = response.json()['data']['session']['id']

# Connect & Login
requests.post(f'http://localhost:8080/api/sessions/{session_id}/connect', headers=headers)
requests.post(f'http://localhost:8080/api/sessions/{session_id}/login', headers=headers)

# Get QR code
response = requests.get(f'http://localhost:8080/api/sessions/{session_id}/qr', headers=headers)
qr_data = response.json()['data']['qr_code']
qr_base64 = qr_data.split(',')[1]
qr_bytes = base64.b64decode(qr_base64)

# Display and save
img = Image.open(BytesIO(qr_bytes))
img.show()
img.save('qr-code.png')
print('QR code saved! Scan it now.')
```

## Node.js Example

```javascript
const axios = require('axios');
const fs = require('fs');

async function createSession() {
  // Login
  const login = await axios.post('http://localhost:8080/api/auth/login', {
    username: 'admin',
    password: 'admin123'
  });
  const token = login.data.data.token;
  const headers = { Authorization: `Bearer ${token}` };

  // Create session
  const session = await axios.post('http://localhost:8080/api/sessions', {
    phone: '6281234567890',
    name: 'Node.js Bot'
  }, { headers });
  const sessionId = session.data.data.session.id;

  // Connect & Login
  await axios.post(`http://localhost:8080/api/sessions/${sessionId}/connect`, {}, { headers });
  await axios.post(`http://localhost:8080/api/sessions/${sessionId}/login`, {}, { headers });

  // Get QR code
  const qr = await axios.get(`http://localhost:8080/api/sessions/${sessionId}/qr`, { headers });
  const qrCode = qr.data.data.qr_code;
  const base64Data = qrCode.replace(/^data:image\/png;base64,/, '');
  
  fs.writeFileSync('qr-code.png', base64Data, 'base64');
  console.log('QR code saved to qr-code.png');
}

createSession();
```

## Check Connection Status

```bash
curl http://localhost:8080/api/sessions \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.data.sessions[] | {id, connected, logged_in}'
```

## Send Message After Connected

```bash
# Wait for connection, then send
curl -X POST http://localhost:8080/api/sessions/$SESSION_ID/send \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "628987654321",
    "message": "Hello from API!"
  }'
```

## WebSocket Alternative (Real-time QR)

```javascript
const ws = new WebSocket('ws://localhost:8080/api/sessions/6281234567890/ws?token=YOUR_TOKEN');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  
  if (data.type === 'qr') {
    // Display QR code
    document.getElementById('qr').src = data.qr;
  }
  
  if (data.type === 'connected') {
    console.log('Session connected!');
  }
};
```

## Complete Flow Diagram

```
1. Login (POST /api/auth/login)
   ↓
2. Create Session (POST /api/sessions)
   ↓
3. Connect (POST /api/sessions/{id}/connect)
   ↓
4. Login (POST /api/sessions/{id}/login)
   ↓
5. Get QR (GET /api/sessions/{id}/qr)
   ↓
6. User scans QR code
   ↓
7. Session connected & authenticated
   ↓
8. Send messages (POST /api/sessions/{id}/send)
```

## Tips

- QR codes expire after 30 seconds
- Session ID = phone number provided during creation
- Store JWT token for subsequent requests
- Check connection status before sending messages
- Use webhooks for incoming messages
- Session persists across container restarts

## Troubleshooting

**QR code not showing:**
- Ensure you called `/connect` first
- Ensure you called `/login` to generate QR
- Check logs: `docker logs whatsapp-multi-session`

**Can't send message:**
- Check if session is connected: `GET /api/sessions`
- Verify `connected: true` and `logged_in: true`
- Re-authenticate if needed

**Session disconnected:**
- Check network connectivity
- Try reconnecting: `POST /api/sessions/{id}/connect`
- May need to rescan QR code
