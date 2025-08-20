# 🚀 **Postman Tests for Loyalty Service**

This directory contains comprehensive Postman tests for your Go Loyalty & Benefits Platform Loyalty Service.

## 📁 **Files**

- **`Loyalty_Service_Tests.postman_collection.json`** - Complete test collection
- **`Loyalty_Service_Environment.postman_environment.json`** - Environment variables
- **`README.md`** - This documentation

## 🎯 **What's Included**

### **Test Categories**

1. **🏥 Health & Monitoring**
   - Health Check endpoint
   - Prometheus metrics endpoint

2. **🌐 Public Endpoints**
   - Get available rewards (no auth required)

3. **🔐 Authentication Tests**
   - Verify protected endpoints reject unauthorized requests
   - Test JWT token validation

4. **👤 User Management**
   - Create test user in Auth Service
   - Login to get JWT token

5. **💎 Protected Endpoints (With Auth)**
   - Get user balance
   - Earn points
   - Check balance updates
   - Test tier progression (Bronze → Silver)
   - Get transaction history
   - Spend points

6. **⚠️ Error Handling**
   - Insufficient points
   - Invalid request data

## 🚀 **Setup Instructions**

### **Step 1: Import Collection**
1. Open **Postman**
2. Click **Import** button
3. Select **`Loyalty_Service_Tests.postman_collection.json`**
4. Click **Import**

### **Step 2: Import Environment**
1. Click **Import** button again
2. Select **`Loyalty_Service_Environment.postman_environment.json`**
3. Click **Import**

### **Step 3: Select Environment**
1. In the top-right corner, select **"Loyalty Service Environment"**
2. Verify the environment is active

### **Step 4: Start Your Services**
Make sure both services are running:

```bash
# Terminal 1: Start Auth Service
cd cmd/auth-svc
go run main.go

# Terminal 2: Start Loyalty Service  
cd cmd/loyalty-svc
go run main.go
```

## 🧪 **Running the Tests**

### **Option 1: Run Individual Tests**
1. Open the collection in Postman
2. Click on any request
3. Click **Send** to execute
4. View test results in the **Test Results** tab

### **Option 2: Run Collection**
1. Right-click on the collection
2. Select **Run collection**
3. Choose which requests to run
4. Click **Run Loyalty Service Tests**

### **Option 3: Run with Newman (CLI)**
```bash
# Install Newman globally
npm install -g newman

# Run the collection
newman run test/postman/Loyalty_Service_Tests.postman_collection.json \
  -e test/postman/Loyalty_Service_Environment.postman_environment.json \
  --reporters cli,json \
  --reporter-json-export test-results.json
```

## 📊 **Test Flow**

The tests are designed to run in sequence:

1. **Health Check** → Verify service is running
2. **Get Rewards** → Test public endpoint
3. **Auth Tests** → Verify protected endpoints reject unauthorized requests
4. **Create User** → Create test user in Auth Service
5. **Login** → Get JWT token
6. **Get Balance** → Check initial state (0 points, Bronze tier)
7. **Earn Points** → Add 100 points
8. **Check Balance** → Verify points increased
9. **Earn More Points** → Add 1000 points to reach Silver tier
10. **Get History** → View transaction log
11. **Spend Points** → Redeem 100 points
12. **Error Tests** → Test edge cases

## 🔧 **Environment Variables**

| Variable | Description | Default Value |
|----------|-------------|---------------|
| `base_url` | Loyalty Service URL | `http://localhost:8082` |
| `auth_base_url` | Auth Service URL | `http://localhost:8081` |
| `test_user_email` | Test user email | Auto-generated |
| `test_user_password` | Test user password | Auto-generated |
| `jwt_token` | JWT token from login | Auto-populated |
| `test_user_id` | Test user ID | Auto-populated |
| `sample_reward_id` | Sample reward ID | Auto-populated |
| `sample_reward_points` | Sample reward points | Auto-populated |

## 🎯 **Test Scenarios**

### **Happy Path Testing**
- ✅ Service health and monitoring
- ✅ Public endpoint access
- ✅ User authentication flow
- ✅ Points earning and spending
- ✅ Tier progression
- ✅ Transaction history

### **Security Testing**
- ✅ JWT token validation
- ✅ Protected endpoint access control
- ✅ Unauthorized request rejection

### **Error Handling Testing**
- ✅ Insufficient points validation
- ✅ Invalid request data handling
- ✅ Proper error response format

### **Data Validation Testing**
- ✅ Response structure validation
- ✅ Data type validation
- ✅ Business logic validation

## 📈 **Test Results**

### **Expected Results**
- **Health Check**: 200 OK with service status
- **Metrics**: 200 OK with Prometheus metrics
- **Rewards**: 200 OK with rewards array
- **Unauthorized**: 401 Unauthorized
- **Protected Endpoints**: 200/201 OK with JWT token
- **Error Cases**: 400 Bad Request with error messages

### **Test Coverage**
- **Status Codes**: 200, 201, 400, 401
- **Response Formats**: JSON validation
- **Business Logic**: Points calculation, tier progression
- **Error Handling**: Validation errors, business rule violations

## 🚨 **Troubleshooting**

### **Common Issues**

#### **Connection Refused**
- Ensure both services are running
- Check ports 8081 (Auth) and 8082 (Loyalty)
- Verify no firewall blocking

#### **401 Unauthorized Errors**
- Check if JWT token is valid
- Verify token hasn't expired
- Ensure Auth Service is running

#### **Test Failures**
- Check service logs for errors
- Verify database schema is set up
- Ensure environment variables are correct

### **Debug Mode**
Enable detailed logging in Postman:
1. Go to **Console** (View → Show Postman Console)
2. Run tests to see detailed request/response logs

## 🔄 **Continuous Integration**

### **GitHub Actions Example**
```yaml
name: API Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Setup Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - name: Install Newman
        run: npm install -g newman
      - name: Run Tests
        run: |
          newman run test/postman/Loyalty_Service_Tests.postman_collection.json \
            -e test/postman/Loyalty_Service_Environment.postman_environment.json \
            --reporters cli,json
```

## 📚 **Additional Resources**

- [Postman Learning Center](https://learning.postman.com/)
- [Newman CLI Documentation](https://learning.postman.com/docs/running-collections/using-newman-cli/)
- [Postman Collection Format](https://schema.getpostman.com/json/collection/v2.1.0/collection.json)

## 🎉 **Success Criteria**

Your tests are successful when:
- ✅ All endpoints return expected status codes
- ✅ JWT authentication works correctly
- ✅ Points earning/spending functions properly
- ✅ Tier progression works automatically
- ✅ Error handling returns proper responses
- ✅ Transaction history is accurate

## 🚀 **Next Steps**

After running these tests:
1. **Fix any failing tests**
2. **Add more edge case tests**
3. **Create performance tests**
4. **Set up automated testing in CI/CD**
5. **Add integration tests with other services**

---

**Happy Testing! 🧪✨**


