package services

import (
	"context"
	"fmt"

	"github.com/ai-atl/nfl-platform/internal/models"
	"github.com/ai-atl/nfl-platform/pkg/gemini"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type InjuryAnalyzerService struct {
	db     *mongo.Database
	gemini *gemini.Client
}

type InjuryImpact struct {
	InjuredPlayer  string            `json:"injured_player"`
	Analysis       string            `json:"analysis"`
	Beneficiaries  []PlayerBenefit   `json:"beneficiaries"`
	Confidence     float64           `json:"confidence"`
}

type PlayerBenefit struct {
	PlayerName       string  `json:"player_name"`
	ExpectedIncrease string  `json:"expected_increase"`
	Reasoning        string  `json:"reasoning"`
}

func NewInjuryAnalyzerService(db *mongo.Database) *InjuryAnalyzerService {
	return &InjuryAnalyzerService{
		db:     db,
		gemini: gemini.NewClient(),
	}
}

// AnalyzeInjuryImpact predicts how an injury affects other players
func (s *InjuryAnalyzerService) AnalyzeInjuryImpact(ctx context.Context, playerID string) (*InjuryImpact, error) {
	// Get injured player
	var player models.Player
	err := s.db.Collection("players").FindOne(ctx, bson.M{"nfl_id": playerID}).Decode(&player)
	if err != nil {
		return nil, fmt.Errorf("player not found: %w", err)
	}
	
	// Get teammates
	var teammates []models.Player
	cursor, err := s.db.Collection("players").Find(ctx, bson.M{
		"team": player.Team,
		"nfl_id": bson.M{"$ne": playerID},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	if err := cursor.All(ctx, &teammates); err != nil {
		return nil, err
	}
	
	// Generate AI analysis
	prompt := s.buildInjuryPrompt(player, teammates)
	response, err := s.gemini.GenerateWithRetry(ctx, prompt, 3)
	if err != nil {
		return nil, err
	}
	
	// Parse response and create impact analysis
	impact := &InjuryImpact{
		InjuredPlayer: player.Name,
		Analysis:      response,
		Beneficiaries: s.extractBeneficiaries(teammates, player),
		Confidence:    0.75,
	}
	
	return impact, nil
}

func (s *InjuryAnalyzerService) buildInjuryPrompt(injured models.Player, teammates []models.Player) string {
	teamStr := fmt.Sprintf("Injured Player: %s (%s - %s)\n", injured.Name, injured.Position, injured.Team)
	teamStr += fmt.Sprintf("EPA per play: %.2f, Snap share: %.1f%%\n\n", injured.EPAPerPlay, injured.SnapShare*100)
	
	teamStr += "Team Depth Chart:\n"
	for _, teammate := range teammates {
		if teammate.Position == injured.Position {
			teamStr += fmt.Sprintf("- %s: %.2f EPA, %.1f%% snaps\n", 
				teammate.Name, teammate.EPAPerPlay, teammate.SnapShare*100)
		}
	}
	
	return fmt.Sprintf(`Analyze this NFL injury impact:

%s

Predict:
1. Which teammates will see increased opportunity?
2. Specific percentage increases in touches/targets/snaps
3. How this affects the team's offensive game plan
4. Fantasy implications for each affected player

Provide quantified predictions (e.g., "+25%% targets") with reasoning.`, teamStr)
}

func (s *InjuryAnalyzerService) extractBeneficiaries(teammates []models.Player, injured models.Player) []PlayerBenefit {
	var benefits []PlayerBenefit
	
	// Simple heuristic: players at same position with decent EPA
	for _, teammate := range teammates {
		if teammate.Position == injured.Position && teammate.EPAPerPlay > 0.10 {
			benefits = append(benefits, PlayerBenefit{
				PlayerName:       teammate.Name,
				ExpectedIncrease: "+30% opportunity",
				Reasoning:        "Primary backup with strong efficiency metrics",
			})
		}
	}
	
	return benefits
}

