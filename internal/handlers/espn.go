package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ai-atl/nfl-platform/internal/models"
	"github.com/ai-atl/nfl-platform/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ESPNHandler struct {
	db              *mongo.Database
	flaskServiceURL string
	advisorService  *services.FantasyAdvisorService
}

func NewESPNHandler(db *mongo.Database, flaskServiceURL string) *ESPNHandler {
	return &ESPNHandler{
		db:              db,
		flaskServiceURL: flaskServiceURL,
		advisorService:  services.NewFantasyAdvisorService(db),
	}
}

type ESPNCredentials struct {
	ESPNS2   string `json:"espn_s2" binding:"required"`
	ESPNSWID string `json:"espn_swid" binding:"required"`
	LeagueID int    `json:"league_id" binding:"required"`
	TeamID   int    `json:"team_id" binding:"required"`
	Year     int    `json:"year" binding:"required"`
}

type ESPNPlayer struct {
	Name            string   `json:"name"`
	Position        string   `json:"position"`
	ProTeam         string   `json:"proTeam"`
	LineupSlot      string   `json:"lineupSlot"`
	ProjectedPoints float64  `json:"projectedPoints"`
	Points          float64  `json:"points"`
	Injured         bool     `json:"injured"`
	InjuryStatus    *string  `json:"injuryStatus"`
	EligibleSlots   []string `json:"eligibleSlots,omitempty"`
	RecommendedSlot string   `json:"recommendedSlot,omitempty"`
	PlayerID        *int     `json:"playerId,omitempty"`
}

type OptimizeLineupResponse struct {
	OptimalLineup  []ESPNPlayer `json:"optimalLineup"`
	Bench          []ESPNPlayer `json:"bench"`
	TotalProjected float64      `json:"totalProjected"`
}

type ESPNStatusResponse struct {
	Connected bool `json:"connected"`
}

type ESPNRosterResponse struct {
	Connected bool         `json:"connected"`
	Players   []ESPNPlayer `json:"players"`
}

// SaveCredentials saves ESPN credentials to user profile
func (h *ESPNHandler) SaveCredentials(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var creds ESPNCredentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Update user document with ESPN credentials
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	update := bson.M{
		"$set": bson.M{
			"espn_s2":   creds.ESPNS2,
			"espn_swid": creds.ESPNSWID,
			"league_id": creds.LeagueID,
			"team_id":   creds.TeamID,
			"year":      creds.Year,
		},
	}

	_, err = h.db.Collection("users").UpdateByID(c.Request.Context(), objectID, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "ESPN credentials saved successfully",
		"connected": true,
	})
}

// GetStatus checks if user has ESPN credentials stored
func (h *ESPNHandler) GetStatus(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var user models.User
	err = h.db.Collection("users").FindOne(c.Request.Context(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}

	connected := user.ESPNS2 != "" && user.ESPNSWID != ""

	c.JSON(http.StatusOK, ESPNStatusResponse{
		Connected: connected,
	})
}

// GetRoster fetches the user's ESPN fantasy roster from Flask service
func (h *ESPNHandler) GetRoster(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get user's ESPN credentials
	var user models.User
	err = h.db.Collection("users").FindOne(c.Request.Context(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}

	if user.ESPNS2 == "" || user.ESPNSWID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ESPN credentials not configured"})
		return
	}

	// Call Flask service to get roster
	flaskURL := fmt.Sprintf("%s/api/espn/roster", h.flaskServiceURL)
	resp, err := http.Get(flaskURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch roster from ESPN service"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ESPN service returned error: " + string(body),
		})
		return
	}

	// Parse the roster response
	var players []ESPNPlayer
	if err := json.NewDecoder(resp.Body).Decode(&players); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse roster data"})
		return
	}

	c.JSON(http.StatusOK, ESPNRosterResponse{
		Connected: true,
		Players:   players,
	})
}

// OptimizeLineup gets the optimal lineup based on projected points
func (h *ESPNHandler) OptimizeLineup(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get user's ESPN credentials
	var user models.User
	err = h.db.Collection("users").FindOne(c.Request.Context(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}

	if user.ESPNS2 == "" || user.ESPNSWID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ESPN credentials not configured"})
		return
	}

	// Call Flask service to get optimized lineup
	flaskURL := fmt.Sprintf("%s/api/espn/optimize-lineup", h.flaskServiceURL)
	resp, err := http.Get(flaskURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch optimized lineup from ESPN service"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ESPN service returned error: " + string(body),
		})
		return
	}

	// Parse the optimize response
	var optimized OptimizeLineupResponse
	if err := json.NewDecoder(resp.Body).Decode(&optimized); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse optimization data"})
		return
	}

	c.JSON(http.StatusOK, optimized)
}

type FreeAgentPlayer struct {
	Name            string      `json:"name"`
	Position        string      `json:"position"`
	ProTeam         string      `json:"proTeam"`
	ProjectedPoints float64     `json:"projectedPoints"`
	Points          float64     `json:"points"`
	Injured         bool        `json:"injured"`
	InjuryStatus    interface{} `json:"injuryStatus"`
	PlayerID        *int        `json:"playerId"`
	PercentOwned    float64     `json:"percentOwned"`
	PercentStarted  float64     `json:"percentStarted"`
}

type FreeAgentsResponse struct {
	Players []FreeAgentPlayer `json:"players"`
	Count   int               `json:"count"`
}

// GetFreeAgents fetches available free agents from ESPN
func (h *ESPNHandler) GetFreeAgents(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get user's ESPN credentials
	var user models.User
	err = h.db.Collection("users").FindOne(c.Request.Context(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}

	if user.ESPNS2 == "" || user.ESPNSWID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ESPN credentials not configured"})
		return
	}

	// Get query parameters
	position := c.Query("position")
	size := c.DefaultQuery("size", "50")

	// Call Flask service to get free agents
	flaskURL := fmt.Sprintf("%s/api/espn/free-agents?size=%s", h.flaskServiceURL, size)
	if position != "" {
		flaskURL += "&position=" + position
	}
	resp, err := http.Get(flaskURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch free agents from ESPN service"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ESPN service returned error: " + string(body),
		})
		return
	}

	// Read and log the response for debugging
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response body"})
		return
	}

	// Log first 500 chars for debugging
	bodyStr := string(body)
	if len(bodyStr) > 500 {
		bodyStr = bodyStr[:500] + "..."
	}
	fmt.Printf("Flask response: %s\n", bodyStr)

	var freeAgents FreeAgentsResponse
	if err := json.Unmarshal(body, &freeAgents); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to parse free agents data: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, freeAgents)
}

type AIStartSitRequest struct {
	PlayerA ESPNPlayer `json:"playerA" binding:"required"`
	PlayerB ESPNPlayer `json:"playerB" binding:"required"`
}

type AIStartSitResponse struct {
	Recommendation string `json:"recommendation"` // "A" or "B"
	Confidence     int    `json:"confidence"`     // 0-100
	Reasoning      string `json:"reasoning"`
	PlayerAName    string `json:"playerAName"`
	PlayerBName    string `json:"playerBName"`
}

// GetAIStartSitAdvice provides AI-powered start/sit recommendations with database enrichment
func (h *ESPNHandler) GetAIStartSitAdvice(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req AIStartSitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Extract player data
	playerAInj := ""
	if req.PlayerA.InjuryStatus != nil {
		playerAInj = *req.PlayerA.InjuryStatus
	}

	playerBInj := ""
	if req.PlayerB.InjuryStatus != nil {
		playerBInj = *req.PlayerB.InjuryStatus
	}

	// Call advisor service with database enrichment
	comparison, err := h.advisorService.GetStartSitAdvice(
		c.Request.Context(),
		req.PlayerA.Name, req.PlayerA.Position, req.PlayerA.ProTeam,
		req.PlayerA.ProjectedPoints, req.PlayerA.Points,
		req.PlayerA.Injured, playerAInj,
		req.PlayerB.Name, req.PlayerB.Position, req.PlayerB.ProTeam,
		req.PlayerB.ProjectedPoints, req.PlayerB.Points,
		req.PlayerB.Injured, playerBInj,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate AI recommendation: " + err.Error(),
		})
		return
	}

	// Build response
	response := AIStartSitResponse{
		Recommendation: comparison.Recommendation,
		Confidence:     comparison.Confidence,
		Reasoning:      comparison.Reasoning,
		PlayerAName:    comparison.PlayerAName,
		PlayerBName:    comparison.PlayerBName,
	}

	c.JSON(http.StatusOK, response)
}
