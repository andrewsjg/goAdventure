package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// isTimeoutError checks if an error is a timeout error.
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline exceeded")
}

// Client is an HTTP client for the Ollama API.
type Client struct {
	BaseURL     string
	Model       string
	Timeout     time.Duration
	Temperature float64 // 0.0 = deterministic, 1.0 = creative (default 0.7)
}

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"` // "system", "user", "assistant"
	Content string `json:"content"`
}

// ChatOptions contains model parameters for the Ollama API.
type ChatOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
}

// ChatRequest is the request body for the Ollama chat API.
type ChatRequest struct {
	Model    string       `json:"model"`
	Messages []Message    `json:"messages"`
	Stream   bool         `json:"stream"`
	Options  *ChatOptions `json:"options,omitempty"`
}

// ChatResponse is the response from the Ollama chat API.
type ChatResponse struct {
	Model     string  `json:"model"`
	CreatedAt string  `json:"created_at"`
	Message   Message `json:"message"`
	Done      bool    `json:"done"`
}

// NewClient creates a new Ollama client with default settings.
func NewClient(baseURL, model string, timeout time.Duration, temperature float64) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "qwen2.5:7b"
	}
	if timeout <= 0 {
		timeout = 240 * time.Second // Default 4 minutes for slower models
	}
	if temperature < 0 {
		temperature = 0.1 // Default to low temperature for more deterministic game play
	}
	return &Client{
		BaseURL:     baseURL,
		Model:       model,
		Timeout:     timeout,
		Temperature: temperature,
	}
}

// Chat sends a chat request to Ollama and returns the response content.
func (c *Client) Chat(messages []Message) (string, error) {
	reqBody := ChatRequest{
		Model:    c.Model,
		Messages: messages,
		Stream:   false,
		Options: &ChatOptions{
			Temperature: c.Temperature,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.BaseURL + "/api/chat"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: c.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		// Check if it's a timeout error
		if isTimeoutError(err) {
			return "", fmt.Errorf("Ollama request timed out after %v. The model may be loading or is slow to respond.\nTry: -ai-timeout 180 for a longer timeout, or use a smaller/faster model", c.Timeout)
		}
		return "", fmt.Errorf("failed to send request to Ollama at %s: %w\nMake sure Ollama is running (ollama serve)", c.BaseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusNotFound {
			return "", fmt.Errorf("model '%s' not found. Run 'ollama pull %s' to download it", c.Model, c.Model)
		}
		return "", fmt.Errorf("Ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return chatResp.Message.Content, nil
}
