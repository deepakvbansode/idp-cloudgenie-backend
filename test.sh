#!/bin/bash

# CloudGenie Backend Service - Test Script
# This script tests the backend service endpoints

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8081}"

echo "=========================================="
echo "CloudGenie Backend Service - Test Suite"
echo "=========================================="
echo ""
echo "Base URL: $BASE_URL"
echo ""

# Function to print test results
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ PASS${NC}: $2"
    else
        echo -e "${RED}✗ FAIL${NC}: $2"
        exit 1
    fi
}

# Test 1: Health Check
echo "Test 1: Health Check"
echo "-------------------"
response=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/v1/health")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$http_code" = "200" ]; then
    print_result 0 "Health check endpoint"
    echo "Response: $body"
else
    print_result 1 "Health check endpoint (HTTP $http_code)"
fi
echo ""

# Test 2: List Tools
echo "Test 2: List Available Tools"
echo "----------------------------"
response=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/v1/tools")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$http_code" = "200" ]; then
    print_result 0 "List tools endpoint"
    echo "Response: $body"
else
    print_result 1 "List tools endpoint (HTTP $http_code)"
fi
echo ""

# Test 3: Chat Request
echo "Test 3: Simple Chat Request"
echo "----------------------------"
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/chat" \
    -H "Content-Type: application/json" \
    -d '{
        "prompt": "What tools are available?"
    }')
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$http_code" = "200" ]; then
    print_result 0 "Chat endpoint (simple query)"
    echo "Response: $body"
else
    print_result 1 "Chat endpoint (HTTP $http_code)"
fi
echo ""

# Test 4: Chat with Tool Request
echo "Test 4: Chat Request with Tool Execution"
echo "----------------------------------------"
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/chat" \
    -H "Content-Type: application/json" \
    -d '{
        "prompt": "Check the health of CloudGenie backend"
    }')
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$http_code" = "200" ]; then
    print_result 0 "Chat endpoint (tool execution)"
    echo "Response: $body"
else
    print_result 1 "Chat endpoint with tools (HTTP $http_code)"
fi
echo ""

# Summary
echo "=========================================="
echo -e "${GREEN}All tests passed!${NC}"
echo "=========================================="
echo ""
echo "The CloudGenie Backend Service is working correctly."
echo ""
echo "You can now:"
echo "  1. Integrate it with your frontend application"
echo "  2. Send chat requests to $BASE_URL/api/v1/chat"
echo "  3. Monitor health at $BASE_URL/api/v1/health"
echo ""
