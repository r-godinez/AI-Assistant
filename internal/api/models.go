package api

import (
	"time"
)

// Request/Response models
type ChatMessage struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type ChatRequest struct {
	Message             string        `json:"message" binding:"required"`
	Model               string        `json:"model"`
	ConversationHistory []ChatMessage `json:"conversation_history"`
	Temperature         float32       `json:"temperature"`
	MaxTokens           int           `json:"max_tokens"`
}

type ChatResponse struct {
	Response    string    `json:"response"`
	ModelUsed   string    `json:"model_used"`
	TokensUsed  *int      `json:"tokens_used,omitempty"`
	ProcessTime float64   `json:"process_time_ms"`
	Timestamp   time.Time `json:"timestamp"`
}

type ModelInfo struct {
	Name           string `json:"name"`
	Size           string `json:"size"`
	ParameterCount string `json:"parameter_count,omitempty"`
	Family         string `json:"family,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type HealthResponse struct {
	Status      string            `json:"status"`
	Service     string            `json:"service"`
	Version     string            `json:"version"`
	Uptime      string            `json:"uptime"`
	Connections map[string]string `json:"connections"`
}

// Database models
type Conversation struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title"`
	Messages  string    `json:"messages"` // JSON serialized ChatMessage array
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ConversationSummary struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Preview   string    `json:"preview"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
