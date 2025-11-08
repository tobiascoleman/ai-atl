package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ai-atl/nfl-platform/internal/config"
	"github.com/ai-atl/nfl-platform/pkg/mongodb"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	fmt.Println("🔧 Fixing Player Season Tracking")
	fmt.Println("This will:")
	fmt.Println("  1. Drop old indexes")
	fmt.Println("  2. Clear player collection")
	fmt.Println("  3. Recreate indexes with (nfl_id + season) compound key")
	fmt.Println("  4. Reload rosters with season tracking")
	fmt.Println()

	// Load environment
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	cfg := config.Load()

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongodb.Connect(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to connect to MongoDB: %w", err))
	}
	defer client.Disconnect(context.Background())

	db := client.Database("nfl_platform")
	playersCollection := db.Collection("players")

	fmt.Println("✓ Connected to MongoDB")

	// Step 1: Drop all indexes on players collection
	fmt.Println("\n📊 Dropping old indexes...")
	if err := playersCollection.Indexes().DropAll(ctx); err != nil {
		log.Printf("⚠️  Warning dropping indexes: %v", err)
	}
	fmt.Println("✓ Old indexes dropped")

	// Step 2: Delete all player documents
	fmt.Println("\n🗑️  Clearing player collection...")
	result, err := playersCollection.DeleteMany(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Failed to clear players: %v", err)
	}
	fmt.Printf("✓ Deleted %d old player documents\n", result.DeletedCount)

	// Step 3: Create new indexes with compound key
	fmt.Println("\n🔨 Creating new indexes...")
	playerIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{"nfl_id", 1}, {"season", 1}}, // Use bson.D for ordered keys
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{"team", 1}, {"position", 1}},
		},
		{
			Keys: bson.D{{"season", 1}},
		},
	}
	_, err = playersCollection.Indexes().CreateMany(ctx, playerIndexes)
	if err != nil {
		log.Fatalf("Failed to create indexes: %v", err)
	}
	fmt.Println("✓ New compound indexes created: (nfl_id + season)")

	fmt.Println("\n✅ Database fixed!")
	fmt.Println("\n🎯 Next step:")
	fmt.Println("   Run: make load-maximum-data")
	fmt.Println("   This will reload all rosters with season tracking")
}
