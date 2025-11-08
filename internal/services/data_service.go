package services

import (
	"context"
	"log"
	"time"

	"github.com/ai-atl/nfl-platform/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// DataService provides methods to query NFL data
type DataService struct {
	db *mongo.Database
}

func NewDataService(db *mongo.Database) *DataService {
	return &DataService{db: db}
}

// ========================================
// PLAYER QUERIES
// ========================================

// GetPlayer retrieves a player by NFL ID and season
func (s *DataService) GetPlayer(ctx context.Context, nflID string, season int) (*models.Player, error) {
	filter := bson.M{
		"nfl_id": nflID,
		"season": season,
	}

	log.Printf("🔍 GetPlayer query: %+v", filter)

	var player models.Player
	err := s.db.Collection("players").FindOne(ctx, filter).Decode(&player)

	if err != nil {
		log.Printf("❌ GetPlayer error: %v (nfl_id=%s, season=%d)", err, nflID, season)
	} else {
		log.Printf("✅ GetPlayer found: %s (%s, %d)", player.Name, player.Team, player.Season)
	}

	return &player, err
}

// GetPlayersByTeam gets all players for a team in a season
func (s *DataService) GetPlayersByTeam(ctx context.Context, team string, season int) ([]models.Player, error) {
	cursor, err := s.db.Collection("players").Find(ctx, bson.M{
		"team":   team,
		"season": season,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var players []models.Player
	if err := cursor.All(ctx, &players); err != nil {
		return nil, err
	}
	return players, nil
}

// GetPlayersByPosition gets players by position for a season
func (s *DataService) GetPlayersByPosition(ctx context.Context, position string, season int) ([]models.Player, error) {
	cursor, err := s.db.Collection("players").Find(ctx, bson.M{
		"position": position,
		"season":   season,
	}, options.Find().SetLimit(100))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var players []models.Player
	if err := cursor.All(ctx, &players); err != nil {
		return nil, err
	}
	return players, nil
}

// GetInjuredPlayers gets players with injury status
func (s *DataService) GetInjuredPlayers(ctx context.Context, season int) ([]models.Player, error) {
	filter := bson.M{
		"season": season,
		"$or": []bson.M{
			{"status": "INA"},
			{"status_description_abbr": bson.M{"$in": []string{"R01", "R04", "R48", "P02"}}},
		},
	}

	cursor, err := s.db.Collection("players").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var players []models.Player
	if err := cursor.All(ctx, &players); err != nil {
		return nil, err
	}
	return players, nil
}

// ========================================
// PLAYER STATS QUERIES
// ========================================

// GetPlayerStats gets seasonal stats for a player
func (s *DataService) GetPlayerStats(ctx context.Context, nflID string, season int) ([]models.PlayerStats, error) {
	filter := bson.M{"nfl_id": nflID}
	if season > 0 {
		filter["season"] = season
	}

	cursor, err := s.db.Collection("player_stats").Find(ctx, filter,
		options.Find().SetSort(bson.D{{"season", -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stats []models.PlayerStats
	if err := cursor.All(ctx, &stats); err != nil {
		return nil, err
	}
	return stats, nil
}

// ========================================
// PLAY-BY-PLAY QUERIES
// ========================================

// GetPlayerPlays gets all plays involving a player
func (s *DataService) GetPlayerPlays(ctx context.Context, playerID string, season int, limit int) ([]models.Play, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"passer_player_id": playerID},
			{"rusher_player_id": playerID},
			{"receiver_player_id": playerID},
		},
	}
	if season > 0 {
		filter["season"] = season
	}

	opts := options.Find().SetLimit(int64(limit))
	cursor, err := s.db.Collection("plays").Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var plays []models.Play
	if err := cursor.All(ctx, &plays); err != nil {
		return nil, err
	}
	return plays, nil
}

// GetTeamPlays gets all plays for a team
func (s *DataService) GetTeamPlays(ctx context.Context, team string, season int, limit int) ([]models.Play, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"possession_team": team},
			{"defense_team": team},
		},
	}
	if season > 0 {
		filter["season"] = season
	}

	opts := options.Find().SetLimit(int64(limit))
	cursor, err := s.db.Collection("plays").Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var plays []models.Play
	if err := cursor.All(ctx, &plays); err != nil {
		return nil, err
	}
	return plays, nil
}

// GetGamePlays gets all plays for a specific game
func (s *DataService) GetGamePlays(ctx context.Context, gameID string) ([]models.Play, error) {
	cursor, err := s.db.Collection("plays").Find(ctx, bson.M{"game_id": gameID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var plays []models.Play
	if err := cursor.All(ctx, &plays); err != nil {
		return nil, err
	}
	return plays, nil
}

// ========================================
// EPA CALCULATIONS
// ========================================

// CalculatePlayerEPA calculates average EPA for a player
func (s *DataService) CalculatePlayerEPA(ctx context.Context, playerID string, season int) (float64, int, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"passer_player_id": playerID},
			{"rusher_player_id": playerID},
			{"receiver_player_id": playerID},
		},
	}
	if season > 0 {
		filter["season"] = season
	}

	cursor, err := s.db.Collection("plays").Find(ctx, filter)
	if err != nil {
		return 0, 0, err
	}
	defer cursor.Close(ctx)

	var plays []models.Play
	if err := cursor.All(ctx, &plays); err != nil {
		return 0, 0, err
	}

	if len(plays) == 0 {
		return 0, 0, nil
	}

	totalEPA := 0.0
	for _, play := range plays {
		totalEPA += play.EPA
	}

	avgEPA := totalEPA / float64(len(plays))
	return avgEPA, len(plays), nil
}

// CalculateTeamEPA calculates average EPA for a team's offense
func (s *DataService) CalculateTeamEPA(ctx context.Context, team string, season int) (float64, int, error) {
	filter := bson.M{"possession_team": team}
	if season > 0 {
		filter["season"] = season
	}

	cursor, err := s.db.Collection("plays").Find(ctx, filter)
	if err != nil {
		return 0, 0, err
	}
	defer cursor.Close(ctx)

	var plays []models.Play
	if err := cursor.All(ctx, &plays); err != nil {
		return 0, 0, err
	}

	if len(plays) == 0 {
		return 0, 0, nil
	}

	totalEPA := 0.0
	for _, play := range plays {
		totalEPA += play.EPA
	}

	avgEPA := totalEPA / float64(len(plays))
	return avgEPA, len(plays), nil
}

// ========================================
// NGS (NEXT GEN STATS) QUERIES
// ========================================

// GetPlayerNGS gets Next Gen Stats for a player
func (s *DataService) GetPlayerNGS(ctx context.Context, playerID string, statType string, season int) ([]models.NextGenStat, error) {
	filter := bson.M{"player_id": playerID}
	if statType != "" {
		filter["stat_type"] = statType
	}
	if season > 0 {
		filter["season"] = season
	}

	cursor, err := s.db.Collection("next_gen_stats").Find(ctx, filter,
		options.Find().SetSort(bson.D{{"season", -1}, {"week", -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stats []models.NextGenStat
	if err := cursor.All(ctx, &stats); err != nil {
		return nil, err
	}
	return stats, nil
}

// GetNGSLeaders gets top players by a specific NGS metric
func (s *DataService) GetNGSLeaders(ctx context.Context, statType string, season int, metric string, limit int) ([]models.NextGenStat, error) {
	filter := bson.M{
		"stat_type": statType,
		"season":    season,
		"week":      0, // Season totals
	}

	opts := options.Find().
		SetSort(bson.D{{metric, -1}}).
		SetLimit(int64(limit))

	cursor, err := s.db.Collection("next_gen_stats").Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stats []models.NextGenStat
	if err := cursor.All(ctx, &stats); err != nil {
		return nil, err
	}
	return stats, nil
}

// ========================================
// GAME QUERIES
// ========================================

// GetGame gets a specific game by ID
func (s *DataService) GetGame(ctx context.Context, gameID string) (*models.Game, error) {
	var game models.Game
	err := s.db.Collection("games").FindOne(ctx, bson.M{"game_id": gameID}).Decode(&game)
	return &game, err
}

// GetGamesBySeason gets games for a season
func (s *DataService) GetGamesBySeason(ctx context.Context, season int, week int) ([]models.Game, error) {
	filter := bson.M{"season": season}
	if week > 0 {
		filter["week"] = week
	}

	cursor, err := s.db.Collection("games").Find(ctx, filter,
		options.Find().SetSort(bson.D{{"week", 1}, {"game_date", 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var games []models.Game
	if err := cursor.All(ctx, &games); err != nil {
		return nil, err
	}
	return games, nil
}

// GetUpcomingGames gets upcoming games for a team
func (s *DataService) GetUpcomingGames(ctx context.Context, team string) ([]models.Game, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"home_team": team},
			{"away_team": team},
		},
		"game_date": bson.M{"$gte": time.Now()},
	}

	cursor, err := s.db.Collection("games").Find(ctx, filter,
		options.Find().SetSort(bson.D{{"game_date", 1}}).SetLimit(5))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var games []models.Game
	if err := cursor.All(ctx, &games); err != nil {
		return nil, err
	}
	return games, nil
}

// ========================================
// AGGREGATE QUERIES
// ========================================

// GetPlayerSummary gets comprehensive player data for ALL seasons
func (s *DataService) GetPlayerSummary(ctx context.Context, nflID string, season int) (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	// Get player info for requested season (or most recent if season=0)
	var player *models.Player
	var err error

	if season > 0 {
		player, err = s.GetPlayer(ctx, nflID, season)
	} else {
		// Get most recent season
		cursor, err := s.db.Collection("players").Find(
			ctx,
			bson.M{"nfl_id": nflID},
			options.Find().SetSort(bson.D{{Key: "season", Value: -1}}).SetLimit(1),
		)
		if err == nil {
			var players []models.Player
			cursor.All(ctx, &players)
			if len(players) > 0 {
				player = &players[0]
			}
			cursor.Close(ctx)
		}
	}

	if player == nil {
		return nil, err
	}

	summary["player"] = player

	// Get ALL seasons for this player
	allSeasonsCursor, _ := s.db.Collection("players").Find(
		ctx,
		bson.M{"nfl_id": nflID},
		options.Find().SetSort(bson.D{{Key: "season", Value: -1}}),
	)
	var allSeasons []models.Player
	if allSeasonsCursor != nil {
		allSeasonsCursor.All(ctx, &allSeasons)
		allSeasonsCursor.Close(ctx)
	}
	summary["all_seasons"] = allSeasons

	// Get ALL stats (all seasons)
	allStats, _ := s.GetPlayerStats(ctx, nflID, 0) // 0 = all seasons
	summary["all_stats"] = allStats

	// Get current season stats
	currentStats, _ := s.GetPlayerStats(ctx, nflID, player.Season)
	summary["stats"] = currentStats

	// Get EPA from player_stats (pre-calculated, much faster!)
	currentSeasonStat := currentStats
	var epa float64
	var playCount int
	if len(currentSeasonStat) > 0 {
		epa = currentSeasonStat[0].EPA
		playCount = currentSeasonStat[0].PlayCount
	}
	summary["epa"] = epa
	summary["play_count"] = playCount

	// Build EPA by season map from all_stats (already have EPA pre-calculated)
	epaBySeasonMap := make(map[int]map[string]interface{})
	var lifetimeEPASum float64
	var lifetimePlaysSum int

	for _, stat := range allStats {
		if stat.PlayCount > 0 {
			epaBySeasonMap[stat.Season] = map[string]interface{}{
				"epa":        stat.EPA,
				"play_count": stat.PlayCount,
			}
			lifetimeEPASum += stat.EPA * float64(stat.PlayCount) // Weight by play count
			lifetimePlaysSum += stat.PlayCount
		}
	}

	// Calculate lifetime average EPA
	var lifetimeEPA float64
	if lifetimePlaysSum > 0 {
		lifetimeEPA = lifetimeEPASum / float64(lifetimePlaysSum)
	}

	summary["epa_by_season"] = epaBySeasonMap
	summary["lifetime_epa"] = lifetimeEPA
	summary["lifetime_plays"] = lifetimePlaysSum

	// Get NGS stats for current season
	ngs, _ := s.GetPlayerNGS(ctx, nflID, "", player.Season)
	summary["ngs"] = ngs

	// Get ALL NGS stats (all seasons)
	allNGS, _ := s.GetPlayerNGS(ctx, nflID, "", 0) // 0 = all seasons
	summary["all_ngs"] = allNGS

	return summary, nil
}

// GetTeamDepthChart gets team's roster by position
func (s *DataService) GetTeamDepthChart(ctx context.Context, team string, season int) (map[string][]models.Player, error) {
	players, err := s.GetPlayersByTeam(ctx, team, season)
	if err != nil {
		return nil, err
	}

	depthChart := make(map[string][]models.Player)
	for _, player := range players {
		depthChart[player.Position] = append(depthChart[player.Position], player)
	}

	return depthChart, nil
}
