package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Simple test client for the loyalty and benefits system
func main() {
	fmt.Println("🚀 Starting Simple Test of Go Loyalty & Benefits Platform")
	fmt.Println("============================================================")

	// Test configuration
	baseURL := "http://localhost"
	userID := "550e8400-e29b-41d4-a716-446655440001" // From init.sql

	// Test 1: Check if services are running
	fmt.Println("\n1️⃣ Testing Service Health...")
	testServiceHealth(baseURL)

	// Test 2: Test Loyalty Service
	fmt.Println("\n2️⃣ Testing Loyalty Service...")
	testLoyaltyService(baseURL, userID)

	// Test 3: Test Catalog Service
	fmt.Println("\n3️⃣ Testing Catalog Service...")
	testCatalogService(baseURL)

	// Test 4: Test Redemption Service
	fmt.Println("\n4️⃣ Testing Redemption Service...")
	testRedemptionService(baseURL, userID)

	fmt.Println("\n✅ Simple test completed!")
	fmt.Println("\n📝 Next steps:")
	fmt.Println("   - Check the logs for any errors")
	fmt.Println("   - Verify data in the database")
	fmt.Println("   - Test the full workflow manually")
}

func testServiceHealth(baseURL string) {
	services := []struct {
		name string
		port string
	}{
		{"Auth Service", "8081"},
		{"Loyalty Service", "8082"},
		{"Catalog Service", "8083"},
		{"Redemption Service", "8084"},
		{"Partner Gateway", "8085"},
		{"Notification Service", "8086"},
	}

	for _, service := range services {
		url := fmt.Sprintf("%s:%s/healthz", baseURL, service.port)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("   ❌ %s: %v\n", service.name, err)
		} else {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				fmt.Printf("   ✅ %s: Healthy\n", service.name)
			} else {
				fmt.Printf("   ⚠️  %s: Status %d\n", service.name, resp.StatusCode)
			}
		}
	}
}

func testLoyaltyService(baseURL string, userID string) {
	// Test creating a transaction
	transactionData := map[string]interface{}{
		"amount":      100.00,
		"mcc":        "5812", // Restaurants
		"merchant_id": "REST-001",
	}

	jsonData, _ := json.Marshal(transactionData)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s:8082/v1/transactions", baseURL), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("   ❌ Create transaction failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		fmt.Printf("   ✅ Transaction created successfully\n")
		
		// Read response
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("   📄 Response: %s\n", string(body))
	} else {
		fmt.Printf("   ❌ Create transaction failed with status: %d\n", resp.StatusCode)
	}

	// Test getting balance
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s:8082/v1/balance", baseURL), nil)
	req.Header.Set("X-User-ID", userID)

	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("   ❌ Get balance failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("   ✅ Balance retrieved successfully\n")
		
		// Read response
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("   📄 Balance: %s\n", string(body))
	} else {
		fmt.Printf("   ❌ Get balance failed with status: %d\n", resp.StatusCode)
	}
}

func testCatalogService(baseURL string) {
	// Test getting benefits
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s:8083/v1/benefits", baseURL), nil)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("   ❌ Get benefits failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("   ✅ Benefits retrieved successfully\n")
		
		// Read response
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("   📄 Benefits: %s\n", string(body))
	} else {
		fmt.Printf("   ❌ Get benefits failed with status: %d\n", resp.StatusCode)
	}

	// Test getting categories
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s:8083/v1/categories", baseURL), nil)
	
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("   ❌ Get categories failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("   ✅ Categories retrieved successfully\n")
		
		// Read response
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("   📄 Categories: %s\n", string(body))
	} else {
		fmt.Printf("   ❌ Get categories failed with status: %d\n", resp.StatusCode)
	}
}

func testRedemptionService(baseURL string, userID string) {
	// Test creating a redemption
	redemptionData := map[string]interface{}{
		"benefit_id": "660e8400-e29b-41d4-a716-446655440000", // $25 Gift Card
		"points":     2000,
	}

	jsonData, _ := json.Marshal(redemptionData)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s:8084/v1/redeem", baseURL), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID)
	req.Header.Set("Idempotency-Key", "test-key-123")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("   ❌ Create redemption failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		fmt.Printf("   ✅ Redemption created successfully\n")
		
		// Read response
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("   📄 Response: %s\n", string(body))
	} else {
		fmt.Printf("   ❌ Create redemption failed with status: %d\n", resp.StatusCode)
	}
}
