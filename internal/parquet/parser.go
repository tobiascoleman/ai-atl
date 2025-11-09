package parquet

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/ai-atl/nfl-platform/internal/models"
	"github.com/apache/arrow/go/v14/arrow"
	"github.com/apache/arrow/go/v14/arrow/array"
	"github.com/apache/arrow/go/v14/arrow/memory"
	"github.com/apache/arrow/go/v14/parquet/file"
	"github.com/apache/arrow/go/v14/parquet/pqarrow"
)

// ParsePlayByPlay reads a Parquet file and returns Play models
func ParsePlayByPlay(data []byte, season int) ([]models.Play, error) {
	reader, err := file.NewParquetReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create parquet reader: %w", err)
	}
	defer reader.Close()

	arrowReader, err := pqarrow.NewFileReader(reader, pqarrow.ArrowReadProperties{}, memory.DefaultAllocator)
	if err != nil {
		return nil, fmt.Errorf("failed to create arrow reader: %w", err)
	}

	table, err := arrowReader.ReadTable(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to read table: %w", err)
	}
	defer table.Release()

	numRows := int(table.NumRows())
	plays := make([]models.Play, 0, numRows)

	// Get column indices
	schema := table.Schema()
	colMap := make(map[string]int)
	for i, field := range schema.Fields() {
		colMap[field.Name] = i
	}

	// Helper function to find which chunk contains a row
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

	// Helper functions to extract values across multiple chunks
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

	getFloat := func(colName string, rowIdx int) float64 {
		if colIdx, ok := colMap[colName]; ok {
			col := table.Column(colIdx)
			chunk, offset := getChunkAndOffset(col, rowIdx)
			if chunk != nil {
				if arr, ok := chunk.(*array.Float64); ok && !arr.IsNull(offset) {
					return arr.Value(offset)
				}
			}
		}
		return 0.0
	}

	getBool := func(colName string, rowIdx int) bool {
		if colIdx, ok := colMap[colName]; ok {
			col := table.Column(colIdx)
			chunk, offset := getChunkAndOffset(col, rowIdx)
			if chunk != nil {
				if arr, ok := chunk.(*array.Boolean); ok && !arr.IsNull(offset) {
					return arr.Value(offset)
				}
			}
		}
		return false
	}

	// Parse each row
	for i := 0; i < numRows; i++ {
		// Try 'play_id' first, fall back to 'id' column
		playID := getString("play_id", i)
		if playID == "" {
			playID = getString("id", i)
		}

		play := models.Play{
			GameID:           getString("game_id", i),
			PlayID:           playID,
			Season:           season,
			Week:             getInt("week", i),
			Quarter:          getInt("qtr", i),
			Down:             getInt("down", i),
			YardsToGo:        getInt("ydstogo", i),
			YardLine:         getInt("yardline_100", i),
			GameSeconds:      getInt("game_seconds_remaining", i),
			Description:      getString("desc", i),
			PlayType:         getString("play_type", i),
			PossessionTeam:   getString("posteam", i),
			DefenseTeam:      getString("defteam", i),
			PasserPlayerID:   getString("passer_player_id", i),
			PasserPlayerName: getString("passer_player_name", i),
			ReceiverPlayerID: getString("receiver_player_id", i),
			RusherPlayerID:   getString("rusher_player_id", i),
			Yards:            getInt("yards_gained", i),
			Touchdown:        getBool("touchdown", i),
			Interception:     getBool("interception", i),
			Fumble:           getBool("fumble", i),
			Sack:             getBool("sack", i),
			EPA:              getFloat("epa", i),
			WPA:              getFloat("wpa", i),
			SuccessPlay:      getBool("success", i),
			AirYards:         getInt("air_yards", i),
			YardsAfterCatch:  getInt("yards_after_catch", i),
			CreatedAt:        time.Now(),
		}

		if play.PlayID != "" {
			plays = append(plays, play)
		}
	}

	return plays, nil
}

// ParseRoster reads a Parquet roster file and returns Player models
func ParseRoster(data []byte, season int) ([]models.Player, error) {
	reader, err := file.NewParquetReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create parquet reader: %w", err)
	}
	defer reader.Close()

	arrowReader, err := pqarrow.NewFileReader(reader, pqarrow.ArrowReadProperties{}, memory.DefaultAllocator)
	if err != nil {
		return nil, fmt.Errorf("failed to create arrow reader: %w", err)
	}

	table, err := arrowReader.ReadTable(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to read table: %w", err)
	}
	defer table.Release()

	numRows := int(table.NumRows())
	players := make([]models.Player, 0, numRows)

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

	for i := 0; i < numRows; i++ {
		player := models.Player{
			NFLID:     getString("gsis_id", i),
			Season:    season, // Track which year this roster is from
			Name:      getString("full_name", i),
			Position:  getString("position", i),
			Team:      getString("team", i),
			UpdatedAt: time.Now(),
		}

		if player.NFLID != "" {
			players = append(players, player)
		}
	}

	return players, nil
}

// ParsePlayerStats reads a Parquet player stats file and returns PlayerStats models
func ParsePlayerStats(data []byte, season int, seasonType string) ([]models.PlayerStats, error) {
	reader, err := file.NewParquetReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create parquet reader: %w", err)
	}
	defer reader.Close()

	arrowReader, err := pqarrow.NewFileReader(reader, pqarrow.ArrowReadProperties{}, memory.DefaultAllocator)
	if err != nil {
		return nil, fmt.Errorf("failed to create arrow reader: %w", err)
	}

	table, err := arrowReader.ReadTable(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to read table: %w", err)
	}
	defer table.Release()

	numRows := int(table.NumRows())
	stats := make([]models.PlayerStats, 0, numRows)

	schema := table.Schema()
	colMap := make(map[string]int)

	// Debug: Print all available columns (first time only)
	fmt.Printf("📋 Available columns in player_stats (season %d): ", season)
	columnNames := make([]string, 0, len(schema.Fields()))
	for i, field := range schema.Fields() {
		colMap[field.Name] = i
		columnNames = append(columnNames, field.Name)
	}
	fmt.Printf("%v\n", columnNames)

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

	getFloat := func(colName string, rowIdx int) float64 {
		if colIdx, ok := colMap[colName]; ok {
			col := table.Column(colIdx)
			chunk, offset := getChunkAndOffset(col, rowIdx)
			if chunk != nil {
				switch arr := chunk.(type) {
				case *array.Float64:
					if !arr.IsNull(offset) {
						return arr.Value(offset)
					}
				case *array.Float32:
					if !arr.IsNull(offset) {
						return float64(arr.Value(offset))
					}
				}
			}
		}
		return 0.0
	}

	for i := 0; i < numRows; i++ {
		// Calculate combined EPA from passing, rushing, and receiving EPA
		passingEPA := getFloat("passing_epa", i)
		rushingEPA := getFloat("rushing_epa", i)
		receivingEPA := getFloat("receiving_epa", i)

		// Sum non-zero EPAs
		combinedEPA := passingEPA + rushingEPA + receivingEPA

		// Count how many plays were involved (for averaging)
		playCount := 0
		if passingEPA != 0 {
			playCount += getInt("attempts", i) // Passing attempts
		}
		if rushingEPA != 0 {
			playCount += getInt("carries", i) // Rushing carries
		}
		if receivingEPA != 0 {
			playCount += getInt("targets", i) // Receiving targets
		}

		playerStats := models.PlayerStats{
			NFLID:      getString("player_id", i),
			Season:     season,
			SeasonType: seasonType,

			// Offensive Stats (CORRECTED COLUMN NAMES)
			PassingYards:  getInt("passing_yards", i),
			PassingTDs:    getInt("passing_tds", i),
			Interceptions: getInt("passing_interceptions", i), // FIXED: was "interceptions"

			RushingYards: getInt("rushing_yards", i),
			RushingTDs:   getInt("rushing_tds", i),

			Receptions:     getInt("receptions", i),
			ReceivingYards: getInt("receiving_yards", i),
			ReceivingTDs:   getInt("receiving_tds", i),
			Targets:        getInt("targets", i),

			// Defensive Stats (CORRECTED COLUMN NAMES)
			Tackles:          getInt("def_tackles_with_assist", i), // FIXED: was "def_tackles_combined"
			TacklesSolo:      getInt("def_tackles_solo", i),
			TacklesAssist:    getInt("def_tackle_assists", i), // Already correct
			TacklesForLoss:   getFloat("def_tackles_for_loss", i),
			Sacks:            getFloat("def_sacks", i),
			SackYards:        getFloat("def_sack_yards", i),
			DefInterceptions: getInt("def_interceptions", i),
			PassDefended:     getInt("def_pass_defended", i), // FIXED: was "def_passes_defended"
			ForcedFumbles:    getInt("def_fumbles_forced", i),
			FumbleRecoveries: getInt("fumble_recovery_opp", i), // FIXED: was "def_fumbles_recovered"
			DefensiveTDs:     getInt("def_tds", i),
			SafetyMD:         getInt("def_safeties", i), // FIXED: was "def_safety"

			// Performance Metrics (from parquet file)
			EPA:       combinedEPA,
			PlayCount: playCount,

			// Fantasy Points
			FantasyPoints:    getFloat("fantasy_points", i),
			FantasyPointsPPR: getFloat("fantasy_points_ppr", i),

			UpdatedAt: time.Now(),
		}

		if playerStats.NFLID != "" {
			stats = append(stats, playerStats)
		}
	}

	return stats, nil
}

// ParseWeeklyStats reads a Parquet weekly player stats file and returns WeeklyStat models
func ParseWeeklyStats(data []byte, season int) ([]models.WeeklyStat, error) {
	reader, err := file.NewParquetReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create parquet reader: %w", err)
	}
	defer reader.Close()

	arrowReader, err := pqarrow.NewFileReader(reader, pqarrow.ArrowReadProperties{}, memory.DefaultAllocator)
	if err != nil {
		return nil, fmt.Errorf("failed to create arrow reader: %w", err)
	}

	table, err := arrowReader.ReadTable(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to read table: %w", err)
	}
	defer table.Release()

	numRows := int(table.NumRows())
	weeklyStats := make([]models.WeeklyStat, 0, numRows)

	schema := table.Schema()
	colMap := make(map[string]int)

	// Debug: Print all available columns (first time only)
	fmt.Printf("📋 Available columns in weekly_stats (season %d): ", season)
	columnNames := make([]string, 0, len(schema.Fields()))
	for i, field := range schema.Fields() {
		colMap[field.Name] = i
		columnNames = append(columnNames, field.Name)
	}
	fmt.Printf("%v\n", columnNames)

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
				case *array.Float64:
					if !arr.IsNull(offset) {
						return int(arr.Value(offset))
					}
				}
			}
		}
		return 0
	}

	getFloat := func(colName string, rowIdx int) float64 {
		if colIdx, ok := colMap[colName]; ok {
			col := table.Column(colIdx)
			chunk, offset := getChunkAndOffset(col, rowIdx)
			if chunk != nil {
				switch arr := chunk.(type) {
				case *array.Float64:
					if !arr.IsNull(offset) {
						return arr.Value(offset)
					}
				case *array.Float32:
					if !arr.IsNull(offset) {
						return float64(arr.Value(offset))
					}
				}
			}
		}
		return 0.0
	}

	for i := 0; i < numRows; i++ {
		// Calculate combined EPA from passing, rushing, and receiving EPA
		passingEPA := getFloat("passing_epa", i)
		rushingEPA := getFloat("rushing_epa", i)
		receivingEPA := getFloat("receiving_epa", i)
		combinedEPA := passingEPA + rushingEPA + receivingEPA

		weeklyStat := models.WeeklyStat{
			NFLID:    getString("player_id", i),
			Week:     getInt("week", i),
			Season:   season,
			Opponent: getString("opponent_team", i),

			// Passing Stats
			PassingYards:  getInt("passing_yards", i),
			PassingTDs:    getInt("passing_tds", i),
			Interceptions: getInt("passing_interceptions", i),

			// Rushing Stats
			Carries:      getInt("carries", i),
			RushingYards: getInt("rushing_yards", i),
			RushingTDs:   getInt("rushing_tds", i),

			// Receiving Stats
			Receptions:     getInt("receptions", i),
			Targets:        getInt("targets", i),
			ReceivingYards: getInt("receiving_yards", i),
			ReceivingTDs:   getInt("receiving_tds", i),

			// Performance Metrics
			EPA: combinedEPA,

			// Fantasy Points
			FantasyPoints:    getFloat("fantasy_points", i),
			FantasyPointsPPR: getFloat("fantasy_points_ppr", i),

			UpdatedAt: time.Now(),
		}

		if weeklyStat.NFLID != "" && weeklyStat.Week > 0 {
			weeklyStats = append(weeklyStats, weeklyStat)
		}
	}

	return weeklyStats, nil
}

// ParseSchedules reads a Parquet schedule file and returns Game models
func ParseSchedules(data []byte) ([]models.Game, error) {
	reader, err := file.NewParquetReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create parquet reader: %w", err)
	}
	defer reader.Close()

	arrowReader, err := pqarrow.NewFileReader(reader, pqarrow.ArrowReadProperties{}, memory.DefaultAllocator)
	if err != nil {
		return nil, fmt.Errorf("failed to create arrow reader: %w", err)
	}

	table, err := arrowReader.ReadTable(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to read table: %w", err)
	}
	defer table.Release()

	numRows := int(table.NumRows())
	games := make([]models.Game, 0, numRows)

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

	getFloat := func(colName string, rowIdx int) float64 {
		if colIdx, ok := colMap[colName]; ok {
			col := table.Column(colIdx)
			chunk, offset := getChunkAndOffset(col, rowIdx)
			if chunk != nil {
				if arr, ok := chunk.(*array.Float64); ok && !arr.IsNull(offset) {
					return arr.Value(offset)
				}
			}
		}
		return 0.0
	}

	parseGameDateTime := func(gamedayStr, gametimeStr string) time.Time {
		if gamedayStr == "" {
			return time.Time{}
		}

		// Eastern Time zone (NFL games are typically in ET)
		etLoc, err := time.LoadLocation("America/New_York")
		if err != nil {
			etLoc = time.UTC
		}

		// Combine gameday and gametime
		dateTimeStr := gamedayStr
		if gametimeStr != "" {
			dateTimeStr = gamedayStr + " " + gametimeStr
		} else {
			dateTimeStr = gamedayStr + " 13:00" // Default to 1pm ET if no time
		}

		// Parse as Eastern Time
		t, err := time.ParseInLocation("2006-01-02 15:04", dateTimeStr, etLoc)
		if err != nil {
			// Fallback: try just the date
			t, err = time.Parse("2006-01-02", gamedayStr)
			if err != nil {
				return time.Time{}
			}
		}

		return t
	}

	for i := 0; i < numRows; i++ {
		homeScore := getInt("home_score", i)
		awayScore := getInt("away_score", i)
		gamedayStr := getString("gameday", i)
		gametimeStr := getString("gametime", i)
		startTime := parseGameDateTime(gamedayStr, gametimeStr)

		// Determine status based on whether game has been played
		// If both scores are 0 and game time is in the future, it's scheduled
		status := "final"
		if homeScore == 0 && awayScore == 0 && !startTime.IsZero() {
			// Compare the actual game time to now
			// Add a buffer: games with no scores scheduled more than 4 hours ago are likely final
			now := time.Now()
			fourHoursAgo := now.Add(-4 * time.Hour)

			if startTime.After(fourHoursAgo) {
				status = "scheduled"
			}
		}

		game := models.Game{
			GameID:    getString("game_id", i),
			Season:    getInt("season", i),
			Week:      getInt("week", i),
			HomeTeam:  getString("home_team", i),
			AwayTeam:  getString("away_team", i),
			StartTime: startTime,
			VegasLine: getFloat("spread_line", i),
			OverUnder: getFloat("total_line", i),
			HomeScore: homeScore,
			AwayScore: awayScore,
			Status:    status,
			UpdatedAt: time.Now(),
		}

		if game.GameID != "" {
			games = append(games, game)
		}
	}

	return games, nil
}

// ParseNextGenStats reads a Parquet NGS file and returns NextGenStat models
func ParseNextGenStats(data []byte, statType string) ([]models.NextGenStat, error) {
	reader, err := file.NewParquetReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create parquet reader: %w", err)
	}
	defer reader.Close()

	arrowReader, err := pqarrow.NewFileReader(reader, pqarrow.ArrowReadProperties{}, memory.DefaultAllocator)
	if err != nil {
		return nil, fmt.Errorf("failed to create arrow reader: %w", err)
	}

	table, err := arrowReader.ReadTable(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to read table: %w", err)
	}
	defer table.Release()

	numRows := int(table.NumRows())
	stats := make([]models.NextGenStat, 0, numRows)

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

	getFloat := func(colName string, rowIdx int) float64 {
		if colIdx, ok := colMap[colName]; ok {
			col := table.Column(colIdx)
			chunk, offset := getChunkAndOffset(col, rowIdx)
			if chunk != nil {
				if arr, ok := chunk.(*array.Float64); ok && !arr.IsNull(offset) {
					return arr.Value(offset)
				}
			}
		}
		return 0.0
	}

	for i := 0; i < numRows; i++ {
		stat := models.NextGenStat{
			PlayerID:   getString("player_gsis_id", i),
			Season:     getInt("season", i),
			Week:       getInt("week", i),
			StatType:   statType,
			PlayerName: getString("player_display_name", i),
			Team:       getString("team_abbr", i),
			Position:   getString("player_position", i),
			UpdatedAt:  time.Now(),
		}

		// Parse stat-specific fields based on type
		switch statType {
		case "passing":
			stat.PassAttempts = getInt("attempts", i)
			stat.PassCompletions = getInt("completions", i)
			stat.PassYards = getInt("pass_yards", i)
			stat.PassTouchdowns = getInt("pass_touchdowns", i)
			stat.Interceptions = getInt("interceptions", i)
			stat.CompletionPercentageAboveExpectation = getFloat("completion_percentage_above_expectation", i)
			stat.AvgTimeToThrow = getFloat("avg_time_to_throw", i)
			stat.AvgCompletedAirYards = getFloat("avg_completed_air_yards", i)
			stat.AvgIntendedAirYards = getFloat("avg_intended_air_yards", i)
			stat.AvgAirYardsDifferential = getFloat("avg_air_yards_differential", i)
			stat.MaxCompletedAirDistance = getFloat("max_completed_air_distance", i)

		case "rushing":
			stat.Carries = getInt("carries", i)
			stat.RushYards = getInt("rush_yards", i)
			stat.RushTouchdowns = getInt("rush_touchdowns", i)
			stat.ExpectedRushYards = getFloat("expected_rush_yards", i)
			stat.RushYardsOverExpected = getFloat("rush_yards_over_expected", i)
			stat.AvgTimeToLOS = getFloat("avg_time_to_los", i)
			stat.Efficiency = getFloat("efficiency", i)

		case "receiving":
			stat.Receptions = getInt("receptions", i)
			stat.Targets = getInt("targets", i)
			stat.ReceivingYards = getInt("yards", i)
			stat.ReceivingTouchdowns = getInt("rec_touchdowns", i)
			stat.AvgCushion = getFloat("avg_cushion", i)
			stat.AvgSeparation = getFloat("avg_separation", i)
			stat.AvgIntendedAirYardsRec = getFloat("avg_intended_air_yards", i)
			stat.CatchPercentage = getFloat("percent_share_of_intended_air_yards", i)
			stat.AvgYAC = getFloat("avg_yac", i)
			stat.AvgExpectedYAC = getFloat("avg_expected_yac", i)
			stat.AvgYACAboveExpectation = getFloat("avg_yac_above_expectation", i)
		}

		if stat.PlayerID != "" {
			stats = append(stats, stat)
		}
	}

	return stats, nil
}
