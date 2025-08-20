# Loyalty System cURL Test Script
# This script provides cURL commands for testing all endpoints

Write-Host "Loyalty System cURL Test Commands" -ForegroundColor Green
Write-Host "==================================" -ForegroundColor Green
Write-Host ""

# Variables for testing
$AUTH_BASE_URL = "http://localhost:8081"
$LOYALTY_BASE_URL = "http://localhost:8082"
$PARTNER_BASE_URL = "http://localhost:8083"
$BENEFITS_BASE_URL = "http://localhost:8084"
$REDEMPTION_BASE_URL = "http://localhost:8085"

# Test user credentials
$TEST_EMAIL = "test@example.com"
$TEST_PASSWORD = "testpassword123"
$TEST_USER_ID = ""

# JWT token storage
$JWT_TOKEN = ""

Write-Host "1. Health Checks" -ForegroundColor Yellow
Write-Host "=================" -ForegroundColor Yellow
Write-Host ""

Write-Host "Auth Service Health Check:" -ForegroundColor Cyan
Write-Host "curl -X GET '$AUTH_BASE_URL/healthz'" -ForegroundColor White
Write-Host ""

Write-Host "Loyalty Service Health Check:" -ForegroundColor Cyan
Write-Host "curl -X GET '$LOYALTY_BASE_URL/healthz'" -ForegroundColor White
Write-Host ""

Write-Host "2. User Registration" -ForegroundColor Yellow
Write-Host "=====================" -ForegroundColor Yellow
Write-Host ""

$registerBody = @{
    email = $TEST_EMAIL
    password = $TEST_PASSWORD
    first_name = "Test"
    last_name = "User"
} | ConvertTo-Json

Write-Host "User Registration:" -ForegroundColor Cyan
Write-Host "curl -X POST '$AUTH_BASE_URL/v1/auth/register' -H 'Content-Type: application/json' -d '$registerBody'" -ForegroundColor White
Write-Host ""

Write-Host "3. User Authentication" -ForegroundColor Yellow
Write-Host "=======================" -ForegroundColor Yellow
Write-Host ""

$loginBody = @{
    email = $TEST_EMAIL
    password = $TEST_PASSWORD
} | ConvertTo-Json

Write-Host "User Login:" -ForegroundColor Cyan
Write-Host "curl -X POST '$AUTH_BASE_URL/v1/auth/login' -H 'Content-Type: application/json' -d '$loginBody'" -ForegroundColor White
Write-Host ""

Write-Host "4. Earn Points (with auto-creation)" -ForegroundColor Yellow
Write-Host "=====================================" -ForegroundColor Yellow
Write-Host ""

$earnBody = @{
    user_id = "REPLACE_WITH_ACTUAL_USER_ID"
    amount = 100
    description = "Purchase completion"
} | ConvertTo-Json

Write-Host "Earn Points:" -ForegroundColor Cyan
Write-Host "curl -X POST '$LOYALTY_BASE_URL/v1/loyalty/earn' -H 'Content-Type: application/json' -H 'Authorization: Bearer REPLACE_WITH_JWT_TOKEN' -d '$earnBody'" -ForegroundColor White
Write-Host ""

Write-Host "5. Spend Points" -ForegroundColor Yellow
Write-Host "=================" -ForegroundColor Yellow
Write-Host ""

$spendBody = @{
    user_id = "REPLACE_WITH_ACTUAL_USER_ID"
    amount = 50
    description = "Reward redemption"
} | ConvertTo-Json

Write-Host "Spend Points:" -ForegroundColor Cyan
Write-Host "curl -X POST '$LOYALTY_BASE_URL/v1/loyalty/spend' -H 'Content-Type: application/json' -H 'Authorization: Bearer REPLACE_WITH_JWT_TOKEN' -d '$spendBody'" -ForegroundColor White
Write-Host ""

Write-Host "6. Get User Balance" -ForegroundColor Yellow
Write-Host "=====================" -ForegroundColor Yellow
Write-Host ""

Write-Host "Get User Balance:" -ForegroundColor Cyan
Write-Host "curl -X GET '$LOYALTY_BASE_URL/v1/loyalty/balance?user_id=REPLACE_WITH_ACTUAL_USER_ID' -H 'Authorization: Bearer REPLACE_WITH_JWT_TOKEN'" -ForegroundColor White
Write-Host ""

Write-Host "7. Get Transaction History" -ForegroundColor Yellow
Write-Host "=============================" -ForegroundColor Yellow
Write-Host ""

Write-Host "Get Transaction History:" -ForegroundColor Cyan
Write-Host "curl -X GET '$LOYALTY_BASE_URL/v1/loyalty/history?user_id=REPLACE_WITH_ACTUAL_USER_ID' -H 'Authorization: Bearer REPLACE_WITH_JWT_TOKEN'" -ForegroundColor White
Write-Host ""

Write-Host "8. List Available Rewards" -ForegroundColor Yellow
Write-Host "============================" -ForegroundColor Yellow
Write-Host ""

Write-Host "List Available Rewards:" -ForegroundColor Cyan
Write-Host "curl -X GET '$LOYALTY_BASE_URL/v1/loyalty/rewards'" -ForegroundColor White
Write-Host ""

Write-Host "9. List Available Benefits" -ForegroundColor Yellow
Write-Host "=============================" -ForegroundColor Yellow
Write-Host ""

Write-Host "List Available Benefits:" -ForegroundColor Cyan
Write-Host "curl -X GET '$BENEFITS_BASE_URL/v1/benefits'" -ForegroundColor White
Write-Host ""

Write-Host "10. Create Redemption Request" -ForegroundColor Yellow
Write-Host "===============================" -ForegroundColor Yellow
Write-Host ""

$redemptionBody = @{
    user_id = "REPLACE_WITH_ACTUAL_USER_ID"
    reward_id = "REPLACE_WITH_REWARD_ID"
    quantity = 1
} | ConvertTo-Json

Write-Host "Create Redemption Request:" -ForegroundColor Cyan
Write-Host "curl -X POST '$REDEMPTION_BASE_URL/v1/redeem' -H 'Content-Type: application/json' -H 'Authorization: Bearer REPLACE_WITH_JWT_TOKEN' -d '$redemptionBody'" -ForegroundColor White
Write-Host ""

Write-Host "11. List Partner Services" -ForegroundColor Yellow
Write-Host "===========================" -ForegroundColor Yellow
Write-Host ""

Write-Host "List Partner Services:" -ForegroundColor Cyan
Write-Host "curl -X GET '$PARTNER_BASE_URL/v1/partners'" -ForegroundColor White
Write-Host ""

Write-Host "================================================" -ForegroundColor Green
Write-Host "Usage Instructions:" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Green
Write-Host ""
Write-Host "1. Replace 'REPLACE_WITH_ACTUAL_USER_ID' with the user ID from registration" -ForegroundColor White
Write-Host "2. Replace 'REPLACE_WITH_JWT_TOKEN' with the JWT token from login" -ForegroundColor White
Write-Host "3. Replace 'REPLACE_WITH_REWARD_ID' with an actual reward ID from the rewards list" -ForegroundColor White
Write-Host ""
Write-Host "Test Flow:" -ForegroundColor Green
Write-Host "1. Start with health checks" -ForegroundColor White
Write-Host "2. Register a user" -ForegroundColor White
Write-Host "3. Login to get JWT token" -ForegroundColor White
Write-Host "4. Test loyalty endpoints with authentication" -ForegroundColor White
Write-Host "5. Test other service endpoints" -ForegroundColor White
Write-Host ""

Write-Host "Note: Make sure all services are running before testing!" -ForegroundColor Yellow
Write-Host "Run .\test\postman\setup_postman.ps1 to check service status" -ForegroundColor Yellow
