package services

import (
	"context"
	"fmt"

	"github.com/ai-atl/nfl-platform/internal/models"
	"github.com/ai-atl/nfl-platform/pkg/gemini"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type GameScriptService struct {
	db     *mongo.Database
	gemini *gemini.Client
}

type GameScriptPrediction struct {
	GameID          string                   `json:"game_id"`
	PredictedFlow   string                   `json:"predicted_flow"`
	PlayerImpacts   []PlayerImpact           `json:"player_impacts"`
	ConfidenceScore float64                  `json:"confidence_score"`
	KeyFactors      []string                 `json:"key_factors"`
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

	// Build comprehensive context
	prompt := s.buildGameScriptPrompt(game)

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

func (s *GameScriptService) buildGameScriptPrompt(game models.Game) string {
	return fmt.Sprintf(`Analyze this NFL matchup and predict the game script:

Game: %s vs %s
Vegas Line: %.1f
Over/Under: %.1f
Start Time: %s

Based on these factors, predict:
1. How will the game flow quarter by quarter (competitive, blowout, defensive struggle)?
2. Which team will likely be playing from ahead/behind?
3. How will this affect play calling and player usage (pass vs run ratios)?
4. Which players' fantasy value increases or decreases based on the expected game script?

Provide specific, actionable predictions with reasoning.`,
		game.HomeTeam,
		game.AwayTeam,
		game.VegasLine,
		game.OverUnder,
		game.StartTime.Format("Mon Jan 2 3:04 PM"),
	)
}

