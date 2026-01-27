package advent

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRandRange tests the random number generator range function
func TestRandRange(t *testing.T) {
	game := &Game{LcgX: 12345} // Initialize with a seed

	tests := []struct {
		name     string
		rndRange int32
		wantMin  int32
		wantMax  int32
	}{
		{"range 5", 5, 0, 4},
		{"range 10", 10, 0, 9},
		{"range 100", 100, 0, 99},
		{"range 1", 1, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run multiple times to check bounds
			for i := 0; i < 100; i++ {
				got := game.randRange(tt.rndRange)
				if got < tt.wantMin || got > tt.wantMax {
					t.Errorf("randRange(%d) = %d, want between %d and %d",
						tt.rndRange, got, tt.wantMin, tt.wantMax)
				}
			}
		})
	}
}

// TestRandRangeZero tests that randRange handles zero gracefully
func TestRandRangeZero(t *testing.T) {
	game := &Game{LcgX: 12345}

	got := game.randRange(0)
	if got != 0 {
		t.Errorf("randRange(0) = %d, want 0", got)
	}
}

// TestRandRangeNegative tests that randRange handles negative values gracefully
func TestRandRangeNegative(t *testing.T) {
	game := &Game{LcgX: 12345}

	got := game.randRange(-5)
	if got != 0 {
		t.Errorf("randRange(-5) = %d, want 0", got)
	}
}

// TestLoadScript tests script file loading
func TestLoadScript(t *testing.T) {
	game := &Game{}

	// Create a temporary script file
	tmpDir := t.TempDir()
	scriptFile := filepath.Join(tmpDir, "test.script")

	content := `# This is a comment
north
south
# Another comment

east
get lamp
`
	if err := os.WriteFile(scriptFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Load the script
	if err := game.LoadScript(scriptFile); err != nil {
		t.Fatalf("LoadScript failed: %v", err)
	}

	// Check commands were loaded correctly (comments and empty lines filtered)
	expected := []string{"north", "south", "east", "get lamp"}
	if len(game.ScriptCommands) != len(expected) {
		t.Errorf("Got %d commands, want %d", len(game.ScriptCommands), len(expected))
	}

	for i, cmd := range expected {
		if i < len(game.ScriptCommands) && game.ScriptCommands[i] != cmd {
			t.Errorf("Command %d: got %q, want %q", i, game.ScriptCommands[i], cmd)
		}
	}
}

// TestLoadScriptNotFound tests loading a non-existent script file
func TestLoadScriptNotFound(t *testing.T) {
	game := &Game{}

	err := game.LoadScript("/nonexistent/path/script.txt")
	if err == nil {
		t.Error("LoadScript should fail for non-existent file")
	}
}

// TestNextScriptCommand tests retrieving commands from script
func TestNextScriptCommand(t *testing.T) {
	game := &Game{
		ScriptCommands: []string{"first", "second", "third"},
		ScriptIndex:    0,
	}

	// Get first command
	cmd, ok := game.NextScriptCommand()
	if !ok || cmd != "first" {
		t.Errorf("First command: got %q, %v; want 'first', true", cmd, ok)
	}

	// Get second command
	cmd, ok = game.NextScriptCommand()
	if !ok || cmd != "second" {
		t.Errorf("Second command: got %q, %v; want 'second', true", cmd, ok)
	}

	// Get third command
	cmd, ok = game.NextScriptCommand()
	if !ok || cmd != "third" {
		t.Errorf("Third command: got %q, %v; want 'third', true", cmd, ok)
	}

	// No more commands
	cmd, ok = game.NextScriptCommand()
	if ok || cmd != "" {
		t.Errorf("Fourth command: got %q, %v; want '', false", cmd, ok)
	}
}

// TestHasScriptCommands tests checking for remaining script commands
func TestHasScriptCommands(t *testing.T) {
	game := &Game{
		ScriptCommands: []string{"one", "two"},
		ScriptIndex:    0,
	}

	if !game.HasScriptCommands() {
		t.Error("HasScriptCommands should return true when commands remain")
	}

	game.ScriptIndex = 2
	if game.HasScriptCommands() {
		t.Error("HasScriptCommands should return false when no commands remain")
	}
}

// TestGetNextLCGValue tests the LCG random number generator
func TestGetNextLCGValue(t *testing.T) {
	game := &Game{LcgX: 0}

	// LCG should produce deterministic sequence
	val1 := game.getNextLCGValue()
	val2 := game.getNextLCGValue()

	if val1 == val2 {
		t.Error("LCG should produce different values on consecutive calls")
	}

	// Values should be within LCG_M
	if val1 >= LCG_M || val2 >= LCG_M {
		t.Errorf("LCG values should be < %d, got %d and %d", LCG_M, val1, val2)
	}
}

// TestPct tests the percentage function
func TestPct(t *testing.T) {
	// pct(100) should always return true
	for i := 0; i < 100; i++ {
		if !pct(100) {
			t.Error("pct(100) should always return true")
		}
	}

	// pct(0) should always return false
	for i := 0; i < 100; i++ {
		if pct(0) {
			t.Error("pct(0) should always return false")
		}
	}
}
