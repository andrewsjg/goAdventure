package ollama

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewPlayer(t *testing.T) {
	client := NewClient("", "", 0, -1)
	player := NewPlayer(client, false)

	if player.client != client {
		t.Error("client not set correctly")
	}
	if player.showThinking != false {
		t.Error("showThinking not set correctly")
	}
	if player.maxHistory != 20 {
		t.Errorf("expected maxHistory 20, got %d", player.maxHistory)
	}
	if len(player.history) != 0 {
		t.Error("history should be empty initially")
	}
}

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name            string
		response        string
		expectedCmd     string
		expectedThink   string
		hasThinking     bool
	}{
		{
			name:        "simple command",
			response:    "NORTH",
			expectedCmd: "NORTH",
		},
		{
			name:        "lowercase command",
			response:    "get lamp",
			expectedCmd: "GET LAMP",
		},
		{
			name:        "command with prefix",
			response:    "> SOUTH",
			expectedCmd: "SOUTH",
		},
		{
			name:        "command with Command: prefix",
			response:    "Command: INVENTORY",
			expectedCmd: "INVENTORY",
		},
		{
			name:            "thinking tags",
			response:        "<think>I should explore north</think>\nNORTH",
			expectedCmd:     "NORTH",
			expectedThink:   "I should explore north",
			hasThinking:     true,
		},
		{
			name:            "thinking tags variant",
			response:        "<thinking>Let me check inventory</thinking>\nINVENTORY",
			expectedCmd:     "INVENTORY",
			expectedThink:   "Let me check inventory",
			hasThinking:     true,
		},
		{
			name:        "multiline with extra text",
			response:    "Let me go north\n\nNORTH",
			expectedCmd: "LET ME GO NORTH", // First non-empty line that's <= 50 chars
		},
		{
			name:        "yes response",
			response:    "y",
			expectedCmd: "Y",
		},
		{
			name:        "no response",
			response:    "no",
			expectedCmd: "NO",
		},
		{
			name:        "two word command",
			response:    "take lamp",
			expectedCmd: "TAKE LAMP",
		},
		{
			name:        "whitespace handling",
			response:    "  EAST  \n",
			expectedCmd: "EAST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, thinking := parseResponse(tt.response)
			if cmd != tt.expectedCmd {
				t.Errorf("expected command %q, got %q", tt.expectedCmd, cmd)
			}
			if tt.hasThinking && thinking != tt.expectedThink {
				t.Errorf("expected thinking %q, got %q", tt.expectedThink, thinking)
			}
		})
	}
}

func TestExtractCommand(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"north", "NORTH"},
		{"GET LAMP", "GET LAMP"},
		{"> inventory", "INVENTORY"},
		{"Command: score", "SCORE"},
		{"", "LOOK"},
		{"This is a very long response that should be truncated to just the first couple words", "THIS IS"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractCommand(tt.input)
			if result != tt.expected {
				t.Errorf("extractCommand(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetCommand(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ChatResponse{
			Message: Message{Role: "assistant", Content: "GET LAMP"},
			Done:    true,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model", 0, 0.1)
	player := NewPlayer(client, false)

	ctx := &GameContext{
		GameOutput:     "You are in a building with a shiny brass lamp.",
		LocationDesc:   "Inside building",
		VisibleObjects: []string{"brass lamp"},
		Inventory:      []string{},
		Score:          0,
		Turns:          1,
	}

	cmd, thinking, err := player.GetCommand(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != "GET LAMP" {
		t.Errorf("expected GET LAMP, got %s", cmd)
	}
	if thinking != "" {
		t.Errorf("expected no thinking, got %s", thinking)
	}

	// Check history was updated
	if len(player.history) != 2 {
		t.Errorf("expected 2 history entries, got %d", len(player.history))
	}
}

func TestReset(t *testing.T) {
	client := NewClient("", "", 0, -1)
	player := NewPlayer(client, false)

	// Add some history
	player.history = append(player.history, Message{Role: "user", Content: "test"})
	player.history = append(player.history, Message{Role: "assistant", Content: "NORTH"})

	if len(player.history) != 2 {
		t.Errorf("expected 2 history entries, got %d", len(player.history))
	}

	player.Reset()

	if len(player.history) != 0 {
		t.Errorf("expected 0 history entries after reset, got %d", len(player.history))
	}
}

func TestHistoryTrimming(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ChatResponse{
			Message: Message{Role: "assistant", Content: "NORTH"},
			Done:    true,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model", 0, 0.1)
	player := NewPlayer(client, false)
	player.maxHistory = 3 // Set small for testing

	ctx := &GameContext{
		GameOutput:   "test output",
		LocationDesc: "Test location",
	}

	// Make enough calls to trigger trimming
	for i := 0; i < 10; i++ {
		_, _, err := player.GetCommand(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	// History should be trimmed (maxHistory * 2 = 6, but we trim by 2 when exceeded)
	if len(player.history) > player.maxHistory*2 {
		t.Errorf("history not trimmed correctly: got %d entries", len(player.history))
	}
}

func TestNewRewardTracker(t *testing.T) {
	rt := NewRewardTracker()

	if rt == nil {
		t.Fatal("NewRewardTracker returned nil")
	}
	if rt.MaxHistory != 10 {
		t.Errorf("expected MaxHistory 10, got %d", rt.MaxHistory)
	}
	if len(rt.History) != 0 {
		t.Error("expected empty history")
	}
}

func TestRecordAction(t *testing.T) {
	rt := NewRewardTracker()

	// Test positive reward
	rt.RecordAction("GET LAMP", 0, 5, false)
	if len(rt.History) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(rt.History))
	}
	if rt.History[0].Outcome != "positive" {
		t.Errorf("expected positive outcome, got %s", rt.History[0].Outcome)
	}

	// Test neutral reward
	rt.RecordAction("NORTH", 5, 5, false)
	if rt.History[1].Outcome != "neutral" {
		t.Errorf("expected neutral outcome, got %s", rt.History[1].Outcome)
	}

	// Test negative reward
	rt.RecordAction("DROP LAMP", 5, 3, false)
	if rt.History[2].Outcome != "negative" {
		t.Errorf("expected negative outcome, got %s", rt.History[2].Outcome)
	}

	// Test death
	rt.RecordAction("JUMP", 3, 0, true)
	if rt.History[3].Outcome != "death" {
		t.Errorf("expected death outcome, got %s", rt.History[3].Outcome)
	}
}

func TestRewardTrackerTrimming(t *testing.T) {
	rt := NewRewardTracker()
	rt.MaxHistory = 3

	// Add more than MaxHistory entries
	for i := 0; i < 5; i++ {
		rt.RecordAction("NORTH", i, i, false)
	}

	if len(rt.History) != 3 {
		t.Errorf("expected 3 history entries after trimming, got %d", len(rt.History))
	}
}

func TestGetFeedback(t *testing.T) {
	rt := NewRewardTracker()

	rt.RecordAction("GET LAMP", 0, 5, false)
	rt.RecordAction("NORTH", 5, 5, false)
	rt.RecordAction("JUMP", 5, 0, true)

	feedback := rt.GetFeedback()

	if len(feedback) != 3 {
		t.Errorf("expected 3 feedback entries, got %d", len(feedback))
	}

	// Check positive feedback contains "Good move"
	if !containsString(feedback[0], "Good move") {
		t.Errorf("expected positive feedback to contain 'Good move', got %s", feedback[0])
	}

	// Check death feedback contains "DIED"
	if !containsString(feedback[2], "DIED") {
		t.Errorf("expected death feedback to contain 'DIED', got %s", feedback[2])
	}
}

func TestGetPositiveActions(t *testing.T) {
	rt := NewRewardTracker()

	rt.RecordAction("GET LAMP", 0, 5, false)
	rt.RecordAction("NORTH", 5, 5, false)
	rt.RecordAction("GET KEYS", 5, 10, false)

	positive := rt.GetPositiveActions()

	if len(positive) != 2 {
		t.Errorf("expected 2 positive actions, got %d", len(positive))
	}
}

func TestGetNegativeActions(t *testing.T) {
	rt := NewRewardTracker()

	rt.RecordAction("GET LAMP", 0, 5, false)
	rt.RecordAction("JUMP", 5, 0, true)
	rt.RecordAction("DROP LAMP", 0, -5, false)

	negative := rt.GetNegativeActions()

	if len(negative) != 2 {
		t.Errorf("expected 2 negative actions, got %d", len(negative))
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
