package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/ai-atl/nfl-platform/internal/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type PlayerHandler struct {
	db *mongo.Database
}

func NewPlayerHandler(db *mongo.Database) *PlayerHandler {
	return &PlayerHandler{db: db}
}

type PlayerWithStats struct {
	models.Player
	// Offensive Stats
	PassingYards   int `json:"passing_yards,omitempty"`
	PassingTDs     int `json:"passing_tds,omitempty"`
	RushingYards   int `json:"rushing_yards,omitempty"`
	RushingTDs     int `json:"rushing_tds,omitempty"`
	ReceivingYards int `json:"receiving_yards,omitempty"`
	ReceivingTDs   int `json:"receiving_tds,omitempty"`
	Receptions     int `json:"receptions,omitempty"`

	// Defensive Stats
	Tackles          int     `json:"tackles,omitempty"`
	TacklesSolo      int     `json:"tackles_solo,omitempty"`
	Sacks            float64 `json:"sacks,omitempty"`
	TacklesForLoss   float64 `json:"tackles_for_loss,omitempty"`
	DefInterceptions int     `json:"def_interceptions,omitempty"`
	PassDefended     int     `json:"pass_defended,omitempty"`
	ForcedFumbles    int     `json:"forced_fumbles,omitempty"`
	FumbleRecoveries int     `json:"fumble_recoveries,omitempty"`

	AvgEPA            float64 `json:"avg_epa"`
	IsCurrentPlayer   bool    `json:"is_current_player"`
	StatusDescription string  `json:"status_description"` // Human-readable status
}

// List returns a list of unique players (one entry per player, showing most recent season)
func (h *PlayerHandler) List(c *gin.Context) {
	collection := h.db.Collection("players")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build match filter
	matchFilter := bson.M{}

	// Team filter
	if team := c.Query("team"); team != "" {
		matchFilter["team"] = team
	}

	// Position filter
	if position := c.Query("position"); position != "" {
		matchFilter["position"] = position
	}

	// Search by name (case-insensitive regex) - BACKEND SEARCH for performance
	if search := c.Query("search"); search != "" {
		matchFilter["name"] = bson.M{
			"$regex":   search,
			"$options": "i", // case-insensitive
		}
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	skip := (page - 1) * limit

	// Sorting - use 'name' as default since it's indexed
	sortField := c.DefaultQuery("sort", "name")
	sortOrder := 1
	if c.Query("order") == "desc" {
		sortOrder = -1
	}

	// Aggregation pipeline to get unique players with their most recent season
	pipeline := mongo.Pipeline{
		// Match filters (uses indexes!)
		{{Key: "$match", Value: matchFilter}},
		// Sort by season descending to get most recent first
		{{Key: "$sort", Value: bson.D{{Key: "season", Value: -1}}}},
		// Group by nfl_id and take the first (most recent) document
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$nfl_id"},
			{Key: "doc", Value: bson.D{{Key: "$first", Value: "$$ROOT"}}},
		}}},
		// Replace root with the document
		{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$doc"}}}},
		// Sort by name (or other field) - uses name index!
		{{Key: "$sort", Value: bson.D{{Key: sortField, Value: sortOrder}}}},
		// Pagination
		{{Key: "$skip", Value: skip}},
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Printf("❌ Aggregation error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch players"})
		return
	}
	defer cursor.Close(ctx)

	var players []models.Player
	if err := cursor.All(ctx, &players); err != nil {
		log.Printf("❌ Decode error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode players"})
		return
	}

	// PERFORMANCE FIX: Batch fetch all stats in ONE query instead of N queries!
	// Build list of nfl_ids to query
	nflIDs := make([]string, len(players))
	playerMap := make(map[string]models.Player) // key: nfl_id_season

	for i, player := range players {
		nflIDs[i] = player.NFLID
		key := player.NFLID + "_" + strconv.Itoa(player.Season)
		playerMap[key] = player
	}

	// Single batch query for all stats (uses nfl_id+season index!)
	statsFilter := bson.M{
		"nfl_id":      bson.M{"$in": nflIDs},
		"season_type": "REGPOST",
	}

	statsCursor, err := h.db.Collection("player_stats").Find(ctx, statsFilter)
	if err != nil {
		log.Printf("❌ Failed to fetch stats: %v", err)
	}
	defer statsCursor.Close(ctx)

	// Build stats lookup map
	statsMap := make(map[string]models.PlayerStats) // key: nfl_id_season
	var allStats []models.PlayerStats
	if err := statsCursor.All(ctx, &allStats); err == nil {
		for _, stat := range allStats {
			key := stat.NFLID + "_" + strconv.Itoa(stat.Season)
			statsMap[key] = stat
		}
	}

	// Enrich players with stats (fast O(1) lookups!)
	enrichedPlayers := make([]PlayerWithStats, 0, len(players))
	for _, player := range players {
		isCurrentPlayer := player.Season == 2025

		// Determine status description
		var statusDesc string
		if !isCurrentPlayer {
			// Player is retired (last season was before 2025)
			statusDesc = fmt.Sprintf("Retired %d", player.Season)
		} else {
			// Active player - use their actual status
			statusDesc = models.GetPlayerStatusDescription(player.Status, player.StatusDescriptionAbbr)
		}

		enriched := PlayerWithStats{
			Player:            player,
			IsCurrentPlayer:   isCurrentPlayer,
			AvgEPA:            0,
			StatusDescription: statusDesc,
		}

		// O(1) lookup instead of N database queries!
		key := player.NFLID + "_" + strconv.Itoa(player.Season)
		if stats, found := statsMap[key]; found {
			// Offensive Stats
			enriched.PassingYards = stats.PassingYards
			enriched.PassingTDs = stats.PassingTDs
			enriched.RushingYards = stats.RushingYards
			enriched.RushingTDs = stats.RushingTDs
			enriched.ReceivingYards = stats.ReceivingYards
			enriched.ReceivingTDs = stats.ReceivingTDs
			enriched.Receptions = stats.Receptions

			// Defensive Stats
			enriched.Tackles = stats.Tackles
			enriched.TacklesSolo = stats.TacklesSolo
			enriched.Sacks = stats.Sacks
			enriched.TacklesForLoss = stats.TacklesForLoss
			enriched.DefInterceptions = stats.DefInterceptions
			enriched.PassDefended = stats.PassDefended
			enriched.ForcedFumbles = stats.ForcedFumbles
			enriched.FumbleRecoveries = stats.FumbleRecoveries

			// Store EPA for frontend
			enriched.AvgEPA = stats.EPA
		}

		enrichedPlayers = append(enrichedPlayers, enriched)
	}

	// Return enriched players (not raw players!)
	c.JSON(http.StatusOK, gin.H{
		"players": enrichedPlayers,
		"count":   len(enrichedPlayers),
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
	objID, err := bson.ObjectIDFromHex(id)
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

	objID, err := bson.ObjectIDFromHex(id)
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

	// Query player stats from player_stats collection
	statsCollection := h.db.Collection("player_stats")
	filter := bson.M{"nfl_id": player.NFLID}
	if season > 0 {
		filter["season"] = season
	}

	cursor, err := statsCollection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats"})
		return
	}
	defer cursor.Close(ctx)

	var stats []models.PlayerStats
	if err = cursor.All(ctx, &stats); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"player_id": player.ID,
		"name":      player.Name,
		"team":      player.Team,
		"position":  player.Position,
		"stats":     stats,
	})
}
