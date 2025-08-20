# Postman Setup Script for Loyalty Service
# This script checks the system status and provides setup instructions

Write-Host "Setting up Postman Tests for Loyalty Service" -ForegroundColor Green
Write-Host "=================================================" -ForegroundColor Green
Write-Host ""

# Check if services are running
Write-Host "Checking if services are running..." -ForegroundColor Yellow
$authRunning = $false
$loyaltyRunning = $false

try {
    $authResponse = Invoke-WebRequest -Uri "http://localhost:8081/healthz" -Method Get -TimeoutSec 5 -ErrorAction SilentlyContinue
    if ($authResponse.StatusCode -eq 200) {
        $authRunning = $true
        Write-Host "SUCCESS: Auth Service is running on port 8081" -ForegroundColor Green
    }
} catch {
    Write-Host "ERROR: Auth Service is NOT running on port 8081" -ForegroundColor Red
    Write-Host "   Start it with: go run cmd/auth-svc/main.go" -ForegroundColor Yellow
}

try {
    $loyaltyResponse = Invoke-WebRequest -Uri "http://localhost:8082/healthz" -Method Get -TimeoutSec 5 -ErrorAction SilentlyContinue
    if ($loyaltyResponse.StatusCode -eq 200) {
        $loyaltyRunning = $true
        Write-Host "SUCCESS: Loyalty Service is running on port 8082" -ForegroundColor Green
    }
} catch {
    Write-Host "ERROR: Loyalty Service is NOT running on port 8082" -ForegroundColor Red
    Write-Host "   Start it with: go run cmd/loyalty-svc/main.go" -ForegroundColor Yellow
}

# Check database connectivity
Write-Host ""
Write-Host "Checking database connectivity..." -ForegroundColor Yellow
try {
    $containerInfo = docker ps --filter "name=loyalty-postgres" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>$null
    if ($containerInfo -and $containerInfo -match "loyalty-postgres") {
        Write-Host "SUCCESS: PostgreSQL container is running" -ForegroundColor Green
        Write-Host "   Container details:" -ForegroundColor Gray
        $containerInfo | ForEach-Object { Write-Host "   $_" -ForegroundColor Gray }
    } else {
        Write-Host "ERROR: PostgreSQL container is not running" -ForegroundColor Red
        Write-Host "   Start it with: make infra-up" -ForegroundColor Yellow
    }
} catch {
    Write-Host "ERROR: Cannot check Docker containers" -ForegroundColor Red
    Write-Host "   Make sure Docker is running" -ForegroundColor Yellow
}

# Check Postman test files
Write-Host ""
Write-Host "Checking Postman test files..." -ForegroundColor Yellow

$collectionPath = "test/postman/Loyalty_Service_Tests.postman_collection.json"
$environmentPath = "test/postman/Loyalty_Service_Environment.postman_environment.json"
$readmePath = "test/postman/README.md"

if (Test-Path $collectionPath) {
    Write-Host "SUCCESS: Postman Collection found: $collectionPath" -ForegroundColor Green
} else {
    Write-Host "ERROR: Postman Collection not found: $collectionPath" -ForegroundColor Red
}

if (Test-Path $environmentPath) {
    Write-Host "SUCCESS: Postman Environment found: $environmentPath" -ForegroundColor Green
} else {
    Write-Host "ERROR: Postman Environment not found: $environmentPath" -ForegroundColor Red
}

if (Test-Path $readmePath) {
    Write-Host "SUCCESS: Postman README found: $readmePath" -ForegroundColor Green
} else {
    Write-Host "ERROR: Postman README not found: $readmePath" -ForegroundColor Red
}

# Test basic endpoints if services are running
Write-Host ""
Write-Host "Testing basic endpoints..." -ForegroundColor Yellow

if ($authRunning) {
    try {
        $authHealth = Invoke-WebRequest -Uri "http://localhost:8081/healthz" -Method Get -TimeoutSec 5
        if ($authHealth.StatusCode -eq 200) {
            Write-Host "SUCCESS: Auth Service health check: $($authHealth.StatusCode)" -ForegroundColor Green
        }
    } catch {
        Write-Host "ERROR: Auth Service health check failed" -ForegroundColor Red
    }
}

if ($loyaltyRunning) {
    try {
        $loyaltyHealth = Invoke-WebRequest -Uri "http://localhost:8082/healthz" -Method Get -TimeoutSec 5
        if ($loyaltyHealth.StatusCode -eq 200) {
            Write-Host "SUCCESS: Loyalty Service health check: $($loyaltyHealth.StatusCode)" -ForegroundColor Green
        }
    } catch {
        Write-Host "ERROR: Loyalty Service health check failed" -ForegroundColor Red
    }
    
    try {
        $rewardsResponse = Invoke-WebRequest -Uri "http://localhost:8082/v1/loyalty/rewards" -Method Get -TimeoutSec 5
        if ($rewardsResponse.StatusCode -eq 200) {
            $rewardsData = $rewardsResponse.Content | ConvertFrom-Json
            $rewardsCount = $rewardsData.data.Count
            Write-Host "SUCCESS: Loyalty Service rewards endpoint: $($rewardsResponse.StatusCode)" -ForegroundColor Green
            Write-Host "   Found $rewardsCount rewards" -ForegroundColor Gray
        }
    } catch {
        Write-Host "ERROR: Loyalty Service rewards endpoint failed" -ForegroundColor Red
    }
}

# Setup Summary
Write-Host ""
Write-Host "Setup Summary" -ForegroundColor Green
Write-Host "===============" -ForegroundColor Green
Write-Host ""

if ($authRunning -and $loyaltyRunning) {
    Write-Host "Next Steps:" -ForegroundColor Green
    Write-Host "1. Import the Postman Collection: $collectionPath" -ForegroundColor White
    Write-Host "2. Import the Postman Environment: $environmentPath" -ForegroundColor White
    Write-Host ""
    Write-Host "3. Select 'Loyalty Service Environment' in Postman" -ForegroundColor White
    Write-Host "4. Run the tests in sequence" -ForegroundColor White
    Write-Host ""
    Write-Host "Documentation:" -ForegroundColor Green
    Write-Host "   Read the detailed setup guide: $readmePath" -ForegroundColor White
    Write-Host ""
    Write-Host "Quick Test Commands:" -ForegroundColor Green
    Write-Host "   Test Auth Service: Invoke-WebRequest -Uri 'http://localhost:8081/healthz'" -ForegroundColor Gray
    Write-Host "   Test Loyalty Service: Invoke-WebRequest -Uri 'http://localhost:8082/healthz'" -ForegroundColor Gray
    Write-Host "   Test Rewards: Invoke-WebRequest -Uri 'http://localhost:8082/v1/loyalty/rewards'" -ForegroundColor Gray
} else {
    Write-Host "Next Steps:" -ForegroundColor Yellow
    Write-Host "1. Start the required services:" -ForegroundColor White
    Write-Host "   - Auth Service: go run cmd/auth-svc/main.go" -ForegroundColor Gray
    Write-Host "   - Loyalty Service: go run cmd/loyalty-svc/main.go" -ForegroundColor Gray
    Write-Host ""
    Write-Host "2. Then run this script again to verify setup" -ForegroundColor White
}

Write-Host ""
Write-Host "Setup complete! Happy testing!" -ForegroundColor Green
