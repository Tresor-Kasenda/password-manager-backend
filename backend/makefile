.PHONY: run build test clean install migrate dev

# Run the application
run:
	@echo "🚀 Starting server..."
	@go run cmd/server/main.go

# Build the application
build:
	@echo "🔨 Building..."
	@go build -o bin/server cmd/server/main.go
	@echo "✅ Build complete: bin/server"

# Install dependencies
install:
	@echo "📦 Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies installed"

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test -v ./...

# Run database migrations
migrate:
	@echo "🗄️  Running migrations..."
	@read -p "PostgreSQL username [postgres]: " PG_USER; \
	PG_USER=$${PG_USER:-postgres}; \
	read -sp "PostgreSQL password: " PG_PASSWORD; \
	echo; \
	read -p "Database name [password_manager]: " DB_NAME; \
	DB_NAME=$${DB_NAME:-password_manager}; \
	PGPASSWORD=$$PG_PASSWORD psql -U $$PG_USER -h localhost -d $$DB_NAME -f migrations/001_initial_schema.sql
	@echo "✅ Migrations completed"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	@rm -rf bin/
	@echo "✅ Clean complete"

# Development mode (with hot reload - requires air)
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "❌ 'air' is not installed. Install it with:"; \
		echo "   go install github.com/cosmtrek/air@latest"; \
	fi

# Setup project
setup:
	@./setup.sh


.PHONY: db-create db-drop db-reset

# Create database
db-create:
	@echo "📦 Creating database..."
	@read -p "PostgreSQL username [postgres]: " PG_USER; \
	PG_USER=$${PG_USER:-postgres}; \
	read -sp "PostgreSQL password: " PG_PASSWORD; \
	echo; \
	read -p "Database name [password_manager]: " DB_NAME; \
	DB_NAME=$${DB_NAME:-password_manager}; \
	PGPASSWORD=$$PG_PASSWORD psql -U $$PG_USER -h localhost -c "CREATE DATABASE $$DB_NAME;" && \
	echo "✅ Database '$$DB_NAME' created successfully" || \
	echo "⚠️  Database might already exist"

# Drop database
db-drop:
	@echo "⚠️  Dropping database..."
	@read -p "PostgreSQL username [postgres]: " PG_USER; \
	PG_USER=$${PG_USER:-postgres}; \
	read -sp "PostgreSQL password: " PG_PASSWORD; \
	echo; \
	read -p "Database name [password_manager]: " DB_NAME; \
	DB_NAME=$${DB_NAME:-password_manager}; \
	read -p "Are you sure you want to drop '$$DB_NAME'? (yes/no): " CONFIRM; \
	if [ "$$CONFIRM" = "yes" ]; then \
		PGPASSWORD=$$PG_PASSWORD psql -U $$PG_USER -h localhost -c "DROP DATABASE IF EXISTS $$DB_NAME;" && \
		echo "✅ Database dropped"; \
	else \
		echo "❌ Operation cancelled"; \
	fi

# Reset database (drop and recreate)
db-reset: db-drop db-create
	@echo "🔄 Database reset complete"

# Help
help:
	@echo "Available commands:"
	@echo "  make run      - Run the application"
	@echo "  make build    - Build the application"
	@echo "  make install  - Install dependencies"
	@echo "  make test     - Run tests"
	@echo "  make migrate  - Run database migrations"
	@echo "  make clean    - Clean build artifacts"
	@echo "  make dev      - Run with hot reload (requires air)"
	@echo "  make setup    - Run initial setup"