package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ai-atl/nfl-platform/internal/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ESPNHandler struct {
	db             *mongo.Database
	flaskServiceURL string
}

func NewESPNHandler(db *mongo.Database, flaskServiceURL string) *ESPNHandler {
	return &ESPNHandler{
		db:             db,
		flaskServiceURL: flaskServiceURL,
	}
}

type ESPNCredentials struct {
	ESPNS2  string `json:"espn_s2" binding:"required"`
	ESPNSWID string `json:"espn_swid" binding:"required"`
	LeagueID int    `json:"league_id" binding:"required"`
	TeamID   int    `json:"team_id" binding:"required"`
	Year     int    `json:"year" binding:"required"`
}

type ESPNPlayer struct {
	Name       string `json:"name"`
	Position   string `json:"position"`
	ProTeam    string `json:"proTeam"`
	LineupSlot string `json:"lineupSlot"`
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
