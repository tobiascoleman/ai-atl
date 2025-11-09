package handlers

import (
	"net/http"

	"github.com/ai-atl/nfl-platform/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type InsightHandler struct {
	db                *mongo.Database
	gameScriptService *services.GameScriptService
	waiverWireService *services.WaiverWireService
}

func NewInsightHandler(db *mongo.Database) *InsightHandler {
	return &InsightHandler{
		db:                db,
		gameScriptService: services.NewGameScriptService(db),
		waiverWireService: services.NewWaiverWireService(db),
	}
}

// GameScript predicts how a game will unfold
func (h *InsightHandler) GameScript(c *gin.Context) {
	gameID := c.Query("game_id")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game_id is required"})
		return
	}

	prediction, err := h.gameScriptService.PredictGameScript(c.Request.Context(), gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, prediction)
}

// InjuryImpact analyzes the impact of an injury on player opportunities
func (h *InsightHandler) InjuryImpact(c *gin.Context) {
	var req struct {
		PlayerID string `json:"player_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement injury impact analysis
	c.JSON(http.StatusOK, gin.H{
		"injured_player": req.PlayerID,
		"analysis":       "Player X will see 30% more targets",
		"beneficiaries": []map[string]interface{}{
			{"player": "John Doe", "increase": "30%"},
		},
	})
}

// Streaks detects hot/cold streaks for a player
func (h *InsightHandler) Streaks(c *gin.Context) {
	playerID := c.Query("player_id")
	if playerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "player_id is required"})
		return
	}

	// TODO: Implement streak detection
	c.JSON(http.StatusOK, gin.H{
		"player_id": playerID,
		"streaks": []map[string]interface{}{
			{
				"type":        "over",
				"stat":        "receiving_yards",
				"line":        75.5,
				"games":       4,
				"ai_analysis": "Player has favorable matchups and increased target share",
			},
		},
	})
}

// TopPerformers returns top over/under performers of the week
func (h *InsightHandler) TopPerformers(c *gin.Context) {
	week := c.DefaultQuery("week", "9")
	performerType := c.DefaultQuery("type", "over")

	// TODO: Calculate from actual data
	c.JSON(http.StatusOK, gin.H{
		"week": week,
		"type": performerType,
		"performers": []map[string]interface{}{
			{
				"player":     "Patrick Mahomes",
				"projected":  24.5,
				"actual":     38.2,
				"difference": 13.7,
				"epa":        15.3,
			},
		},
	})
}

// WaiverGems finds undervalued players with breakout potential
func (h *InsightHandler) WaiverGems(c *gin.Context) {
	position := c.DefaultQuery("position", "ALL")
	limit := 10 // Top 10 candidates

	gems, err := h.waiverWireService.FindWaiverGems(c.Request.Context(), position, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"gems":  gems,
		"count": len(gems),
	})
}

// PersonalizedWaiverGems provides waiver recommendations based on user's ESPN roster
func (h *InsightHandler) PersonalizedWaiverGems(c *gin.Context) {
	var req struct {
		Roster   []services.RosterPlayer `json:"roster" binding:"required"`
		Position string                  `json:"position"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default to "ALL" if not specified
	if req.Position == "" {
		req.Position = "ALL"
	}

	limit := 10
	gems, err := h.waiverWireService.FindPersonalizedWaiverGems(c.Request.Context(), req.Roster, req.Position, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"gems":  gems,
		"count": len(gems),
	})
}
