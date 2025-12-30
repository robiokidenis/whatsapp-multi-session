# Forward and Reply Message API Documentation

## Overview
This document describes the new Forward and Reply message functionality added to the WhatsApp Multi-Session API.

## Authentication
All endpoints require authentication using JWT Bearer token:
```
Authorization: Bearer <your-jwt-token>
```

## New Endpoints

### 1. Forward Message
Forward an existing WhatsApp message to another recipient.

**Endpoint:** `POST /api/sessions/{sessionId}/forward`

**Request Body:**
```json
{
  "to": "628123456789",          // Recipient phone number (without +)
  "message_id": "BAE5ABC123..."  // ID of the message to forward
}
```

**Response:**
```json
{
  "success": true,
  "id": "BAE5DEF456...",
  "message": "Message forwarded successfully"
}
```

**Error Responses:**
- `400 Bad Request` - Missing required fields
- `401 Unauthorized` - Invalid or missing authentication
- `403 Forbidden` - Session is disabled
- `404 Not Found` - Session not found
- `500 Internal Server Error` - Failed to forward message

### 2. Reply Message
Send a reply to a specific WhatsApp message.

**Endpoint:** `POST /api/sessions/{sessionId}/reply`

**Request Body:**
```json
{
  "to": "628123456789",                    // Recipient phone number
  "message": "This is my reply",           // Reply message text
  "quoted_message_id": "BAE5ABC123..."     // ID of the message being replied to
}
```

**Response:**
```json
{
  "success": true,
  "id": "BAE5GHI789...",
  "message": "Reply sent successfully"
}
```

**Error Responses:**
- `400 Bad Request` - Missing required fields
- `401 Unauthorized` - Invalid or missing authentication
- `403 Forbidden` - Session is disabled
- `404 Not Found` - Session not found
- `500 Internal Server Error` - Failed to send reply

## Implementation Details

### Forward Message
The forward functionality uses WhatsApp's context info to mark the message as forwarded. The forwarded message will appear with a "Forwarded" label in WhatsApp.

### Reply Message
The reply functionality creates a quoted message that references the original message. The reply will appear linked to the original message in WhatsApp's interface.

## Usage Examples

### Using cURL

#### Forward a Message
```bash
curl -X POST "http://localhost:8080/api/sessions/{sessionId}/forward" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "628123456789",
    "message_id": "BAE5ABC123DEF456"
  }'
```

#### Reply to a Message
```bash
curl -X POST "http://localhost:8080/api/sessions/{sessionId}/reply" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "628123456789",
    "message": "Thanks for your message!",
    "quoted_message_id": "BAE5ABC123DEF456"
  }'
```

### Using JavaScript/Node.js

```javascript
// Forward a message
const forwardMessage = async (sessionId, to, messageId) => {
  const response = await fetch(`${API_URL}/api/sessions/${sessionId}/forward`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      to: to,
      message_id: messageId
    })
  });
  return response.json();
};

// Reply to a message
const replyMessage = async (sessionId, to, message, quotedMessageId) => {
  const response = await fetch(`${API_URL}/api/sessions/${sessionId}/reply`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      to: to,
      message: message,
      quoted_message_id: quotedMessageId
    })
  });
  return response.json();
};
```

### Using PHP

```php
<?php
// Forward a message
function forwardMessage($sessionId, $to, $messageId, $token) {
    $url = "http://localhost:8080/api/sessions/{$sessionId}/forward";
    
    $data = [
        'to' => $to,
        'message_id' => $messageId
    ];
    
    $options = [
        'http' => [
            'header' => [
                "Authorization: Bearer {$token}",
                "Content-Type: application/json"
            ],
            'method' => 'POST',
            'content' => json_encode($data)
        ]
    ];
    
    $context = stream_context_create($options);
    $result = file_get_contents($url, false, $context);
    return json_decode($result, true);
}

// Reply to a message
function replyMessage($sessionId, $to, $message, $quotedMessageId, $token) {
    $url = "http://localhost:8080/api/sessions/{$sessionId}/reply";
    
    $data = [
        'to' => $to,
        'message' => $message,
        'quoted_message_id' => $quotedMessageId
    ];
    
    $options = [
        'http' => [
            'header' => [
                "Authorization: Bearer {$token}",
                "Content-Type: application/json"
            ],
            'method' => 'POST',
            'content' => json_encode($data)
        ]
    ];
    
    $context = stream_context_create($options);
    $result = file_get_contents($url, false, $context);
    return json_decode($result, true);
}
?>
```

## Notes

1. **Message IDs**: Message IDs are returned when sending any message through the API. Store these IDs if you need to forward or reply to messages later.

2. **Phone Number Format**: Phone numbers should be provided without the '+' sign. For example, use "628123456789" instead of "+628123456789".

3. **Session Requirements**: The session must be:
   - Connected to WhatsApp
   - Authenticated (logged in)
   - Enabled (not disabled)

4. **Rate Limiting**: Be mindful of WhatsApp's rate limits when forwarding or replying to multiple messages.

5. **Message History**: The original message being forwarded or replied to must still exist in the WhatsApp chat history.

## Testing

Use the provided test script to test the functionality:
```bash
./test-forward-reply.sh
```

Make sure to update the phone numbers in the script with valid WhatsApp numbers before testing.