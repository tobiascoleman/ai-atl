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
	if uri == "" { uri = os.Getenv("MONGODB_URI") }
	if uri == "" { log.Fatal("MONGO_URI not set") }
	client, _ := mongo.Connect(options.Client().ApplyURI(uri))
	defer client.Disconnect(context.Background())
	db := client.Database("nfl_platform")
	result, err := db.Collection("plays").DeleteMany(context.Background(), bson.M{})
	if err != nil { log.Fatal(err) }
	fmt.Printf("Deleted %d plays\n", result.DeletedCount)
}
