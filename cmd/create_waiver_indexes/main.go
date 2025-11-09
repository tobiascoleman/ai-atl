package main

import (
	"context"
	"log"
	"time"

	"github.com/ai-atl/nfl-platform/internal/config"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Load config from .env
	cfg := config.Load()
	uri := cfg.MongoURI

	log.Println("Connecting to MongoDB...")
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("nfl_data")
	playsCollection := db.Collection("plays")

	log.Println("Creating indexes for waiver wire performance optimization...")
	log.Println("This may take several minutes on a large collection...")

	// Index for QB queries (passer_player_id)
	log.Println("\n1. Creating index: season_week_passer...")
	_, err = playsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "season", Value: 1},
			{Key: "week", Value: 1},
			{Key: "passer_player_id", Value: 1},
		},
		Options: options.Index().SetName("season_week_passer"),
	})
	if err != nil {
		log.Printf("Warning: %v (may already exist)\n", err)
	} else {
		log.Println("✓ Created season_week_passer index")
	}

	// Index for RB queries (rusher_player_id)
	log.Println("\n2. Creating index: season_week_rusher...")
	_, err = playsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "season", Value: 1},
			{Key: "week", Value: 1},
			{Key: "rusher_player_id", Value: 1},
		},
		Options: options.Index().SetName("season_week_rusher"),
	})
	if err != nil {
		log.Printf("Warning: %v (may already exist)\n", err)
	} else {
		log.Println("✓ Created season_week_rusher index")
	}

	// Index for WR/TE queries (receiver_player_id)
	log.Println("\n3. Creating index: season_week_receiver...")
	_, err = playsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "season", Value: 1},
			{Key: "week", Value: 1},
			{Key: "receiver_player_id", Value: 1},
		},
		Options: options.Index().SetName("season_week_receiver"),
	})
	if err != nil {
		log.Printf("Warning: %v (may already exist)\n", err)
	} else {
		log.Println("✓ Created season_week_receiver index")
	}

	// Index for defensive matchup queries
	log.Println("\n4. Creating index: season_defense_team...")
	_, err = playsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "season", Value: 1},
			{Key: "defense_team", Value: 1},
		},
		Options: options.Index().SetName("season_defense_team"),
	})
	if err != nil {
		log.Printf("Warning: %v (may already exist)\n", err)
	} else {
		log.Println("✓ Created season_defense_team index")
	}

	// Index for EPA calculations
	log.Println("\n5. Creating index: season_week_epa...")
	_, err = playsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "season", Value: 1},
			{Key: "week", Value: 1},
			{Key: "epa", Value: 1},
		},
		Options: options.Index().SetName("season_week_epa"),
	})
	if err != nil {
		log.Printf("Warning: %v (may already exist)\n", err)
	} else {
		log.Println("✓ Created season_week_epa index")
	}

	log.Println("\n✅ Index creation complete!")
	log.Println("Waiver wire queries should now be much faster.")
	log.Println("You can now re-enable the full analysis in waiver_wire.go")
}
