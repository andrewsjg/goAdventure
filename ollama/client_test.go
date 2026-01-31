package ollama

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	// Test defaults
	c := NewClient("", "", 0, -1)
	if c.BaseURL != "http://localhost:11434" {
		t.Errorf("expected default BaseURL, got %s", c.BaseURL)
	}
	if c.Model != "qwen2.5:7b" {
		t.Errorf("expected default Model, got %s", c.Model)
	}

	// Test custom values
	c = NewClient("http://example.com", "llama3.1:8b", 0, 0.5)
	if c.BaseURL != "http://example.com" {
		t.Errorf("expected custom BaseURL, got %s", c.BaseURL)
	}
	if c.Model != "llama3.1:8b" {
		t.Errorf("expected custom Model, got %s", c.Model)
	}
}

func TestChat(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Errorf("expected path /api/chat, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		// Decode request to verify it
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if req.Model != "test-model" {
			t.Errorf("expected model test-model, got %s", req.Model)
		}

		// Return a response
		resp := ChatResponse{
			Message: Message{Role: "assistant", Content: "NORTH"},
			Done:    true,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model", 0, 0.1)
	messages := []Message{
		{Role: "user", Content: "You are at the entrance to a cave."},
	}

	response, err := client.Chat(messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response != "NORTH" {
		t.Errorf("expected NORTH, got %s", response)
	}
}

func TestChatModelNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("model not found"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "nonexistent-model", 0, 0.1)
	_, err := client.Chat([]Message{{Role: "user", Content: "test"}})

	if err == nil {
		t.Fatal("expected error for missing model")
	}
	if !contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %s", err.Error())
	}
}

func TestChatServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model", 0, 0.1)
	_, err := client.Chat([]Message{{Role: "user", Content: "test"}})

	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
