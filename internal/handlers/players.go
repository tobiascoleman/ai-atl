package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/ai-atl/nfl-platform/internal/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PlayerHandler struct {
	db *mongo.Database
}

func NewPlayerHandler(db *mongo.Database) *PlayerHandler {
	return &PlayerHandler{db: db}
}

// List returns a list of players with optional filters
func (h *PlayerHandler) List(c *gin.Context) {
	collection := h.db.Collection("players")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build filter
	filter := bson.M{}
	
	if team := c.Query("team"); team != "" {
		filter["team"] = team
	}
	if position := c.Query("position"); position != "" {
		filter["position"] = position
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	skip := (page - 1) * limit

	// Sorting
	sortField := c.DefaultQuery("sort", "name")
	sortOrder := 1
	if c.Query("order") == "desc" {
		sortOrder = -1
	}

	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: sortField, Value: sortOrder}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch players"})
		return
	}
	defer cursor.Close(ctx)

	var players []models.Player
	if err := cursor.All(ctx, &players); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode players"})
		return
	}

	// Get total count
	total, _ := collection.CountDocuments(ctx, filter)

	c.JSON(http.StatusOK, gin.H{
		"players": players,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// Get returns a single player by ID
func (h *PlayerHandler) Get(c *gin.Context) {
	collection := h.db.Collection("players")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		// Try finding by NFL ID instead
		var player models.Player
		err = collection.FindOne(ctx, bson.M{"nfl_id": id}).Decode(&player)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
			return
		}
		c.JSON(http.StatusOK, player)
		return
	}

	var player models.Player
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&player)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
		return
	}

	c.JSON(http.StatusOK, player)
}

// GetStats returns player statistics for a specific season and week
func (h *PlayerHandler) GetStats(c *gin.Context) {
	collection := h.db.Collection("players")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := c.Param("id")
	season, _ := strconv.Atoi(c.DefaultQuery("season", strconv.Itoa(time.Now().Year())))
	week, _ := strconv.Atoi(c.Query("week"))

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player ID"})
		return
	}

	var player models.Player
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&player)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
		return
	}

	// Filter weekly stats
	var filteredStats []models.WeeklyStat
	for _, stat := range player.WeeklyStats {
		if stat.Season == season {
			if week == 0 || stat.Week == week {
				filteredStats = append(filteredStats, stat)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"player_id": player.ID,
		"name":      player.Name,
		"stats":     filteredStats,
	})
}

