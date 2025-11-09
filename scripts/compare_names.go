package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	ctx := context.Background()

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = os.Getenv("MONGODB_URI")
	}
	if mongoURI == "" {
		log.Fatal("MONGO_URI not set")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("nfl_platform")

	// Get sample player names from players collection
	fmt.Println("=== Sample names from PLAYERS collection ===")
	cursor, _ := db.Collection("players").Find(ctx, bson.M{"position": "QB"}, options.Find().SetLimit(3))
	var players []struct {
		Name string `bson:"name"`
	}
	cursor.All(ctx, &players)
	for _, p := range players {
		fmt.Printf("  '%s'\n", p.Name)
	}

	// Get sample player names from plays collection
	fmt.Println("\n=== Sample passer names from PLAYS collection ===")
	cursor2, _ := db.Collection("plays").Find(ctx, bson.M{
		"passer_player_name": bson.M{"$exists": true, "$ne": ""},
		"season":             2025,
	}, options.Find().SetLimit(5))
	var plays []struct {
		PasserName string `bson:"passer_player_name"`
	}
	cursor2.All(ctx, &plays)
	for _, p := range plays {
		fmt.Printf("  '%s'\n", p.PasserName)
	}

	// Try a specific known player
	fmt.Println("\n=== Checking specific player: Geno Smith ===")
	count, _ := db.Collection("plays").CountDocuments(ctx, bson.M{
		"passer_player_name": "Geno Smith",
		"season":             2025,
	})
	fmt.Printf("Plays found with passer_player_name='Geno Smith': %d\n", count)

	count2, _ := db.Collection("plays").CountDocuments(ctx, bson.M{
		"passer_player_name": bson.M{"$regex": "Geno", "$options": "i"},
		"season":             2025,
	})
	fmt.Printf("Plays found with passer_player_name matching 'Geno': %d\n", count2)

	// Check total plays
	fmt.Println("\n=== Total plays stats ===")
	total, _ := db.Collection("plays").CountDocuments(ctx, bson.M{})
	fmt.Printf("Total plays in collection: %d\n", total)

	withPasser, _ := db.Collection("plays").CountDocuments(ctx, bson.M{
		"passer_player_name": bson.M{"$exists": true, "$ne": ""},
	})
	fmt.Printf("Plays with passer_player_name: %d\n", withPasser)

	withRusher, _ := db.Collection("plays").CountDocuments(ctx, bson.M{
		"rusher_player_name": bson.M{"$exists": true, "$ne": ""},
	})
	fmt.Printf("Plays with rusher_player_name: %d\n", withRusher)

	withReceiver, _ := db.Collection("plays").CountDocuments(ctx, bson.M{
		"receiver_player_name": bson.M{"$exists": true, "$ne": ""},
	})
	fmt.Printf("Plays with receiver_player_name: %d\n", withReceiver)
}
