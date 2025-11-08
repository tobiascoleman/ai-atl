package services

import (
	"context"
	"fmt"

	"github.com/ai-atl/nfl-platform/internal/models"
	"github.com/ai-atl/nfl-platform/pkg/gemini"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ChatbotService struct {
	db     *mongo.Database
	gemini *gemini.Client
}

func NewChatbotService(db *mongo.Database) *ChatbotService {
	return &ChatbotService{
		db:     db,
		gemini: gemini.NewClient(),
	}
}

// Ask handles a question from the user and returns an AI-generated response
func (s *ChatbotService) Ask(ctx context.Context, userID string, question string) (string, error) {
	// Get user's lineup context
	objID, _ := bson.ObjectIDFromHex(userID)
	
	var lineups []models.FantasyLineup
	cursor, err := s.db.Collection("lineups").Find(ctx, bson.M{"user_id": objID})
	if err == nil {
		cursor.All(ctx, &lineups)
	}

	// Build context-aware prompt
	prompt := s.buildChatbotPrompt(question, lineups)

	// Get AI response
	response, err := s.gemini.GenerateWithRetry(ctx, prompt, 3)
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	return response, nil
}

func (s *ChatbotService) buildChatbotPrompt(question string, lineups []models.FantasyLineup) string {
	contextInfo := "No lineup information available."
	if len(lineups) > 0 {
		// Get the most recent lineup
		lineup := lineups[len(lineups)-1]
		contextInfo = fmt.Sprintf("User's current lineup: %v", lineup.Positions)
	}

	return fmt.Sprintf(`You are an expert NFL fantasy football advisor with access to advanced EPA metrics and player data.

User Context:
%s

User Question: %s

Provide specific, actionable fantasy football advice based on:
1. Recent player performance and trends
2. Matchup analysis
3. Injury reports
4. Advanced metrics (EPA, target share, snap counts)

Be conversational but data-driven. Explain your reasoning.`,
		contextInfo,
		question,
	)
}

