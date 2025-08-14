.PHONY: help infra-up infra-down build test lint run-% clean docker-build docker-push

# Default target
help:
	@echo "Go Loyalty & Benefits Platform - Available Commands:"
	@echo ""
	@echo "Infrastructure:"
	@echo "  infra-up      - Start all infrastructure services (Docker Compose)"
	@echo "  infra-down    - Stop all infrastructure services"
	@echo "  infra-logs    - View infrastructure logs"
	@echo ""
	@echo "Development:"
	@echo "  build         - Build all Go binaries"
	@echo "  test          - Run all tests"
	@echo "  lint          - Run linter"
	@echo "  clean         - Clean build artifacts"
	@echo ""
	@echo "Services:"
	@echo "  run-auth      - Run authentication service"
	@echo "  run-loyalty   - Run loyalty service"
	@echo "  run-catalog   - Run catalog service"
	@echo "  run-redemption - Run redemption service"
	@echo "  run-partner   - Run partner gateway"
	@echo "  run-notify    - Run notification service"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build  - Build all Docker images"
	@echo "  docker-push   - Push Docker images to registry"
	@echo ""
	@echo "Database:"
	@echo "  db-migrate    - Run database migrations"
	@echo "  db-seed       - Seed database with sample data"
	@echo ""
	@echo "Monitoring:"
	@echo "  open-jaeger   - Open Jaeger UI in browser"
	@echo "  open-grafana  - Open Grafana UI in browser"
	@echo "  open-kafka-ui - Open Kafka UI in browser"

# Infrastructure management
infra-up:
	@echo "Starting infrastructure services..."
	docker-compose -f deploy/compose/docker-compose.yml up -d
	@echo "Infrastructure services started. Waiting for health checks..."
	@echo "Services will be available at:"
	@echo "  - PostgreSQL: localhost:5432"
	@echo "  - MongoDB: localhost:27017"
	@echo "  - Redis: localhost:6379"
	@echo "  - Kafka: localhost:9092"
	@echo "  - Jaeger: http://localhost:16686"
	@echo "  - Prometheus: http://localhost:9090"
	@echo "  - Grafana: http://localhost:3000 (admin/admin)"
	@echo "  - Kafka UI: http://localhost:8080"

infra-down:
	@echo "Stopping infrastructure services..."
	docker-compose -f deploy/compose/docker-compose.yml down -v
	@echo "Infrastructure services stopped"

infra-logs:
	docker-compose -f deploy/compose/docker-compose.yml logs -f

# Development commands
build:
	@echo "Building Go binaries..."
	go mod tidy
	go build ./...

test:
	@echo "Running tests..."
	go test ./... -race -count=1 -v

lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run ./...; \
	fi

clean:
	@echo "Cleaning build artifacts..."
	go clean
	rm -rf bin/
	find . -name "*.exe" -delete

# Service runners
run-auth:
	@echo "Starting Auth Service..."
	@if [ -f .env ]; then export $$(cat .env | xargs); fi; \
	go run ./cmd/auth-svc

run-loyalty:
	@echo "Starting Loyalty Service..."
	@if [ -f .env ]; then export $$(cat .env | xargs); fi; \
	go run ./cmd/loyalty-svc

run-catalog:
	@echo "Starting Catalog Service..."
	@if [ -f .env ]; then export $$(cat .env | xargs); fi; \
	go run ./cmd/catalog-svc

run-redemption:
	@echo "Starting Redemption Service..."
	@if [ -f .env ]; then export $$(cat .env | xargs); fi; \
	go run ./cmd/redemption-svc

run-partner:
	@echo "Starting Partner Gateway..."
	@if [ -f .env ]; then export $$(cat .env | xargs); fi; \
	go run ./cmd/partner-gateway

run-notify:
	@echo "Starting Notification Service..."
	@if [ -f .env ]; then export $$(cat .env | xargs); fi; \
	go run ./cmd/notify-svc

# Docker commands
docker-build:
	@echo "Building Docker images..."
	docker build -t go-loyalty-benefits/auth-svc:latest ./cmd/auth-svc
	docker build -t go-loyalty-benefits/loyalty-svc:latest ./cmd/loyalty-svc
	docker build -t go-loyalty-benefits/catalog-svc:latest ./cmd/catalog-svc
	docker build -t go-loyalty-benefits/redemption-svc:latest ./cmd/redemption-svc
	docker build -t go-loyalty-benefits/partner-gateway:latest ./cmd/partner-gateway
	docker build -t go-loyalty-benefits/notify-svc:latest ./cmd/notify-svc

docker-push:
	@echo "Pushing Docker images..."
	docker push go-loyalty-benefits/auth-svc:latest
	docker push go-loyalty-benefits/loyalty-svc:latest
	docker push go-loyalty-benefits/catalog-svc:latest
	docker push go-loyalty-benefits/redemption-svc:latest
	docker push go-loyalty-benefits/partner-gateway:latest
	docker push go-loyalty-benefits/notify-svc:latest

# Database commands
db-migrate:
	@echo "Running database migrations..."
	@echo "TODO: Implement migration runner"

db-seed:
	@echo "Seeding database with sample data..."
	@echo "TODO: Implement data seeder"

# Monitoring shortcuts
open-jaeger:
	@echo "Opening Jaeger UI..."
	@if command -v xdg-open >/dev/null 2>&1; then \
		xdg-open http://localhost:16686; \
	elif command -v open >/dev/null 2>&1; then \
		open http://localhost:16686; \
	else \
		echo "Please open http://localhost:16686 in your browser"; \
	fi

open-grafana:
	@echo "Opening Grafana UI..."
	@if command -v xdg-open >/dev/null 2>&1; then \
		xdg-open http://localhost:3000; \
	elif command -v open >/dev/null 2>&1; then \
		open http://localhost:3000; \
	else \
		echo "Please open http://localhost:3000 in your browser (admin/admin)"; \
	fi

open-kafka-ui:
	@echo "Opening Kafka UI..."
	@if command -v xdg-open >/dev/null 2>&1; then \
		xdg-open http://localhost:8080; \
	elif command -v open >/dev/null 2>&1; then \
		open http://localhost:8080; \
	else \
		echo "Please open http://localhost:8080 in your browser"; \
	fi

# Development setup
setup-dev:
	@echo "Setting up development environment..."
	@if ! command -v go >/dev/null 2>&1; then \
		echo "Go is not installed. Please install Go 1.22+ first."; \
		exit 1; \
	fi
	@echo "Go version: $$(go version)"
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development environment setup complete!"

# Health checks
health-check:
	@echo "Checking service health..."
	@echo "Auth Service: $$(curl -s http://localhost:8081/healthz | jq -r '.status' 2>/dev/null || echo 'unavailable')"
	@echo "Loyalty Service: $$(curl -s http://localhost:8082/healthz | jq -r '.status' 2>/dev/null || echo 'unavailable')"
	@echo "Catalog Service: $$(curl -s http://localhost:8083/healthz | jq -r '.status' 2>/dev/null || echo 'unavailable')"
	@echo "Redemption Service: $$(curl -s http://localhost:8084/healthz | jq -r '.status' 2>/dev/null || echo 'unavailable')"
	@echo "Partner Gateway: $$(curl -s http://localhost:8085/healthz | jq -r '.status' 2>/dev/null || echo 'unavailable')"
	@echo "Notification Service: $$(curl -s http://localhost:8086/healthz | jq -r '.status' 2>/dev/null || echo 'unavailable')"
