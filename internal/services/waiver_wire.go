package services

import (
	"context"
	"fmt"

	"github.com/ai-atl/nfl-platform/internal/models"
	"github.com/ai-atl/nfl-platform/pkg/gemini"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type WaiverWireService struct {
	db     *mongo.Database
	gemini *gemini.Client
}

type WaiverGem struct {
	Player       models.Player `json:"player"`
	Ownership    float64       `json:"ownership"`
	EPAPerPlay   float64       `json:"epa_per_play"`
	AIAnalysis   string        `json:"ai_analysis"`
	Recommendation string      `json:"recommendation"`
}

func NewWaiverWireService(db *mongo.Database) *WaiverWireService {
	return &WaiverWireService{
		db:     db,
		gemini: gemini.NewClient(),
	}
}

// FindWaiverGems identifies undervalued players on waivers
func (s *WaiverWireService) FindWaiverGems(ctx context.Context, ownershipThreshold float64) ([]WaiverGem, error) {
	// Find low-owned players with high EPA
	collection := s.db.Collection("players")
	
	filter := bson.M{
		"epa_per_play": bson.M{"$gt": 0.15}, // Above average EPA
		// ownership field would come from a fantasy platform integration
	}
	
	opts := options.Find().
		SetSort(bson.D{{"epa_per_play", -1}}).
		SetLimit(10)
	
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
		analysis, err := s.analyzeWaiverTarget(ctx, player)
		if err != nil {
			analysis = "Analysis unavailable"
		}
		
		gems = append(gems, WaiverGem{
			Player:         player,
			Ownership:      calculateOwnership(player), // Placeholder
			EPAPerPlay:     player.EPAPerPlay,
			AIAnalysis:     analysis,
			Recommendation: determineRecommendation(player),
		})
	}
	
	return gems, nil
}

func (s *WaiverWireService) analyzeWaiverTarget(ctx context.Context, player models.Player) (string, error) {
	prompt := fmt.Sprintf(`Analyze this waiver wire target:

Player: %s (%s - %s)
EPA per play: %.2f
Success rate: %.2f
Snap share: %.2f%%
Target share: %.2f%%

Provide a brief analysis (2-3 sentences) on:
1. Why this player is undervalued
2. Expected usage and opportunity
3. Recommendation for fantasy managers`,
		player.Name,
		player.Position,
		player.Team,
		player.EPAPerPlay,
		player.SuccessRate,
		player.SnapShare*100,
		player.TargetShare*100,
	)
	
	response, err := s.gemini.GenerateWithRetry(ctx, prompt, 3)
	if err != nil {
		return "", err
	}
	
	return response, nil
}

func calculateOwnership(player models.Player) float64 {
	// Placeholder - would integrate with fantasy platform API
	return 35.5
}

func determineRecommendation(player models.Player) string {
	if player.EPAPerPlay > 0.25 {
		return "Strong Add - High Priority"
	} else if player.EPAPerPlay > 0.18 {
		return "Add - Good Value"
	}
	return "Monitor - Potential Upside"
}

