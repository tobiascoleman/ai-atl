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
	uri := os.Getenv("MONGO_URI")
	if uri == "" { log.Fatal("MONGO_URI not set") }
	client, _ := mongo.Connect(options.Client().ApplyURI(uri))
	defer client.Disconnect(context.Background())
	db := client.Database("nfl_platform")
	
	// Check total 2025 plays
	count2025, _ := db.Collection("plays").CountDocuments(context.Background(), bson.M{"season": 2025})
	fmt.Printf("Total 2025 plays: %d\n", count2025)
	
	// Sample a play to see player ID format
	var sample bson.M
	db.Collection("plays").FindOne(context.Background(), bson.M{
		"season": 2025,
		"week": 10,
		"passer_player_id": bson.M{"$ne": ""},
	}).Decode(&sample)
	fmt.Printf("Sample passer_player_id: %v\n", sample["passer_player_id"])
	
	// Sample a player to see their NFLID format
	var player bson.M
	db.Collection("players").FindOne(context.Background(), bson.M{"season": 2025}).Decode(&player)
	fmt.Printf("Sample player nfl_id: %v\n", player["nfl_id"])
	fmt.Printf("Sample player name: %v\n", player["name"])
}
