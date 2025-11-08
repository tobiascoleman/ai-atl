package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/ai-atl/nfl-platform/internal/models"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database(dbName)

	// Load sample players
	log.Println("Loading sample players...")
	players := []interface{}{
		models.Player{
			ID:          bson.NewObjectID(),
			NFLID:       "00-0033873",
			Name:        "Patrick Mahomes",
			Team:        "KC",
			Position:    "QB",
			EPAPerPlay:  0.35,
			SuccessRate: 0.58,
			UpdatedAt:   time.Now(),
		},
		models.Player{
			ID:          bson.NewObjectID(),
			NFLID:       "00-0036355",
			Name:        "Justin Jefferson",
			Team:        "MIN",
			Position:    "WR",
			EPAPerPlay:  0.28,
			SuccessRate: 0.52,
			TargetShare: 0.31,
			UpdatedAt:   time.Now(),
		},
		models.Player{
			ID:          bson.NewObjectID(),
			NFLID:       "00-0035704",
			Name:        "Christian McCaffrey",
			Team:        "SF",
			Position:    "RB",
			EPAPerPlay:  0.25,
			SuccessRate: 0.55,
			SnapShare:   0.78,
			UpdatedAt:   time.Now(),
		},
	}

	_, err = db.Collection("players").InsertMany(ctx, players)
	if err != nil {
		log.Printf("Warning: Failed to insert players: %v", err)
	} else {
		log.Println("Sample players loaded successfully!")
	}

	// Load sample game
	log.Println("Loading sample game...")
	game := models.Game{
		ID:        bson.NewObjectID(),
		GameID:    "2024_09_KC_BUF",
		Season:    2024,
		Week:      9,
		HomeTeam:  "KC",
		AwayTeam:  "BUF",
		StartTime: time.Now().Add(24 * time.Hour),
		Status:    "scheduled",
		VegasLine: -3.5,
		OverUnder: 52.5,
		UpdatedAt: time.Now(),
	}

	_, err = db.Collection("games").InsertOne(ctx, game)
	if err != nil {
		log.Printf("Warning: Failed to insert game: %v", err)
	} else {
		log.Println("Sample game loaded successfully!")
	}

	log.Println("Sample data loading complete!")
}
