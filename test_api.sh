#!/bin/bash

echo "🚀 Testing WhatsApp Multi-Session API"
echo "====================================="

# Start server in background
echo "Starting server..."
./whatsapp-multi-session &
SERVER_PID=$!

# Wait for server to start
echo "Waiting for server to start..."
sleep 3

echo ""
echo "1. Testing Login API (should return success: true)"
echo "------------------------------------------------"
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')

echo "$LOGIN_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$LOGIN_RESPONSE"

# Extract token from response
TOKEN=$(echo "$LOGIN_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin)['token'])" 2>/dev/null || echo "")

echo ""
echo "2. Testing Sessions API (with authentication)"
echo "--------------------------------------------"
if [ ! -z "$TOKEN" ]; then
    curl -s -X GET http://localhost:8080/api/sessions \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $TOKEN" \
      | python3 -m json.tool 2>/dev/null || echo "No sessions yet"
else
    echo "❌ No token received, cannot test authenticated endpoints"
fi

echo ""
echo "3. Testing Debug Endpoint"
echo "------------------------"
curl -s http://localhost:8080/debug | python3 -m json.tool 2>/dev/null

echo ""
echo "4. Testing Frontend (should return HTML)"
echo "---------------------------------------"
curl -s http://localhost:8080/ | head -5

# Kill server
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo ""
echo ""
echo "✅ API Tests Complete!"
echo ""
echo "To use the application:"
echo "1. Run: ./start_server.sh"
echo "2. Open: http://localhost:8080"
echo "3. Login with: admin / admin123"