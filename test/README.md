# Testing Your Auth Service

This directory contains comprehensive test scripts to verify your auth service is working correctly.

## 🚀 Quick Start

### 1. Start Your Auth Service
```bash
# In one terminal, start your auth service
go run cmd/auth-svc/main.go
```

### 2. Run the Tests
```bash
# PowerShell (Windows)
.\test\test_auth_service.ps1

# Bash (Linux/Mac/WSL)
./test/test_auth_service.sh
```

## 🧪 What Gets Tested

### **Basic Functionality**
- ✅ **Health Check** - `/healthz` endpoint
- ✅ **Metrics** - `/metrics` endpoint for Prometheus
- ✅ **User Registration** - Create new user accounts
- ✅ **User Login** - Authenticate existing users
- ✅ **Protected Endpoints** - JWT token validation

### **Security & Validation**
- ✅ **Duplicate Registration** - Prevents duplicate emails
- ✅ **Invalid Login** - Rejects wrong passwords
- ✅ **Input Validation** - Rejects malformed data
- ✅ **JWT Protection** - Requires valid tokens for protected routes

### **Database Operations**
- ✅ **User Creation** - Stores users in PostgreSQL
- ✅ **Password Hashing** - Passwords are properly hashed
- ✅ **Data Integrity** - All required fields are stored

## 🔍 Manual Testing

If you prefer to test manually, here are the key endpoints:

### **Health Check**
```bash
curl http://localhost:8081/healthz
```

### **Metrics**
```bash
curl http://localhost:8081/metrics
```

### **User Registration**
```bash
curl -X POST http://localhost:8081/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

### **User Login**
```bash
curl -X POST http://localhost:8081/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

### **Protected Endpoint (with token)**
```bash
# Use the token from login/registration response
curl -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  http://localhost:8081/v1/auth/me
```

## 🧹 Cleanup

After testing, clean up test data:

```bash
# Remove test user from database
docker exec -it loyalty-postgres psql -U loyalty -d loyalty \
  -c "DELETE FROM users WHERE email = 'test@example.com';"
```

## 🐛 Troubleshooting

### **Service Not Running**
- Ensure your auth service is running on port 8081
- Check for port conflicts with `netstat -an | findstr 8081`

### **Database Connection Issues**
- Verify PostgreSQL is running: `docker ps | grep postgres`
- Check database credentials in your `.env` file
- Ensure database is initialized with `init.sql`

### **JWT Issues**
- Verify JWT_SECRET is set in your `.env` file
- Check that JWT environment variables are being loaded correctly

### **Test Failures**
- Look at the specific error messages in the test output
- Check the auth service logs for detailed error information
- Verify all required environment variables are set

## 📊 Expected Results

When everything is working correctly, you should see:

```
🧪 Testing Auth Service...
=====================================

1️⃣ Testing Health Check...
✅ Health check passed: ok

2️⃣ Testing Metrics Endpoint...
✅ Metrics endpoint working (1271 bytes)

3️⃣ Testing User Registration...
✅ User registration successful
ℹ️  User ID: [uuid]
ℹ️  Token length: [length]

4️⃣ Testing User Login...
✅ User login successful
ℹ️  Token length: [length]

5️⃣ Testing Protected Endpoint (without token)...
✅ Protected endpoint correctly rejected request without token (401 Unauthorized)

6️⃣ Testing Protected Endpoint (with token)...
✅ Protected endpoint accessible with valid token
ℹ️  User ID: [uuid]
ℹ️  User Email: test@example.com

7️⃣ Testing Duplicate Registration...
✅ Duplicate registration correctly rejected (409 Conflict)

8️⃣ Testing Invalid Login...
✅ Invalid login correctly rejected (401 Unauthorized)

9️⃣ Testing Invalid Registration Data...
✅ Invalid registration data correctly rejected (400 Bad Request)

🔟 Verifying Database State...
ℹ️  Expected: User 'test@example.com' exists in database
ℹ️  Expected: User has hashed password (not plaintext)
ℹ️  Expected: User has created_at and updated_at timestamps

📊 Test Summary
===============
✅ Health Check: Working
✅ Metrics Endpoint: Working
✅ User Registration: Working
✅ User Login: Working
✅ Protected Endpoints: Working
✅ Duplicate Registration: Properly rejected
✅ Invalid Login: Properly rejected
✅ Input Validation: Working

🎉 All tests completed!
Your auth service is working correctly!
```

## 🔄 Next Steps

After passing all tests:

1. **Test other services** - Apply similar testing to loyalty, catalog, etc.
2. **Integration testing** - Test service-to-service communication
3. **Load testing** - Test performance under load
4. **Security testing** - Test for common vulnerabilities
5. **Monitoring** - Set up alerts based on the metrics you're now collecting

## 📚 Additional Resources

- [Go Testing Best Practices](https://golang.org/doc/code.html#Testing)
- [HTTP Status Codes](https://httpstatuses.com/)
- [JWT Debugger](https://jwt.io/)
- [PostgreSQL Commands](https://www.postgresql.org/docs/current/app-psql.html)
