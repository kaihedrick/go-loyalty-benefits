# Cursor Rules for Go Loyalty & Benefits Platform

## 🎯 Project Overview
This is a Go-based microservices platform for loyalty programs and benefits management. The system uses PostgreSQL, MongoDB, Redis, Kafka, and follows microservices architecture patterns.

## 🧠 AI Assistant Guidelines

### General Responsibilities:
- Guide the development of idiomatic, maintainable, and high-performance Go code.
- Enforce modular design and separation of concerns through Clean Architecture.
- Promote test-driven development, robust observability, and scalable patterns across services.

### Architecture Patterns:
- Apply **Clean Architecture** by structuring code into handlers/controllers, services/use cases, repositories/data access, and domain models.
- Use **domain-driven design** principles where applicable.
- Prioritize **interface-driven development** with explicit dependency injection.
- Prefer **composition over inheritance**; favor small, purpose-specific interfaces.
- Ensure that all public functions interact with interfaces, not concrete types, to enhance flexibility and testability.

### Project Structure Guidelines:
- Use a consistent project layout:
  - cmd/: application entrypoints
  - internal/: core application logic (not exposed externally)
  - pkg/: shared utilities and packages
  - api/: gRPC/REST transport definitions and handlers
  - configs/: configuration schemas and loading
  - test/: test utilities, mocks, and integration tests
- Group code by feature when it improves clarity and cohesion.
- Keep logic decoupled from framework-specific code.

### Development Best Practices:
- Write **short, focused functions** with a single responsibility.
- Always **check and handle errors explicitly**, using wrapped errors for traceability ('fmt.Errorf("context: %w", err)').
- Avoid **global state**; use constructor functions to inject dependencies.
- Leverage **Go's context propagation** for request-scoped values, deadlines, and cancellations.
- Use **goroutines safely**; guard shared state with channels or sync primitives.
- **Defer closing resources** and handle them carefully to avoid leaks.

### Security and Resilience:
- Apply **input validation and sanitization** rigorously, especially on inputs from external sources.
- Use secure defaults for **JWT, cookies**, and configuration settings.
- Isolate sensitive operations with clear **permission boundaries**.
- Implement **retries, exponential backoff, and timeouts** on all external calls.
- Use **circuit breakers and rate limiting** for service protection.
- Consider implementing **distributed rate-limiting** to prevent abuse across services (e.g., using Redis).

### Documentation and Standards:
- Document public functions and packages with **GoDoc-style comments**.
- Provide concise **READMEs** for services and libraries.
- Maintain a 'CONTRIBUTING.md' and 'ARCHITECTURE.md' to guide team practices.
- Enforce naming consistency and formatting with 'go fmt', 'goimports', and 'golangci-lint'.

### Observability with OpenTelemetry:
- Use **OpenTelemetry** for distributed tracing, metrics, and structured logging.
- Start and propagate tracing **spans** across all service boundaries (HTTP, gRPC, DB, external APIs).
- Always attach 'context.Context' to spans, logs, and metric exports.
- Use **otel.Tracer** for creating spans and **otel.Meter** for collecting metrics.
- Record important attributes like request parameters, user ID, and error messages in spans.
- Use **log correlation** by injecting trace IDs into structured logs.
- Export data to **OpenTelemetry Collector**, **Jaeger**, or **Prometheus**.

### Tracing and Monitoring Best Practices:
- Trace all **incoming requests** and propagate context through internal and external calls.
- Use **middleware** to instrument HTTP and gRPC endpoints automatically.
- Annotate slow, critical, or error-prone paths with **custom spans**.
- Monitor application health via key metrics: **request latency, throughput, error rate, resource usage**.
- Define **SLIs** (e.g., request latency < 300ms) and track them with **Prometheus/Grafana** dashboards.
- Alert on key conditions (e.g., high 5xx rates, DB errors, Redis timeouts) using a robust alerting pipeline.
- Avoid excessive **cardinality** in labels and traces; keep observability overhead minimal.
- Use **log levels** appropriately (info, warn, error) and emit **JSON-formatted logs** for ingestion by observability tools.
- Include unique **request IDs** and trace context in all logs for correlation.

### Performance:
- Use **benchmarks** to track performance regressions and identify bottlenecks.
- Minimize **allocations** and avoid premature optimization; profile before tuning.
- Instrument key areas (DB, external calls, heavy computation) to monitor runtime behavior.

### Concurrency and Goroutines:
- Ensure safe use of **goroutines**, and guard shared state with channels or sync primitives.
- Implement **goroutine cancellation** using context propagation to avoid leaks and deadlocks.

### Tooling and Dependencies:
- Rely on **stable, minimal third-party libraries**; prefer the standard library where feasible.
- Use **Go modules** for dependency management and reproducibility.
- Version-lock dependencies for deterministic builds.
- Integrate **linting, testing, and security checks** in CI pipelines.

### Key Conventions:
1. Prioritize **readability, simplicity, and maintainability**.
2. Design for **change**: isolate business logic and minimize framework lock-in.
3. Emphasize clear **boundaries** and **dependency inversion**.
4. Ensure all behavior is **observable, testable, and documented**.
5. **Automate workflows** for testing, building, and deployment.

## 🏗️ Architecture Principles

### Microservices Design
- Each service should be independently deployable
- Services communicate via HTTP APIs and Kafka events
- Database per service pattern (PostgreSQL for relational data, MongoDB for document data)
- Event-driven architecture for loose coupling

### Code Organization
- `cmd/` - Service entry points
- `internal/` - Private application code
- `pkg/` - Public libraries that can be imported by other projects
- `deploy/` - Infrastructure and deployment configurations

## 🐹 Go Best Practices

### Code Style
- Use `gofmt` for code formatting
- Follow Go naming conventions (PascalCase for exported, camelCase for private)
- Use meaningful variable names, avoid abbreviations
- Keep functions small and focused (single responsibility)

### Error Handling
- Always check errors, never ignore them
- Use `fmt.Errorf` with `%w` for error wrapping
- Return errors from functions, don't log them inside
- Use custom error types for specific error conditions

### Database Operations
- Use prepared statements to prevent SQL injection
- Handle `sql.ErrNoRows` appropriately (it's not always an error)
- Use transactions for multi-step operations
- Implement proper connection pooling

### Configuration
- Use environment variables for configuration
- Avoid hardcoded values in code
- Use Viper for configuration management
- Support `.env` files for local development

## 🔒 Security Guidelines

### Authentication & Authorization
- Use JWT tokens for stateless authentication
- Implement proper password hashing with bcrypt
- Validate all input data
- Use HTTPS in production
- Implement rate limiting

### Data Protection
- Never log sensitive information (passwords, tokens, PII)
- Use parameterized queries to prevent SQL injection
- Validate and sanitize all user inputs
- Implement proper access controls

## 📊 Observability

### Logging
- Use structured logging with logrus
- Include correlation IDs for request tracing
- Log at appropriate levels (debug, info, warn, error)
- Include context in log messages

### Metrics
- Expose Prometheus metrics on `/metrics` endpoint
- Track HTTP request counts, response times, error rates
- Monitor database connection pool health
- Track business metrics (user registrations, transactions)

### Tracing
- Use OpenTelemetry for distributed tracing
- Include trace IDs in logs
- Track service-to-service calls

## 🧪 Testing

### Unit Tests
- Write tests for all business logic
- Use table-driven tests for multiple scenarios
- Mock external dependencies
- Aim for >80% code coverage
- Use tools like 'go test -cover' to ensure adequate test coverage

### Integration Tests
- Test database operations with test containers
- Test HTTP endpoints with real HTTP requests
- Test Kafka message handling
- Separate **fast unit tests** from slower integration and E2E tests
- Ensure **test coverage** for every exported function, with behavioral checks

## 🚀 Performance

### Database
- Use connection pooling
- Implement proper indexing strategies
- Use transactions for data consistency
- Consider read replicas for heavy read workloads

### HTTP
- Use middleware for common operations (logging, CORS, auth)
- Implement proper timeout handling
- Use gzip compression for responses
- Consider caching strategies

## 🔄 Event-Driven Patterns

### Kafka Usage
- Use topics for different event types
- Implement idempotency for event processing
- Use dead letter queues for failed events
- Consider event schema evolution

### Outbox Pattern
- Store events in database before publishing
- Use background workers to publish events
- Ensure at-least-once delivery semantics

## 📝 Documentation

### Code Comments
- Comment exported functions and types
- Explain complex business logic
- Document API endpoints with examples
- Keep comments up-to-date with code changes

### API Documentation
- Use OpenAPI/Swagger for API documentation
- Include request/response examples
- Document error codes and messages
- Provide integration examples

## 🛠️ Development Workflow

### Dependencies
- Use Go modules for dependency management
- Run `go mod tidy` regularly
- Pin dependency versions for stability
- Avoid vendoring unless necessary

### Build & Deploy
- Use Makefiles for common operations
- Support both local development and containerized deployment
- Use multi-stage Docker builds
- Implement health checks for all services

## 🚨 Anti-Patterns to Avoid

- Don't use global variables
- Don't ignore errors
- Don't use `panic` in production code
- Don't hardcode configuration values
- Don't use `database/sql` directly (use `pgx` for PostgreSQL)
- Don't log sensitive information
- Don't use unbounded goroutines
- Don't forget to implement proper timeouts

## 🔧 Tools & Commands

### Essential Commands
```bash
# Format code
go fmt ./...

# Run tests
go test ./...

# Build
go build ./cmd/auth-svc

# Run service
go run cmd/auth-svc/main.go

# Update dependencies
go mod tidy
go mod download

# Check for issues
go vet ./...
golangci-lint run
```

### Docker Commands
```bash
# Start infrastructure
docker-compose -f deploy/compose/docker-compose.yml up -d

# View logs
docker-compose -f deploy/compose/docker-compose.yml logs -f

# Stop infrastructure
docker-compose -f deploy/compose/docker-compose.yml down
```

## 🐛 CRITICAL BUGS & LESSONS LEARNED

### 🚨 Configuration Loading Issues (RESOLVED)

#### Problem 1: Environment Variables Not Loading
**Symptoms**: Database connection errors with empty credentials
**Root Cause**: Viper configuration loading order and `.env` file path issues
**Solution**: 
```go
// CRITICAL: Load .env BEFORE calling viper.AutomaticEnv()
if err := godotenv.Load(); err != nil {
    // Fallback to manual parsing
    if envFile, err := findEnvFile(); err == nil {
        parseAndSetEnvVars(envFile)
    }
}
viper.AutomaticEnv()
```

#### Problem 2: .env File Path Resolution
**Symptoms**: "Config File .env Not Found" errors
**Root Cause**: Viper looking in wrong directory (cmd/service instead of project root)
**Solution**: Add multiple search paths
```go
viper.AddConfigPath(".")
viper.AddConfigPath("..")
viper.AddConfigPath("../..")
```

#### Problem 3: JWT Configuration Not Loading
**Symptoms**: Empty JWT secret causing authentication failures
**Root Cause**: Inconsistent environment variable naming (JWT_SECRET vs AUTH-SVC_SECURITY_JWT_SECRET)
**Solution**: Standardize naming convention in .env file

### 🚨 Database Schema Mismatches (RESOLVED)

#### Problem 4: SQL Query Failures
**Symptoms**: "Failed to check existing user: no rows in result set"
**Root Cause**: Go struct fields didn't match PostgreSQL table schema
**Solution**: Update User struct to include all database columns
```go
type User struct {
    ID           string    `json:"id"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"-"`
    Role         string    `json:"role"`
    FirstName    *string   `json:"first_name,omitempty"`  // ADDED
    LastName     *string   `json:"last_name,omitempty"`   // ADDED
    Phone        *string   `json:"phone,omitempty"`       // ADDED
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

#### Problem 5: Database Not Initialized
**Symptoms**: Tables don't exist when service starts
**Root Cause**: init.sql script not automatically executed
**Solution**: Manual database initialization required
```bash
# Copy init script to container
docker cp deploy/compose/init.sql <container_name>:/tmp/

# Execute in container
docker exec -it <container_name> psql -U loyalty -d loyalty -f /tmp/init.sql
```

### 🚨 Error Handling Logic Issues (RESOLVED)

#### Problem 6: sql.ErrNoRows Treated as Fatal Error
**Symptoms**: Registration failing with "Internal server error" for new users
**Root Cause**: sql.ErrNoRows means "user not found" - expected for new registrations
**Solution**: Proper error handling
```go
existingUser, err := s.getUserByEmail(r.Context(), req.Email)
if err != nil {
    if err == sql.ErrNoRows {
        // This is expected for new registrations - continue
        s.logger.Infof("User with email %s does not exist (expected for new registrations)", req.Email)
        return
    } else {
        // This is a real error - return 500
        s.logger.Errorf("Failed to check existing user: %v", err)
        render.Status(r, http.StatusInternalServerError)
        render.JSON(w, r, map[string]string{"error": "Internal server error"})
        return
    }
}
```

### 🚨 Compilation & Runtime Issues (RESOLVED)

#### Problem 7: pgxpool.Config Field Names
**Symptoms**: Compilation errors with unknown fields
**Root Cause**: Incorrect field names in pgxpool.Config
**Solution**: Use correct field names
```go
// WRONG:
MaxConnsLifetime: time.Hour,
MaxConnsIdleTime: time.Minute * 5,

// CORRECT:
MaxConnLifetime: time.Hour,
MaxConnIdleTime: time.Minute * 5,
```

#### Problem 8: Unused Imports
**Symptoms**: Compilation errors about unused packages
**Root Cause**: Leftover imports from refactoring
**Solution**: Run `go mod tidy` and remove unused imports

#### Problem 9: Port Conflicts
**Symptoms**: "Only one usage of each socket address is normally permitted"
**Root Cause**: Multiple services trying to use same port
**Solution**: Ensure unique ports per service in configuration

### 🚨 API Routing Issues (RESOLVED)

#### Problem 10: 404 Page Not Found
**Symptoms**: API endpoints returning 404
**Root Cause**: Incorrect URL paths (e.g., `/register` instead of `/v1/auth/register`)
**Solution**: Use correct API paths from service routing
```bash
# WRONG:
curl -X POST http://localhost:8081/register

# CORRECT:
curl -X POST http://localhost:8081/v1/auth/register
```

## 🎯 Debugging Checklist

When encountering issues, check in this order:

1. **Configuration Loading**: Are environment variables being read correctly?
2. **Database Connection**: Can the service connect to PostgreSQL?
3. **Database Schema**: Do tables exist and match Go structs?
4. **Error Handling**: Are expected errors (like sql.ErrNoRows) handled properly?
5. **API Routing**: Are endpoints accessible at correct paths?
6. **Port Conflicts**: Are services using unique ports?
7. **Dependencies**: Are all Go modules properly downloaded?

## 📚 Learning Resources

- [Go by Example](https://gobyexample.com/)
- [Go Web Examples](https://gowebexamples.com/)
- [Go Best Practices](https://github.com/golang/go/wiki/CodeReviewComments)
- [Microservices Patterns](https://microservices.io/patterns/)

---

**Remember**: This is a living document. Update it as new bugs are discovered and resolved!
