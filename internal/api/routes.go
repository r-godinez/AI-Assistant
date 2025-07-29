package api

import (
	"ai-assistant/internal/ai"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, ollamaClient *ai.OllamaClient) {
	handlers := NewHandlers(ollamaClient)

	// Serve static files
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*")

	// Web interface
	r.GET("/", handlers.ServeWeb)

	// API routes
	api := r.Group("/api")
	{
		api.GET("/health", handlers.HealthCheck)
		api.GET("/models", handlers.ListModels)
		api.POST("/chat", handlers.Chat)
		api.POST("/chat/stream", handlers.StreamChat)
	}

	// Handle 404 for unknown routes
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "route_not_found",
			Code:    http.StatusNotFound,
			Message: "The requested endpoint does not exist",
		})
	})
}
