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
)

type VoteHandler struct {
	db *mongo.Database
}

func NewVoteHandler(db *mongo.Database) *VoteHandler {
	return &VoteHandler{db: db}
}

// Create creates a new vote
func (h *VoteHandler) Create(c *gin.Context) {
	userID, _ := c.Get("user_id")
	objID, _ := primitive.ObjectIDFromHex(userID.(string))

	var vote models.Vote
	if err := c.ShouldBindJSON(&vote); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vote.ID = primitive.NewObjectID()
	vote.UserID = objID
	vote.CreatedAt = time.Now()

	collection := h.db.Collection("votes")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, vote)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create vote"})
		return
	}

	c.JSON(http.StatusCreated, vote)
}

// GetConsensus returns community consensus for a player
func (h *VoteHandler) GetConsensus(c *gin.Context) {
	playerID := c.Query("player_id")
	week, _ := strconv.Atoi(c.Query("week"))

	if playerID == "" || week == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "player_id and week are required"})
		return
	}

	collection := h.db.Collection("votes")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"player_id": playerID,
		"week":      week,
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch votes"})
		return
	}
	defer cursor.Close(ctx)

	var votes []models.Vote
	if err := cursor.All(ctx, &votes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode votes"})
		return
	}

	// Count vote types
	consensus := map[string]int{
		"over": 0,
		"under": 0,
		"lock": 0,
		"fade": 0,
	}

	for _, vote := range votes {
		consensus[vote.PredictionType]++
	}

	total := len(votes)
	percentages := map[string]float64{}
	for key, count := range consensus {
		if total > 0 {
			percentages[key] = float64(count) / float64(total) * 100
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"player_id":   playerID,
		"week":        week,
		"total_votes": total,
		"consensus":   consensus,
		"percentages": percentages,
	})
}

