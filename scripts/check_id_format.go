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
	
	var play bson.M
	db.Collection("plays").FindOne(context.Background(), bson.M{
		"season": 2025,
		"passer_player_id": bson.M{"$ne": "", "$exists": true},
	}).Decode(&play)
	fmt.Printf("Plays passer_player_id: '%v'\n", play["passer_player_id"])
	
	var player bson.M
	db.Collection("players").FindOne(context.Background(), bson.M{"season": 2025, "position": "QB"}).Decode(&player)
	fmt.Printf("Players nfl_id: '%v'\n", player["nfl_id"])
	fmt.Printf("Players name: '%v'\n", player["name"])
}
