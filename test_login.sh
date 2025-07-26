#!/bin/bash

echo "Testing Login API..."
echo "===================="

# Start server in background
./whatsapp-multi-session &
SERVER_PID=$!

# Wait for server to start
sleep 3

echo "Testing login with admin credentials..."
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  | python3 -m json.tool 2>/dev/null || echo "Raw response"

echo ""
echo "Testing invalid login..."
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"invalid","password":"wrong"}' \
  | python3 -m json.tool 2>/dev/null || echo "Raw response"

# Kill server
kill $SERVER_PID 2>/dev/null

echo ""
echo "Test complete!"