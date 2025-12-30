#!/bin/bash

# Test script for error handling with different HTTP status codes

# Replace with your actual auth token
TOKEN="your-auth-token-here"

echo "=== Testing Error Handling for Different Session States ==="
echo

# Test 1: Send message to non-existent session (should return 404)
echo "Test 1: Send to non-existent session (expecting 404)..."
curl -i -X POST http://localhost:8080/api/send \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "9999999999",
    "to": "6281234567890",
    "message": "Test message"
  }'

echo -e "\n\n"

# Test 2: Connect non-existent session (should return 404)
echo "Test 2: Connect non-existent session (expecting 404)..."
curl -i -X POST http://localhost:8080/api/sessions/nonexistent/connect \
  -H "Authorization: Bearer $TOKEN"

echo -e "\n\n"

# Test 3: Send message to disconnected session (should return 503)
echo "Test 3: Send to disconnected session (expecting 503)..."
echo "First, disconnect a session if connected..."
curl -i -X POST http://localhost:8080/api/sessions/4760020378/disconnect \
  -H "Authorization: Bearer $TOKEN"

echo -e "\n"

echo "Now try to send message..."
curl -i -X POST http://localhost:8080/api/send \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "4760020378",
    "to": "6281234567890",
    "message": "Test message to disconnected session"
  }'

echo -e "\n\n"

# Test 4: Login already logged-in session (should return 400)
echo "Test 4: Login already logged-in session (expecting 400)..."
curl -i -X POST http://localhost:8080/api/sessions/7480961146/login \
  -H "Authorization: Bearer $TOKEN"

echo -e "\n\n"

# Test 5: Logout non-logged-in session (should return 400)
echo "Test 5: Logout non-logged-in session (expecting 400)..."
curl -i -X POST http://localhost:8080/api/sessions/4760020378/logout \
  -H "Authorization: Bearer $TOKEN"

echo -e "\n"