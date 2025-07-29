package ai

import (
	"ai-assistant/internal/api"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type OllamaClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewOllamaClient(baseURL string) *OllamaClient {
	return &OllamaClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // AI responses can take time
		},
	}
}

func (c *OllamaClient) HealthCheck() error {
	resp, err := c.httpClient.Get(c.baseURL + "/api/version")
	if err != nil {
		return fmt.Errorf("ollama health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	return nil
}

func (c *OllamaClient) ListModels() ([]api.ModelInfo, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/tags")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name    string `json:"name"`
			Size    int64  `json:"size"`
			Digest  string `json:"digest"`
			Details struct {
				Format            string   `json:"format"`
				Family            string   `json:"family"`
				Families          []string `json:"families"`
				ParameterSize     string   `json:"parameter_size"`
				QuantizationLevel string   `json:"quantization_level"`
			} `json:"details"`
			ModifiedAt time.Time `json:"modified_at"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode models response: %w", err)
	}

	models := make([]api.ModelInfo, len(result.Models))
	for i, model := range result.Models {
		models[i] = api.ModelInfo{
			Name:           model.Name,
			Size:           formatBytes(model.Size),
			ParameterCount: model.Details.ParameterSize,
			Family:         model.Details.Family,
		}
	}

	return models, nil
}

func (c *OllamaClient) Chat(model string, messages []api.ChatMessage, temperature float32) (*api.ChatResponse, error) {
	startTime := time.Now()

	if temperature == 0 {
		temperature = 0.7
	}

	// Convert our ChatMessage format to Ollama's expected format
	ollamaMessages := make([]map[string]string, len(messages))
	for i, msg := range messages {
		ollamaMessages[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	payload := map[string]interface{}{
		"model":    model,
		"messages": ollamaMessages,
		"stream":   false,
		"options": map[string]interface{}{
			"temperature": temperature,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/chat",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send chat request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var result struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		CreatedAt          time.Time `json:"created_at"`
		Done               bool      `json:"done"`
		TotalDuration      int64     `json:"total_duration"`
		LoadDuration       int64     `json:"load_duration"`
		PromptEvalCount    int       `json:"prompt_eval_count"`
		PromptEvalDuration int64     `json:"prompt_eval_duration"`
		EvalCount          int       `json:"eval_count"`
		EvalDuration       int64     `json:"eval_duration"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode chat response: %w", err)
	}

	processTime := time.Since(startTime).Seconds() * 1000 // Convert to milliseconds

	return &api.ChatResponse{
		Response:    result.Message.Content,
		ModelUsed:   model,
		TokensUsed:  &result.EvalCount,
		ProcessTime: processTime,
		Timestamp:   time.Now(),
	}, nil
}

func (c *OllamaClient) StreamChat(model string, messages []api.ChatMessage, temperature float32, responseChan chan<- string, errorChan chan<- error) {
	defer close(responseChan)
	defer close(errorChan)

	if temperature == 0 {
		temperature = 0.7
	}

	ollamaMessages := make([]map[string]string, len(messages))
	for i, msg := range messages {
		ollamaMessages[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	payload := map[string]interface{}{
		"model":    model,
		"messages": ollamaMessages,
		"stream":   true,
		"options": map[string]interface{}{
			"temperature": temperature,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		errorChan <- fmt.Errorf("failed to marshal request: %w", err)
		return
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/chat",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		errorChan <- fmt.Errorf("failed to send chat request: %w", err)
		return
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	for {
		var result struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Done bool `json:"done"`
		}

		if err := decoder.Decode(&result); err != nil {
			if err.Error() != "EOF" {
				errorChan <- fmt.Errorf("failed to decode stream response: %w", err)
			}
			return
		}

		if result.Message.Content != "" {
			responseChan <- result.Message.Content
		}

		if result.Done {
			return
		}
	}
}

// Utility function to format bytes
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
