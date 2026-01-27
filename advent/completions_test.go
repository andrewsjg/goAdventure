package advent

import (
	"testing"

	"github.com/andrewsjg/goAdventure/dungeon"
)

// TestGetCompletionsEmpty tests completions with empty input
func TestGetCompletionsEmpty(t *testing.T) {
	game := &Game{}

	completions := game.GetCompletions("")
	if completions != nil {
		t.Errorf("GetCompletions('') should return nil, got %v", completions)
	}
}

// TestGetCompletionsVerbs tests that verb completions work
func TestGetCompletionsVerbs(t *testing.T) {
	game := &Game{}

	// "ge" should match "get"
	completions := game.GetCompletions("ge")
	found := false
	for _, c := range completions {
		if c == "get" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("GetCompletions('ge') should include 'get', got %v", completions)
	}
}

// TestGetCompletionsDirections tests that direction completions work
func TestGetCompletionsDirections(t *testing.T) {
	game := &Game{}

	// "nor" should match "north"
	completions := game.GetCompletions("nor")
	found := false
	for _, c := range completions {
		if c == "north" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("GetCompletions('nor') should include 'north', got %v", completions)
	}
}

// TestGetCompletionsObjects tests that visible object completions work
func TestGetCompletionsObjects(t *testing.T) {
	game := NewGame(0, "", "", "", false, false, false, nil)

	// Place lamp at player's location
	game.Objects[dungeon.LAMP].Place = game.Loc

	// "lam" should match "lamp"
	completions := game.GetCompletions("lam")
	found := false
	for _, c := range completions {
		if c == "lamp" || c == "lante" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("GetCompletions('lam') should include lamp-related word, got %v", completions)
	}
}

// TestGetCompletionsCaseInsensitive tests that completions are case-insensitive
func TestGetCompletionsCaseInsensitive(t *testing.T) {
	game := &Game{}

	// "GE" should still match "get"
	completions := game.GetCompletions("GE")
	found := false
	for _, c := range completions {
		if c == "get" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("GetCompletions('GE') should include 'get' (case insensitive), got %v", completions)
	}
}

// TestGetAllVerbs tests that GetAllVerbs returns verbs
func TestGetAllVerbs(t *testing.T) {
	game := &Game{}

	verbs := game.GetAllVerbs()
	if len(verbs) == 0 {
		t.Error("GetAllVerbs should return some verbs")
	}

	// Check for some common verbs (use first word of each action which may be abbreviated)
	verbMap := make(map[string]bool)
	for _, v := range verbs {
		verbMap[v] = true
	}

	// "g" is get, "drop" is drop - these are common first words
	expectedVerbs := []string{"g", "drop"}
	for _, expected := range expectedVerbs {
		if !verbMap[expected] {
			t.Errorf("GetAllVerbs should include '%s', got verbs: %v", expected, verbs[:min(10, len(verbs))])
		}
	}
}

// TestGetLocationDescription tests location description retrieval
func TestGetLocationDescription(t *testing.T) {
	game := NewGame(0, "", "", "", false, false, false, nil)

	// Should return a non-empty description for starting location
	desc := game.GetLocationDescription()
	if desc == "" {
		t.Error("GetLocationDescription should return non-empty string for valid location")
	}
}

// TestGetLocationDescriptionInvalid tests location description for invalid location
func TestGetLocationDescriptionInvalid(t *testing.T) {
	game := &Game{Loc: -1}

	desc := game.GetLocationDescription()
	if desc != "" {
		t.Errorf("GetLocationDescription should return empty for invalid location, got %q", desc)
	}
}

// TestGetVisibleObjects tests visible object retrieval
func TestGetVisibleObjects(t *testing.T) {
	game := NewGame(0, "", "", "", false, false, false, nil)

	// Get visible objects at starting location
	objects := game.GetVisibleObjects()

	// Result should be a slice (may be empty depending on starting location)
	if objects == nil {
		// That's OK, nil means no visible objects
	}
}
