package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/parquet-go/parquet-go"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Play struct matching NFLverse play-by-play schema
type NFLVersePlay struct {
	GameID             string  `parquet:"game_id" json:"game_id"`
	PlayID             string  `parquet:"play_id" json:"play_id"`
	Season             int     `parquet:"season" json:"season"`
	Week               int     `parquet:"week" json:"week"`
	Quarter            int     `parquet:"quarter_seconds_remaining" json:"quarter"`
	Down               int     `parquet:"down" json:"down"`
	YardsToGo          int     `parquet:"ydstogo" json:"ydstogo"`
	YardLine           int     `parquet:"yardline_100" json:"yardline_100"`
	Description        string  `parquet:"desc" json:"desc"`
	PlayType           string  `parquet:"play_type" json:"play_type"`
	PossessionTeam     string  `parquet:"posteam" json:"posteam"`
	DefenseTeam        string  `parquet:"defteam" json:"defteam"`
	PasserPlayerID     string  `parquet:"passer_player_id"`
	PasserPlayerName   string  `parquet:"passer_player_name"`
	ReceiverPlayerID   string  `parquet:"receiver_player_id"`
	ReceiverPlayerName string  `parquet:"receiver_player_name"`
	RusherPlayerID     string  `parquet:"rusher_player_id"`
	RusherPlayerName   string  `parquet:"rusher_player_name"`
	Yards              int     `parquet:"yards_gained" json:"yards_gained"`
	Touchdown          int     `parquet:"touchdown" json:"touchdown"`
	Interception       int     `parquet:"interception" json:"interception"`
	Fumble             int     `parquet:"fumble" json:"fumble"`
	Sack               int     `parquet:"sack" json:"sack"`
	EPA                float64 `parquet:"epa" json:"epa"`
	WPA                float64 `parquet:"wpa" json:"wpa"`
	Success            int     `parquet:"success" json:"success"`
	AirYards           int     `parquet:"air_yards" json:"air_yards"`
	YardsAfterCatch    int     `parquet:"yards_after_catch" json:"yards_after_catch"`
}

func main() {
	// Get MongoDB URI from env
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = os.Getenv("MONGODB_URI")
	}
	if uri == "" {
		log.Fatal("MONGO_URI or MONGODB_URI not set")
	}

	// Connect to MongoDB
	ctx := context.Background()
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("nfl_platform")
	collection := db.Collection("plays")

	// Load 2025 season play-by-play data
	season := 2025
	url := fmt.Sprintf("https://github.com/nflverse/nflverse-data/releases/download/pbp/play_by_play_%d.parquet", season)

	fmt.Printf("Downloading play-by-play data for %d season...\n", season)
	fmt.Printf("URL: %s\n", url)

	// Download the parquet file
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("Failed to download: HTTP %d", resp.StatusCode)
	}

	// Save to temp file
	tempFile, err := os.CreateTemp("", "pbp_*.parquet")
	if err != nil {
		log.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	fmt.Println("Downloading...")
	written, err := io.Copy(tempFile, resp.Body)
	if err != nil {
		log.Fatalf("Failed to save file: %v", err)
	}
	fmt.Printf("Downloaded %.2f MB\n", float64(written)/(1024*1024))

	// Reopen for reading
	tempFile.Close()
	fileInfo, err := os.Stat(tempFile.Name())
	if err != nil {
		log.Fatalf("Failed to stat temp file: %v", err)
	}

	file, err := os.Open(tempFile.Name())
	if err != nil {
		log.Fatalf("Failed to open temp file: %v", err)
	}
	defer file.Close()

	fmt.Printf("Reading parquet file (%.2f MB)...\n", float64(fileInfo.Size())/(1024*1024))
	pf, err := parquet.OpenFile(file, fileInfo.Size())
	if err != nil {
		log.Fatalf("Failed to open parquet: %v", err)
	}

	fmt.Printf("Found %d rows\n", pf.NumRows())

	// Read and insert in batches
	batchSize := 5000
	totalInserted := 0
	var batch []interface{}

	reader := parquet.NewReader(pf)
	for {
		var play NFLVersePlay
		err := reader.Read(&play)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading row: %v", err)
			continue
		}

		// Convert to BSON document
		doc := bson.M{
			"game_id":              play.GameID,
			"play_id":              play.PlayID,
			"season":               play.Season,
			"week":                 play.Week,
			"quarter":              play.Quarter,
			"down":                 play.Down,
			"yards_to_go":          play.YardsToGo,
			"yard_line":            play.YardLine,
			"description":          play.Description,
			"play_type":            play.PlayType,
			"possession_team":      play.PossessionTeam,
			"defense_team":         play.DefenseTeam,
			"passer_player_id":     play.PasserPlayerID,
			"passer_player_name":   play.PasserPlayerName,
			"receiver_player_id":   play.ReceiverPlayerID,
			"receiver_player_name": play.ReceiverPlayerName,
			"rusher_player_id":     play.RusherPlayerID,
			"rusher_player_name":   play.RusherPlayerName,
			"yards":                play.Yards,
			"touchdown":            play.Touchdown == 1,
			"interception":         play.Interception == 1,
			"fumble":               play.Fumble == 1,
			"sack":                 play.Sack == 1,
			"epa":                  play.EPA,
			"wpa":                  play.WPA,
			"success_play":         play.Success == 1,
			"air_yards":            play.AirYards,
			"yards_after_catch":    play.YardsAfterCatch,
		}

		batch = append(batch, doc)

		if len(batch) >= batchSize {
			// Insert batch
			insertCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			_, err := collection.InsertMany(insertCtx, batch)
			cancel()

			if err != nil {
				log.Printf("Failed to insert batch: %v", err)
			} else {
				totalInserted += len(batch)
				fmt.Printf("Inserted %d plays (total: %d)\n", len(batch), totalInserted)
			}
			batch = nil
		}
	}

	// Insert remaining plays
	if len(batch) > 0 {
		insertCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		_, err := collection.InsertMany(insertCtx, batch)
		cancel()

		if err != nil {
			log.Printf("Failed to insert final batch: %v", err)
		} else {
			totalInserted += len(batch)
			fmt.Printf("Inserted %d plays (total: %d)\n", len(batch), totalInserted)
		}
	}

	fmt.Printf("\n✅ Successfully loaded %d plays from %d season\n", totalInserted, season)
	fmt.Println("\nNow creating indexes for fast queries...")

	// Create indexes for fast EPA queries
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "season", Value: 1},
				{Key: "week", Value: 1},
				{Key: "passer_player_id", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "season", Value: 1},
				{Key: "week", Value: 1},
				{Key: "rusher_player_id", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "season", Value: 1},
				{Key: "week", Value: 1},
				{Key: "receiver_player_id", Value: 1},
			},
		},
	}

	indexCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	for i, idx := range indexes {
		_, err := collection.Indexes().CreateOne(indexCtx, idx)
		if err != nil {
			log.Printf("Failed to create index %d: %v", i+1, err)
		} else {
			fmt.Printf("Created index %d\n", i+1)
		}
	}

	fmt.Println("\n✅ All done! EPA data is ready to use.")
}
