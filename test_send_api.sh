#!/bin/bash

# Test script for /api/send endpoint

# Replace with your actual auth token
TOKEN="your-auth-token-here"

# Test 1: Send message using session ID
echo "Test 1: Sending message using session ID..."
curl -X POST http://localhost:8080/api/send \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "4760020378",
    "to": "6281234567890",
    "message": "Test message using session ID"
  }'

echo -e "\n\n"

# Test 2: Send message using actual phone number
echo "Test 2: Sending message using actual phone number..."
curl -X POST http://localhost:8080/api/send \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "6285156018708",
    "to": "6281234567890",
    "message": "Test message using actual phone number"
  }'

echo -e "\n\n"

# Test 3: Send message using phone with @s.whatsapp.net
echo "Test 3: Sending message using phone with JID format..."
curl -X POST http://localhost:8080/api/send \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "6285156018708@s.whatsapp.net",
    "to": "6281234567890",
    "message": "Test message using phone JID format"
  }'

echo -e "\n"