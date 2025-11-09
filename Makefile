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

# Fix player season tracking (run this if you have old data)
fix-player-seasons:
	@echo "🔧 This will clear and rebuild the players collection"
	@echo "⚠️  All existing player data will be deleted"
	@echo ""
	@echo "Press Ctrl+C to cancel, or wait 3 seconds to continue..."
	@sleep 3
	go run scripts/fix_player_seasons.go

# Load MAXIMUM data from NFLverse (ALL 27 seasons: 1999-2025!)
# This will download ~10GB of data and take 30-60 minutes
# EPA is automatically parsed from the parquet files!
load-maximum-data:
	@echo "⚠️  WARNING: This will download ALL NFLverse data (1999-2025)"
	@echo "📦 Expected size: ~10GB"
	@echo "⏱️  Expected time: 30-60 minutes"
	@echo "✨ EPA will be automatically parsed from parquet files"
	@echo ""
	@echo "Press Ctrl+C to cancel, or wait 5 seconds to continue..."
	@sleep 5
	go run scripts/load_maximum_data.go

# Quick reload of just player_stats with corrected column names (much faster!)
reload-player-stats:
	@echo "🔄 Reloading player_stats with corrected column names"
	@echo "⏱️  Expected time: 5-10 minutes"
	@echo "✨ Includes: EPA, correct interceptions, defensive stats"
	@echo ""
	go run scripts/reload_player_stats.go

# Quick reload of just games/schedules with proper status handling
reload-games:
	@echo "🔄 Reloading games with proper status (scheduled/final)"
	@echo "⏱️  Expected time: < 1 minute"
	@echo "✨ Includes: 2025 schedule, proper game dates, status tracking"
	@echo ""
	go run scripts/reload_games.go

# Full setup for new developers
setup: deps docker-up
	@echo "Waiting for MongoDB to be ready..."
	@sleep 3
	@echo "Setup complete! Run 'make run' to start the server"

# Development mode with auto-reload (requires 'air')
dev:
	air

