package handlers

import (
	"net/http"

	"github.com/ai-atl/nfl-platform/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type ChatbotHandler struct {
	db             *mongo.Database
	chatbotService *services.ChatbotService
}

func NewChatbotHandler(db *mongo.Database) *ChatbotHandler {
	return &ChatbotHandler{
		db:             db,
		chatbotService: services.NewChatbotService(db),
	}
}

type ChatRequest struct {
	Question string `json:"question" binding:"required"`
}

// Ask handles a question to the AI chatbot
func (h *ChatbotHandler) Ask(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.chatbotService.Ask(c.Request.Context(), userID.(string), req.Question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"question": req.Question,
		"response": response,
	})
}

// History returns chat history for the user
func (h *ChatbotHandler) History(c *gin.Context) {
	// TODO: Implement chat history retrieval
	c.JSON(http.StatusOK, gin.H{
		"history": []map[string]interface{}{
			{
				"question":  "Who should I start at RB?",
				"response":  "Based on matchups and recent performance...",
				"timestamp": "2024-11-08T10:30:00Z",
			},
		},
	})
}

