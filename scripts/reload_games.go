package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/ai-atl/nfl-platform/internal/models"
	"github.com/ai-atl/nfl-platform/internal/parquet"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const nflverseBaseURL = "https://github.com/nflverse/nflverse-data/releases/download"

func main() {
	fmt.Println("🔄 Reloading games data...")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️  Warning: .env file not found: %v", err)
	}

	// Connect to MongoDB
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Println("⚠️  MONGO_URI not set in environment, falling back to localhost")
		mongoURI = "mongodb://localhost:27017"
	} else {
		// Mask the connection string for security
		maskedURI := mongoURI
		if len(mongoURI) > 20 {
			maskedURI = mongoURI[:20] + "...***"
		}
		log.Printf("✓ Found MONGODB_URI: %s", maskedURI)
	}

	fmt.Printf("→ Connecting to MongoDB...\n")

	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database("nfl_platform")
	ctx := context.Background()

	// Download games.parquet
	url := nflverseBaseURL + "/schedules/games.parquet"
	fmt.Printf("→ Downloading games from: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to download: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	fmt.Printf("→ Downloaded %d bytes\n", len(data))

	// Parse games
	fmt.Println("→ Parsing games...")
	games, err := parquet.ParseSchedules(data)
	if err != nil {
		log.Fatalf("Failed to parse: %v", err)
	}

	fmt.Printf("→ Parsed %d games\n", len(games))

	// Count scheduled vs final
	scheduled := 0
	final := 0
	for _, g := range games {
		if g.Status == "scheduled" {
			scheduled++
		} else {
			final++
		}
	}
	fmt.Printf("   - %d scheduled games\n", scheduled)
	fmt.Printf("   - %d completed games\n", final)

	// Clear existing games
	fmt.Println("→ Clearing existing games collection...")
	collection := db.Collection("games")
	_, err = collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Failed to clear collection: %v", err)
	}

	// Insert new games
	fmt.Println("→ Inserting games into MongoDB...")
	docs := make([]interface{}, len(games))
	for i, game := range games {
		docs[i] = game
	}

	opts := options.InsertMany().SetOrdered(false)
	result, err := collection.InsertMany(ctx, docs, opts)
	if err != nil {
		// Check if it's a bulk write error (some succeeded)
		if bulkErr, ok := err.(mongo.BulkWriteException); ok {
			if result != nil && len(result.InsertedIDs) > 0 {
				fmt.Printf("✓ Inserted %d games (with some errors)\n", len(result.InsertedIDs))
				return
			}
			fmt.Printf("⚠️  Errors: %d\n", len(bulkErr.WriteErrors))
			return
		}
		log.Fatalf("Failed to insert: %v", err)
	}

	fmt.Printf("✅ Successfully loaded %d games!\n", len(result.InsertedIDs))

	// Show some example scheduled games for 2025
	fmt.Println("\n📅 Sample 2025 scheduled games:")
	cursor, err := collection.Find(ctx,
		bson.M{"season": 2025, "status": "scheduled"},
		options.Find().SetSort(bson.D{{"start_time", 1}}).SetLimit(5))
	if err != nil {
		log.Printf("Warning: couldn't query samples: %v", err)
		return
	}
	defer cursor.Close(ctx)

	var sampleGames []models.Game
	if err := cursor.All(ctx, &sampleGames); err != nil {
		log.Printf("Warning: couldn't decode samples: %v", err)
		return
	}

	for _, g := range sampleGames {
		fmt.Printf("   Week %d: %s @ %s (%s)\n",
			g.Week, g.AwayTeam, g.HomeTeam, g.StartTime.Format("Mon Jan 2, 2006"))
	}
}
