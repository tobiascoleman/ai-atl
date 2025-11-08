package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ai-atl/nfl-platform/internal/models"
	"github.com/ai-atl/nfl-platform/pkg/gemini"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ChatbotService struct {
	db          *mongo.Database
	gemini      *gemini.Client
	dataService *DataService
}

func NewChatbotService(db *mongo.Database) *ChatbotService {
	return &ChatbotService{
		db:          db,
		gemini:      gemini.NewClient(),
		dataService: NewDataService(db),
	}
}

// QueryIntent represents what data the user is asking about
type QueryIntent struct {
	PlayerNames []string `json:"player_names"`
	Teams       []string `json:"teams"`
	Positions   []string `json:"positions"`
	StatTypes   []string `json:"stat_types"` // e.g., "passing", "rushing", "receiving", "epa"
	Season      int      `json:"season"`
	NeedsData   bool     `json:"needs_data"`
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

	// Extract query intent from the question
	intent, err := s.extractQueryIntent(ctx, question)
	if err != nil {
		// If extraction fails, continue without data context
		intent = &QueryIntent{NeedsData: false}
	}

	// Retrieve relevant stats from database if needed
	var statsContext string
	if intent.NeedsData {
		statsContext, err = s.retrieveRelevantStats(ctx, intent)
		if err != nil {
			// Log error but continue - we can still answer without perfect data
			statsContext = "Unable to retrieve some requested stats from database."
		}
	}

	// Build context-aware prompt with database stats
	prompt := s.buildChatbotPrompt(question, lineups, statsContext)

	// Get AI response
	response, err := s.gemini.GenerateWithRetry(ctx, prompt, 3)
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	return response, nil
}

// extractQueryIntent uses AI to extract what data the user is asking about
func (s *ChatbotService) extractQueryIntent(ctx context.Context, question string) (*QueryIntent, error) {
	extractionPrompt := fmt.Sprintf(`Analyze this fantasy football question and extract the data requirements.
Return ONLY a valid JSON object with this structure:
{
  "player_names": ["name1", "name2"],
  "teams": ["team1"],
  "positions": ["QB", "RB"],
  "stat_types": ["passing", "rushing", "receiving", "epa"],
  "needs_data": true
}

Rules:
- player_names: Full player names mentioned (e.g., ["Patrick Mahomes", "Travis Kelce"])
- teams: Team abbreviations (e.g., ["KC", "BUF"])
- positions: Positions mentioned (QB, RB, WR, TE, K, DEF)
- stat_types: Types of stats (passing, rushing, receiving, epa, injuries)
- season: Year mentioned or 2024 if current season, 2025 if future
- needs_data: true if specific players/teams/stats mentioned, false for general questions

Question: %s

Return only the JSON object, no explanation.`, question)

	response, err := s.gemini.Generate(ctx, extractionPrompt)
	if err != nil {
		return nil, err
	}

	// Clean up response to extract JSON
	response = strings.TrimSpace(response)
	response = strings.Trim(response, "`")
	if strings.HasPrefix(response, "json") {
		response = strings.TrimPrefix(response, "json")
		response = strings.TrimSpace(response)
	}

	var intent QueryIntent
	if err := json.Unmarshal([]byte(response), &intent); err != nil {
		return nil, fmt.Errorf("failed to parse intent JSON: %w", err)
	}

	return &intent, nil
}

// retrieveRelevantStats fetches stats from MongoDB based on query intent
func (s *ChatbotService) retrieveRelevantStats(ctx context.Context, intent *QueryIntent) (string, error) {
	var statsBuilder strings.Builder
	statsBuilder.WriteString("\n=== RELEVANT DATABASE STATS ===\n\n")

	currentSeason := time.Now().Year()
	if intent.Season == 0 {
		intent.Season = currentSeason
	}

	// Fetch player-specific data
	for _, playerName := range intent.PlayerNames {
		// Try to find player by name
		players, err := s.findPlayersByName(ctx, playerName, intent.Season)
		if err != nil || len(players) == 0 {
			continue
		}

		player := players[0] // Use first match
		statsBuilder.WriteString(fmt.Sprintf("## %s (%s - %s)\n", player.Name, player.Position, player.Team))

		// Get injury status
		if player.Status == "INA" || player.StatusDescriptionAbbr != "" {
			statsBuilder.WriteString(fmt.Sprintf("- **Injury Status**: %s (Week %d)\n", player.StatusDescriptionAbbr, player.Week))
		}

		// Get season stats
		stats, err := s.dataService.GetPlayerStats(ctx, player.NFLID, intent.Season)
		if err == nil && len(stats) > 0 {
			for _, stat := range stats {
				statsBuilder.WriteString(fmt.Sprintf("- **%d %s Stats**:\n", stat.Season, stat.SeasonType))
				if stat.PassingYards > 0 {
					statsBuilder.WriteString(fmt.Sprintf("  - Passing: %d yards, %d TDs, %d INTs\n", 
						stat.PassingYards, stat.PassingTDs, stat.Interceptions))
				}
				if stat.RushingYards > 0 {
					statsBuilder.WriteString(fmt.Sprintf("  - Rushing: %d yards, %d TDs\n", 
						stat.RushingYards, stat.RushingTDs))
				}
				if stat.Receptions > 0 {
					statsBuilder.WriteString(fmt.Sprintf("  - Receiving: %d rec, %d yards, %d TDs, %d targets\n", 
						stat.Receptions, stat.ReceivingYards, stat.ReceivingTDs, stat.Targets))
				}
			}
		}

		// Get EPA if requested
		if s.containsStatType(intent.StatTypes, "epa") {
			epa, playCount, err := s.dataService.CalculatePlayerEPA(ctx, player.NFLID, intent.Season)
			if err == nil && playCount > 0 {
				statsBuilder.WriteString(fmt.Sprintf("- **EPA**: %.3f (over %d plays)\n", epa, playCount))
			}
		}

		statsBuilder.WriteString("\n")
	}

	// Fetch team-specific data
	for _, team := range intent.Teams {
		statsBuilder.WriteString(fmt.Sprintf("## Team: %s\n", team))

		// Get team EPA
		epa, playCount, err := s.dataService.CalculateTeamEPA(ctx, team, intent.Season)
		if err == nil && playCount > 0 {
			statsBuilder.WriteString(fmt.Sprintf("- **Team EPA**: %.3f (over %d plays)\n", epa, playCount))
		}

		// Get injured players on team
		if s.containsStatType(intent.StatTypes, "injuries") {
			players, err := s.dataService.GetPlayersByTeam(ctx, team, intent.Season)
			if err == nil {
				var injured []string
				for _, p := range players {
					if p.Status == "INA" || p.StatusDescriptionAbbr != "" {
						injured = append(injured, fmt.Sprintf("%s (%s)", p.Name, p.Position))
					}
				}
				if len(injured) > 0 {
					statsBuilder.WriteString(fmt.Sprintf("- **Injured Players**: %s\n", strings.Join(injured, ", ")))
				}
			}
		}

		statsBuilder.WriteString("\n")
	}

	// Fetch position-specific data
	for _, position := range intent.Positions {
		players, err := s.dataService.GetPlayersByPosition(ctx, position, intent.Season)
		if err == nil && len(players) > 0 {
			statsBuilder.WriteString(fmt.Sprintf("## Top %s Players (limited to first 10)\n", position))
			count := 0
			for _, p := range players {
				if count >= 10 {
					break
				}
				statsBuilder.WriteString(fmt.Sprintf("- %s (%s)\n", p.Name, p.Team))
				count++
			}
			statsBuilder.WriteString("\n")
		}
	}

	result := statsBuilder.String()
	if result == "\n=== RELEVANT DATABASE STATS ===\n\n" {
		return "No relevant stats found in database for this query.", nil
	}

	return result, nil
}

// findPlayersByName searches for players by name (case-insensitive partial match)
func (s *ChatbotService) findPlayersByName(ctx context.Context, name string, season int) ([]models.Player, error) {
	// Try exact match first
	var players []models.Player
	opts := options.Find().SetLimit(5)
	
	cursor, err := s.db.Collection("players").Find(ctx, bson.M{
		"name":   bson.M{"$regex": name, "$options": "i"},
		"season": season,
	}, opts)
	
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &players); err != nil {
		return nil, err
	}

	return players, nil
}

// containsStatType checks if a stat type is in the list
func (s *ChatbotService) containsStatType(statTypes []string, target string) bool {
	for _, st := range statTypes {
		if strings.EqualFold(st, target) {
			return true
		}
	}
	return true // Default to including all stats if not specified
}

func (s *ChatbotService) buildChatbotPrompt(question string, lineups []models.FantasyLineup, statsContext string) string {
	contextInfo := "No lineup information available."
	if len(lineups) > 0 {
		// Get the most recent lineup
		lineup := lineups[len(lineups)-1]
		contextInfo = fmt.Sprintf("User's current lineup: %v", lineup.Positions)
	}

	// Add database stats context if available
	dataContext := ""
	if statsContext != "" {
		dataContext = fmt.Sprintf("\n\nDatabase Stats:\n%s", statsContext)
	}

	return fmt.Sprintf(`You are an expert NFL fantasy football advisor with access to advanced EPA metrics and player data.

User Context:
%s%s

User Question: %s

Provide specific, actionable fantasy football advice based on:
1. The actual stats from our database shown above
2. Recent player performance and trends
3. Matchup analysis
4. Injury reports
5. Advanced metrics (EPA, target share, snap counts)

IMPORTANT: Reference the specific stats provided from our database in your answer. Use the actual numbers shown.

Be conversational but data-driven. Explain your reasoning.`,
		contextInfo,
		dataContext,
		question,
	)
}

