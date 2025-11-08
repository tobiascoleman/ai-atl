package services

import (
	"context"
	"fmt"

	"github.com/ai-atl/nfl-platform/internal/models"
	"github.com/ai-atl/nfl-platform/pkg/gemini"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type StreakDetectorService struct {
	db     *mongo.Database
	gemini *gemini.Client
}

type Streak struct {
	PlayerID      string  `json:"player_id"`
	PlayerName    string  `json:"player_name"`
	StreakType    string  `json:"streak_type"` // "over", "under", "hot", "cold"
	StatLine      string  `json:"stat_line"`
	GamesInStreak int     `json:"games_in_streak"`
	AIExplanation string  `json:"ai_explanation"`
	Confidence    float64 `json:"confidence"`
}

func NewStreakDetectorService(db *mongo.Database) *StreakDetectorService {
	return &StreakDetectorService{
		db:     db,
		gemini: gemini.NewClient(),
	}
}

// DetectStreaks identifies hot/cold streaks for a player
func (s *StreakDetectorService) DetectStreaks(ctx context.Context, playerID string, lookbackGames int) ([]Streak, error) {
	// Get player
	var player models.Player
	err := s.db.Collection("players").FindOne(ctx, bson.M{"nfl_id": playerID}).Decode(&player)
	if err != nil {
		return nil, fmt.Errorf("player not found: %w", err)
	}
	
	if len(player.WeeklyStats) < lookbackGames {
		lookbackGames = len(player.WeeklyStats)
	}
	
	// Analyze recent games
	recentGames := player.WeeklyStats
	if len(recentGames) > lookbackGames {
		recentGames = recentGames[len(recentGames)-lookbackGames:]
	}
	
	// Detect patterns
	streaks := []Streak{}
	
	// Check for performance streaks
	if performanceStreak := s.detectPerformanceStreak(recentGames); performanceStreak != nil {
		performanceStreak.PlayerID = player.NFLID
		performanceStreak.PlayerName = player.Name
		
		// Get AI explanation
		explanation, err := s.explainStreak(ctx, player, recentGames, *performanceStreak)
		if err == nil {
			performanceStreak.AIExplanation = explanation
		}
		
		streaks = append(streaks, *performanceStreak)
	}
	
	// Check for over/under streaks on various stat lines
	if overStreak := s.detectOverUnderStreak(recentGames, "receiving_yards", 75.5); overStreak != nil {
		overStreak.PlayerID = player.NFLID
		overStreak.PlayerName = player.Name
		streaks = append(streaks, *overStreak)
	}
	
	return streaks, nil
}

func (s *StreakDetectorService) detectPerformanceStreak(games []models.WeeklyStat) *Streak {
	if len(games) < 3 {
		return nil
	}
	
	// Calculate if player is consistently outperforming or underperforming projections
	outperforms := 0
	underperforms := 0
	
	for _, game := range games {
		if game.ActualPoints > game.ProjectedPoints {
			outperforms++
		} else {
			underperforms++
		}
	}
	
	if outperforms >= 3 {
		return &Streak{
			StreakType:    "hot",
			StatLine:      "fantasy_points",
			GamesInStreak: outperforms,
			Confidence:    0.80,
		}
	} else if underperforms >= 3 {
		return &Streak{
			StreakType:    "cold",
			StatLine:      "fantasy_points",
			GamesInStreak: underperforms,
			Confidence:    0.80,
		}
	}
	
	return nil
}

func (s *StreakDetectorService) detectOverUnderStreak(games []models.WeeklyStat, statName string, line float64) *Streak {
	if len(games) < 3 {
		return nil
	}
	
	overs := 0
	for _, game := range games {
		// Simplified - would check specific stat
		if float64(game.Yards) > line {
			overs++
		}
	}
	
	if overs >= 3 {
		return &Streak{
			StreakType:    "over",
			StatLine:      fmt.Sprintf("%s %.1f", statName, line),
			GamesInStreak: overs,
			Confidence:    0.75,
		}
	}
	
	return nil
}

func (s *StreakDetectorService) explainStreak(ctx context.Context, player models.Player, games []models.WeeklyStat, streak Streak) (string, error) {
	prompt := fmt.Sprintf(`Analyze this player's performance streak:

Player: %s (%s - %s)
Streak: %s for %d consecutive games
Stat line: %s

Recent game data:
%s

Explain:
1. WHY is this streak happening? (matchups, role changes, team strategy)
2. Is it sustainable? What factors support continuation?
3. What could break the streak?

Provide a concise analysis (3-4 sentences) with actionable insights for fantasy managers.`,
		player.Name,
		player.Position,
		player.Team,
		streak.StreakType,
		streak.GamesInStreak,
		streak.StatLine,
		formatGamesForPrompt(games),
	)
	
	response, err := s.gemini.GenerateWithRetry(ctx, prompt, 3)
	if err != nil {
		return "", err
	}
	
	return response, nil
}

func formatGamesForPrompt(games []models.WeeklyStat) string {
	result := ""
	for _, game := range games {
		result += fmt.Sprintf("Week %d: %d yards, %d TDs, %.1f fantasy points\n", 
			game.Week, game.Yards, game.Touchdowns, game.ActualPoints)
	}
	return result
}

