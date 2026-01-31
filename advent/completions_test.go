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

// TestStateCheckMethods tests the game state checking methods
func TestStateCheckMethods(t *testing.T) {
	game := NewGame(0, "", "", "", false, false, false, nil)

	// At the start, player should be at the starting location (road), not the building
	if !game.IsAtStart() {
		t.Errorf("Game should start at LOC_START, Loc=%d, LOC_START=%d", game.Loc, dungeon.LOC_START)
	}

	// Player starts outside the cave
	if game.IsInCave() {
		t.Error("Player should not start in the cave")
	}

	// Player doesn't have items at start
	if game.HasLamp() {
		t.Error("Player should not have lamp at start")
	}
	if game.HasKeys() {
		t.Error("Player should not have keys at start")
	}
}

// TestCanSeeItems tests visibility checks for items
func TestCanSeeItems(t *testing.T) {
	game := NewGame(0, "", "", "", false, false, false, nil)

	// Player starts at LOC_START (road), lamp and keys are at LOC_BUILDING
	// So we can't see them from the starting location
	if game.CanSeeLamp() {
		t.Error("Should not see lamp at starting location (it's in the building)")
	}
	if game.CanSeeKeys() {
		t.Error("Should not see keys at starting location (they're in the building)")
	}

	// Move to building and check again
	game.Loc = int32(dungeon.LOC_BUILDING)
	if !game.CanSeeLamp() {
		t.Error("Should see lamp at building")
	}
	if !game.CanSeeKeys() {
		t.Error("Should see keys at building")
	}
}

// TestGenerateHintsWaitingForInstructions tests hint when game first starts
func TestGenerateHintsWaitingForInstructions(t *testing.T) {
	game := NewGame(0, "", "", "", false, false, false, nil)
	// NewGame starts with Settings.NewGame = true

	hints := game.GenerateHints()

	// Should suggest answering N to instructions
	if len(hints) == 0 {
		t.Error("Should generate hint for instructions question")
	}

	foundNHint := false
	for _, hint := range hints {
		if contains(hint, "N") {
			foundNHint = true
			break
		}
	}
	if !foundNHint {
		t.Errorf("Should have hint to answer N, got hints: %v", hints)
	}
}

// TestGenerateHintsAtStart tests hint generation at the starting location
func TestGenerateHintsAtStart(t *testing.T) {
	game := NewGame(0, "", "", "", false, false, false, nil)
	game.Settings.NewGame = false // Past the instructions question

	hints := game.GenerateHints()

	// At the start without items, we should get a hint to go to the building
	if len(hints) == 0 {
		t.Error("Should generate hints at starting location")
	}

	// Should suggest going east to the building
	foundEastHint := false
	for _, hint := range hints {
		if contains(hint, "EAST") {
			foundEastHint = true
			break
		}
	}
	if !foundEastHint {
		t.Errorf("Should have hint about EAST (go to building), got hints: %v", hints)
	}
}

// TestGenerateHintsAtBuilding tests hint generation at the building
func TestGenerateHintsAtBuilding(t *testing.T) {
	game := NewGame(0, "", "", "", false, false, false, nil)
	game.Settings.NewGame = false          // Past the instructions question
	game.Loc = int32(dungeon.LOC_BUILDING) // Move to building

	hints := game.GenerateHints()

	// At the building without items, we should get hints to pick them up
	if len(hints) == 0 {
		t.Error("Should generate hints at building without items")
	}

	// Should suggest getting lamp
	foundLampHint := false
	for _, hint := range hints {
		if contains(hint, "LAMP") {
			foundLampHint = true
			break
		}
	}
	if !foundLampHint {
		t.Errorf("Should have hint about LAMP, got hints: %v", hints)
	}
}

// TestGenerateHintsWithItems tests that hints change after picking up items
func TestGenerateHintsWithItems(t *testing.T) {
	game := NewGame(0, "", "", "", false, false, false, nil)
	game.Settings.NewGame = false // Past the instructions question

	// Pick up lamp and keys
	game.Objects[dungeon.LAMP].Place = -1 // -1 means carried
	game.Objects[dungeon.KEYS].Place = -1

	hints := game.GenerateHints()

	// At starting location with lamp and keys, should suggest going south
	foundSouthHint := false
	for _, hint := range hints {
		if contains(hint, "SOUTH") {
			foundSouthHint = true
			break
		}
	}
	if !foundSouthHint {
		t.Errorf("With lamp and keys at start, should have hint about SOUTH, got hints: %v", hints)
	}
}

// TestGenerateHintsAtBuildingWithItems tests hints when player has items at building
func TestGenerateHintsAtBuildingWithItems(t *testing.T) {
	game := NewGame(0, "", "", "", false, false, false, nil)
	game.Settings.NewGame = false // Past the instructions question
	game.Loc = int32(dungeon.LOC_BUILDING)

	// Pick up lamp, keys, food, and bottle (all items at building)
	game.Objects[dungeon.LAMP].Place = -1
	game.Objects[dungeon.KEYS].Place = -1
	game.Objects[dungeon.FOOD].Place = -1
	game.Objects[dungeon.BOTTLE].Place = -1

	hints := game.GenerateHints()

	// At building with all items, should suggest leaving west
	foundWestHint := false
	for _, hint := range hints {
		if contains(hint, "WEST") {
			foundWestHint = true
			break
		}
	}
	if !foundWestHint {
		t.Errorf("With all items at building, should have hint about WEST, got hints: %v", hints)
	}
}

// TestGenerateHintsAtBuildingWithSomeItems tests hints when only some items picked up
func TestGenerateHintsAtBuildingWithSomeItems(t *testing.T) {
	game := NewGame(0, "", "", "", false, false, false, nil)
	game.Settings.NewGame = false // Past the instructions question
	game.Loc = int32(dungeon.LOC_BUILDING)

	// Pick up lamp and keys, but leave food and bottle
	game.Objects[dungeon.LAMP].Place = -1
	game.Objects[dungeon.KEYS].Place = -1

	hints := game.GenerateHints()

	// Should still suggest getting food/bottle before leaving
	foundFoodOrBottleHint := false
	for _, hint := range hints {
		if contains(hint, "FOOD") || contains(hint, "BOTTLE") {
			foundFoodOrBottleHint = true
			break
		}
	}
	if !foundFoodOrBottleHint {
		t.Errorf("With some items remaining, should suggest getting FOOD or BOTTLE, got hints: %v", hints)
	}
}

// Helper function for string contains
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
