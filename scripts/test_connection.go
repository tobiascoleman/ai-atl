package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func main() {
	fmt.Println("=== MongoDB Connection Diagnostic ===")
	fmt.Println()

	// Load .env file
	fmt.Println("1. Loading .env file...")
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	} else {
		fmt.Println("   ✓ .env file loaded")
	}
	fmt.Println()

	// Check environment variables
	fmt.Println("2. Checking environment variables...")
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("   ✗ MONGO_URI not set!")
	}

	// Mask the password for display
	maskedURI := mongoURI
	if len(mongoURI) > 50 {
		maskedURI = mongoURI[:30] + "***MASKED***" + mongoURI[len(mongoURI)-20:]
	}
	fmt.Printf("   ✓ MONGO_URI found: %s\n", maskedURI)
	fmt.Println()

	// Attempt connection
	fmt.Println("3. Attempting MongoDB connection...")
	fmt.Println("   Setting up client options...")

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(mongoURI).
		SetServerAPIOptions(serverAPI).
		SetConnectTimeout(30 * time.Second).
		SetServerSelectionTimeout(30 * time.Second)

	fmt.Println("   ✓ Client options configured")
	fmt.Println()

	fmt.Println("4. Connecting to MongoDB Atlas...")
	client, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Fatalf("   ✗ Failed to create client: %v", err)
	}
	defer client.Disconnect(context.Background())
	fmt.Println("   ✓ Client created")
	fmt.Println()

	// Ping the database
	fmt.Println("5. Pinging MongoDB...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalf("   ✗ Ping failed: %v", err)
	}
	fmt.Println("   ✓ Ping successful!")
	fmt.Println()

	// List databases
	fmt.Println("6. Listing databases...")
	databases, err := client.ListDatabaseNames(ctx, map[string]interface{}{})
	if err != nil {
		log.Printf("   ✗ Failed to list databases: %v", err)
	} else {
		fmt.Println("   ✓ Available databases:")
		for _, db := range databases {
			fmt.Printf("     - %s\n", db)
		}
	}
	fmt.Println()

	// Test specific database
	fmt.Println("7. Testing AI-ATL database...")
	db := client.Database("nfl_platform")
	collections, err := db.ListCollectionNames(ctx, map[string]interface{}{})
	if err != nil {
		log.Printf("   ⚠ Warning: Could not list collections: %v", err)
	} else {
		if len(collections) == 0 {
			fmt.Println("   ℹ Database exists but has no collections yet (this is normal for new databases)")
		} else {
			fmt.Println("   ✓ Existing collections:")
			for _, coll := range collections {
				fmt.Printf("     - %s\n", coll)
			}
		}
	}
	fmt.Println()

	fmt.Println("=== SUCCESS! ===")
	fmt.Println("MongoDB Atlas connection is working correctly!")
	fmt.Println()
	fmt.Println("Your backend should now work. Try running:")
	fmt.Println("  go run cmd/api/main.go")
}
