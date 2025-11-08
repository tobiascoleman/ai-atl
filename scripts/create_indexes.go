package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/ai-atl/nfl-platform/pkg/mongodb"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "nfl_platform"
	}

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongodb.Connect(ctx, mongoURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	// Create indexes
	log.Println("Creating MongoDB indexes...")
	if err := mongodb.CreateIndexes(ctx, client.Database(dbName)); err != nil {
		log.Fatalf("Failed to create indexes: %v", err)
	}

	log.Println("Indexes created successfully!")
}
