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

const (
	nflverseBaseURL = "https://github.com/nflverse/nflverse-data/releases/download"
	startYear       = 2017 // Player stats start in 2017
	endYear         = 2025
)

func main() {
	log.Println("🔄 Reloading player_stats with CORRECTED column names...")
	log.Println("   This includes:")
	log.Println("   ✓ passing_interceptions (was 'interceptions')")
	log.Println("   ✓ EPA from parquet files (passing_epa + rushing_epa + receiving_epa)")
	log.Println("   ✓ All corrected defensive stats")
	log.Println()

	// Load environment
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("❌ MONGO_URI not set in .env")
	}

	// Connect to MongoDB
	ctx := context.Background()
	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("❌ Failed to connect to MongoDB:", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("nfl_platform")
	collection := db.Collection("player_stats")

	// Step 1: Clear existing player_stats
	log.Println("🗑️  Clearing existing player_stats...")
	result, err := collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		log.Fatal("❌ Failed to clear player_stats:", err)
	}
	log.Printf("   Deleted %d existing records\n", result.DeletedCount)

	// Step 2: Reload player_stats for all years
	totalInserted := 0
	for year := startYear; year <= endYear; year++ {
		log.Printf("\n📥 Loading player_stats for %d...", year)

		url := fmt.Sprintf("%s/stats_player/stats_player_regpost_%d.parquet", nflverseBaseURL, year)

		// Download
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("   ⚠️  Failed to download: %v", err)
			continue
		}

		if resp.StatusCode != 200 {
			log.Printf("   ⚠️  HTTP %d (data may not exist for this year)", resp.StatusCode)
			resp.Body.Close()
			continue
		}

		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Printf("   ⚠️  Failed to read: %v", err)
			continue
		}

		log.Printf("   ✓ Downloaded %d bytes", len(data))

		// Parse with CORRECTED column names
		stats, err := parquet.ParsePlayerStats(data, year, "REGPOST")
		if err != nil {
			log.Printf("   ⚠️  Failed to parse: %v", err)
			continue
		}

		log.Printf("   ✓ Parsed %d player records", len(stats))

		// Insert into MongoDB
		if len(stats) > 0 {
			// Convert to []interface{} for bulk insert
			docs := make([]interface{}, len(stats))
			for i, stat := range stats {
				docs[i] = stat
			}

			insertResult, err := collection.InsertMany(ctx, docs)
			if err != nil {
				log.Printf("   ⚠️  Failed to insert: %v", err)
				continue
			}

			inserted := len(insertResult.InsertedIDs)
			totalInserted += inserted
			log.Printf("   ✅ Inserted %d records", inserted)

			// Show sample with EPA
			if inserted > 0 && year == 2023 {
				// Find a QB with EPA
				var sample models.PlayerStats
				err := collection.FindOne(ctx, bson.M{
					"season": 2023,
					"epa":    bson.M{"$ne": 0},
				}).Decode(&sample)
				if err == nil {
					log.Printf("\n   📊 Sample record (2023):")
					log.Printf("      Player ID: %s", sample.NFLID)
					log.Printf("      Passing Yards: %d", sample.PassingYards)
					log.Printf("      Passing TDs: %d", sample.PassingTDs)
					log.Printf("      Interceptions: %d", sample.Interceptions)
					log.Printf("      EPA: %.3f ⭐", sample.EPA)
					log.Printf("      Play Count: %d", sample.PlayCount)
				}
			}
		}
	}

	log.Println()
	log.Println("=" + string(make([]byte, 60)))
	log.Printf("✅ Reload complete!")
	log.Printf("   Total records inserted: %d", totalInserted)
	log.Println()
	log.Println("🎉 player_stats now has:")
	log.Println("   ✓ Correct interceptions (passing_interceptions)")
	log.Println("   ✓ Pre-calculated EPA from parquet files")
	log.Println("   ✓ Correct defensive stats")
	log.Println()
	log.Println("🚀 Restart your backend and the player modal will show correct data!")
}
