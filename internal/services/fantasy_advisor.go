package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/ai-atl/nfl-platform/internal/models"
	"github.com/ai-atl/nfl-platform/pkg/gemini"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type FantasyAdvisorService struct {
	db          *mongo.Database
	gemini      *gemini.Client
	dataService *DataService
}

func NewFantasyAdvisorService(db *mongo.Database) *FantasyAdvisorService {
	return &FantasyAdvisorService{
		db:          db,
		gemini:      gemini.NewClient(),
		dataService: NewDataService(db),
	}
}

// PlayerComparison contains enriched data for comparing two players
type PlayerComparison struct {
	PlayerAName    string
	PlayerBName    string
	PlayerAData    *EnrichedPlayerData
	PlayerBData    *EnrichedPlayerData
	Recommendation string // "A" or "B"
	Confidence     int    // 0-100
	Reasoning      string
}

// EnrichedPlayerData contains all the data needed for AI fantasy advice
type EnrichedPlayerData struct {
	// Basic info from ESPN
	Name            string
	Position        string
	Team            string
	ProjectedPoints float64
	SeasonAverage   float64
	InjuryStatus    string
	IsInjured       bool

	// Database enrichments
	RecentGames      []GamePerformance
	AvgEPA           float64
	PlayerTrend      string // "hot", "cold", "neutral"
	TrendDescription string
	OpponentTeam     string
	OpponentRank     int // Defensive rank vs this position (1=best, 32=worst)
	MatchupAnalysis  string
}

type GamePerformance struct {
	Week           int
	Opponent       string
	PassingYards   int
	PassingTDs     int
	Interceptions  int
	RushingYards   int
	RushingTDs     int
	Receptions     int
	Targets        int
	ReceivingYards int
	ReceivingTDs   int
	FantasyPoints  float64
	EPA            float64
}

// GetStartSitAdvice provides AI-powered start/sit recommendations with database enrichment
func (s *FantasyAdvisorService) GetStartSitAdvice(ctx context.Context, playerAName, playerAPos, playerATeam string, playerAProj, playerASeason float64, playerAInj bool, playerAInjStatus string,
	playerBName, playerBPos, playerBTeam string, playerBProj, playerBSeason float64, playerBInj bool, playerBInjStatus string) (*PlayerComparison, error) {

	currentSeason := 2024 // TODO: Make dynamic
	currentWeek := 10     // TODO: Calculate from current date

	// Enrich Player A
	enrichedA := s.enrichPlayerData(ctx, playerAName, playerAPos, playerATeam, playerAProj, playerASeason, playerAInj, playerAInjStatus, currentSeason, currentWeek)

	// Enrich Player B
	enrichedB := s.enrichPlayerData(ctx, playerBName, playerBPos, playerBTeam, playerBProj, playerBSeason, playerBInj, playerBInjStatus, currentSeason, currentWeek)

	// Build comprehensive prompt with database context
	prompt := s.buildComparisonPrompt(enrichedA, enrichedB)

	// Get AI recommendation
	response, err := s.gemini.GenerateWithRetry(ctx, prompt, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AI recommendation: %w", err)
	}

	// Parse response
	comparison := &PlayerComparison{
		PlayerAName: playerAName,
		PlayerBName: playerBName,
		PlayerAData: enrichedA,
		PlayerBData: enrichedB,
	}

	s.parseAIResponse(response, comparison)

	return comparison, nil
}

// enrichPlayerData fetches all relevant data from MongoDB
func (s *FantasyAdvisorService) enrichPlayerData(ctx context.Context, name, position, team string, projPoints, seasonAvg float64, injured bool, injStatus string, season, currentWeek int) *EnrichedPlayerData {
	enriched := &EnrichedPlayerData{
		Name:            name,
		Position:        position,
		Team:            team,
		ProjectedPoints: projPoints,
		SeasonAverage:   seasonAvg,
		IsInjured:       injured,
		InjuryStatus:    injStatus,
	}

	// Find player in database
	player, err := s.findPlayerByName(ctx, name, team, season)
	if err != nil {
		// Player not found in DB - return ESPN data only
		return enriched
	}

	// Get recent game performances (last 5 games)
	recentGames, avgEPA := s.getRecentGamePerformances(ctx, player.NFLID, position, season, currentWeek, 5)
	enriched.RecentGames = recentGames
	enriched.AvgEPA = avgEPA

	// Analyze player trend
	enriched.PlayerTrend, enriched.TrendDescription = s.analyzePlayerTrend(recentGames)

	// Get next opponent and defensive matchup
	opponent := s.getNextOpponent(ctx, team, season, currentWeek)
	if opponent != "" {
		enriched.OpponentTeam = opponent
		rank, analysis := s.getDefensiveMatchup(ctx, opponent, position, season, currentWeek)
		enriched.OpponentRank = rank
		enriched.MatchupAnalysis = analysis
	}

	return enriched
}

// findPlayerByName searches for a player by name and team
func (s *FantasyAdvisorService) findPlayerByName(ctx context.Context, name, team string, season int) (*models.Player, error) {
	// Try exact match first
	var player models.Player
	err := s.db.Collection("players").FindOne(ctx, bson.M{
		"name":   name,
		"team":   team,
		"season": season,
	}).Decode(&player)

	if err == nil {
		return &player, nil
	}

	// Try fuzzy match
	err = s.db.Collection("players").FindOne(ctx, bson.M{
		"name":   bson.M{"$regex": fmt.Sprintf(".*%s.*", name), "$options": "i"},
		"team":   team,
		"season": season,
	}).Decode(&player)

	return &player, err
}

// getRecentGamePerformances fetches last N games for a player from plays collection
func (s *FantasyAdvisorService) getRecentGamePerformances(ctx context.Context, nflID, position string, season, currentWeek, numGames int) ([]GamePerformance, float64) {
	// Build position-specific match condition
	var playerMatch bson.M
	switch position {
	case "QB":
		playerMatch = bson.M{"passer_player_id": nflID}
	case "RB":
		playerMatch = bson.M{
			"$or": []bson.M{
				{"rusher_player_id": nflID},
				{"receiver_player_id": nflID},
			},
		}
	case "WR", "TE":
		playerMatch = bson.M{"receiver_player_id": nflID}
	default:
		return nil, 0
	}

	// Aggregate plays by week
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"season": season,
			"week":   bson.M{"$lt": currentWeek, "$gte": bson.M{"$max": []interface{}{1, currentWeek - 6}}},
		}}},
		{{Key: "$match", Value: playerMatch}},
		{{Key: "$group", Value: bson.M{
			"_id":      "$week",
			"opponent": bson.M{"$first": "$defense_team"},
			"passing_yards": bson.M{"$sum": bson.M{
				"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$passer_player_id", nflID}},
					"$yards",
					0,
				},
			}},
			"passing_tds": bson.M{"$sum": bson.M{
				"$cond": []interface{}{
					bson.M{"$and": []interface{}{
						bson.M{"$eq": []interface{}{"$passer_player_id", nflID}},
						"$touchdown",
					}},
					1,
					0,
				},
			}},
			"interceptions": bson.M{"$sum": bson.M{
				"$cond": []interface{}{
					bson.M{"$and": []interface{}{
						bson.M{"$eq": []interface{}{"$passer_player_id", nflID}},
						"$interception",
					}},
					1,
					0,
				},
			}},
			"rushing_yards": bson.M{"$sum": bson.M{
				"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$rusher_player_id", nflID}},
					"$yards",
					0,
				},
			}},
			"rushing_tds": bson.M{"$sum": bson.M{
				"$cond": []interface{}{
					bson.M{"$and": []interface{}{
						bson.M{"$eq": []interface{}{"$rusher_player_id", nflID}},
						"$touchdown",
					}},
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
			"targets": bson.M{"$sum": bson.M{
				"$cond": []interface{}{
					bson.M{"$ne": []interface{}{"$receiver_player_id", ""}},
					1,
					0,
				},
			}},
			"receiving_yards": bson.M{"$sum": bson.M{
				"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$receiver_player_id", nflID}},
					"$yards",
					0,
				},
			}},
			"receiving_tds": bson.M{"$sum": bson.M{
				"$cond": []interface{}{
					bson.M{"$and": []interface{}{
						bson.M{"$eq": []interface{}{"$receiver_player_id", nflID}},
						"$touchdown",
					}},
					1,
					0,
				},
			}},
			"avg_epa": bson.M{"$avg": "$epa"},
		}}},
		{{Key: "$sort", Value: bson.M{"_id": -1}}},
		{{Key: "$limit", Value: numGames}},
	}

	cursor, err := s.db.Collection("plays").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0
	}
	defer cursor.Close(ctx)

	var games []GamePerformance
	totalEPA := 0.0
	epaCount := 0

	for cursor.Next(ctx) {
		var result struct {
			Week           int     `bson:"_id"`
			Opponent       string  `bson:"opponent"`
			PassingYards   int     `bson:"passing_yards"`
			PassingTDs     int     `bson:"passing_tds"`
			Interceptions  int     `bson:"interceptions"`
			RushingYards   int     `bson:"rushing_yards"`
			RushingTDs     int     `bson:"rushing_tds"`
			Receptions     int     `bson:"receptions"`
			Targets        int     `bson:"targets"`
			ReceivingYards int     `bson:"receiving_yards"`
			ReceivingTDs   int     `bson:"receiving_tds"`
			AvgEPA         float64 `bson:"avg_epa"`
		}

		if err := cursor.Decode(&result); err != nil {
			continue
		}

		// Calculate PPR fantasy points
		fantasyPoints := s.calculateFantasyPoints(result.PassingYards, result.PassingTDs, result.Interceptions,
			result.RushingYards, result.RushingTDs, result.ReceivingYards, result.ReceivingTDs, result.Receptions)

		games = append(games, GamePerformance{
			Week:           result.Week,
			Opponent:       result.Opponent,
			PassingYards:   result.PassingYards,
			PassingTDs:     result.PassingTDs,
			Interceptions:  result.Interceptions,
			RushingYards:   result.RushingYards,
			RushingTDs:     result.RushingTDs,
			Receptions:     result.Receptions,
			Targets:        result.Targets,
			ReceivingYards: result.ReceivingYards,
			ReceivingTDs:   result.ReceivingTDs,
			FantasyPoints:  fantasyPoints,
			EPA:            result.AvgEPA,
		})

		totalEPA += result.AvgEPA
		epaCount++
	}

	avgEPA := 0.0
	if epaCount > 0 {
		avgEPA = totalEPA / float64(epaCount)
	}

	return games, avgEPA
}

// calculateFantasyPoints uses standard PPR scoring
func (s *FantasyAdvisorService) calculateFantasyPoints(passYards, passTDs, ints, rushYards, rushTDs, recYards, recTDs, receptions int) float64 {
	points := 0.0

	// Passing (1 pt per 25 yards, 4 pts per TD, -2 per INT)
	points += float64(passYards) * 0.04
	points += float64(passTDs) * 4.0
	points -= float64(ints) * 2.0

	// Rushing (1 pt per 10 yards, 6 pts per TD)
	points += float64(rushYards) * 0.1
	points += float64(rushTDs) * 6.0

	// Receiving (1 pt per 10 yards, 6 pts per TD, 1 pt per reception for PPR)
	points += float64(recYards) * 0.1
	points += float64(recTDs) * 6.0
	points += float64(receptions) * 1.0

	return points
}

// analyzePlayerTrend determines if player is hot, cold, or neutral
func (s *FantasyAdvisorService) analyzePlayerTrend(games []GamePerformance) (string, string) {
	if len(games) < 2 {
		return "neutral", "Limited recent data available"
	}

	// Calculate average of last 3 games
	numRecent := 3
	if len(games) < numRecent {
		numRecent = len(games)
	}

	recentAvg := 0.0
	for i := 0; i < numRecent; i++ {
		recentAvg += games[i].FantasyPoints
	}
	recentAvg /= float64(numRecent)

	// Check for upward trend (each game better than previous)
	if numRecent >= 3 {
		trending := true
		for i := 0; i < numRecent-1; i++ {
			if games[i].FantasyPoints <= games[i+1].FantasyPoints {
				trending = false
				break
			}
		}
		if trending && recentAvg > 15 {
			return "hot", fmt.Sprintf("🔥 On fire! Averaging %.1f pts with upward trend", recentAvg)
		}
	}

	// Classify based on recent average
	if recentAvg >= 18 {
		return "hot", fmt.Sprintf("🔥 Hot streak - averaging %.1f pts over last %d games", recentAvg, numRecent)
	} else if recentAvg <= 8 {
		return "cold", fmt.Sprintf("❄️ Cold streak - only %.1f pts per game recently", recentAvg)
	}

	return "neutral", fmt.Sprintf("📊 Averaging %.1f pts over last %d games", recentAvg, numRecent)
}

// getNextOpponent finds the next opponent for a team
func (s *FantasyAdvisorService) getNextOpponent(ctx context.Context, team string, season, currentWeek int) string {
	var game models.Game
	err := s.db.Collection("games").FindOne(ctx, bson.M{
		"season": season,
		"week":   currentWeek,
		"$or": []bson.M{
			{"home_team": team},
			{"away_team": team},
		},
	}).Decode(&game)

	if err != nil {
		return ""
	}

	if game.HomeTeam == team {
		return game.AwayTeam
	}
	return game.HomeTeam
}

// getDefensiveMatchup analyzes how good the opponent's defense is against this position
func (s *FantasyAdvisorService) getDefensiveMatchup(ctx context.Context, defenseTeam, position string, season, currentWeek int) (int, string) {
	// Query plays where this team was on defense
	var matchCondition bson.M

	switch position {
	case "QB":
		matchCondition = bson.M{"passer_player_id": bson.M{"$ne": ""}}
	case "RB":
		matchCondition = bson.M{"rusher_player_id": bson.M{"$ne": ""}}
	case "WR", "TE":
		matchCondition = bson.M{"receiver_player_id": bson.M{"$ne": ""}}
	default:
		return 16, "Unknown position"
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"season":       season,
			"week":         bson.M{"$lt": currentWeek},
			"defense_team": defenseTeam,
		}}},
		{{Key: "$match", Value: matchCondition}},
		{{Key: "$group", Value: bson.M{
			"_id":         nil,
			"total_plays": bson.M{"$sum": 1},
			"total_yards": bson.M{"$sum": "$yards"},
			"total_tds":   bson.M{"$sum": bson.M{"$cond": []interface{}{"$touchdown", 1, 0}}},
			"avg_epa":     bson.M{"$avg": "$epa"},
		}}},
	}

	cursor, err := s.db.Collection("plays").Aggregate(ctx, pipeline)
	if err != nil {
		return 16, "Matchup data unavailable"
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return 16, "Matchup data unavailable"
	}

	var result struct {
		TotalPlays int     `bson:"total_plays"`
		TotalYards int     `bson:"total_yards"`
		TotalTDs   int     `bson:"total_tds"`
		AvgEPA     float64 `bson:"avg_epa"`
	}

	if err := cursor.Decode(&result); err != nil {
		return 16, "Matchup data unavailable"
	}

	// Classify defense strength based on EPA
	var rank int
	var strength string

	if result.AvgEPA < -0.15 {
		rank = 3
		strength = "elite"
	} else if result.AvgEPA < -0.05 {
		rank = 8
		strength = "strong"
	} else if result.AvgEPA < 0.05 {
		rank = 16
		strength = "average"
	} else if result.AvgEPA < 0.15 {
		rank = 24
		strength = "weak"
	} else {
		rank = 30
		strength = "very weak"
	}

	positionStr := map[string]string{
		"QB": "quarterbacks",
		"RB": "running backs",
		"WR": "wide receivers",
		"TE": "tight ends",
	}[position]

	var analysis string
	if strength == "elite" || strength == "strong" {
		analysis = fmt.Sprintf("⚠️ Tough matchup: %s defense ranks #%d vs %s (%s, %.3f EPA)",
			defenseTeam, rank, positionStr, strength, result.AvgEPA)
	} else if strength == "weak" || strength == "very weak" {
		analysis = fmt.Sprintf("✅ Great matchup: %s defense ranks #%d vs %s (%s, %.3f EPA)",
			defenseTeam, rank, positionStr, strength, result.AvgEPA)
	} else {
		analysis = fmt.Sprintf("📊 Average matchup: %s defense ranks #%d vs %s (%.3f EPA)",
			defenseTeam, rank, positionStr, result.AvgEPA)
	}

	return rank, analysis
}

// buildComparisonPrompt creates a comprehensive prompt with database context
func (s *FantasyAdvisorService) buildComparisonPrompt(playerA, playerB *EnrichedPlayerData) string {
	var prompt strings.Builder

	prompt.WriteString("You are an expert fantasy football advisor with access to comprehensive NFL play-by-play data, recent game logs, and defensive matchup analysis.\n\n")
	prompt.WriteString("Compare these two players and recommend which one to START this week. Base your decision on:\n")
	prompt.WriteString("1. Recent performance trends (last 3-5 games)\n")
	prompt.WriteString("2. This week's defensive matchup quality\n")
	prompt.WriteString("3. Statistical efficiency (EPA - Expected Points Added)\n")
	prompt.WriteString("4. Health status and injury concerns\n")
	prompt.WriteString("5. Projected points and consistency\n\n")

	// Player A details
	prompt.WriteString(fmt.Sprintf("=== PLAYER A: %s ===\n", playerA.Name))
	prompt.WriteString(fmt.Sprintf("Position: %s | Team: %s\n", playerA.Position, playerA.Team))
	prompt.WriteString(fmt.Sprintf("ESPN Projected Points: %.1f\n", playerA.ProjectedPoints))
	prompt.WriteString(fmt.Sprintf("Season Average: %.1f PPG\n", playerA.SeasonAverage))
	prompt.WriteString(fmt.Sprintf("Health: %s", playerA.InjuryStatus))
	if playerA.IsInjured {
		prompt.WriteString(" ⚠️ INJURED")
	}
	prompt.WriteString("\n\n")

	if len(playerA.RecentGames) > 0 {
		prompt.WriteString(fmt.Sprintf("Recent Trend: %s\n", playerA.TrendDescription))
		prompt.WriteString(fmt.Sprintf("Average EPA: %.3f per play\n", playerA.AvgEPA))
		prompt.WriteString("Last 3 Games:\n")
		for i, game := range playerA.RecentGames {
			if i >= 3 {
				break
			}
			prompt.WriteString(fmt.Sprintf("  Week %d vs %s: %.1f pts", game.Week, game.Opponent, game.FantasyPoints))
			if playerA.Position == "QB" && game.PassingYards > 0 {
				prompt.WriteString(fmt.Sprintf(" (%d pass yds, %d TD, %d INT)", game.PassingYards, game.PassingTDs, game.Interceptions))
			} else if playerA.Position == "RB" {
				prompt.WriteString(fmt.Sprintf(" (%d rush yds, %d rush TD, %d rec)", game.RushingYards, game.RushingTDs, game.Receptions))
			} else if playerA.Position == "WR" || playerA.Position == "TE" {
				prompt.WriteString(fmt.Sprintf(" (%d rec, %d yds, %d TD)", game.Receptions, game.ReceivingYards, game.ReceivingTDs))
			}
			prompt.WriteString("\n")
		}
		prompt.WriteString("\n")
	}

	if playerA.MatchupAnalysis != "" {
		prompt.WriteString(fmt.Sprintf("This Week's Matchup: %s\n", playerA.MatchupAnalysis))
	}
	prompt.WriteString("\n")

	// Player B details
	prompt.WriteString(fmt.Sprintf("=== PLAYER B: %s ===\n", playerB.Name))
	prompt.WriteString(fmt.Sprintf("Position: %s | Team: %s\n", playerB.Position, playerB.Team))
	prompt.WriteString(fmt.Sprintf("ESPN Projected Points: %.1f\n", playerB.ProjectedPoints))
	prompt.WriteString(fmt.Sprintf("Season Average: %.1f PPG\n", playerB.SeasonAverage))
	prompt.WriteString(fmt.Sprintf("Health: %s", playerB.InjuryStatus))
	if playerB.IsInjured {
		prompt.WriteString(" ⚠️ INJURED")
	}
	prompt.WriteString("\n\n")

	if len(playerB.RecentGames) > 0 {
		prompt.WriteString(fmt.Sprintf("Recent Trend: %s\n", playerB.TrendDescription))
		prompt.WriteString(fmt.Sprintf("Average EPA: %.3f per play\n", playerB.AvgEPA))
		prompt.WriteString("Last 3 Games:\n")
		for i, game := range playerB.RecentGames {
			if i >= 3 {
				break
			}
			prompt.WriteString(fmt.Sprintf("  Week %d vs %s: %.1f pts", game.Week, game.Opponent, game.FantasyPoints))
			if playerB.Position == "QB" && game.PassingYards > 0 {
				prompt.WriteString(fmt.Sprintf(" (%d pass yds, %d TD, %d INT)", game.PassingYards, game.PassingTDs, game.Interceptions))
			} else if playerB.Position == "RB" {
				prompt.WriteString(fmt.Sprintf(" (%d rush yds, %d rush TD, %d rec)", game.RushingYards, game.RushingTDs, game.Receptions))
			} else if playerB.Position == "WR" || playerB.Position == "TE" {
				prompt.WriteString(fmt.Sprintf(" (%d rec, %d yds, %d TD)", game.Receptions, game.ReceivingYards, game.ReceivingTDs))
			}
			prompt.WriteString("\n")
		}
		prompt.WriteString("\n")
	}

	if playerB.MatchupAnalysis != "" {
		prompt.WriteString(fmt.Sprintf("This Week's Matchup: %s\n", playerB.MatchupAnalysis))
	}
	prompt.WriteString("\n")

	prompt.WriteString("=== YOUR TASK ===\n")
	prompt.WriteString("Provide your recommendation in EXACTLY this format:\n\n")
	prompt.WriteString("RECOMMENDATION: [A or B]\n")
	prompt.WriteString("CONFIDENCE: [number from 0-100]\n")
	prompt.WriteString("REASONING: [2-3 sentences explaining your choice, referencing specific stats, trends, and matchup quality]\n\n")
	prompt.WriteString("Be data-driven and concise. Reference specific numbers from the data above.")

	return prompt.String()
}

// parseAIResponse extracts structured data from AI response
func (s *FantasyAdvisorService) parseAIResponse(response string, comparison *PlayerComparison) {
	lines := strings.Split(response, "\n")

	// Default values
	comparison.Recommendation = "A"
	comparison.Confidence = 50
	comparison.Reasoning = response

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "RECOMMENDATION:") {
			rec := strings.TrimSpace(strings.TrimPrefix(line, "RECOMMENDATION:"))
			rec = strings.ToUpper(rec)
			if strings.Contains(rec, "B") {
				comparison.Recommendation = "B"
			} else {
				comparison.Recommendation = "A"
			}
		} else if strings.HasPrefix(line, "CONFIDENCE:") {
			confStr := strings.TrimSpace(strings.TrimPrefix(line, "CONFIDENCE:"))
			confStr = strings.ReplaceAll(confStr, "%", "")
			var conf int
			fmt.Sscanf(confStr, "%d", &conf)
			if conf >= 0 && conf <= 100 {
				comparison.Confidence = conf
			}
		} else if strings.HasPrefix(line, "REASONING:") {
			reasoning := strings.TrimSpace(strings.TrimPrefix(line, "REASONING:"))
			comparison.Reasoning = reasoning

			// Collect any additional reasoning lines
			for i := 0; i < len(lines); i++ {
				if strings.HasPrefix(lines[i], "REASONING:") {
					for j := i + 1; j < len(lines); j++ {
						nextLine := strings.TrimSpace(lines[j])
						if nextLine != "" && !strings.HasPrefix(nextLine, "RECOMMENDATION:") && !strings.HasPrefix(nextLine, "CONFIDENCE:") {
							comparison.Reasoning += " " + nextLine
						} else {
							break
						}
					}
					break
				}
			}
		}
	}
}
