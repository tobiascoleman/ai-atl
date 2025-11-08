package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type StatsHandler struct {
	db *mongo.Database
}

func NewStatsHandler(db *mongo.Database) *StatsHandler {
	return &StatsHandler{db: db}
}

type DashboardStats struct {
	TotalPlayers      int64 `json:"total_players"`
	TotalGames        int64 `json:"total_games"`
	TotalPlays        int64 `json:"total_plays"`
	InjuredPlayers    int64 `json:"injured_players"`
	NextGenStats      int64 `json:"next_gen_stats"`
	ActiveTeams       int64 `json:"active_teams"`
	CurrentSeasonYear int   `json:"current_season_year"`
}

// GetDashboardStats returns real statistics from the database
func (h *StatsHandler) GetDashboardStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stats := DashboardStats{
		CurrentSeasonYear: 2025,
	}

	// Count total unique players (across all seasons)
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$nfl_id"},
		}}},
		bson.D{{Key: "$count", Value: "total"}},
	}
	cursor, err := h.db.Collection("players").Aggregate(ctx, pipeline)
	if err == nil {
		var result []bson.M
		if err := cursor.All(ctx, &result); err == nil && len(result) > 0 {
			if total, ok := result[0]["total"].(int32); ok {
				stats.TotalPlayers = int64(total)
			}
		}
	}

	// Count total games
	stats.TotalGames, _ = h.db.Collection("games").CountDocuments(ctx, bson.M{})

	// Count total plays
	stats.TotalPlays, _ = h.db.Collection("plays").CountDocuments(ctx, bson.M{})

	// Count injured players (current season)
	injuryFilter := bson.M{
		"season": 2025,
		"$or": []bson.M{
			{"status": "INA"},
			{"status_description_abbr": bson.M{"$in": []string{"R01", "R04", "R48", "P02"}}},
		},
	}
	stats.InjuredPlayers, _ = h.db.Collection("players").CountDocuments(ctx, injuryFilter)

	// Count Next Gen Stats entries
	stats.NextGenStats, _ = h.db.Collection("next_gen_stats").CountDocuments(ctx, bson.M{})

	// Count active teams
	stats.ActiveTeams, _ = h.db.Collection("teams").CountDocuments(ctx, bson.M{})

	c.JSON(http.StatusOK, stats)
}
