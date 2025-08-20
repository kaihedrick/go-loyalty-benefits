# Loyalty Service

The Loyalty Service is the core component of the Go Loyalty & Benefits Platform, responsible for managing user loyalty points, tiers, and rewards.

## 🎯 **Features**

### **Core Functionality**
- **Points Management** - Earn, spend, and track loyalty points
- **Tier System** - Automatic tier progression (Bronze → Silver → Gold → Platinum)
- **Rewards Catalog** - Browse and redeem available rewards
- **Transaction History** - Complete audit trail of all point activities
- **JWT Authentication** - Secure endpoints with token validation

### **Tier Progression**
- **Bronze**: 0-999 points
- **Silver**: 1,000-4,999 points  
- **Gold**: 5,000-9,999 points
- **Platinum**: 10,000+ points

## 🚀 **Quick Start**

### **1. Start the Service**
```bash
go run cmd/loyalty-svc/main.go
```

The service will start on port 8082 (configurable via `LOYALTY-SVC_APP_HTTP_ADDR`).

### **2. Set Up Database**
```bash
# Run the loyalty schema
docker exec -i loyalty-postgres psql -U loyalty -d loyalty < deploy/compose/loyalty_schema.sql
```

### **3. Test the Service**
```bash
# PowerShell (Windows)
.\test\test_loyalty_service.ps1

# Bash (Linux/Mac/WSL)
./test/test_loyalty_service.sh
```

## 📡 **API Endpoints**

### **Public Endpoints**

#### **GET /v1/loyalty/rewards**
Get all available rewards for redemption.

**Response:**
```json
{
  "success": true,
  "message": "Rewards retrieved successfully",
  "data": [
    {
      "id": "reward-001",
      "name": "Free Coffee",
      "description": "Redeem for a free coffee at any participating location",
      "points_cost": 100,
      "category": "Food & Beverage",
      "is_active": true
    }
  ]
}
```

### **Protected Endpoints (Require JWT Token)**

#### **POST /v1/loyalty/earn**
Earn points for a user.

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Request Body:**
```json
{
  "user_id": "user-123",
  "amount": 100,
  "description": "Purchase at Coffee Shop"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Points earned successfully",
  "data": {
    "transaction": {
      "id": "tx-001",
      "user_id": "user-123",
      "type": "earn",
      "amount": 100,
      "description": "Purchase at Coffee Shop",
      "created_at": "2025-08-19T17:00:00Z"
    },
    "user": {
      "id": "user-123",
      "email": "user@example.com",
      "points": 1100,
      "tier": "Silver",
      "created_at": "2025-08-19T10:00:00Z",
      "updated_at": "2025-08-19T17:00:00Z"
    }
  }
}
```

#### **POST /v1/loyalty/spend**
Spend points for a user.

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Request Body:**
```json
{
  "user_id": "user-123",
  "amount": 100,
  "description": "Redeemed free coffee"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Points spent successfully",
  "data": {
    "transaction": {
      "id": "tx-002",
      "user_id": "user-123",
      "type": "spend",
      "amount": 100,
      "description": "Redeemed free coffee",
      "created_at": "2025-08-19T17:30:00Z"
    },
    "user": {
      "id": "user-123",
      "email": "user@example.com",
      "points": 1000,
      "tier": "Silver",
      "created_at": "2025-08-19T10:00:00Z",
      "updated_at": "2025-08-19T17:30:00Z"
    }
  }
}
```

#### **GET /v1/loyalty/balance**
Get the current user's loyalty balance and tier.

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Response:**
```json
{
  "success": true,
  "message": "Balance retrieved successfully",
  "data": {
    "id": "user-123",
    "email": "user@example.com",
    "points": 1000,
    "tier": "Silver",
    "created_at": "2025-08-19T10:00:00Z",
    "updated_at": "2025-08-19T17:30:00Z"
  }
}
```

#### **GET /v1/loyalty/history**
Get the user's transaction history.

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Response:**
```json
{
  "success": true,
  "message": "History retrieved successfully",
  "data": [
    {
      "id": "tx-001",
      "user_id": "user-123",
      "type": "earn",
      "amount": 100,
      "description": "Purchase at Coffee Shop",
      "created_at": "2025-08-19T17:00:00Z"
    },
    {
      "id": "tx-002",
      "user_id": "user-123",
      "type": "spend",
      "amount": 100,
      "description": "Redeemed free coffee",
      "created_at": "2025-08-19T17:30:00Z"
    }
  ]
}
```

## 🗄️ **Database Schema**

### **Tables**

#### **loyalty_users**
Stores user loyalty profiles and current point balances.

```sql
CREATE TABLE loyalty_users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    points INTEGER DEFAULT 0 NOT NULL,
    tier VARCHAR(50) DEFAULT 'Bronze' NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
```

#### **loyalty_transactions**
Tracks all point earning and spending activities.

```sql
CREATE TABLE loyalty_transactions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('earn', 'spend')),
    amount INTEGER NOT NULL CHECK (amount > 0),
    description TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES loyalty_users(id) ON DELETE CASCADE
);
```

#### **loyalty_rewards**
Catalog of available rewards for redemption.

```sql
CREATE TABLE loyalty_rewards (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    points_cost INTEGER NOT NULL CHECK (points_cost > 0),
    category VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT true NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
```

### **Automatic Features**
- **Tier Calculation**: Automatically updates user tier based on point balance
- **Timestamp Updates**: Automatically updates `updated_at` fields
- **Data Integrity**: Foreign key constraints and check constraints

## 🔐 **Authentication**

All protected endpoints require a valid JWT token in the `Authorization` header:

```
Authorization: Bearer <JWT_TOKEN>
```

The token must be obtained from the Auth Service and contain valid claims for:
- `user_id`: The user's unique identifier
- `email`: The user's email address  
- `role`: The user's role in the system

## 🧪 **Testing**

### **Run Tests**
```bash
# PowerShell (Windows)
.\test\test_loyalty_service.ps1

# Bash (Linux/Mac/WSL)
./test/test_loyalty_service.sh
```

### **Test Coverage**
- ✅ Health check endpoint
- ✅ Metrics endpoint
- ✅ Public rewards endpoint
- ✅ Protected endpoint authentication
- ✅ JWT token validation

## 🔧 **Configuration**

### **Environment Variables**

| Variable | Description | Default |
|----------|-------------|---------|
| `LOYALTY-SVC_APP_NAME` | Service name | `loyalty-svc` |
| `LOYALTY-SVC_APP_HTTP_ADDR` | HTTP address | `:8082` |
| `LOYALTY-SVC_APP_LOG_LEVEL` | Log level | `info` |
| `AUTH-SVC_DATABASE_POSTGRES_*` | Database configuration | See env.example |
| `JWT_SECRET` | JWT signing secret | Required |
| `JWT_ISSUER` | JWT issuer claim | `go-loyalty` |
| `JWT_AUDIENCE` | JWT audience claim | `go-loyalty-clients` |
| `JWT_EXPIRATION` | JWT expiration time | `24h` |

## 🚀 **Next Steps**

### **Immediate Enhancements**
1. **Integration Testing** - Test with real JWT tokens from Auth Service
2. **Points Calculation** - Implement MCC-based point multipliers
3. **Reward Redemption** - Connect with Redemption Service
4. **Event Emission** - Send events to Kafka for other services

### **Future Features**
1. **Partner Integration** - External merchant point earning
2. **Promotional Campaigns** - Bonus point events
3. **Point Expiration** - Time-based point validity
4. **Analytics Dashboard** - User behavior insights

## 📚 **Related Services**

- **Auth Service** - User authentication and JWT generation
- **Catalog Service** - Product catalog and inventory
- **Redemption Service** - Reward redemption processing
- **Notification Service** - User notifications and alerts

## 🐛 **Troubleshooting**

### **Common Issues**

#### **Service Won't Start**
- Check if port 8082 is available
- Verify database connection
- Check environment variable configuration

#### **Database Connection Issues**
- Ensure PostgreSQL is running
- Verify database credentials
- Check if loyalty schema is created

#### **Authentication Errors**
- Verify JWT token is valid
- Check token expiration
- Ensure token contains required claims

### **Logs**
The service uses structured JSON logging. Look for:
- Configuration debug information
- Database connection status
- Request/response logging
- Error details with context

## 📖 **API Documentation**

For complete API documentation, see the OpenAPI/Swagger specification or test the endpoints directly using the provided test scripts.
