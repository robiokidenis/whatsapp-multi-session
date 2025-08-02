# Typing Indicator Testing Guide

## Prerequisites
1. **Two WhatsApp accounts** - One for the bot session, one for testing
2. **Both accounts must support typing indicators** - Check by sending messages between them using regular WhatsApp
3. **Session must be logged in** - QR code scanned and connected
4. **Valid authentication token** - From successful login

## Testing Steps

### 1. Setup Session
```bash
# 1. Login to get token
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'

# Response will contain token, copy it
```

### 2. Create and Connect Session
```bash
# 2. Create session
curl -X POST http://localhost:3000/api/sessions \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alex Johnson",
    "webhook_url": ""
  }'

# 3. Connect session (will generate QR code)
curl -X POST http://localhost:3000/api/sessions/SESSION_ID/connect \
  -H "Authorization: Bearer YOUR_TOKEN"

# 4. Scan QR code with WhatsApp account
# Wait for connection and login to complete
```

### 3. Test Typing Indicators

#### Method 1: Direct Typing Test
```bash
# Send typing indicator
curl -X POST http://localhost:3000/api/sessions/SESSION_ID/typing \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "628123456789@s.whatsapp.net"
  }'

# Expected response:
# {"success": true, "message": "Typing indicator sent"}

# Stop typing indicator (after a few seconds)
curl -X POST http://localhost:3000/api/sessions/SESSION_ID/stop-typing \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "628123456789@s.whatsapp.net"
  }'
```

#### Method 2: Test with Presence Control
```bash
# 1. Set online presence first
curl -X POST http://localhost:3000/api/sessions/SESSION_ID/presence \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "available"
  }'

# 2. Then test typing
curl -X POST http://localhost:3000/api/sessions/SESSION_ID/typing \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "628123456789@s.whatsapp.net"
  }'
```

## Expected Behavior

### ✅ Success Indicators
1. **API Response**: Returns 200 status with success message
2. **Logs**: Check server logs for:
   - `"Set push name 'Test Bot' for session SESSION_ID"`
   - `"Set online presence for session SESSION_ID"`
   - `"Sent typing indicator to 628123456789@s.whatsapp.net from session SESSION_ID"`
3. **WhatsApp Client**: The recipient should see "typing..." indicator

### ❌ Troubleshooting

#### If typing indicator doesn't appear:

1. **Check Session Status**
```bash
curl -X GET http://localhost:3000/api/sessions/SESSION_ID \
  -H "Authorization: Bearer YOUR_TOKEN"

# Verify: "connected": true, "logged_in": true
```

2. **Check Phone Number Format**
- Use full international format: `628123456789@s.whatsapp.net`
- No spaces, dashes, or plus signs in the number part
- Must include `@s.whatsapp.net` for individual chats
- Use `@g.us` for group chats

3. **Test with Different Recipients**
- Try with multiple WhatsApp accounts
- Some accounts may have typing indicators disabled

4. **Check Server Logs**
Look for these log messages:
```
DEBUG: Set push name 'Test Bot' for session SESSION_ID
DEBUG: Set online presence for session SESSION_ID  
INFO: Sent typing indicator to 628123456789@s.whatsapp.net from session SESSION_ID
```

If you see warnings like:
```
WARN: Failed to set online presence for session SESSION_ID: [error]
```
This indicates a deeper connection issue.

5. **Manual Presence Setting**
```bash
# Try setting presence manually
curl -X POST http://localhost:3000/api/sessions/SESSION_ID/presence \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": "available"}'
```

## Known Limitations

1. **Delivery Reliability**: Typing indicators may not always be delivered due to WhatsApp's protocol
2. **Client Compatibility**: Some WhatsApp clients don't display typing indicators consistently
3. **Account Restrictions**: Some accounts may have typing indicators disabled
4. **Rate Limiting**: Don't send typing indicators too frequently
5. **Auto-Expiry**: Typing indicators automatically expire after ~5 seconds

## Testing with Postman

Use the updated Postman collection (v2.1.0):
1. Run "Login" request (token is set automatically)
2. Update `session_id` variable with your session ID
3. Use "Send Typing" request in Session Management folder
4. Use "Stop Typing" request to clear the indicator

## JID Format Examples

| Type | Format | Example |
|------|--------|---------|
| Individual | `number@s.whatsapp.net` | `628123456789@s.whatsapp.net` |
| Group | `groupid@g.us` | `628123456789-1598765432@g.us` |
| Business | `number@s.whatsapp.net` | `628123456789@s.whatsapp.net` |

## Debug Mode

Enable debug logging by setting log level to "DEBUG" in your configuration to see detailed typing indicator operations.