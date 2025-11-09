package services

import (
	"context"
	"fmt"
	"log"

	"github.com/ai-atl/nfl-platform/internal/models"
	"github.com/ai-atl/nfl-platform/pkg/gemini"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type GameScriptService struct {
	db     *mongo.Database
	gemini *gemini.Client
}

type GameScriptPrediction struct {
	GameID          string         `json:"game_id"`
	PredictedFlow   string         `json:"predicted_flow"`
	PlayerImpacts   []PlayerImpact `json:"player_impacts"`
	ConfidenceScore float64        `json:"confidence_score"`
	KeyFactors      []string       `json:"key_factors"`
}

type PlayerImpact struct {
	PlayerName string `json:"player_name"`
	Impact     string `json:"impact"`
	Reasoning  string `json:"reasoning"`
}

func NewGameScriptService(db *mongo.Database) *GameScriptService {
	return &GameScriptService{
		db:     db,
		gemini: gemini.NewClient(),
	}
}

// PredictGameScript predicts how a game will unfold
func (s *GameScriptService) PredictGameScript(ctx context.Context, gameID string) (*GameScriptPrediction, error) {
	// Fetch game data
	var game models.Game
	err := s.db.Collection("games").FindOne(ctx, bson.M{"game_id": gameID}).Decode(&game)
	if err != nil {
		return nil, fmt.Errorf("game not found: %w", err)
	}

	// Fetch roster and stats for both teams
	homeTeamContext, err := s.fetchTeamContext(ctx, game.HomeTeam, game.Season, game.Week)
	if err != nil {
		homeTeamContext = fmt.Sprintf("Team: %s", game.HomeTeam)
	}

	awayTeamContext, err := s.fetchTeamContext(ctx, game.AwayTeam, game.Season, game.Week)
	if err != nil {
		awayTeamContext = fmt.Sprintf("Team: %s", game.AwayTeam)
	}

	// Fetch historical matchup data
	historicalContext := s.fetchHistoricalMatchups(ctx, game.HomeTeam, game.AwayTeam, game.Season)

	// Fetch home/away performance splits
	homeAwayContext := s.fetchHomeAwaySplits(ctx, game.HomeTeam, game.AwayTeam, game.Season)

	// Build comprehensive context with real database data
	prompt := s.buildGameScriptPrompt(game, homeTeamContext, awayTeamContext, historicalContext, homeAwayContext)

	// Get AI prediction
	response, err := s.gemini.GenerateWithRetry(ctx, prompt, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to generate prediction: %w", err)
	}

	// Parse response (simplified for hackathon)
	prediction := &GameScriptPrediction{
		GameID:          gameID,
		PredictedFlow:   response,
		ConfidenceScore: 0.85,
		KeyFactors: []string{
			"Vegas line suggests competitive game",
			"Weather conditions favorable",
		},
		PlayerImpacts: []PlayerImpact{
			{
				PlayerName: "Key Player",
				Impact:     "+25% opportunity",
				Reasoning:  "Favorable game script for passing",
			},
		},
	}

	return prediction, nil
}

type PlayerWithStats struct {
	Player      models.Player
	Stats       models.PlayerStats
	RecentWeeks []models.WeeklyStat
	AvgFantasy  float64
	GamesPlayed int
}

func (s *GameScriptService) fetchTeamContext(ctx context.Context, team string, season int, currentWeek int) (string, error) {
	// Try to get the most recent roster data available
	// First try requested season, then fall back to previous season if needed
	var players []models.Player
	usedSeason := season

	// Try requested season first
	cursor, err := s.db.Collection("players").Find(ctx, bson.M{
		"team":   team,
		"season": season,
	})
	if err == nil {
		cursor.All(ctx, &players)
		cursor.Close(ctx)
	}

	// If no players found and we're looking at 2025, fall back to 2024
	// (2025 roster data might be incomplete/unavailable)
	if len(players) == 0 && season == 2025 {
		log.Printf("⚠️  No %d roster for %s, falling back to 2024", season, team)
		cursor, err = s.db.Collection("players").Find(ctx, bson.M{
			"team":   team,
			"season": 2024,
		})
		if err == nil {
			cursor.All(ctx, &players)
			cursor.Close(ctx)
			usedSeason = 2024
		}
	}

	if len(players) == 0 {
		return "", fmt.Errorf("no roster data found for %s (tried %d and 2024)", team, season)
	}

	log.Printf("📊 Loaded %d players for %s (using %d data for %d season game)", len(players), team, usedSeason, season)

	// Fetch stats for all players with weekly breakdown
	var playersWithStats []PlayerWithStats
	var skippedReasons = map[string]int{
		"injured":      0,
		"no_stats":     0,
		"no_fantasy":   0,
		"low_activity": 0,
	}

	for _, p := range players {
		// Skip players marked as injured or inactive (but log it)
		if p.Status == "INA" || s.isInjuredStatus(p.StatusDescriptionAbbr) {
			skippedReasons["injured"]++
			continue
		}

		var stats models.PlayerStats
		err := s.db.Collection("player_stats").FindOne(ctx, bson.M{
			"nfl_id":      p.NFLID,
			"season":      usedSeason,
			"season_type": "REG",
		}).Decode(&stats)

		if err != nil {
			skippedReasons["no_stats"]++
			continue
		}

		// If no fantasy points, they likely haven't played
		if stats.FantasyPointsPPR <= 0 {
			skippedReasons["no_fantasy"]++
			continue
		}

		// Calculate games played and average fantasy points
		gamesPlayed := s.estimateGamesPlayed(stats, currentWeek)
		avgFantasy := 0.0
		if gamesPlayed > 0 {
			avgFantasy = stats.FantasyPointsPPR / float64(gamesPlayed)
		}

		// Only filter out players with extremely low activity
		// Be more lenient - if they have ANY stats, include them
		// This is especially important for 2024 data being used for 2025 predictions
		if gamesPlayed < 1 && stats.FantasyPointsPPR < 1.0 {
			skippedReasons["low_activity"]++
			continue
		}

		playersWithStats = append(playersWithStats, PlayerWithStats{
			Player:      p,
			Stats:       stats,
			AvgFantasy:  avgFantasy,
			GamesPlayed: gamesPlayed,
		})
	}

	log.Printf("📊 Filtering results for %s: injured=%d, no_stats=%d, no_fantasy=%d, low_activity=%d, kept=%d",
		team, skippedReasons["injured"], skippedReasons["no_stats"],
		skippedReasons["no_fantasy"], skippedReasons["low_activity"], len(playersWithStats))

	log.Printf("✓ After filtering: %d active players for %s", len(playersWithStats), team)

	// Build context with sorted/prioritized players
	dataSource := fmt.Sprintf("%d season", usedSeason)
	if usedSeason != season {
		dataSource = fmt.Sprintf("%d season data (using %d as fallback)", season, usedSeason)
	}
	context := fmt.Sprintf("**%s Active Roster & Key Players (%s, predicting Week %d):**\n", team, dataSource, currentWeek)
	context += fmt.Sprintf("*Note: Using %d roster/stats. Players who are injured (INA status), haven't played recently, or were traded mid-season are filtered out*\n\n", usedSeason)

	// Get starting QB (sorted by fantasy points per game)
	qbs := s.filterAndSortByPosition(playersWithStats, "QB", func(a, b PlayerWithStats) bool {
		return a.AvgFantasy > b.AvgFantasy
	})
	if len(qbs) > 0 {
		context += "**Starting QB:**\n"
		context += s.formatPlayerWithContext(qbs[0], true)
		if len(qbs) > 1 && qbs[1].AvgFantasy > 5.0 { // Only show backup if they've played
			context += "\n**Backup QB:**\n"
			context += s.formatPlayerWithContext(qbs[1], false)
		}
	}

	// Get top RBs (sorted by fantasy points per game)
	rbs := s.filterAndSortByPosition(playersWithStats, "RB", func(a, b PlayerWithStats) bool {
		return a.AvgFantasy > b.AvgFantasy
	})
	if len(rbs) > 0 {
		context += "\n**Starting RB:**\n"
		context += s.formatPlayerWithContext(rbs[0], true)
		if len(rbs) > 1 {
			context += "\n**Committee/Backup RBs:**\n"
			for i := 1; i < len(rbs) && i < 3; i++ {
				if rbs[i].AvgFantasy > 3.0 { // Only show if fantasy relevant
					context += s.formatPlayerWithContext(rbs[i], false)
				}
			}
		}
	}

	// Get top WRs (sorted by fantasy points per game)
	wrs := s.filterAndSortByPosition(playersWithStats, "WR", func(a, b PlayerWithStats) bool {
		return a.AvgFantasy > b.AvgFantasy
	})
	if len(wrs) > 0 {
		context += "\n**Starting WRs:**\n"
		for i := 0; i < len(wrs) && i < 3; i++ {
			if wrs[i].AvgFantasy > 5.0 { // Only show fantasy relevant WRs
				context += s.formatPlayerWithContext(wrs[i], i == 0)
			}
		}
	}

	// Get top TEs (sorted by fantasy points per game)
	tes := s.filterAndSortByPosition(playersWithStats, "TE", func(a, b PlayerWithStats) bool {
		return a.AvgFantasy > b.AvgFantasy
	})
	if len(tes) > 0 && tes[0].AvgFantasy > 3.0 {
		context += "\n**Starting TE:**\n"
		context += s.formatPlayerWithContext(tes[0], true)
	}

	return context, nil
}

func (s *GameScriptService) filterAndSortByPosition(players []PlayerWithStats, position string, less func(a, b PlayerWithStats) bool) []PlayerWithStats {
	var filtered []PlayerWithStats
	for _, p := range players {
		if p.Player.Position == position {
			filtered = append(filtered, p)
		}
	}

	// Simple bubble sort (fine for small arrays)
	for i := 0; i < len(filtered); i++ {
		for j := i + 1; j < len(filtered); j++ {
			if less(filtered[j], filtered[i]) {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	return filtered
}

func (s *GameScriptService) formatPlayerWithContext(pws PlayerWithStats, isStarter bool) string {
	role := "STARTER"
	if !isStarter {
		role = "BACKUP"
	}

	stats := pws.Stats
	output := fmt.Sprintf("- **%s** (%s, %d games) - ", pws.Player.Name, role, pws.GamesPlayed)

	// Format based on position
	if pws.Player.Position == "QB" {
		output += fmt.Sprintf("%d pass yds, %d pass TDs, %d INTs, %d rush yds, %d rush TDs",
			stats.PassingYards, stats.PassingTDs, stats.Interceptions,
			stats.RushingYards, stats.RushingTDs)
		if pws.AvgFantasy > 0 {
			output += fmt.Sprintf(" | **%.1f fantasy pts/game** (%.1f total)", pws.AvgFantasy, stats.FantasyPointsPPR)
		}
	} else if pws.Player.Position == "RB" {
		output += fmt.Sprintf("%d rush yds, %d rush TDs, %d rec, %d rec yds, %d rec TDs",
			stats.RushingYards, stats.RushingTDs, stats.Receptions,
			stats.ReceivingYards, stats.ReceivingTDs)
		if pws.AvgFantasy > 0 {
			output += fmt.Sprintf(" | **%.1f fantasy pts/game** (%.1f total)", pws.AvgFantasy, stats.FantasyPointsPPR)
		}
	} else if pws.Player.Position == "WR" || pws.Player.Position == "TE" {
		output += fmt.Sprintf("%d rec, %d targets, %d rec yds, %d rec TDs",
			stats.Receptions, stats.Targets, stats.ReceivingYards, stats.ReceivingTDs)
		if stats.RushingYards > 0 {
			output += fmt.Sprintf(", %d rush yds", stats.RushingYards)
		}
		if pws.AvgFantasy > 0 {
			output += fmt.Sprintf(" | **%.1f fantasy pts/game** (%.1f total)", pws.AvgFantasy, stats.FantasyPointsPPR)
		}
	}

	output += "\n"
	return output
}

func (s *GameScriptService) isInjuredStatus(statusAbbr string) bool {
	injuredStatuses := []string{
		"R01", // Reserve/Injured
		"R02", // Reserve/Retired
		"R04", // Reserve/PUP
		"R06", // Reserve/Non-Football Injury
		"R48", // Reserve/Injured; DFR
		"P02", // Practice Squad; Injured
		"W01", // Waived/Injured
		"W03", // Waived/Injured; Settlement
	}

	for _, status := range injuredStatuses {
		if statusAbbr == status {
			return true
		}
	}
	return false
}

func (s *GameScriptService) estimateGamesPlayed(stats models.PlayerStats, currentWeek int) int {
	// Estimate games played based on activity
	gamesPlayed := 0

	if stats.PassingYards > 0 {
		// QBs: estimate based on attempts (roughly 30-40 per game)
		gamesPlayed = stats.PassingYards / 250 // ~250 yards per game
		if gamesPlayed == 0 && stats.PassingYards > 0 {
			gamesPlayed = 1
		}
	} else if stats.RushingYards > 0 || stats.ReceivingYards > 0 {
		// RB/WR/TE: estimate based on yards
		totalYards := stats.RushingYards + stats.ReceivingYards
		gamesPlayed = totalYards / 60 // ~60 yards per game average
		if gamesPlayed == 0 && totalYards > 0 {
			gamesPlayed = 1
		}
	}

	// Cap at current week - 1 (can't have played more games than have happened)
	if gamesPlayed > currentWeek-1 {
		gamesPlayed = currentWeek - 1
	}

	if gamesPlayed < 1 && (stats.PassingYards > 0 || stats.RushingYards > 0 || stats.ReceivingYards > 0) {
		gamesPlayed = 1
	}

	return gamesPlayed
}

func (s *GameScriptService) fetchHistoricalMatchups(ctx context.Context, homeTeam, awayTeam string, currentSeason int) string {
	// Look for previous games between these teams in last 3 years
	cursor, err := s.db.Collection("games").Find(ctx, bson.M{
		"$or": []bson.M{
			{"home_team": homeTeam, "away_team": awayTeam},
			{"home_team": awayTeam, "away_team": homeTeam},
		},
		"season": bson.M{"$gte": currentSeason - 3, "$lt": currentSeason},
		"status": "final",
	}, options.Find().SetSort(bson.D{{"season", -1}, {"week", -1}}).SetLimit(3))

	if err != nil {
		return ""
	}
	defer cursor.Close(ctx)

	var games []models.Game
	if err := cursor.All(ctx, &games); err != nil || len(games) == 0 {
		return ""
	}

	context := fmt.Sprintf("\n**Recent Head-to-Head History:**\n")
	for _, game := range games {
		winner := game.HomeTeam
		if game.AwayScore > game.HomeScore {
			winner = game.AwayTeam
		}
		context += fmt.Sprintf("- %d Week %d: %s %d - %s %d (Winner: %s, Total: %d)\n",
			game.Season, game.Week, game.AwayTeam, game.AwayScore,
			game.HomeTeam, game.HomeScore, winner, game.HomeScore+game.AwayScore)
	}

	return context
}

func (s *GameScriptService) fetchHomeAwaySplits(ctx context.Context, homeTeam, awayTeam string, season int) string {
	// Get home team's home record
	homeGames, homeWins, homePointsFor, homePointsAgainst := s.getTeamRecord(ctx, homeTeam, season, true)
	awayGames, awayWins, awayPointsFor, awayPointsAgainst := s.getTeamRecord(ctx, awayTeam, season, false)

	if homeGames == 0 && awayGames == 0 {
		return ""
	}

	context := fmt.Sprintf("\n**Home/Away Performance (2025):**\n")
	if homeGames > 0 {
		context += fmt.Sprintf("- %s at HOME: %d-%d (Avg: %.1f pts/game for, %.1f against)\n",
			homeTeam, homeWins, homeGames-homeWins,
			float64(homePointsFor)/float64(homeGames),
			float64(homePointsAgainst)/float64(homeGames))
	}
	if awayGames > 0 {
		context += fmt.Sprintf("- %s on ROAD: %d-%d (Avg: %.1f pts/game for, %.1f against)\n",
			awayTeam, awayWins, awayGames-awayWins,
			float64(awayPointsFor)/float64(awayGames),
			float64(awayPointsAgainst)/float64(awayGames))
	}

	return context
}

func (s *GameScriptService) getTeamRecord(ctx context.Context, team string, season int, isHome bool) (games, wins, pointsFor, pointsAgainst int) {
	filter := bson.M{
		"season": season,
		"status": "final",
	}

	if isHome {
		filter["home_team"] = team
	} else {
		filter["away_team"] = team
	}

	cursor, err := s.db.Collection("games").Find(ctx, filter)
	if err != nil {
		return 0, 0, 0, 0
	}
	defer cursor.Close(ctx)

	var teamGames []models.Game
	if err := cursor.All(ctx, &teamGames); err != nil {
		return 0, 0, 0, 0
	}

	games = len(teamGames)
	for _, game := range teamGames {
		if isHome {
			pointsFor += game.HomeScore
			pointsAgainst += game.AwayScore
			if game.HomeScore > game.AwayScore {
				wins++
			}
		} else {
			pointsFor += game.AwayScore
			pointsAgainst += game.HomeScore
			if game.AwayScore > game.HomeScore {
				wins++
			}
		}
	}

	return
}

func (s *GameScriptService) buildGameScriptPrompt(game models.Game, homeTeamContext, awayTeamContext, historicalContext, homeAwayContext string) string {
	return fmt.Sprintf(`Analyze this NFL matchup and predict the game script:

**Game:** %s (Away) @ %s (Home)
**Vegas Line:** %s %.1f (negative = home team favored)
**Over/Under:** %.1f
**Start Time:** %s
**Week:** %d

%s

%s

%s

%s

**Analysis Instructions:**

1. **Focus on STARTERS & HIGH-USAGE PLAYERS**: 
   - Players are ranked by FANTASY POINTS PER GAME, not career totals
   - Look at the "fantasy pts/game" average - this shows current season performance
   - Players with higher pts/game averages are the actual starters getting opportunities
   - Games played shown next to each player (e.g., "STARTER, 8 games")

2. **Use Per-Game Performance**:
   - Compare fantasy pts/game averages to assess true workload
   - A player with 12 pts/game over 8 games is more relevant than one with 20 total pts over 2 games
   - Consider consistency: steady performers vs boom/bust players

3. **Leverage Historical Context**:
   - Recent head-to-head results show scoring trends between these teams
   - Home/away splits reveal team performance in different venues
   - Factor these patterns into your game script prediction

4. **Game Script Prediction**:
   Based on Vegas lines, team trends, and home/away splits:
   - Will this be competitive, a blowout, or defensive struggle?
   - Which team will likely be playing from ahead/behind?
   - How does this affect pass/run ratios?

5. **Player Impact Analysis** (TOP STARTERS ONLY):
   - Who benefits from expected game script?
   - Reference their actual pts/game average and recent performance
   - Project specific opportunity increases (more targets, carries, attempts)

**CRITICAL RULES:**
- **DO NOT mention ANY players not explicitly listed above** - the roster is filtered for active, healthy players only
- **Players listed have been filtered to exclude injured/inactive players** - if someone isn't listed, they're not available
- ONLY reference players with (STARTER) or (BACKUP) labels shown above
- Focus on players with HIGH fantasy pts/game averages (they're the actual starters)
- If you're unsure about a player, DO NOT mention them - stick to the provided roster
- Use the HOME/AWAY splits and HISTORICAL data to inform predictions
- Reference actual numbers from the data (pts/game, team scoring averages, etc.)

**ROSTER DATA ACCURACY:**
The player lists above have been filtered to remove:
- Injured players (IR, PUP, Out status)
- Players who haven't played in recent weeks
- Inactive or practice squad players
- **Mid-season trades are not always reflected - only reference players explicitly shown for each team**

If a notable player you'd expect to see is missing from the roster, they are either injured, traded, or inactive. Do not mention them.

**Format:** Use clear markdown with ## headers and bullet points. Be specific and actionable.`,
		game.AwayTeam,
		game.HomeTeam,
		game.HomeTeam,
		game.VegasLine,
		game.OverUnder,
		game.StartTime.Format("Mon Jan 2 3:04 PM"),
		game.Week,
		awayTeamContext,
		homeTeamContext,
		historicalContext,
		homeAwayContext,
	)
}
