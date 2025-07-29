package api

import (
	"ai-assistant/internal/ai"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	ollamaClient *ai.OllamaClient
	startTime    time.Time
}

func NewHandlers(ollamaClient *ai.OllamaClient) *Handlers {
	return &Handlers{
		ollamaClient: ollamaClient,
		startTime:    time.Now(),
	}
}

func (h *Handlers) HealthCheck(c *gin.Context) {
	uptime := time.Since(h.startTime).String()

	connections := map[string]string{
		"ollama": "unknown",
	}

	// Test Ollama connection
	if err := h.ollamaClient.HealthCheck(); err != nil {
		connections["ollama"] = "disconnected"
	} else {
		connections["ollama"] = "connected"
	}

	c.JSON(http.StatusOK, HealthResponse{
		Status:      "healthy",
		Service:     "ai-assistant",
		Version:     "1.0.0",
		Uptime:      uptime,
		Connections: connections,
	})
}

func (h *Handlers) ListModels(c *gin.Context) {
	models, err := h.ollamaClient.ListModels()
	if err != nil {
		log.Printf("Error listing models: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "failed_to_list_models",
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"models": models,
		"count":  len(models),
	})
}

func (h *Handlers) Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	// Set default model if not provided
	model := req.Model
	if model == "" {
		model = "llama3.2"
	}

	// Add current message to conversation history
	messages := append(req.ConversationHistory, ChatMessage{
		Role:      "user",
		Content:   req.Message,
		Timestamp: time.Now(),
	})

	log.Printf("Chat request - Model: %s, Messages: %d, Temperature: %.2f",
		model, len(messages), req.Temperature)

	response, err := h.ollamaClient.Chat(model, messages, req.Temperature)
	if err != nil {
		log.Printf("Chat error: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "chat_failed",
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	log.Printf("Chat response - Tokens: %v, Time: %.2fms",
		response.TokensUsed, response.ProcessTime)

	c.JSON(http.StatusOK, response)
}

func (h *Handlers) StreamChat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		})
		return
	}

	model := req.Model
	if model == "" {
		model = "llama3.2"
	}

	messages := append(req.ConversationHistory, ChatMessage{
		Role:      "user",
		Content:   req.Message,
		Timestamp: time.Now(),
	})

	// Set up Server-Sent Events
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	responseChan := make(chan string)
	errorChan := make(chan error)

	// Start streaming chat in goroutine
	go h.ollamaClient.StreamChat(model, messages, req.Temperature, responseChan, errorChan)

	c.Stream(func(w gin.ResponseWriter) bool {
		select {
		case content, ok := <-responseChan:
			if !ok {
				c.SSEvent("done", "")
				return false
			}
			c.SSEvent("message", content)
			return true

		case err, ok := <-errorChan:
			if ok && err != nil {
				c.SSEvent("error", err.Error())
			}
			return false
		}
	})
}

func (h *Handlers) ServeWeb(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "AI Assistant",
	})
}
