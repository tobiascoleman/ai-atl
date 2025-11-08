#!/bin/bash

# Chatbot Testing Script with Database Integration
# This script demonstrates how the chatbot now uses MongoDB data

echo "=== Chatbot Database Integration Test ==="
echo ""

# Configuration
API_URL="http://localhost:8080/api/v1"
EMAIL="test@example.com"
PASSWORD="password123"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Login and get token
echo -e "${BLUE}Step 1: Logging in...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo -e "${YELLOW}Login failed. Creating new account...${NC}"
  
  # Try to register
  REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\",\"name\":\"Test User\"}")
  
  TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.token')
  
  if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo "Failed to get authentication token. Is the server running?"
    exit 1
  fi
fi

echo -e "${GREEN}✓ Logged in successfully${NC}"
echo ""

# Function to ask chatbot a question
ask_chatbot() {
  local question=$1
  echo -e "${BLUE}Question: ${NC}$question"
  echo ""
  
  RESPONSE=$(curl -s -X POST "$API_URL/chatbot/ask" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"question\":\"$question\"}")
  
  echo -e "${GREEN}Response:${NC}"
  echo $RESPONSE | jq -r '.response' | fold -w 80 -s
  echo ""
  echo "---"
  echo ""
}

# Test 1: Player-specific question
echo -e "${YELLOW}=== Test 1: Player Performance Query ===${NC}"
ask_chatbot "How is Patrick Mahomes performing this season?"

# Test 2: Player comparison
echo -e "${YELLOW}=== Test 2: Player Comparison ===${NC}"
ask_chatbot "Should I start Travis Kelce or George Kittle this week?"

# Test 3: Injury check
echo -e "${YELLOW}=== Test 3: Injury Status Check ===${NC}"
ask_chatbot "Is Christian McCaffrey injured?"

# Test 4: Team analysis
echo -e "${YELLOW}=== Test 4: Team Performance ===${NC}"
ask_chatbot "How is the Kansas City Chiefs offense doing this year?"

# Test 5: Position-based query
echo -e "${YELLOW}=== Test 5: Position Rankings ===${NC}"
ask_chatbot "Who are the top running backs in 2024?"

# Test 6: EPA-specific question
echo -e "${YELLOW}=== Test 6: EPA Analysis ===${NC}"
ask_chatbot "What is Lamar Jackson's EPA this season?"

# Test 7: General fantasy advice (no specific data needed)
echo -e "${YELLOW}=== Test 7: General Advice ===${NC}"
ask_chatbot "What should I look for in a waiver wire pickup?"

echo -e "${GREEN}=== All tests complete! ===${NC}"
echo ""
echo "Notice how the responses with specific players/teams include actual stats from the database!"
