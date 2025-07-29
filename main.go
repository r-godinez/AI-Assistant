package main

import (
	"ai-assistant/internal/ai"
	"ai-assistant/internal/api"
	"ai-assistant/internal/config"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize Ollama client
	ollamaClient := ai.NewOllamaClient(cfg.OllamaURL)

	// Test Ollama connection
	if err := ollamaClient.HealthCheck(); err != nil {
		log.Printf("⚠️  Warning: Ollama not available: %v", err)
		log.Println("💡 Make sure Ollama is running: ollama serve")
	} else {
		log.Println("✅ Ollama connection successful")
	}

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	r := gin.Default()

	// Setup routes
	api.SetupRoutes(r, ollamaClient)

	// Start server
	log.Printf("🚀 AI Assistant starting on http://localhost:%s", cfg.Port)
	log.Printf("📱 Web interface: http://localhost:%s", cfg.Port)
	log.Printf("🔗 API docs: http://localhost:%s/api/health", cfg.Port)

	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
