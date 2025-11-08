package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ai-atl/nfl-platform/internal/config"
	"github.com/ai-atl/nfl-platform/internal/handlers"
	"github.com/ai-atl/nfl-platform/internal/middleware"
	"github.com/ai-atl/nfl-platform/internal/services"
	"github.com/ai-atl/nfl-platform/pkg/mongodb"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	mongoClient *mongo.Client
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var err error
	mongoClient, err = mongodb.Connect(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	log.Println("Connected to MongoDB successfully!")

	// Initialize Gin router
	router := gin.Default()

	db := mongoClient.Database(cfg.DBName)
	yahooService := services.NewYahooService(db, cfg)
	fantasyHandler := handlers.NewFantasyHandler(cfg, yahooService)
	espnHandler := handlers.NewESPNHandler(db, "http://localhost:5002")

	// Middleware
	router.Use(middleware.CORS())
	router.Use(middleware.RequestLogger())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "nfl-platform-api",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes
		auth := v1.Group("/auth")
		{
			authHandler := handlers.NewAuthHandler(db)
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Yahoo OAuth callback (public)
		v1.GET("/fantasy/oauth/callback", fantasyHandler.Callback)

		// Protected routes (require JWT)
		protected := v1.Group("")
		protected.Use(middleware.AuthRequired())
		{
			fantasy := protected.Group("/fantasy")
			{
				fantasy.GET("/status", fantasyHandler.Status)
				fantasy.GET("/oauth/url", fantasyHandler.GetAuthURL)
				fantasy.GET("/teams", fantasyHandler.Teams)
			}

			// ESPN Fantasy routes
			espn := protected.Group("/espn")
			{
				espn.POST("/credentials", espnHandler.SaveCredentials)
				espn.GET("/status", espnHandler.GetStatus)
				espn.GET("/roster", espnHandler.GetRoster)
			}

			// Players
			players := protected.Group("/players")
			{
				playerHandler := handlers.NewPlayerHandler(db)
				players.GET("", playerHandler.List)
				players.GET("/:id", playerHandler.Get)
				players.GET("/:id/stats", playerHandler.GetStats)
			}

			// Lineups
			lineups := protected.Group("/lineups")
			{
				lineupHandler := handlers.NewLineupHandler(db)
				lineups.GET("", lineupHandler.List)
				lineups.POST("", lineupHandler.Create)
				lineups.GET("/:id", lineupHandler.Get)
				lineups.PUT("/:id", lineupHandler.Update)
				lineups.DELETE("/:id", lineupHandler.Delete)
				lineups.POST("/optimize", lineupHandler.Optimize)
			}

			// Insights (AI-powered features)
			insights := protected.Group("/insights")
			{
				insightHandler := handlers.NewInsightHandler(db)
				insights.GET("/game_script", insightHandler.GameScript)
				insights.POST("/injury_impact", insightHandler.InjuryImpact)
				insights.GET("/streaks", insightHandler.Streaks)
				insights.GET("/top_performers", insightHandler.TopPerformers)
				insights.GET("/waiver_gems", insightHandler.WaiverGems)
			}

			// Trade Analyzer
			trades := protected.Group("/trades")
			{
				tradeHandler := handlers.NewTradeHandler(db)
				trades.POST("/analyze", tradeHandler.Analyze)
			}

			// Chatbot
			chatbot := protected.Group("/chatbot")
			{
				chatbotHandler := handlers.NewChatbotHandler(db)
				chatbot.POST("/ask", chatbotHandler.Ask)
				chatbot.GET("/history", chatbotHandler.History)
			}

			// Voting
			votes := protected.Group("/votes")
			{
				voteHandler := handlers.NewVoteHandler(db)
				votes.POST("", voteHandler.Create)
				votes.GET("/consensus", voteHandler.GetConsensus)
			}
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
