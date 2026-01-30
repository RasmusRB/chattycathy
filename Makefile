.PHONY: all up down restart logs api app db-seed db-reset build clean help dev

# Default target - start everything in Docker
all: up

# =============================================================================
# Docker Commands
# =============================================================================

# Start all services (production mode)
up:
	@if docker compose ps --status running 2>/dev/null | grep -q "chattycathy"; then \
		echo "Services are already running. Use 'make restart' to restart or 'make logs' to view logs."; \
	else \
		echo "Starting all services..."; \
		docker compose up -d; \
		echo ""; \
		echo "Services started!"; \
		echo "  App:       http://localhost:3000"; \
		echo "  API:       http://localhost:8080"; \
		echo "  Traefik:   http://localhost:8081"; \
		echo "  DB:        localhost:5432"; \
		echo ""; \
		echo "Run 'make logs' to view logs"; \
		echo "Run 'make db-seed' to populate dummy data"; \
	fi

# Start with hot reload (development mode)
dev:
	@echo "Starting services with hot reload..."
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d --build
	@echo ""
	@echo "Development services started with hot reload!"
	@echo "  App:       http://localhost:3000 (hot reload enabled)"
	@echo "  API:       http://localhost:8080"
	@echo "  Traefik:   http://localhost:8081"
	@echo ""
	@echo "Run 'make logs-app' to view frontend logs"

# Start with build
up-build:
	@echo "Building and starting all services..."
	docker compose up -d --build

# Stop all services
down:
	@echo "Stopping all services..."
	docker compose down

# Restart all services
restart: down up

# View logs
logs:
	docker compose logs -f

logs-api:
	docker compose logs -f api

logs-db:
	docker compose logs -f db

logs-traefik:
	docker compose logs -f traefik

logs-app:
	docker compose logs -f app

# =============================================================================
# Database Commands
# =============================================================================

# Seed database with dummy data
db-seed:
	@echo "Seeding database with dummy data..."
	@docker compose exec db psql -U postgres -d chattycathy -c "\
		INSERT INTO pings (message, created_at) \
		SELECT 'pong', NOW() - (random() * interval '30 days') \
		FROM generate_series(1, 10) \
		WHERE NOT EXISTS (SELECT 1 FROM pings LIMIT 1);" 2>/dev/null || \
		echo "Table doesn't exist yet. Start the API first: make up"
	@echo "Database seeding complete"

# Reset database (removes all data)
db-reset:
	@echo "Resetting database..."
	docker compose down -v
	@$(MAKE) up
	@echo "Database reset complete"

# Connect to database shell
db-shell:
	docker compose exec db psql -U postgres -d chattycathy

# =============================================================================
# Development Commands (without Docker)
# =============================================================================

# Run API locally (requires local postgres)
api-local:
	@echo "Starting API locally..."
	cd api && go run cmd/server/main.go

# Run App locally
app-local:
	@echo "Starting App locally..."
	cd app && bun run dev

# Run both locally
dev-local:
	@echo "Starting local development..."
	@$(MAKE) -j2 api-local app-local

# =============================================================================
# Build Commands
# =============================================================================

# Build all Docker images
build:
	@echo "Building Docker images..."
	docker compose build

# Build API image only
build-api:
	@echo "Building API image..."
	docker compose build api

# Build API binary locally
build-api-local:
	@echo "Building API binary..."
	cd api && go build -o bin/server cmd/server/main.go

# =============================================================================
# Utility Commands
# =============================================================================

# Install dependencies locally
deps:
	@echo "Installing API dependencies..."
	cd api && go mod download
	@echo "Installing App dependencies..."
	# cd app && npm install

# Clean up
clean:
	@echo "Cleaning up..."
	docker compose down -v --rmi local
	rm -rf api/bin
	# rm -rf app/node_modules app/dist
	@echo "Cleanup complete"

# Show running containers
ps:
	docker compose ps

# Help
help:
	@echo "ChattyCathy Development Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Docker Commands:"
	@echo "  up           - Start all services in Docker (default)"
	@echo "  up-build     - Build and start all services"
	@echo "  down         - Stop all services"
	@echo "  restart      - Restart all services"
	@echo "  logs         - View all logs (follow mode)"
	@echo "  logs-api     - View API logs only"
	@echo "  logs-db      - View database logs only"
	@echo "  ps           - Show running containers"
	@echo ""
	@echo "Database Commands:"
	@echo "  db-seed      - Seed database with dummy data"
	@echo "  db-reset     - Reset database (removes all data)"
	@echo "  db-shell     - Connect to database shell"
	@echo ""
	@echo "Build Commands:"
	@echo "  build        - Build all Docker images"
	@echo "  build-api    - Build API image only"
	@echo ""
	@echo "Local Development (without Docker):"
	@echo "  api-local    - Run API locally"
	@echo "  app-local    - Run App locally"
	@echo "  dev-local    - Run both locally"
	@echo ""
	@echo "Utility:"
	@echo "  deps         - Install dependencies"
	@echo "  clean        - Remove containers, volumes, and build artifacts"
	@echo "  help         - Show this help"
