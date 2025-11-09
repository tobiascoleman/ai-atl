package services

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ai-atl/nfl-platform/internal/models"
	"github.com/ai-atl/nfl-platform/pkg/gemini"
	"github.com/ai-atl/nfl-platform/pkg/sleeper"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type WaiverWireService struct {
	db            *mongo.Database
	gemini        *gemini.Client
	dataService   *DataService
	sleeperClient *sleeper.Client
}

type WaiverGem struct {
	// Basic player info
	PlayerName string `json:"playerName"`
	Position   string `json:"position"`
	Team       string `json:"team"`

	// Key metrics
	BreakoutScore    float64 `json:"breakoutScore"`    // 0-100
	TargetShareTrend string  `json:"targetShareTrend"` // "increasing", "stable", "decreasing"
	SnapCountPct     float64 `json:"snapCountPct"`     // Recent snap percentage
	EPAPerPlay       float64 `json:"epaPerPlay"`

	// Opportunity analysis
	DepthChartStatus string `json:"depthChartStatus"` // "starter injured", "increased role", "backup"
	UpcomingSchedule string `json:"upcomingSchedule"` // "favorable", "average", "difficult"
	ScheduleRank     int    `json:"scheduleRank"`     // 1-32, lower is easier

	// Recent performance
	LastThreeGames []GameStats `json:"lastThreeGames"`
	TrendingUp     bool        `json:"trendingUp"`

	// AI analysis
	AIAnalysis     string `json:"aiAnalysis"`
	Recommendation string `json:"recommendation"` // "Must Add", "Strong Add", "Monitor", "Pass"
}

type GameStats struct {
	Week          int     `json:"week"`
	Opponent      string  `json:"opponent"`
	SnapPct       float64 `json:"snapPct"`
	Targets       int     `json:"targets"`
	TargetShare   float64 `json:"targetShare"`
	Production    string  `json:"production"` // e.g., "5 rec, 72 yds, 1 TD"
	FantasyPoints float64 `json:"fantasyPoints"`
}

type ScheduleAnalysis struct {
	NextThreeOpponents []string
	AvgDefensiveRank   float64
	Difficulty         string
}

func NewWaiverWireService(db *mongo.Database) *WaiverWireService {
	return &WaiverWireService{
		db:            db,
		gemini:        gemini.NewClient(),
		dataService:   NewDataService(db),
		sleeperClient: sleeper.NewClient(),
	}
}

// FindWaiverGems identifies undervalued players with breakout potential
func (s *WaiverWireService) FindWaiverGems(ctx context.Context, position string, limit int) ([]WaiverGem, error) {
	season := 2025
	currentWeek := 10

	// Get all players for the position (limit initial query for performance)
	var positionFilter bson.M
	maxPlayersToAnalyze := 20 // Reduced to 20 for faster analysis

	if position != "" && position != "ALL" {
		positionFilter = bson.M{"position": position, "season": season}
	} else {
		positionFilter = bson.M{
			"position": bson.M{"$in": []string{"QB", "RB", "WR", "TE"}},
			"season":   season,
		}
		maxPlayersToAnalyze = 30 // Reduced to 30 for ALL positions
	}

	// Limit query for performance
	findOptions := options.Find().SetLimit(int64(maxPlayersToAnalyze))

	cursor, err := s.db.Collection("players").Find(ctx, positionFilter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var players []models.Player
	if err := cursor.All(ctx, &players); err != nil {
		return nil, err
	}

	fmt.Printf("Analyzing %d players for position %s...\n", len(players), position)

	// Analyze each player for breakout potential
	var gems []WaiverGem
	for i, player := range players {
		if i%10 == 0 {
			fmt.Printf("Progress: %d/%d players analyzed\n", i, len(players))
		}

		gem := s.analyzeBreakoutPotential(ctx, player, season, currentWeek)
		if gem != nil && gem.BreakoutScore > 0 { // Include all players for now (no real data yet)
			gems = append(gems, *gem)
		}
	}

	fmt.Printf("Found %d candidates with score > 0\n", len(gems))

	// Sort by breakout score
	sort.Slice(gems, func(i, j int) bool {
		return gems[i].BreakoutScore > gems[j].BreakoutScore
	})

	// Limit results
	if limit > 0 && len(gems) > limit {
		gems = gems[:limit]
	}

	// Generate AI analysis for top candidates (reduced to top 5 for speed)
	fmt.Printf("Generating AI analysis for top %d candidates...\n", min(5, len(gems)))
	for i := range gems {
		if i < 5 { // Only analyze top 5 to save API calls and time
			gems[i].AIAnalysis = s.generateAIAnalysis(ctx, &gems[i])
		} else {
			gems[i].AIAnalysis = "High breakout potential based on metrics"
		}
	}

	return gems, nil
}

// RosterPlayer represents a player on user's ESPN roster
type RosterPlayer struct {
	Name            string  `json:"name"`
	Position        string  `json:"position"`
	ProjectedPoints float64 `json:"projectedPoints"`
	LineupSlot      string  `json:"lineupSlot"`
}

// FindPersonalizedWaiverGems analyzes waiver wire based on user's roster needs
func (s *WaiverWireService) FindPersonalizedWaiverGems(ctx context.Context, roster []RosterPlayer, limit int) ([]WaiverGem, error) {
	// Analyze roster strength by position
	positionStrength := s.analyzeRosterStrength(roster)

	// Find weak positions that need upgrades
	weakPositions := s.identifyWeakPositions(positionStrength)

	fmt.Printf("Roster analysis: Weak positions: %v\n", weakPositions)

	// Get waiver gems for ALL positions in one query (much faster than 4 separate calls)
	allGems, err := s.FindWaiverGems(ctx, "ALL", 30)
	if err != nil {
		return nil, err
	}

	// Prioritize based on roster needs
	for i := range allGems {
		allGems[i].BreakoutScore = s.adjustScoreForRosterFit(allGems[i], positionStrength, weakPositions)

		// Add roster context to existing AI analysis instead of regenerating
		if allGems[i].Position != "" {
			posAvg := positionStrength[allGems[i].Position]
			if posAvg > 0 {
				allGems[i].AIAnalysis = fmt.Sprintf("🎯 TEAM NEED: Your %s average %.1f pts/week. %s",
					allGems[i].Position, posAvg, allGems[i].AIAnalysis)
			}
		}
	}

	// Sort by adjusted score
	sort.Slice(allGems, func(i, j int) bool {
		return allGems[i].BreakoutScore > allGems[j].BreakoutScore
	})

	// Return top candidates
	if limit > 0 && len(allGems) > limit {
		allGems = allGems[:limit]
	}

	return allGems, nil
}

// analyzeRosterStrength calculates average projected points by position
func (s *WaiverWireService) analyzeRosterStrength(roster []RosterPlayer) map[string]float64 {
	positionTotals := make(map[string]float64)
	positionCounts := make(map[string]int)

	for _, player := range roster {
		if player.LineupSlot != "BE" && player.LineupSlot != "IR" { // Only count starters
			positionTotals[player.Position] += player.ProjectedPoints
			positionCounts[player.Position]++
		}
	}

	averages := make(map[string]float64)
	for pos, total := range positionTotals {
		if count := positionCounts[pos]; count > 0 {
			averages[pos] = total / float64(count)
		}
	}

	return averages
}

// identifyWeakPositions finds positions that need upgrades
func (s *WaiverWireService) identifyWeakPositions(positionStrength map[string]float64) []string {
	// League average projected points by position (approximate)
	leagueAverages := map[string]float64{
		"QB": 18.0,
		"RB": 12.0,
		"WR": 10.0,
		"TE": 8.0,
	}

	weak := []string{}
	for pos, avg := range positionStrength {
		if leagueAvg, ok := leagueAverages[pos]; ok {
			if avg < leagueAvg*0.85 { // If 15% below league average
				weak = append(weak, pos)
			}
		}
	}

	return weak
}

// adjustScoreForRosterFit boosts score if position is a team need
func (s *WaiverWireService) adjustScoreForRosterFit(gem WaiverGem, positionStrength map[string]float64, weakPositions []string) float64 {
	score := gem.BreakoutScore

	// Boost score if this position is weak on user's team
	for _, weakPos := range weakPositions {
		if gem.Position == weakPos {
			score += 20.0 // Significant boost for addressing weakness
			break
		}
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

// generatePersonalizedAnalysis creates team-specific AI analysis
func (s *WaiverWireService) generatePersonalizedAnalysis(ctx context.Context, gem *WaiverGem, roster []RosterPlayer, positionStrength map[string]float64) string {
	// Find current starters at same position
	currentStarters := []RosterPlayer{}
	for _, player := range roster {
		if player.Position == gem.Position && player.LineupSlot != "BE" && player.LineupSlot != "IR" {
			currentStarters = append(currentStarters, player)
		}
	}

	posAvg := positionStrength[gem.Position]

	prompt := fmt.Sprintf(`Analyze this waiver wire pickup for a fantasy team:

Player: %s (%s, %s)
Breakout Score: %.0f/100
EPA per Play: %.2f
Snap %%: %.1f
Trend: %s

Current %s starters on roster:
`, gem.PlayerName, gem.Position, gem.Team, gem.BreakoutScore, gem.EPAPerPlay, gem.SnapCountPct, gem.TargetShareTrend, gem.Position)

	if len(currentStarters) > 0 {
		for _, starter := range currentStarters {
			prompt += fmt.Sprintf("- %s (%.1f projected pts)\n", starter.Name, starter.ProjectedPoints)
		}
		prompt += fmt.Sprintf("Team's %s average: %.1f pts/week\n\n", gem.Position, posAvg)
	} else {
		prompt += fmt.Sprintf("No current starters at %s\n\n", gem.Position)
	}

	prompt += `In 2-3 sentences, explain:
1. How this player would improve their team
2. Whether they should start immediately or be a bench stash
3. Specific roster move recommendation (who to drop/bench)`

	response, err := s.gemini.Generate(ctx, prompt)
	if err != nil {
		return gem.AIAnalysis // Fallback to generic analysis
	}

	return response
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// analyzeBreakoutPotential performs comprehensive analysis on a player
func (s *WaiverWireService) analyzeBreakoutPotential(ctx context.Context, player models.Player, season, currentWeek int) *WaiverGem {
	gem := &WaiverGem{
		PlayerName: player.Name,
		Position:   player.Position,
		Team:       player.Team,
	}

	// TEMPORARY: Use player collection data instead of plays collection for speed
	// The plays collection queries are too slow even with indexes (scanning millions of records)
	// We'll use aggregated stats from the players collection which is much faster

	// Use player's aggregated stats (these are already calculated in players collection)
	gem.LastThreeGames = []GameStats{} // Empty for now - would need different data source

	// Get real snap count from Sleeper API
	// Try last 3 weeks to find the most recent game this player played
	gem.SnapCountPct = 0.0
	for week := 10; week >= 8; week-- {
		snapPct, err := s.sleeperClient.GetPlayerSnapCount(ctx, player.Name, "2025", week)
		if err == nil && snapPct > 0 {
			gem.SnapCountPct = snapPct
			break // Found recent snap data
		}
	}

	// Get EPA per play from plays collection for 2025 season (using player name)
	gem.EPAPerPlay = s.getPlayerEPAPerPlay(ctx, player.Name, 2025)

	// Set default trends without expensive query
	gem.TargetShareTrend = "stable"
	gem.TrendingUp = false

	// Infer depth chart status from snap count percentage
	if gem.SnapCountPct >= 70 {
		gem.DepthChartStatus = "starter"
	} else if gem.SnapCountPct >= 40 {
		gem.DepthChartStatus = "rotational"
	} else if gem.SnapCountPct > 0 {
		gem.DepthChartStatus = "backup"
	} else {
		gem.DepthChartStatus = "unknown"
	}

	gem.UpcomingSchedule = "average" // Default - would need schedule API
	gem.ScheduleRank = 16            // Default middle rank

	// Calculate breakout score (0-100)
	gem.BreakoutScore = s.calculateBreakoutScore(gem)

	// Determine recommendation
	gem.Recommendation = s.determineRecommendation(gem.BreakoutScore)

	return gem
}

// formatPlayerNameForNFLVerse converts "FirstName LastName" to "F.LastName" format
// to match NFLverse play-by-play data naming convention
func formatPlayerNameForNFLVerse(fullName string) string {
	parts := strings.Fields(fullName)
	if len(parts) < 2 {
		return fullName // Return as-is if can't parse
	}

	firstName := parts[0]
	lastName := parts[len(parts)-1]

	// Handle names like "Patrick Mahomes II" -> "P.Mahomes"
	// or "JuJu Smith-Schuster" -> "J.Smith-Schuster"
	initial := string(firstName[0])
	return fmt.Sprintf("%s.%s", initial, lastName)
}

// getPlayerEPAPerPlay calculates EPA per play from plays collection for recent weeks
func (s *WaiverWireService) getPlayerEPAPerPlay(ctx context.Context, playerName string, season int) float64 {
	// Calculate from plays collection using recent weeks (6-10) with timeout
	// NFLverse uses abbreviated names like "K.Murray" instead of "Kyler Murray"
	abbreviatedName := formatPlayerNameForNFLVerse(playerName)

	queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"season": season,
			"week":   bson.M{"$gte": 6, "$lte": 10}, // Recent 5 weeks
			"$or": []bson.M{
				{"passer_player_name": abbreviatedName},
				{"rusher_player_name": abbreviatedName},
				{"receiver_player_name": abbreviatedName},
			},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":        nil,
			"total_epa":  bson.M{"$sum": "$epa"},
			"play_count": bson.M{"$sum": 1},
		}}},
	}

	cursor, err := s.db.Collection("plays").Aggregate(queryCtx, pipeline)
	if err != nil {
		fmt.Printf("EPA query error for %s: %v\n", playerName, err)
		return 0.0
	}
	defer cursor.Close(ctx)

	var result struct {
		TotalEPA  float64 `bson:"total_epa"`
		PlayCount int     `bson:"play_count"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err == nil && result.PlayCount > 0 {
			epaPerPlay := result.TotalEPA / float64(result.PlayCount)
			fmt.Printf("EPA for %s (%s): %.3f (%d plays)\n", playerName, abbreviatedName, epaPerPlay, result.PlayCount)
			return epaPerPlay
		}
	}

	fmt.Printf("No EPA data for %s (looked for %s)\n", playerName, abbreviatedName)
	return 0.0
}

// getRecentGameStats fetches last N games with snap counts and target share
func (s *WaiverWireService) getRecentGameStats(ctx context.Context, nflID, position string, season, currentWeek, numGames int) []GameStats {
	var matchCondition bson.M

	switch position {
	case "QB":
		matchCondition = bson.M{"passer_player_id": nflID}
	case "RB":
		matchCondition = bson.M{
			"$or": []bson.M{
				{"rusher_player_id": nflID},
				{"receiver_player_id": nflID},
			},
		}
	case "WR", "TE":
		matchCondition = bson.M{"receiver_player_id": nflID}
	default:
		return nil
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"season": season,
			"week":   bson.M{"$lt": currentWeek, "$gte": currentWeek - numGames - 2},
		}}},
		{{Key: "$match", Value: matchCondition}},
		{{Key: "$group", Value: bson.M{
			"_id":      "$week",
			"opponent": bson.M{"$first": "$defense_team"},
			"targets": bson.M{"$sum": bson.M{
				"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$receiver_player_id", nflID}},
					1,
					0,
				},
			}},
			"receptions": bson.M{"$sum": bson.M{
				"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$receiver_player_id", nflID}},
					1,
					0,
				},
			}},
			"rec_yards": bson.M{"$sum": bson.M{
				"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$receiver_player_id", nflID}},
					"$yards",
					0,
				},
			}},
			"rec_tds": bson.M{"$sum": bson.M{
				"$cond": []interface{}{
					bson.M{"$and": []interface{}{
						bson.M{"$eq": []interface{}{"$receiver_player_id", nflID}},
						"$touchdown",
					}},
					1,
					0,
				},
			}},
			"rush_yards": bson.M{"$sum": bson.M{
				"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$rusher_player_id", nflID}},
					"$yards",
					0,
				},
			}},
			"rush_tds": bson.M{"$sum": bson.M{
				"$cond": []interface{}{
					bson.M{"$and": []interface{}{
						bson.M{"$eq": []interface{}{"$rusher_player_id", nflID}},
						"$touchdown",
					}},
					1,
					0,
				},
			}},
			"total_plays": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.M{"_id": -1}}},
		{{Key: "$limit", Value: numGames}},
	}

	cursor, err := s.db.Collection("plays").Aggregate(ctx, pipeline)
	if err != nil {
		return nil
	}
	defer cursor.Close(ctx)

	var games []GameStats
	for cursor.Next(ctx) {
		var result struct {
			Week       int    `bson:"_id"`
			Opponent   string `bson:"opponent"`
			Targets    int    `bson:"targets"`
			Receptions int    `bson:"receptions"`
			RecYards   int    `bson:"rec_yards"`
			RecTDs     int    `bson:"rec_tds"`
			RushYards  int    `bson:"rush_yards"`
			RushTDs    int    `bson:"rush_tds"`
			TotalPlays int    `bson:"total_plays"`
		}

		if err := cursor.Decode(&result); err != nil {
			continue
		}

		// Calculate fantasy points (PPR)
		fantasyPts := float64(result.RecYards+result.RushYards)*0.1 +
			float64(result.RecTDs+result.RushTDs)*6.0 +
			float64(result.Receptions)*1.0

		// Build production string
		production := ""
		if result.Receptions > 0 {
			production = fmt.Sprintf("%d rec, %d yds", result.Receptions, result.RecYards)
			if result.RecTDs > 0 {
				production += fmt.Sprintf(", %d TD", result.RecTDs)
			}
		} else if result.RushYards > 0 {
			production = fmt.Sprintf("%d rush yds", result.RushYards)
			if result.RushTDs > 0 {
				production += fmt.Sprintf(", %d TD", result.RushTDs)
			}
		}

		// Estimate snap percentage (plays involved / ~60 offensive plays per game)
		snapPct := float64(result.TotalPlays) / 60.0 * 100
		if snapPct > 100 {
			snapPct = 100
		}

		// Estimate target share (targets / ~30 team pass attempts)
		targetShare := float64(result.Targets) / 30.0 * 100
		if targetShare > 100 {
			targetShare = 100
		}

		games = append(games, GameStats{
			Week:          result.Week,
			Opponent:      result.Opponent,
			SnapPct:       snapPct,
			Targets:       result.Targets,
			TargetShare:   targetShare,
			Production:    production,
			FantasyPoints: fantasyPts,
		})
	}

	return games
}

// analyzeTargetShareTrend determines if usage is increasing
func (s *WaiverWireService) analyzeTargetShareTrend(games []GameStats) string {
	if len(games) < 2 {
		return "insufficient data"
	}

	// Check if most recent game has higher target share than average
	if len(games) >= 3 {
		recent := games[0].TargetShare
		older := (games[1].TargetShare + games[2].TargetShare) / 2

		if recent > older*1.2 {
			return "increasing"
		} else if recent < older*0.8 {
			return "decreasing"
		}
	}

	return "stable"
}

// checkDepthChartStatus checks for injured starters ahead on depth chart
func (s *WaiverWireService) checkDepthChartStatus(ctx context.Context, player models.Player, season int) string {
	// Check if any teammates at same position are injured
	cursor, err := s.db.Collection("players").Find(ctx, bson.M{
		"team":     player.Team,
		"position": player.Position,
		"season":   season,
		"status":   "INA", // Injured/Inactive
	})

	if err != nil {
		return "unknown"
	}
	defer cursor.Close(ctx)

	var injuredPlayers []models.Player
	cursor.All(ctx, &injuredPlayers)

	if len(injuredPlayers) > 0 {
		return fmt.Sprintf("Starter injured (%s)", injuredPlayers[0].Name)
	}

	return "Normal role"
}

// analyzeUpcomingSchedule looks at next 3 opponents' defensive strength
func (s *WaiverWireService) analyzeUpcomingSchedule(ctx context.Context, team, position string, season, currentWeek int) ScheduleAnalysis {
	analysis := ScheduleAnalysis{
		NextThreeOpponents: []string{},
		AvgDefensiveRank:   16,
		Difficulty:         "average",
	}

	// Get next 3 games
	cursor, err := s.db.Collection("games").Find(ctx, bson.M{
		"season": season,
		"week":   bson.M{"$gte": currentWeek, "$lte": currentWeek + 2},
		"$or": []bson.M{
			{"home_team": team},
			{"away_team": team},
		},
	})

	if err != nil {
		return analysis
	}
	defer cursor.Close(ctx)

	var games []models.Game
	cursor.All(ctx, &games)

	totalDefensiveEPA := 0.0
	for _, game := range games {
		opponent := game.AwayTeam
		if opponent == team {
			opponent = game.HomeTeam
		}
		analysis.NextThreeOpponents = append(analysis.NextThreeOpponents, opponent)

		// Get opponent's defensive EPA vs this position
		defEPA := s.getDefensiveEPA(ctx, opponent, position, season, currentWeek)
		totalDefensiveEPA += defEPA
	}

	if len(games) > 0 {
		avgEPA := totalDefensiveEPA / float64(len(games))

		// Rank based on EPA (lower EPA = better defense = harder matchup)
		if avgEPA < -0.1 {
			analysis.Difficulty = "difficult"
			analysis.AvgDefensiveRank = 8
		} else if avgEPA < 0 {
			analysis.Difficulty = "average"
			analysis.AvgDefensiveRank = 16
		} else if avgEPA > 0.1 {
			analysis.Difficulty = "favorable"
			analysis.AvgDefensiveRank = 28
		} else {
			analysis.Difficulty = "average"
			analysis.AvgDefensiveRank = 18
		}
	}

	return analysis
}

// getDefensiveEPA calculates how good a defense is vs a position
func (s *WaiverWireService) getDefensiveEPA(ctx context.Context, defenseTeam, position string, season, currentWeek int) float64 {
	var matchCondition bson.M

	switch position {
	case "QB":
		matchCondition = bson.M{"passer_player_id": bson.M{"$ne": ""}}
	case "RB":
		matchCondition = bson.M{"rusher_player_id": bson.M{"$ne": ""}}
	case "WR", "TE":
		matchCondition = bson.M{"receiver_player_id": bson.M{"$ne": ""}}
	default:
		return 0
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"season":       season,
			"week":         bson.M{"$lt": currentWeek},
			"defense_team": defenseTeam,
		}}},
		{{Key: "$match", Value: matchCondition}},
		{{Key: "$group", Value: bson.M{
			"_id":     nil,
			"avg_epa": bson.M{"$avg": "$epa"},
		}}},
	}

	cursor, err := s.db.Collection("plays").Aggregate(ctx, pipeline)
	if err != nil {
		return 0
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var result struct {
			AvgEPA float64 `bson:"avg_epa"`
		}
		cursor.Decode(&result)
		return result.AvgEPA
	}

	return 0
}

// calculateBreakoutScore computes 0-100 score based on all factors
func (s *WaiverWireService) calculateBreakoutScore(gem *WaiverGem) float64 {
	score := 0.0

	// EPA component (0-25 points)
	if gem.EPAPerPlay > 0.3 {
		score += 25
	} else if gem.EPAPerPlay > 0.2 {
		score += 20
	} else if gem.EPAPerPlay > 0.1 {
		score += 15
	} else if gem.EPAPerPlay > 0 {
		score += 10
	}

	// Trend component (0-25 points)
	if gem.TargetShareTrend == "increasing" {
		score += 25
	} else if gem.TargetShareTrend == "stable" {
		score += 10
	}

	// Snap count component (0-20 points)
	if gem.SnapCountPct > 70 {
		score += 20
	} else if gem.SnapCountPct > 50 {
		score += 15
	} else if gem.SnapCountPct > 30 {
		score += 10
	}

	// Depth chart opportunity (0-15 points)
	if strings.Contains(gem.DepthChartStatus, "injured") {
		score += 15
	} else if strings.Contains(gem.DepthChartStatus, "increased") {
		score += 10
	}

	// Schedule component (0-15 points)
	if gem.UpcomingSchedule == "favorable" {
		score += 15
	} else if gem.UpcomingSchedule == "average" {
		score += 8
	} else {
		score += 3
	}

	// Recent performance momentum
	if len(gem.LastThreeGames) >= 2 {
		if gem.LastThreeGames[0].FantasyPoints > gem.LastThreeGames[1].FantasyPoints {
			score += 5 // Bonus for improving
		}
	}

	return score
}

// determineRecommendation maps score to action
func (s *WaiverWireService) determineRecommendation(score float64) string {
	if score >= 80 {
		return "🔥 Must Add"
	} else if score >= 70 {
		return "⭐ Strong Add"
	} else if score >= 55 {
		return "📈 Good Add"
	} else if score >= 40 {
		return "👀 Monitor"
	}
	return "❌ Pass"
}

// calculatePlayerEPA gets EPA for recent weeks
func (s *WaiverWireService) calculatePlayerEPA(ctx context.Context, playerID string) float64 {
	filter := bson.M{
		"$or": []bson.M{
			{"passer_player_id": playerID},
			{"rusher_player_id": playerID},
			{"receiver_player_id": playerID},
		},
		"season": 2024,
	}

	cursor, err := s.db.Collection("plays").Find(ctx, filter)
	if err != nil {
		return 0.0
	}
	defer cursor.Close(ctx)

	var plays []models.Play
	if err := cursor.All(ctx, &plays); err != nil {
		return 0.0
	}

	if len(plays) == 0 {
		return 0.0
	}

	totalEPA := 0.0
	for _, play := range plays {
		totalEPA += play.EPA
	}

	return totalEPA / float64(len(plays))
}

// generateAIAnalysis creates comprehensive AI analysis
func (s *WaiverWireService) generateAIAnalysis(ctx context.Context, gem *WaiverGem) string {
	var recentPerf strings.Builder
	for i, game := range gem.LastThreeGames {
		if i >= 3 {
			break
		}
		recentPerf.WriteString(fmt.Sprintf("Week %d vs %s: %s (%.1f pts, %.0f%% snaps)\n",
			game.Week, game.Opponent, game.Production, game.FantasyPoints, game.SnapPct))
	}

	prompt := fmt.Sprintf(`You are an expert fantasy football waiver wire analyst. Analyze this breakout candidate:

Player: %s (%s - %s)
Breakout Score: %.0f/100

KEY METRICS:
- EPA per play: %.3f (efficiency)
- Snap Count: %.0f%% (recent usage)
- Target Share Trend: %s
- Depth Chart: %s
- Upcoming Schedule: %s (next 3 weeks rank #%d)

RECENT PERFORMANCE:
%s

Provide a 2-3 sentence analysis covering:
1. Why this player has breakout potential NOW
2. The specific opportunity (injury, role change, or favorable matchups)
3. A clear action recommendation for fantasy managers

Be data-driven, concise, and actionable.`,
		gem.PlayerName, gem.Position, gem.Team,
		gem.BreakoutScore,
		gem.EPAPerPlay,
		gem.SnapCountPct,
		gem.TargetShareTrend,
		gem.DepthChartStatus,
		gem.UpcomingSchedule, gem.ScheduleRank,
		recentPerf.String(),
	)

	response, err := s.gemini.GenerateWithRetry(ctx, prompt, 2)
	if err != nil {
		return "AI analysis unavailable"
	}

	return response
}
