#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "Testing WhatsApp Multi-Session API Endpoints"
echo "============================================"

# Base URL
BASE_URL="http://localhost:8080/api"

# Login and get token
echo -e "\n${GREEN}1. Testing Login Endpoint${NC}"
LOGIN_RESP=$(curl -s -X POST $BASE_URL/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}')

TOKEN=$(echo $LOGIN_RESP | python3 -c "import sys, json; print(json.load(sys.stdin)['token'])" 2>/dev/null)

if [ -z "$TOKEN" ]; then
    echo -e "${RED}Failed to get token${NC}"
    echo $LOGIN_RESP
    exit 1
fi

echo -e "${GREEN}✓ Login successful${NC}"
echo "Token: ${TOKEN:0:30}..."

# Test protected endpoints
echo -e "\n${GREEN}2. Testing Get Sessions${NC}"
curl -s -X GET $BASE_URL/sessions \
  -H "Authorization: Bearer $TOKEN" | json_pp

echo -e "\n${GREEN}3. Testing Create Session${NC}"
SESSION_RESP=$(curl -s -X POST $BASE_URL/sessions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "API Test Session",
    "webhook_url": "https://example.com/webhook"
  }')
echo $SESSION_RESP | json_pp

# Extract session ID
SESSION_ID=$(echo $SESSION_RESP | python3 -c "import sys, json; print(json.load(sys.stdin)['id'])" 2>/dev/null)

if [ ! -z "$SESSION_ID" ]; then
    echo -e "\n${GREEN}4. Testing Update Session Name${NC}"
    curl -s -X PUT $BASE_URL/sessions/$SESSION_ID/name \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d '{"name": "Updated Session Name"}' | json_pp

    echo -e "\n${GREEN}5. Testing Update Session Webhook${NC}"
    curl -s -X PUT $BASE_URL/sessions/$SESSION_ID/webhook \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d '{"webhook_url": "https://new-webhook.com/hook"}' | json_pp

    echo -e "\n${GREEN}6. Testing Connect Session${NC}"
    curl -s -X POST $BASE_URL/sessions/$SESSION_ID/connect \
      -H "Authorization: Bearer $TOKEN" | json_pp
fi

echo -e "\n${GREEN}7. Testing Admin Endpoints${NC}"
curl -s -X GET $BASE_URL/admin/users \
  -H "Authorization: Bearer $TOKEN" | json_pp

echo -e "\n${GREEN}Testing Complete!${NC}"