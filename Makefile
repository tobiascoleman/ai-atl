.PHONY: run build test clean docker-up docker-down

# Run the application
run:
	go run cmd/api/main.go

# Build the application
build:
	go build -o nfl-api cmd/api/main.go

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -cover ./...

# Clean build artifacts
clean:
	rm -f nfl-api

# Start MongoDB and Redis with Docker
docker-up:
	docker run -d -p 27017:27017 --name nfl-mongodb mongo:latest || true
	docker run -d -p 6379:6379 --name nfl-redis redis:latest || true

# Stop Docker containers
docker-down:
	docker stop nfl-mongodb nfl-redis || true
	docker rm nfl-mongodb nfl-redis || true

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Create MongoDB indexes
create-indexes:
	go run scripts/create_indexes.go

# Load sample data (for development)
load-sample-data:
	go run scripts/load_sample_data.go

# Full setup for new developers
setup: deps docker-up
	@echo "Waiting for MongoDB to be ready..."
	@sleep 3
	@echo "Setup complete! Run 'make run' to start the server"

# Development mode with auto-reload (requires 'air')
dev:
	air

