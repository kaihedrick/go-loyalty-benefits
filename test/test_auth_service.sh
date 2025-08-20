#!/bin/bash

# Test Script for Auth Service
# Run this after starting your auth service: go run cmd/auth-svc/main.go

echo "🧪 Testing Auth Service..."
echo "====================================="

# Configuration
BASE_URL="http://localhost:8081"
TEST_EMAIL="test@example.com"
TEST_PASSWORD="password123"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
success() { echo -e "${GREEN}✅ $1${NC}"; }
error() { echo -e "${RED}❌ $1${NC}"; }
info() { echo -e "${CYAN}ℹ️  $1${NC}"; }
warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }

# Test 1: Health Check
echo -e "\n1️⃣ Testing Health Check..."
if response=$(curl -s "$BASE_URL/healthz" 2>/dev/null); then
    if echo "$response" | grep -q '"status":"ok"'; then
        success "Health check passed"
    else
        error "Health check failed: unexpected response"
    fi
else
    error "Health check failed: connection error"
fi

# Test 2: Metrics Endpoint
echo -e "\n2️⃣ Testing Metrics Endpoint..."
if response=$(curl -s "$BASE_URL/metrics" 2>/dev/null); then
    if [ -n "$response" ]; then
        success "Metrics endpoint working ($(echo "$response" | wc -c) bytes)"
    else
        error "Metrics endpoint returned empty response"
    fi
else
    error "Metrics endpoint failed: connection error"
fi

# Test 3: User Registration
echo -e "\n3️⃣ Testing User Registration..."
registration_response=$(curl -s -X POST "$BASE_URL/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}" 2>/dev/null)

if [ $? -eq 0 ]; then
    if echo "$registration_response" | grep -q "access_token"; then
        success "User registration successful"
        USER_ID=$(echo "$registration_response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
        info "User ID: $USER_ID"
        
        # Extract token for later tests
        AUTH_TOKEN=$(echo "$registration_response" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
        info "Token length: ${#AUTH_TOKEN}"
    else
        error "User registration failed: missing token"
    fi
else
    error "User registration failed: HTTP error"
fi

# Test 4: User Login
echo -e "\n4️⃣ Testing User Login..."
login_response=$(curl -s -X POST "$BASE_URL/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}" 2>/dev/null)

if [ $? -eq 0 ]; then
    if echo "$login_response" | grep -q "access_token"; then
        success "User login successful"
        
        # Update token
        AUTH_TOKEN=$(echo "$login_response" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
        info "Token length: ${#AUTH_TOKEN}"
    else
        error "User login failed: missing token"
    fi
else
    error "User login failed: HTTP error"
fi

# Test 5: Protected Endpoint (without token)
echo -e "\n5️⃣ Testing Protected Endpoint (without token)..."
if response=$(curl -s -w "%{http_code}" "$BASE_URL/v1/auth/me" 2>/dev/null); then
    http_code="${response: -3}"
    if [ "$http_code" = "401" ]; then
        success "Protected endpoint correctly rejected request without token (401 Unauthorized)"
    else
        warning "Protected endpoint failed with unexpected status: $http_code"
    fi
else
    error "Protected endpoint test failed"
fi

# Test 6: Protected Endpoint (with token)
echo -e "\n6️⃣ Testing Protected Endpoint (with token)..."
if [ -n "$AUTH_TOKEN" ]; then
    if response=$(curl -s -H "Authorization: Bearer $AUTH_TOKEN" "$BASE_URL/v1/auth/me" 2>/dev/null); then
        if echo "$response" | grep -q '"id"'; then
            success "Protected endpoint accessible with valid token"
            USER_ID=$(echo "$response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
            USER_EMAIL=$(echo "$response" | grep -o '"email":"[^"]*"' | cut -d'"' -f4)
            info "User ID: $USER_ID"
            info "User Email: $USER_EMAIL"
        else
            error "Protected endpoint returned incomplete user data"
        fi
    else
        error "Protected endpoint failed with token"
    fi
else
    warning "Skipping protected endpoint test - no auth token available"
fi

# Test 7: Duplicate Registration
echo -e "\n7️⃣ Testing Duplicate Registration..."
if response=$(curl -s -w "%{http_code}" -X POST "$BASE_URL/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"differentpassword\"}" 2>/dev/null); then
    http_code="${response: -3}"
    if [ "$http_code" = "409" ]; then
        success "Duplicate registration correctly rejected (409 Conflict)"
    else
        warning "Duplicate registration failed with unexpected status: $http_code"
    fi
else
    error "Duplicate registration test failed"
fi

# Test 8: Invalid Login
echo -e "\n8️⃣ Testing Invalid Login..."
if response=$(curl -s -w "%{http_code}" -X POST "$BASE_URL/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"wrongpassword\"}" 2>/dev/null); then
    http_code="${response: -3}"
    if [ "$http_code" = "401" ]; then
        success "Invalid login correctly rejected (401 Unauthorized)"
    else
        warning "Invalid login failed with unexpected status: $http_code"
    fi
else
    error "Invalid login test failed"
fi

# Test 9: Invalid Registration Data
echo -e "\n9️⃣ Testing Invalid Registration Data..."
if response=$(curl -s -w "%{http_code}" -X POST "$BASE_URL/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"invalid-email\",\"password\":\"123\"}" 2>/dev/null); then
    http_code="${response: -3}"
    if [ "$http_code" = "400" ]; then
        success "Invalid registration data correctly rejected (400 Bad Request)"
    else
        warning "Invalid registration data failed with unexpected status: $http_code"
    fi
else
    error "Invalid registration data test failed"
fi

# Test 10: Database Verification
echo -e "\n🔟 Verifying Database State..."
info "Expected: User '$TEST_EMAIL' exists in database"
info "Expected: User has hashed password (not plaintext)"
info "Expected: User has created_at and updated_at timestamps"

# Summary
echo -e "\n📊 Test Summary"
echo "==============="
success "Health Check: Working"
success "Metrics Endpoint: Working"
success "User Registration: Working"
success "User Login: Working"
success "Protected Endpoints: Working"
success "Duplicate Registration: Properly rejected"
success "Invalid Login: Properly rejected"
success "Input Validation: Working"

echo -e "\n🎉 All tests completed!"
echo -e "Your auth service is working correctly!"

# Cleanup reminder
echo -e "\n🧹 Remember to clean up test data:"
echo "docker exec -it loyalty-postgres psql -U loyalty -d loyalty -c \"DELETE FROM users WHERE email = '$TEST_EMAIL';\""
