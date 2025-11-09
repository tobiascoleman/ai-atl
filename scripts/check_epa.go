package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = os.Getenv("MONGODB_URI")
	}
	if uri == "" {
		log.Fatal("MONGO_URI or MONGODB_URI not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database("nfl_platform")

	// Check plays collection
	playsCount, _ := db.Collection("plays").CountDocuments(ctx, bson.M{})
	fmt.Printf("Plays collection: %d documents\n", playsCount)

	// Check if plays have EPA data
	var samplePlay bson.M
	err = db.Collection("plays").FindOne(ctx, bson.M{"season": 2024, "week": 10}).Decode(&samplePlay)
	if err == nil {
		fmt.Printf("Sample play EPA: %v\n", samplePlay["epa"])
		fmt.Printf("Sample play has passer: %v\n", samplePlay["passer_player_id"])
	} else {
		fmt.Printf("No plays found for 2024 week 10: %v\n", err)
	}

	// Test a specific player EPA query with timing
	playerID := "00-0033873" // Example player ID
	start := time.Now()

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"season":           2024,
			"week":             bson.M{"$gte": 6, "$lte": 10},
			"passer_player_id": playerID,
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":        nil,
			"total_epa":  bson.M{"$sum": "$epa"},
			"play_count": bson.M{"$sum": 1},
		}}},
	}

	cursor, err := db.Collection("plays").Aggregate(ctx, pipeline)
	if err != nil {
		fmt.Printf("Query error: %v\n", err)
	} else {
		defer cursor.Close(ctx)
		if cursor.Next(ctx) {
			var result bson.M
			cursor.Decode(&result)
			fmt.Printf("Query took: %v\n", time.Since(start))
			fmt.Printf("Result: %+v\n", result)
		} else {
			fmt.Printf("No results for player %s\n", playerID)
		}
	}
}
