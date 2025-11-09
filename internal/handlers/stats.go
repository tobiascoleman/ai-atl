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

// GetDashboardStats returns statistics from the database (optimized with estimated counts)
func (h *StatsHandler) GetDashboardStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // Reduced timeout
	defer cancel()

	stats := DashboardStats{
		CurrentSeasonYear: 2025,
		ActiveTeams:       32, // NFL has 32 teams (static)
	}

	// PERFORMANCE: Run counts in parallel using goroutines
	type countResult struct {
		key   string
		count int64
		err   error
	}

	resultsChan := make(chan countResult, 5)

	// Count total unique players (most recent season only - much faster!)
	go func() {
		count, err := h.db.Collection("players").CountDocuments(ctx, bson.M{"season": 2025})
		resultsChan <- countResult{"players", count, err}
	}()

	// Count total games (use estimatedDocumentCount for speed)
	go func() {
		count, err := h.db.Collection("games").EstimatedDocumentCount(ctx)
		resultsChan <- countResult{"games", count, err}
	}()

	// Count plays collection (estimated for speed)
	go func() {
		count, err := h.db.Collection("play_by_play").EstimatedDocumentCount(ctx)
		resultsChan <- countResult{"plays", count, err}
	}()

	// Count injured players (indexed query)
	go func() {
		injuryFilter := bson.M{
			"season": 2025,
			"$or": []bson.M{
				{"status": "INA"},
				{"status_description_abbr": bson.M{"$in": []string{"R01", "R04", "R48", "P02"}}},
			},
		}
		count, err := h.db.Collection("players").CountDocuments(ctx, injuryFilter)
		resultsChan <- countResult{"injured", count, err}
	}()

	// Count NGS entries (estimated for speed)
	go func() {
		count, err := h.db.Collection("next_gen_stats").EstimatedDocumentCount(ctx)
		resultsChan <- countResult{"ngs", count, err}
	}()

	// Collect results from goroutines
	for i := 0; i < 5; i++ {
		result := <-resultsChan
		if result.err == nil {
			switch result.key {
			case "players":
				stats.TotalPlayers = result.count
			case "games":
				stats.TotalGames = result.count
			case "plays":
				stats.TotalPlays = result.count
			case "injured":
				stats.InjuredPlayers = result.count
			case "ngs":
				stats.NextGenStats = result.count
			}
		}
	}

	c.JSON(http.StatusOK, stats)
}
