# Postman Collection Updates

## Changes Made (v2.1.0)

### 🆕 New Endpoint Added
- **GET** `/api/sessions/{sessionId}/conversations` - Get all conversations/chats for a session
  - Retrieves both individual chats and group chats
  - Returns conversation metadata including JID, name, group status, and chat states
  - Follows the same authentication pattern as other session endpoints

### 🔧 Authentication Improvements
- Added `auth_token` variable for easier token management
- Collection now uses `{{auth_token}}` as the primary bearer token variable
- Automatic token management: Login request now automatically sets the auth token
- Added test script to Login endpoint that:
  - Extracts JWT token from successful login response
  - Sets both `auth_token` and `jwt_token` variables automatically
  - Provides console feedback on success/failure

### 📝 Usage Instructions
1. **First Time Setup**: 
   - Import the collection into Postman
   - Update the `base_url` variable if your server runs on a different port
   
2. **Authentication**:
   - Run the "Login" request in the Authentication folder
   - The auth token will be automatically set for all subsequent requests
   - No need to manually copy/paste tokens anymore!

3. **Testing the New Conversations Endpoint**:
   - First, ensure you have a session created and logged in to WhatsApp
   - Use the "Get Conversations" request in the Session Management folder
   - Replace `{{session_id}}` with your actual session ID

### 🔄 Migration from Previous Version
- Existing collections will continue to work
- The `jwt_token` variable is still supported for backward compatibility
- New automatic authentication makes the workflow much smoother

### 📊 Response Format for Conversations
```json
{
  "success": true,
  "message": "Conversations retrieved successfully",
  "data": {
    "conversations": [
      {
        "jid": "628123456789@s.whatsapp.net",
        "name": "Contact Name",
        "is_group": false,
        "unread_count": 0,
        "is_pinned": false,
        "is_muted": false,
        "is_archived": false
      }
    ],
    "count": 1
  }
}
```

### 🔍 Notes
- The conversations endpoint returns basic conversation metadata
- For detailed message history, use the existing message-related endpoints
- Group information is also included in the conversations list
- Chat states (pinned, muted, archived) are currently set to default values as WhatsApp doesn't expose this data directly through the basic APIs