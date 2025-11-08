package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/ai-atl/nfl-platform/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type DataHandler struct {
	service *services.DataService
}

func NewDataHandler(db *mongo.Database) *DataHandler {
	return &DataHandler{
		service: services.NewDataService(db),
	}
}

// ========================================
// PLAYER ENDPOINTS
// ========================================

// GetPlayer - GET /api/data/players/:nfl_id?season=2024
func (h *DataHandler) GetPlayer(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	nflID := c.Param("nfl_id")
	season, _ := strconv.Atoi(c.DefaultQuery("season", "2025"))

	player, err := h.service.GetPlayer(ctx, nflID, season)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
		return
	}

	c.JSON(http.StatusOK, player)
}

// GetPlayersByTeam - GET /api/data/teams/:team/players?season=2024
func (h *DataHandler) GetPlayersByTeam(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	team := c.Param("team")
	season, _ := strconv.Atoi(c.DefaultQuery("season", "2025"))

	players, err := h.service.GetPlayersByTeam(ctx, team, season)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch players"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"team":    team,
		"season":  season,
		"count":   len(players),
		"players": players,
	})
}

// GetPlayersByPosition - GET /api/data/positions/:position?season=2024
func (h *DataHandler) GetPlayersByPosition(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	position := c.Param("position")
	season, _ := strconv.Atoi(c.DefaultQuery("season", "2025"))

	players, err := h.service.GetPlayersByPosition(ctx, position, season)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch players"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"position": position,
		"season":   season,
		"count":    len(players),
		"players":  players,
	})
}

// GetInjuredPlayers - GET /api/data/injuries?season=2024
func (h *DataHandler) GetInjuredPlayers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	season, _ := strconv.Atoi(c.DefaultQuery("season", "2025"))

	players, err := h.service.GetInjuredPlayers(ctx, season)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch injured players"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"season":  season,
		"count":   len(players),
		"players": players,
	})
}

// ========================================
// STATS ENDPOINTS
// ========================================

// GetPlayerStats - GET /api/data/players/:nfl_id/stats?season=2024
func (h *DataHandler) GetPlayerStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	nflID := c.Param("nfl_id")
	season, _ := strconv.Atoi(c.Query("season"))

	stats, err := h.service.GetPlayerStats(ctx, nflID, season)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"nfl_id": nflID,
		"season": season,
		"count":  len(stats),
		"stats":  stats,
	})
}

// ========================================
// EPA ENDPOINTS
// ========================================

// GetPlayerEPA - GET /api/data/players/:nfl_id/epa?season=2024
func (h *DataHandler) GetPlayerEPA(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	nflID := c.Param("nfl_id")
	season, _ := strconv.Atoi(c.Query("season"))

	epa, playCount, err := h.service.CalculatePlayerEPA(ctx, nflID, season)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate EPA"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"nfl_id":     nflID,
		"season":     season,
		"epa":        epa,
		"play_count": playCount,
	})
}

// GetTeamEPA - GET /api/data/teams/:team/epa?season=2024
func (h *DataHandler) GetTeamEPA(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	team := c.Param("team")
	season, _ := strconv.Atoi(c.Query("season"))

	epa, playCount, err := h.service.CalculateTeamEPA(ctx, team, season)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate EPA"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"team":       team,
		"season":     season,
		"epa":        epa,
		"play_count": playCount,
	})
}

// ========================================
// PLAYS ENDPOINTS
// ========================================

// GetPlayerPlays - GET /api/data/players/:nfl_id/plays?season=2024&limit=100
func (h *DataHandler) GetPlayerPlays(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	nflID := c.Param("nfl_id")
	season, _ := strconv.Atoi(c.Query("season"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	plays, err := h.service.GetPlayerPlays(ctx, nflID, season, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plays"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"nfl_id": nflID,
		"season": season,
		"count":  len(plays),
		"plays":  plays,
	})
}

// GetTeamPlays - GET /api/data/teams/:team/plays?season=2024&limit=100
func (h *DataHandler) GetTeamPlays(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	team := c.Param("team")
	season, _ := strconv.Atoi(c.Query("season"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	plays, err := h.service.GetTeamPlays(ctx, team, season, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plays"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"team":   team,
		"season": season,
		"count":  len(plays),
		"plays":  plays,
	})
}

// GetGamePlays - GET /api/data/games/:game_id/plays
func (h *DataHandler) GetGamePlays(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	gameID := c.Param("game_id")

	plays, err := h.service.GetGamePlays(ctx, gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plays"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"game_id": gameID,
		"count":   len(plays),
		"plays":   plays,
	})
}

// ========================================
// NGS ENDPOINTS
// ========================================

// GetPlayerNGS - GET /api/data/players/:nfl_id/ngs?stat_type=passing&season=2024
func (h *DataHandler) GetPlayerNGS(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	nflID := c.Param("nfl_id")
	statType := c.Query("stat_type")
	season, _ := strconv.Atoi(c.Query("season"))

	stats, err := h.service.GetPlayerNGS(ctx, nflID, statType, season)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch NGS stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"nfl_id":    nflID,
		"stat_type": statType,
		"season":    season,
		"count":     len(stats),
		"stats":     stats,
	})
}

// GetNGSLeaders - GET /api/data/ngs/leaders?stat_type=passing&season=2024&metric=completion_percentage_above_expectation&limit=10
func (h *DataHandler) GetNGSLeaders(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	statType := c.Query("stat_type")
	season, _ := strconv.Atoi(c.Query("season"))
	metric := c.Query("metric")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	stats, err := h.service.GetNGSLeaders(ctx, statType, season, metric, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch NGS leaders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stat_type": statType,
		"season":    season,
		"metric":    metric,
		"count":     len(stats),
		"leaders":   stats,
	})
}

// ========================================
// GAME ENDPOINTS
// ========================================

// GetGame - GET /api/data/games/:game_id
func (h *DataHandler) GetGame(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	gameID := c.Param("game_id")

	game, err := h.service.GetGame(ctx, gameID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	c.JSON(http.StatusOK, game)
}

// GetGamesBySeason - GET /api/data/games?season=2024&week=1
func (h *DataHandler) GetGamesBySeason(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	season, _ := strconv.Atoi(c.Query("season"))
	week, _ := strconv.Atoi(c.Query("week"))

	games, err := h.service.GetGamesBySeason(ctx, season, week)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch games"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"season": season,
		"week":   week,
		"count":  len(games),
		"games":  games,
	})
}

// GetUpcomingGames - GET /api/data/teams/:team/upcoming
func (h *DataHandler) GetUpcomingGames(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	team := c.Param("team")

	games, err := h.service.GetUpcomingGames(ctx, team)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch games"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"team":  team,
		"count": len(games),
		"games": games,
	})
}

// ========================================
// AGGREGATE ENDPOINTS
// ========================================

// GetPlayerSummary - GET /api/data/players/:nfl_id/summary?season=2024
func (h *DataHandler) GetPlayerSummary(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	nflID := c.Param("nfl_id")
	season, _ := strconv.Atoi(c.DefaultQuery("season", "2024"))

	summary, err := h.service.GetPlayerSummary(ctx, nflID, season)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch player summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetTeamDepthChart - GET /api/data/teams/:team/depth-chart?season=2024
func (h *DataHandler) GetTeamDepthChart(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	team := c.Param("team")
	season, _ := strconv.Atoi(c.DefaultQuery("season", "2025"))

	depthChart, err := h.service.GetTeamDepthChart(ctx, team, season)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch depth chart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"team":        team,
		"season":      season,
		"depth_chart": depthChart,
	})
}
