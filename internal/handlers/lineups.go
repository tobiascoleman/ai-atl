package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/ai-atl/nfl-platform/internal/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type LineupHandler struct {
	db *mongo.Database
}

func NewLineupHandler(db *mongo.Database) *LineupHandler {
	return &LineupHandler{db: db}
}

// List returns all lineups for the authenticated user
func (h *LineupHandler) List(c *gin.Context) {
	userID, _ := c.Get("user_id")
	objID, _ := bson.ObjectIDFromHex(userID.(string))

	collection := h.db.Collection("lineups")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"user_id": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch lineups"})
		return
	}
	defer cursor.Close(ctx)

	var lineups []models.FantasyLineup
	if err := cursor.All(ctx, &lineups); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode lineups"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"lineups": lineups})
}

// Create creates a new lineup
func (h *LineupHandler) Create(c *gin.Context) {
	userID, _ := c.Get("user_id")
	objID, _ := bson.ObjectIDFromHex(userID.(string))

	var lineup models.FantasyLineup
	if err := c.ShouldBindJSON(&lineup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lineup.ID = bson.NewObjectID()
	lineup.UserID = objID
	lineup.CreatedAt = time.Now()
	lineup.UpdatedAt = time.Now()

	collection := h.db.Collection("lineups")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, lineup)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create lineup"})
		return
	}

	c.JSON(http.StatusCreated, lineup)
}

// Get returns a specific lineup
func (h *LineupHandler) Get(c *gin.Context) {
	id := c.Param("id")
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lineup ID"})
		return
	}

	collection := h.db.Collection("lineups")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var lineup models.FantasyLineup
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&lineup)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lineup not found"})
		return
	}

	c.JSON(http.StatusOK, lineup)
}

// Update updates an existing lineup
func (h *LineupHandler) Update(c *gin.Context) {
	id := c.Param("id")
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lineup ID"})
		return
	}

	var updates bson.M
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates["updated_at"] = time.Now()

	collection := h.db.Collection("lineups")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": updates})
	if err != nil || result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lineup not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lineup updated"})
}

// Delete removes a lineup
func (h *LineupHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lineup ID"})
		return
	}

	collection := h.db.Collection("lineups")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil || result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lineup not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lineup deleted"})
}

// Optimize generates an optimized lineup using AI (placeholder)
func (h *LineupHandler) Optimize(c *gin.Context) {
	// TODO: Implement AI optimization
	c.JSON(http.StatusOK, gin.H{
		"message": "AI optimization feature coming soon",
		"suggested_lineup": gin.H{
			"QB":  "player_123",
			"RB1": "player_456",
			"RB2": "player_789",
		},
	})
}

