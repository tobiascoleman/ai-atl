package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ai-atl/nfl-platform/internal/config"
	"github.com/ai-atl/nfl-platform/internal/models"
	"github.com/ai-atl/nfl-platform/internal/parquet"
	"github.com/ai-atl/nfl-platform/pkg/mongodb"
	"github.com/apache/arrow/go/v14/arrow"
	"github.com/apache/arrow/go/v14/arrow/array"
	"github.com/apache/arrow/go/v14/arrow/memory"
	"github.com/apache/arrow/go/v14/parquet/file"
	"github.com/apache/arrow/go/v14/parquet/pqarrow"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	nflverseBaseURL = "https://github.com/nflverse/nflverse-data/releases/download"
	cacheDir        = "./nflverse_cache"
)

// Correct URL patterns - verified from nflverse-data releases
var dataURLs = map[string]string{
	// Core data
	"schedules": nflverseBaseURL + "/schedules/games.parquet",
	"teams":     nflverseBaseURL + "/teams/teams_colors_logos.parquet",
	"players":   nflverseBaseURL + "/players/players.parquet",
	"officials": nflverseBaseURL + "/officials/officials.parquet",

	// Play-by-play (1999-2024)
	"pbp":               nflverseBaseURL + "/pbp/play_by_play_%d.parquet",
	"pbp_participation": nflverseBaseURL + "/pbp_participation/pbp_participation_%d.parquet",

	// Rosters
	"roster_yearly": nflverseBaseURL + "/rosters/roster_%d.parquet",
	"roster_weekly": nflverseBaseURL + "/weekly_rosters/roster_weekly_%d.parquet",

	// Player stats (2017+)
	"player_stats_regpost": nflverseBaseURL + "/stats_player/stats_player_regpost_%d.parquet",
	"player_stats_weekly":  nflverseBaseURL + "/stats_player/stats_player_week_%d.parquet",

	// Team stats (1999+) - multiple types
	"team_stats_post":    nflverseBaseURL + "/stats_team/stats_team_post_%d.parquet",
	"team_stats_reg":     nflverseBaseURL + "/stats_team/stats_team_reg_%d.parquet",
	"team_stats_postreg": nflverseBaseURL + "/stats_team/stats_team_postreg_%d.parquet",
	"team_stats_week":    nflverseBaseURL + "/stats_team/stats_team_week_%d.parquet",

	// QB stats (ESPN)
	"qbr_week":   nflverseBaseURL + "/espn_data/qbr_week_level.parquet",
	"qbr_season": nflverseBaseURL + "/espn_data/qbr_season_level.parquet",

	// Injuries (yearly) & Next Gen Stats (all years in one file)
	"injuries":      nflverseBaseURL + "/injuries/injuries_%d.parquet",
	"ngs_passing":   nflverseBaseURL + "/nextgen_stats/ngs_passing.parquet",
	"ngs_rushing":   nflverseBaseURL + "/nextgen_stats/ngs_rushing.parquet",
	"ngs_receiving": nflverseBaseURL + "/nextgen_stats/ngs_receiving.parquet",
}

type DataLoader struct {
	db         *mongo.Database
	httpClient *http.Client
	mu         sync.Mutex
	stats      LoadStats
}

type LoadStats struct {
	TotalFiles    int
	Downloaded    int
	Processed     int
	Errors        int
	PlayersLoaded int
	GamesLoaded   int
	PlaysLoaded   int
	NGSLoaded     int
	StartTime     time.Time
}

func main() {
	fmt.Println("=== NFLverse Maximum Data Loader ===")
	fmt.Println("Loading ALL available data (1999-2025)")
	fmt.Println("This will take approximately 30-60 minutes")
	fmt.Println()

	// Load environment
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	cfg := config.Load()

	// Connect to MongoDB
	ctx := context.Background()
	client, err := mongodb.Connect(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to connect to MongoDB: %w", err))
	}
	defer client.Disconnect(ctx)

	db := client.Database("nfl_platform")
	log.Println("✓ Connected to MongoDB (database: nfl_platform)")

	// Create cache directory
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Fatalf("Failed to create cache directory: %v", err)
	}

	// Initialize loader
	loader := &DataLoader{
		db: db,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
		stats: LoadStats{
			StartTime: time.Now(),
		},
	}

	// Start loading
	loader.LoadAll(ctx)

	// Print final stats
	loader.PrintFinalStats()
}

func (l *DataLoader) LoadAll(ctx context.Context) {
	fmt.Println("\n📊 Phase 1: Loading Schedules & Teams")
	fmt.Println(strings.Repeat("=", 50))
	//l.LoadSchedules(ctx)
	//l.LoadTeams(ctx)

	fmt.Println("\n📊 Phase 2: Loading Rosters (2020-2025)")
	fmt.Println(strings.Repeat("=", 50))
	//l.LoadRosters(ctx, 2020, 2025)

	fmt.Println("\n📊 Phase 3: Loading Weekly Rosters for Injury Status (2024-2025)")
	fmt.Println(strings.Repeat("=", 50))
	//l.LoadWeeklyRosters(ctx, 2024, 2025)

	fmt.Println("\n📊 Phase 4: Loading Player Stats (2020-2025)")
	fmt.Println(strings.Repeat("=", 50))
	//l.LoadPlayerStats(ctx, 2020, 2025)

	fmt.Println("\n📊 Phase 4.5: Loading Weekly Player Stats (2020-2025)")
	fmt.Println(strings.Repeat("=", 50))
	//l.LoadWeeklyStats(ctx, 2020, 2025)

	fmt.Println("\n📊 Phase 5: Loading Play-by-Play Data (ALL 27 SEASONS!) 🏈")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("This is the biggest dataset - will take 15-20 minutes")
	l.LoadPlayByPlay(ctx, 1999, 2025)

	fmt.Println("\n📊 Phase 6: Loading Next Gen Stats (All Seasons)")
	fmt.Println(strings.Repeat("=", 50))
	//l.LoadNextGenStats(ctx, 2020, 2025)

	fmt.Println("\n✅ All data loaded!")
}

func (l *DataLoader) LoadSchedules(ctx context.Context) {
	fmt.Println("→ Downloading schedules (games.parquet)...")

	url := dataURLs["schedules"]
	data, err := l.downloadFile(url, "games.parquet")
	if err != nil {
		log.Printf("❌ Failed to download schedules: %v", err)
		l.stats.Errors++
		return
	}

	fmt.Println("→ Parsing schedules...")
	games := l.parseSchedules(data)

	fmt.Printf("→ Inserting %d games into MongoDB...\n", len(games))
	inserted := l.insertGames(ctx, games)
	l.stats.GamesLoaded += inserted

	fmt.Printf("✓ Loaded %d games\n", inserted)
}

func (l *DataLoader) LoadTeams(ctx context.Context) {
	fmt.Println("→ Downloading teams...")

	url := dataURLs["teams"]
	_, err := l.downloadFile(url, "teams.parquet")
	if err != nil {
		log.Printf("❌ Failed to download teams: %v", err)
		l.stats.Errors++
		return
	}

	fmt.Println("✓ Teams data cached (use for UI logos/colors)")
}

func (l *DataLoader) LoadRosters(ctx context.Context, startYear, endYear int) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // Limit concurrent downloads

	for year := startYear; year <= endYear; year++ {
		wg.Add(1)
		go func(y int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			l.loadRosterYear(ctx, y)
		}(year)
	}

	wg.Wait()
}

func (l *DataLoader) loadRosterYear(ctx context.Context, year int) {
	fmt.Printf("→ Loading rosters %d...\n", year)

	url := fmt.Sprintf(dataURLs["roster_yearly"], year)
	data, err := l.downloadFile(url, fmt.Sprintf("roster_%d.parquet", year))
	if err != nil {
		log.Printf("❌ Failed to download roster %d: %v", year, err)
		l.mu.Lock()
		l.stats.Errors++
		l.mu.Unlock()
		return
	}

	players := l.parseRoster(data, year)
	inserted := l.insertPlayers(ctx, players)

	l.mu.Lock()
	l.stats.PlayersLoaded += inserted
	l.mu.Unlock()

	fmt.Printf("✓ Loaded %d players from %d\n", inserted, year)
}

func (l *DataLoader) LoadWeeklyRosters(ctx context.Context, startYear, endYear int) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // Limit concurrent downloads

	for year := startYear; year <= endYear; year++ {
		wg.Add(1)
		go func(y int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			l.loadWeeklyRosterYear(ctx, y)
		}(year)
	}

	wg.Wait()
}

func (l *DataLoader) loadWeeklyRosterYear(ctx context.Context, year int) {
	fmt.Printf("→ Loading weekly rosters %d (injury status)...\n", year)

	url := fmt.Sprintf(dataURLs["roster_weekly"], year)
	data, err := l.downloadFile(url, fmt.Sprintf("roster_weekly_%d.parquet", year))
	if err != nil {
		log.Printf("❌ Failed to download weekly roster %d: %v", year, err)
		l.mu.Lock()
		l.stats.Errors++
		l.mu.Unlock()
		return
	}

	// Parse weekly rosters which include injury status
	weeklyRosters := l.parseWeeklyRoster(data, year)
	fmt.Printf("  📦 Parsed %d weekly roster entries\n", len(weeklyRosters))

	// Update players with injury status
	updated := l.updatePlayerInjuryStatus(ctx, weeklyRosters)

	l.mu.Lock()
	l.stats.PlayersLoaded += updated
	l.mu.Unlock()

	fmt.Printf("✓ Updated %d players with injury status from %d\n", updated, year)
}

func (l *DataLoader) LoadPlayerStats(ctx context.Context, startYear, endYear int) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5)

	for year := startYear; year <= endYear; year++ {
		wg.Add(1)
		go func(y int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			l.loadPlayerStatsYear(ctx, y)
		}(year)
	}

	wg.Wait()
}

func (l *DataLoader) loadPlayerStatsYear(ctx context.Context, year int) {
	// Player stats only available from 2017+
	if year < 2017 {
		return
	}

	fmt.Printf("→ Loading player stats %d...\n", year)

	url := fmt.Sprintf(dataURLs["player_stats_regpost"], year)
	data, err := l.downloadFile(url, fmt.Sprintf("player_stats_regpost_%d.parquet", year))
	if err != nil {
		log.Printf("❌ Failed to download player stats %d: %v", year, err)
		l.mu.Lock()
		l.stats.Errors++
		l.mu.Unlock()
		return
	}

	// Parse the stats
	stats := l.parsePlayerStats(data, year, "REGPOST") // Regular + Post season combined
	inserted := l.insertPlayerStats(ctx, stats)

	l.mu.Lock()
	l.stats.PlayersLoaded += inserted // Reuse counter for stats
	l.mu.Unlock()

	fmt.Printf("✓ Loaded %d player stats from %d\n", inserted, year)
}

func (l *DataLoader) LoadWeeklyStats(ctx context.Context, startYear, endYear int) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5)

	for year := startYear; year <= endYear; year++ {
		wg.Add(1)
		go func(y int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			l.loadWeeklyStatsYear(ctx, y)
		}(year)
	}

	wg.Wait()
}

func (l *DataLoader) loadWeeklyStatsYear(ctx context.Context, year int) {
	// Weekly stats only available from 2017+
	if year < 2017 {
		return
	}

	fmt.Printf("→ Loading weekly stats %d...\n", year)

	url := fmt.Sprintf(dataURLs["player_stats_weekly"], year)
	data, err := l.downloadFile(url, fmt.Sprintf("player_stats_weekly_%d.parquet", year))
	if err != nil {
		log.Printf("❌ Failed to download weekly stats %d: %v", year, err)
		l.mu.Lock()
		l.stats.Errors++
		l.mu.Unlock()
		return
	}

	// Parse the weekly stats
	weeklyStats := l.parseWeeklyStats(data, year)
	inserted := l.insertWeeklyStats(ctx, weeklyStats)

	l.mu.Lock()
	l.stats.PlayersLoaded += inserted // Reuse counter
	l.mu.Unlock()

	fmt.Printf("✓ Loaded %d weekly stat records from %d\n", inserted, year)
}

func (l *DataLoader) LoadPlayByPlay(ctx context.Context, startYear, endYear int) {
	fmt.Printf("\n🏈 Loading %d seasons of play-by-play data\n", endYear-startYear+1)
	fmt.Println("This is ~1 million plays - progress will be shown every 5 years")

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 3) // Limit to 3 concurrent PBP downloads (large files)

	for year := startYear; year <= endYear; year++ {
		wg.Add(1)
		go func(y int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			l.loadPlayByPlayYear(ctx, y)
		}(year)
	}

	wg.Wait()
}

func (l *DataLoader) loadPlayByPlayYear(ctx context.Context, year int) {
	fmt.Printf("→ Loading play-by-play %d...\n", year)

	url := fmt.Sprintf(dataURLs["pbp"], year)
	data, err := l.downloadFile(url, fmt.Sprintf("pbp_%d.parquet", year))
	if err != nil {
		log.Printf("❌ Failed to download PBP %d: %v", year, err)
		l.mu.Lock()
		l.stats.Errors++
		l.mu.Unlock()
		return
	}

	plays := l.parsePlayByPlay(data, year)
	inserted := l.insertPlays(ctx, plays)

	l.mu.Lock()
	l.stats.PlaysLoaded += inserted
	l.mu.Unlock()

	fmt.Printf("✓ Loaded %d plays from %d (Total: %d plays)\n", inserted, year, l.stats.PlaysLoaded)
}

func (l *DataLoader) LoadInjuries(ctx context.Context, startYear, endYear int) {
	for year := startYear; year <= endYear; year++ {
		fmt.Printf("→ Loading injuries %d...\n", year)

		url := fmt.Sprintf(dataURLs["injuries"], year)
		_, err := l.downloadFile(url, fmt.Sprintf("injuries_%d.parquet", year))
		if err != nil {
			log.Printf("⚠ Injuries %d not available: %v", year, err)
			continue
		}

		fmt.Printf("✓ Cached injuries %d\n", year)
	}
}

func (l *DataLoader) LoadNextGenStats(ctx context.Context, startYear, endYear int) {
	// NGS files contain ALL years in a single file (not per-year)
	statTypes := map[string]string{
		"passing":   "ngs_passing",
		"rushing":   "ngs_rushing",
		"receiving": "ngs_receiving",
	}

	for statName, urlKey := range statTypes {
		fmt.Printf("→ Loading NGS %s (all seasons)...\n", statName)

		url := dataURLs[urlKey]
		data, err := l.downloadFile(url, fmt.Sprintf("ngs_%s.parquet", statName))
		if err != nil {
			log.Printf("⚠ NGS %s not available: %v", statName, err)
			l.mu.Lock()
			l.stats.Errors++
			l.mu.Unlock()
			continue
		}

		// Parse the NGS stats
		stats, err := parquet.ParseNextGenStats(data, statName)
		if err != nil {
			log.Printf("⚠ Failed to parse NGS %s: %v", statName, err)
			l.mu.Lock()
			l.stats.Errors++
			l.mu.Unlock()
			continue
		}
		if len(stats) == 0 {
			log.Printf("⚠ No NGS %s stats parsed", statName)
			continue
		}

		// Insert into MongoDB
		inserted := l.insertNGSStats(ctx, stats)

		l.mu.Lock()
		l.stats.NGSLoaded += inserted
		l.mu.Unlock()

		fmt.Printf("✓ Loaded %d NGS %s stats (all years)\n", inserted, statName)
	}
}

func (l *DataLoader) insertNGSStats(ctx context.Context, stats []models.NextGenStat) int {
	if len(stats) == 0 {
		return 0
	}

	collection := l.db.Collection("next_gen_stats")

	// Upsert stats with compound key (player_id + season + week + stat_type)
	inserted := 0
	for _, stat := range stats {
		filter := bson.M{
			"player_id": stat.PlayerID,
			"season":    stat.Season,
			"week":      stat.Week,
			"stat_type": stat.StatType,
		}
		update := bson.M{"$set": stat}

		opts := options.UpdateOne().SetUpsert(true)
		_, err := collection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			log.Printf("Error upserting NGS stat: %v", err)
			continue
		}
		inserted++
	}

	return inserted
}

// Helper functions

func (l *DataLoader) downloadFile(url, filename string) ([]byte, error) {
	cachePath := fmt.Sprintf("%s/%s", cacheDir, filename)

	// Check cache first
	if data, err := os.ReadFile(cachePath); err == nil {
		l.mu.Lock()
		l.stats.Downloaded++
		l.mu.Unlock()
		return data, nil
	}

	// Download
	resp, err := l.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Cache it
	os.WriteFile(cachePath, data, 0644)

	l.mu.Lock()
	l.stats.Downloaded++
	l.mu.Unlock()

	return data, nil
}

// Real Parquet parsers using Apache Arrow

func (l *DataLoader) parseSchedules(data []byte) []models.Game {
	games, err := parquet.ParseSchedules(data)
	if err != nil {
		log.Printf("Error parsing schedules: %v", err)
		return []models.Game{}
	}
	return games
}

func (l *DataLoader) parseRoster(data []byte, year int) []models.Player {
	players, err := parquet.ParseRoster(data, year)
	if err != nil {
		log.Printf("Error parsing roster %d: %v", year, err)
		return []models.Player{}
	}
	return players
}

func (l *DataLoader) parsePlayerStats(data []byte, year int, seasonType string) []models.PlayerStats {
	stats, err := parquet.ParsePlayerStats(data, year, seasonType)
	if err != nil {
		log.Printf("Error parsing player stats %d: %v", year, err)
		return []models.PlayerStats{}
	}
	return stats
}

func (l *DataLoader) parseWeeklyStats(data []byte, year int) []models.WeeklyStat {
	weeklyStats, err := parquet.ParseWeeklyStats(data, year)
	if err != nil {
		log.Printf("Error parsing weekly stats %d: %v", year, err)
		return []models.WeeklyStat{}
	}
	return weeklyStats
}

func (l *DataLoader) parsePlayByPlay(data []byte, year int) []models.Play {
	plays, err := parquet.ParsePlayByPlay(data, year)
	if err != nil {
		log.Printf("Error parsing play-by-play %d: %v", year, err)
		return []models.Play{}
	}
	return plays
}

func (l *DataLoader) insertGames(ctx context.Context, games []models.Game) int {
	if len(games) == 0 {
		return 0
	}

	collection := l.db.Collection("games")

	// Batch insert
	docs := make([]interface{}, len(games))
	for i, game := range games {
		docs[i] = game
	}

	opts := options.InsertMany().SetOrdered(false) // Continue on duplicates
	result, err := collection.InsertMany(ctx, docs, opts)
	if err != nil {
		// Check if it's a bulk write exception (duplicate keys are expected)
		if bulkErr, ok := err.(mongo.BulkWriteException); ok {
			// Return the number of successfully inserted documents
			// In v2, InsertedIDs will contain the successful inserts
			if result != nil && len(result.InsertedIDs) > 0 {
				return len(result.InsertedIDs)
			}
			// Calculate from error
			return len(games) - len(bulkErr.WriteErrors)
		}
		log.Printf("Error inserting games: %v", err)
		return 0
	}

	return len(result.InsertedIDs)
}

func (l *DataLoader) insertPlayers(ctx context.Context, players []models.Player) int {
	if len(players) == 0 {
		return 0
	}

	collection := l.db.Collection("players")

	// Upsert players with compound key (nfl_id + season)
	// This allows tracking player movement across seasons
	inserted := 0
	for _, player := range players {
		filter := bson.M{
			"nfl_id": player.NFLID,
			"season": player.Season,
		}
		update := bson.M{"$set": player}

		opts := options.UpdateOne().SetUpsert(true)
		_, err := collection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			log.Printf("Error upserting player: %v", err)
			continue
		}
		inserted++
	}

	return inserted
}

func (l *DataLoader) parseWeeklyRoster(data []byte, season int) []models.WeeklyRosterEntry {
	// Parse weekly roster from Parquet using the parquet parser
	// Weekly rosters include status and status_description_abbr columns
	reader, err := file.NewParquetReader(bytes.NewReader(data))
	if err != nil {
		log.Printf("Failed to create parquet reader: %v", err)
		return nil
	}
	defer reader.Close()

	arrowReader, err := pqarrow.NewFileReader(reader, pqarrow.ArrowReadProperties{}, memory.DefaultAllocator)
	if err != nil {
		log.Printf("Failed to create arrow reader: %v", err)
		return nil
	}

	table, err := arrowReader.ReadTable(context.Background())
	if err != nil {
		log.Printf("Failed to read table: %v", err)
		return nil
	}
	defer table.Release()

	numRows := int(table.NumRows())
	entries := make([]models.WeeklyRosterEntry, 0, numRows)

	// Create column map
	schema := table.Schema()
	colMap := make(map[string]int)
	for i, field := range schema.Fields() {
		colMap[field.Name] = i
	}

	getChunkAndOffset := func(col *arrow.Column, rowIdx int) (arrow.Array, int) {
		offset := rowIdx
		for _, chunk := range col.Data().Chunks() {
			if offset < chunk.Len() {
				return chunk, offset
			}
			offset -= chunk.Len()
		}
		return nil, 0
	}

	getString := func(colName string, rowIdx int) string {
		if colIdx, ok := colMap[colName]; ok {
			col := table.Column(colIdx)
			chunk, offset := getChunkAndOffset(col, rowIdx)
			if chunk != nil {
				if arr, ok := chunk.(*array.String); ok && !arr.IsNull(offset) {
					return arr.Value(offset)
				}
			}
		}
		return ""
	}

	getInt := func(colName string, rowIdx int) int {
		if colIdx, ok := colMap[colName]; ok {
			col := table.Column(colIdx)
			chunk, offset := getChunkAndOffset(col, rowIdx)
			if chunk != nil {
				switch arr := chunk.(type) {
				case *array.Int64:
					if !arr.IsNull(offset) {
						return int(arr.Value(offset))
					}
				case *array.Int32:
					if !arr.IsNull(offset) {
						return int(arr.Value(offset))
					}
				}
			}
		}
		return 0
	}

	for i := 0; i < numRows; i++ {
		entry := models.WeeklyRosterEntry{
			NFLID:                 getString("gsis_id", i), // Use gsis_id, not player_id!
			Season:                season,
			Week:                  getInt("week", i),
			Team:                  getString("team", i),
			Status:                getString("status", i),
			StatusDescriptionAbbr: getString("status_description_abbr", i),
		}

		if entry.NFLID != "" {
			entries = append(entries, entry)
		}
	}

	// Debug: log first few entries
	if len(entries) > 0 {
		fmt.Printf("  📋 Sample weekly roster entry: NFLID=%s, Week=%d, Status=%s, StatusAbbr=%s\n",
			entries[0].NFLID, entries[0].Week, entries[0].Status, entries[0].StatusDescriptionAbbr)
	}

	return entries
}

func (l *DataLoader) updatePlayerInjuryStatus(ctx context.Context, weeklyRosters []models.WeeklyRosterEntry) int {
	if len(weeklyRosters) == 0 {
		fmt.Println("  ⚠️  No weekly roster entries to process")
		return 0
	}

	collection := l.db.Collection("players")

	// Group by player ID and get the most recent week
	playerStatusMap := make(map[string]models.WeeklyRosterEntry)
	for _, entry := range weeklyRosters {
		key := entry.NFLID + "_" + strconv.Itoa(entry.Season)
		if existing, ok := playerStatusMap[key]; ok {
			// Keep the most recent week
			if entry.Week > existing.Week {
				playerStatusMap[key] = entry
			}
		} else {
			playerStatusMap[key] = entry
		}
	}

	fmt.Printf("  📊 Parsed %d weekly entries → %d unique players\n", len(weeklyRosters), len(playerStatusMap))

	updated := 0
	matched := 0
	for _, entry := range playerStatusMap {
		filter := bson.M{
			"nfl_id": entry.NFLID,
			"season": entry.Season,
		}

		update := bson.M{
			"$set": bson.M{
				"status":                  entry.Status,
				"status_description_abbr": entry.StatusDescriptionAbbr,
				"week":                    entry.Week,
				"updated_at":              time.Now(),
			},
		}

		result, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Printf("Error updating player injury status: %v", err)
			continue
		}

		if result.MatchedCount > 0 {
			matched++
		}
		if result.ModifiedCount > 0 {
			updated++
		}
	}

	fmt.Printf("  📍 Matched: %d players, Modified: %d players\n", matched, updated)

	return updated
}

func (l *DataLoader) insertPlayerStats(ctx context.Context, stats []models.PlayerStats) int {
	if len(stats) == 0 {
		return 0
	}

	collection := l.db.Collection("player_stats")

	// Upsert stats with compound key (nfl_id + season + season_type)
	inserted := 0
	for _, stat := range stats {
		filter := bson.M{
			"nfl_id":      stat.NFLID,
			"season":      stat.Season,
			"season_type": stat.SeasonType,
		}
		update := bson.M{"$set": stat}

		opts := options.UpdateOne().SetUpsert(true)
		_, err := collection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			log.Printf("Error upserting player stats: %v", err)
			continue
		}
		inserted++
	}

	return inserted
}

func (l *DataLoader) insertWeeklyStats(ctx context.Context, weeklyStats []models.WeeklyStat) int {
	if len(weeklyStats) == 0 {
		return 0
	}

	collection := l.db.Collection("player_weekly_stats")

	// Upsert weekly stats with compound key (nfl_id + season + week)
	inserted := 0
	for _, stat := range weeklyStats {
		filter := bson.M{
			"nfl_id": stat.NFLID,
			"season": stat.Season,
			"week":   stat.Week,
		}
		update := bson.M{"$set": stat}

		opts := options.UpdateOne().SetUpsert(true)
		_, err := collection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			log.Printf("Error upserting weekly stats: %v", err)
			continue
		}
		inserted++
	}

	return inserted
}

func (l *DataLoader) insertPlays(ctx context.Context, plays []models.Play) int {
	if len(plays) == 0 {
		return 0
	}

	collection := l.db.Collection("plays")

	// Batch insert with duplicate handling
	batchSize := 1000
	inserted := 0

	for i := 0; i < len(plays); i += batchSize {
		end := i + batchSize
		if end > len(plays) {
			end = len(plays)
		}

		batch := plays[i:end]
		docs := make([]interface{}, len(batch))
		for j, play := range batch {
			docs[j] = play
		}

		opts := options.InsertMany().SetOrdered(false)
		result, err := collection.InsertMany(ctx, docs, opts)
		if err != nil {
			if bulkErr, ok := err.(mongo.BulkWriteException); ok {
				// Count successful inserts despite errors
				if result != nil && len(result.InsertedIDs) > 0 {
					inserted += len(result.InsertedIDs)
				} else {
					inserted += len(batch) - len(bulkErr.WriteErrors)
				}
			}
			continue
		}

		inserted += len(result.InsertedIDs)
	}

	return inserted
}

func (l *DataLoader) PrintFinalStats() {
	duration := time.Since(l.stats.StartTime)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 LOADING COMPLETE!")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("\n⏱️  Total Time: %s\n", duration.Round(time.Second))
	fmt.Printf("\n📥 Downloaded: %d files\n", l.stats.Downloaded)
	fmt.Printf("✅ Games Loaded: %d\n", l.stats.GamesLoaded)
	fmt.Printf("✅ Players Loaded: %d\n", l.stats.PlayersLoaded)
	fmt.Printf("✅ Plays Loaded: %d\n", l.stats.PlaysLoaded)
	fmt.Printf("❌ Errors: %d\n", l.stats.Errors)

	fmt.Println("\n🎯 Next Steps:")
	fmt.Println("1. Start backend: go run cmd/api/main.go")
	fmt.Println("2. Start frontend: cd frontend && npm run dev")
	fmt.Println("3. Test AI features with real data!")
	fmt.Println()
}
