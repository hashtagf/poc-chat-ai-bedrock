#!/bin/bash

# Simple API test script
# Start the server first with: go run cmd/server/main.go

BASE_URL="http://localhost:8080"

echo "Testing Chat Backend API"
echo "========================"
echo ""

# Test health check
echo "1. Testing health check..."
curl -s "$BASE_URL/health"
echo ""
echo ""

# Test create session
echo "2. Creating a new session..."
SESSION_RESPONSE=$(curl -s -X POST "$BASE_URL/api/sessions")
echo "$SESSION_RESPONSE" | jq .
SESSION_ID=$(echo "$SESSION_RESPONSE" | jq -r .id)
echo "Session ID: $SESSION_ID"
echo ""

# Test get session
echo "3. Getting session details..."
curl -s "$BASE_URL/api/sessions/$SESSION_ID" | jq .
echo ""

# Test list sessions
echo "4. Listing all sessions..."
curl -s "$BASE_URL/api/sessions" | jq .
echo ""

echo "API tests completed!"
echo ""
echo "To test WebSocket streaming, use a WebSocket client and connect to:"
echo "ws://localhost:8080/api/chat/stream"
echo ""
echo "Send a message like:"
echo '{"session_id": "'$SESSION_ID'", "content": "Hello, world!"}'
