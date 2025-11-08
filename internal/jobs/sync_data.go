package jobs

import (
	"context"
	"log"
	"time"

	"github.com/ai-atl/nfl-platform/pkg/nflverse"
	"go.mongodb.org/mongo-driver/mongo"
)

// SyncNFLverseData syncs data from NFLverse to MongoDB
func SyncNFLverseData(ctx context.Context, db *mongo.Database) error {
	log.Println("Starting NFLverse data sync...")
	
	client := nflverse.NewClient()
	currentSeason := time.Now().Year()
	
	// Sync player stats
	log.Printf("Fetching player stats for season %d", currentSeason)
	playerData, err := client.FetchPlayerStats(ctx, currentSeason)
	if err != nil {
		log.Printf("Warning: Failed to fetch player stats: %v", err)
	} else {
		log.Printf("Fetched %d bytes of player data", len(playerData))
		// TODO: Parse and insert into MongoDB
	}
	
	// Sync rosters
	log.Printf("Fetching rosters for season %d", currentSeason)
	rosterData, err := client.FetchRosters(ctx, currentSeason)
	if err != nil {
		log.Printf("Warning: Failed to fetch rosters: %v", err)
	} else {
		log.Printf("Fetched %d bytes of roster data", len(rosterData))
		// TODO: Parse and insert into MongoDB
	}
	
	// Sync injuries
	log.Printf("Fetching injuries for season %d", currentSeason)
	injuryData, err := client.FetchInjuries(ctx, currentSeason)
	if err != nil {
		log.Printf("Warning: Failed to fetch injuries: %v", err)
	} else {
		log.Printf("Fetched %d bytes of injury data", len(injuryData))
		// TODO: Parse and insert into MongoDB
	}
	
	log.Println("NFLverse data sync completed")
	return nil
}

// SchedulePeriodicSync sets up periodic data syncing
func SchedulePeriodicSync(db *mongo.Database, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		if err := SyncNFLverseData(ctx, db); err != nil {
			log.Printf("Sync error: %v", err)
		}
		cancel()
	}
}

