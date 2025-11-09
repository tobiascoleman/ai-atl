package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI not set in .env")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("nfl_platform")

	log.Println("🏗️  Creating MongoDB indexes for performance...")

	// PLAYERS COLLECTION INDEXES
	playersCollection := db.Collection("players")

	// Index for name lookups and sorting
	_, err = playersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "name", Value: 1}},
	})
	if err != nil {
		log.Printf("❌ Failed to create name index: %v", err)
	} else {
		log.Println("✅ Created index on players.name")
	}

	// Index for nfl_id lookups (very common)
	_, err = playersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "nfl_id", Value: 1}},
	})
	if err != nil {
		log.Printf("❌ Failed to create nfl_id index: %v", err)
	} else {
		log.Println("✅ Created index on players.nfl_id")
	}

	// Index for team filtering
	_, err = playersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "team", Value: 1}},
	})
	if err != nil {
		log.Printf("❌ Failed to create team index: %v", err)
	} else {
		log.Println("✅ Created index on players.team")
	}

	// Index for position filtering
	_, err = playersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "position", Value: 1}},
	})
	if err != nil {
		log.Printf("❌ Failed to create position index: %v", err)
	} else {
		log.Println("✅ Created index on players.position")
	}

	// Compound index for team + position filtering with name sort
	_, err = playersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "team", Value: 1},
			{Key: "position", Value: 1},
			{Key: "name", Value: 1},
		},
	})
	if err != nil {
		log.Printf("❌ Failed to create compound team+position+name index: %v", err)
	} else {
		log.Println("✅ Created compound index on players (team, position, name)")
	}

	// Index for season filtering (for most recent data)
	_, err = playersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "season", Value: -1}},
	})
	if err != nil {
		log.Printf("❌ Failed to create season index: %v", err)
	} else {
		log.Println("✅ Created index on players.season")
	}

	// PLAYER_STATS COLLECTION INDEXES
	playerStatsCollection := db.Collection("player_stats")

	// Index for nfl_id lookups (primary query field)
	_, err = playerStatsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "nfl_id", Value: 1}},
	})
	if err != nil {
		log.Printf("❌ Failed to create nfl_id index on player_stats: %v", err)
	} else {
		log.Println("✅ Created index on player_stats.nfl_id")
	}

	// Index for season lookups
	_, err = playerStatsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "season", Value: -1}},
	})
	if err != nil {
		log.Printf("❌ Failed to create season index on player_stats: %v", err)
	} else {
		log.Println("✅ Created index on player_stats.season")
	}

	// Compound index for nfl_id + season (most common query)
	_, err = playerStatsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "nfl_id", Value: 1},
			{Key: "season", Value: -1},
		},
	})
	if err != nil {
		log.Printf("❌ Failed to create compound nfl_id+season index: %v", err)
	} else {
		log.Println("✅ Created compound index on player_stats (nfl_id, season)")
	}

	// Index for EPA sorting
	_, err = playerStatsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "epa", Value: -1}},
	})
	if err != nil {
		log.Printf("❌ Failed to create epa index: %v", err)
	} else {
		log.Println("✅ Created index on player_stats.epa")
	}

	// PLAY_BY_PLAY COLLECTION INDEXES
	playsCollection := db.Collection("plays")

	// Index for game_id lookups
	_, err = playsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "game_id", Value: 1}},
	})
	if err != nil {
		log.Printf("❌ Failed to create game_id index: %v", err)
	} else {
		log.Println("✅ Created index on plays.game_id")
	}

	// Index for season lookups
	_, err = playsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "season", Value: -1}},
	})
	if err != nil {
		log.Printf("❌ Failed to create season index on plays: %v", err)
	} else {
		log.Println("✅ Created index on plays.season")
	}

	// GAMES/SCHEDULES COLLECTION INDEXES
	gamesCollection := db.Collection("games")

	// Index for season lookups
	_, err = gamesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "season", Value: -1}},
	})
	if err != nil {
		log.Printf("❌ Failed to create season index on games: %v", err)
	} else {
		log.Println("✅ Created index on games.season")
	}

	// Index for gameday (date) sorting
	_, err = gamesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "gameday", Value: -1}},
	})
	if err != nil {
		log.Printf("❌ Failed to create gameday index: %v", err)
	} else {
		log.Println("✅ Created index on games.gameday")
	}

	// Compound index for team queries
	_, err = gamesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "home_team", Value: 1},
			{Key: "season", Value: -1},
		},
	})
	if err != nil {
		log.Printf("❌ Failed to create home_team+season index: %v", err)
	} else {
		log.Println("✅ Created compound index on games (home_team, season)")
	}

	_, err = gamesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "away_team", Value: 1},
			{Key: "season", Value: -1},
		},
	})
	if err != nil {
		log.Printf("❌ Failed to create away_team+season index: %v", err)
	} else {
		log.Println("✅ Created compound index on games (away_team, season)")
	}

	// NEXT_GEN_STATS COLLECTION INDEXES
	ngsCollection := db.Collection("next_gen_stats")

	// Index for player_id lookups
	_, err = ngsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "player_id", Value: 1}},
	})
	if err != nil {
		log.Printf("❌ Failed to create player_id index on NGS: %v", err)
	} else {
		log.Println("✅ Created index on next_gen_stats.player_id")
	}

	// Compound index for player_id + season + stat_type
	_, err = ngsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "player_id", Value: 1},
			{Key: "season", Value: -1},
			{Key: "stat_type", Value: 1},
		},
	})
	if err != nil {
		log.Printf("❌ Failed to create compound NGS index: %v", err)
	} else {
		log.Println("✅ Created compound index on next_gen_stats (player_id, season, stat_type)")
	}

	// USERS COLLECTION INDEXES (for auth)
	usersCollection := db.Collection("users")

	// Unique index for email (auth lookup)
	_, err = usersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("❌ Failed to create email index: %v", err)
	} else {
		log.Println("✅ Created unique index on users.email")
	}

	log.Println("\n🎉 Index creation complete!")
	log.Println("💡 Query performance should now be MUCH faster!")
}
