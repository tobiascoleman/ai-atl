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

type WaiverWireService struct {
	db     *mongo.Database
	gemini *gemini.Client
}

type WaiverGem struct {
	Player         models.Player `json:"player"`
	Ownership      float64       `json:"ownership"`
	EPAPerPlay     float64       `json:"epa_per_play"`
	AIAnalysis     string        `json:"ai_analysis"`
	Recommendation string        `json:"recommendation"`
}

func NewWaiverWireService(db *mongo.Database) *WaiverWireService {
	return &WaiverWireService{
		db:     db,
		gemini: gemini.NewClient(),
	}
}

// FindWaiverGems identifies undervalued players on waivers
func (s *WaiverWireService) FindWaiverGems(ctx context.Context, ownershipThreshold float64) ([]WaiverGem, error) {
	// Find current season players
	collection := s.db.Collection("players")

	filter := bson.M{
		"season": 2025, // Current season
	}

	opts := options.Find().SetLimit(20)

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var players []models.Player
	if err := cursor.All(ctx, &players); err != nil {
		return nil, err
	}

	// Generate AI analysis for each player
	var gems []WaiverGem
	for _, player := range players {
		// Calculate EPA from plays
		epa := s.calculatePlayerEPA(ctx, player.NFLID)

		analysis, err := s.analyzeWaiverTarget(ctx, player, epa)
		if err != nil {
			analysis = "Analysis unavailable"
		}

		gems = append(gems, WaiverGem{
			Player:         player,
			Ownership:      calculateOwnership(player), // Placeholder
			EPAPerPlay:     epa,
			AIAnalysis:     analysis,
			Recommendation: determineRecommendation(epa),
		})
	}

	return gems, nil
}

func (s *WaiverWireService) analyzeWaiverTarget(ctx context.Context, player models.Player, epa float64) (string, error) {
	prompt := fmt.Sprintf(`Analyze this waiver wire target:

Player: %s (%s - %s)
EPA per play: %.3f
Season: %d

Provide a brief analysis (2-3 sentences) on:
1. Why this player might be undervalued
2. Expected usage and opportunity
3. Recommendation for fantasy managers`,
		player.Name,
		player.Position,
		player.Team,
		epa,
		player.Season,
	)

	response, err := s.gemini.GenerateWithRetry(ctx, prompt, 3)
	if err != nil {
		return "", err
	}

	return response, nil
}

func (s *WaiverWireService) calculatePlayerEPA(ctx context.Context, playerID string) float64 {
	// Calculate EPA from plays collection
	collection := s.db.Collection("plays")

	// Look for plays where this player was involved
	filter := bson.M{
		"$or": []bson.M{
			{"passer_player_id": playerID},
			{"rusher_player_id": playerID},
			{"receiver_player_id": playerID},
		},
		"season": 2024, // Most recent complete season
	}

	cursor, err := collection.Find(ctx, filter)
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

	// Calculate average EPA
	totalEPA := 0.0
	for _, play := range plays {
		totalEPA += play.EPA
	}

	return totalEPA / float64(len(plays))
}

func calculateOwnership(player models.Player) float64 {
	// Placeholder - would integrate with fantasy platform API
	return 35.5
}

func determineRecommendation(epa float64) string {
	if epa > 0.25 {
		return "Strong Add - High Priority"
	} else if epa > 0.18 {
		return "Add - Good Value"
	} else if epa > 0.10 {
		return "Monitor - Potential Upside"
	}
	return "Hold - Limited Value"
}
