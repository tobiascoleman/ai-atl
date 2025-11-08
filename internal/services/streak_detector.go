package services

import (
	"context"
	"fmt"

	"github.com/ai-atl/nfl-platform/internal/models"
	"github.com/ai-atl/nfl-platform/pkg/gemini"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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
	err := s.db.Collection("players").FindOne(ctx, bson.M{
		"nfl_id": playerID,
		"season": 2025, // Current season
	}).Decode(&player)
	if err != nil {
		return nil, fmt.Errorf("player not found: %w", err)
	}
	
	// Get player stats history
	statsCollection := s.db.Collection("player_stats")
	cursor, err := statsCollection.Find(ctx, bson.M{
		"nfl_id": playerID,
	}, options.Find().SetSort(bson.D{{"season", -1}}).SetLimit(int64(lookbackGames)))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stats: %w", err)
	}
	defer cursor.Close(ctx)
	
	var stats []models.PlayerStats
	if err := cursor.All(ctx, &stats); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}
	
	// Detect patterns
	streaks := []Streak{}
	
	// Check for performance streaks
	if performanceStreak := s.detectPerformanceStreak(stats); performanceStreak != nil {
		performanceStreak.PlayerID = player.NFLID
		performanceStreak.PlayerName = player.Name
		
		// Get AI explanation
		explanation, err := s.explainStreak(ctx, player, stats, *performanceStreak)
		if err == nil {
			performanceStreak.AIExplanation = explanation
		}
		
		streaks = append(streaks, *performanceStreak)
	}
	
	return streaks, nil
}

func (s *StreakDetectorService) detectPerformanceStreak(stats []models.PlayerStats) *Streak {
	if len(stats) < 2 {
		return nil
	}
	
	// Compare recent seasons to detect trends
	// Most recent vs previous season
	if len(stats) >= 2 {
		current := stats[0]
		previous := stats[1]
		
		// Calculate total production
		currentTotal := current.PassingYards + current.RushingYards + current.ReceivingYards +
			(current.PassingTDs+current.RushingTDs+current.ReceivingTDs)*10
		previousTotal := previous.PassingYards + previous.RushingYards + previous.ReceivingYards +
			(previous.PassingTDs+previous.RushingTDs+previous.ReceivingTDs)*10
		
		improvement := float64(currentTotal-previousTotal) / float64(previousTotal)
		
		if improvement > 0.15 { // 15% improvement
			return &Streak{
				StreakType:    "hot",
				StatLine:      "overall_production",
				GamesInStreak: len(stats),
				Confidence:    0.75,
			}
		} else if improvement < -0.15 { // 15% decline
			return &Streak{
				StreakType:    "cold",
				StatLine:      "overall_production",
				GamesInStreak: len(stats),
				Confidence:    0.75,
			}
		}
	}
	
	return nil
}

func (s *StreakDetectorService) explainStreak(ctx context.Context, player models.Player, stats []models.PlayerStats, streak Streak) (string, error) {
	prompt := fmt.Sprintf(`Analyze this player's performance trend:

Player: %s (%s - %s)
Trend: %s over %d seasons
Stat line: %s

Recent season data:
%s

Explain:
1. WHY is this trend happening? (role changes, team strategy, aging)
2. Is it sustainable? What factors support continuation?
3. Fantasy outlook for next season?

Provide a concise analysis (3-4 sentences) with actionable insights for fantasy managers.`,
		player.Name,
		player.Position,
		player.Team,
		streak.StreakType,
		streak.GamesInStreak,
		streak.StatLine,
		formatStatsForPrompt(stats),
	)
	
	response, err := s.gemini.GenerateWithRetry(ctx, prompt, 3)
	if err != nil {
		return "", err
	}
	
	return response, nil
}

func formatStatsForPrompt(stats []models.PlayerStats) string {
	result := ""
	for _, stat := range stats {
		result += fmt.Sprintf("Season %d (%s): Pass: %d yds/%d TDs, Rush: %d yds/%d TDs, Rec: %d rec/%d yds/%d TDs\n",
			stat.Season, stat.SeasonType,
			stat.PassingYards, stat.PassingTDs,
			stat.RushingYards, stat.RushingTDs,
			stat.Receptions, stat.ReceivingYards, stat.ReceivingTDs)
	}
	return result
}

