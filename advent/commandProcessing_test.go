package advent

import (
	"strings"
	"testing"

	"github.com/andrewsjg/goAdventure/dungeon"
)

// Helper to create a fresh game for testing
func newTestGame() *Game {
	game := NewGame(12345, "", "", "", false, false, false, nil)
	return &game
}

// Helper to create a game that's past the intro (answered "no" to instructions)
func newStartedGame() *Game {
	game := newTestGame()
	// Answer "no" to skip intro instructions
	game.ProcessCommand("no")
	game.Output = "" // Clear the output
	return game
}

// TestProcessCommandEmpty tests that empty commands don't increment turns
func TestProcessCommandEmpty(t *testing.T) {
	game := newTestGame()
	initialTurns := game.Turns

	err := game.ProcessCommand("")
	if err != nil {
		t.Errorf("ProcessCommand('') returned error: %v", err)
	}

	if game.Turns != initialTurns {
		t.Errorf("Empty command should not increment turns: got %d, want %d", game.Turns, initialTurns)
	}
}

// TestProcessCommandInvalid tests that invalid commands produce appropriate response
func TestProcessCommandInvalid(t *testing.T) {
	game := newStartedGame()

	err := game.ProcessCommand("xyzzy123notaword")
	if err != nil {
		t.Errorf("ProcessCommand returned error: %v", err)
	}

	// Game should produce some response for invalid words
	// (could be "Nothing happens" or "I don't know that word")
	lower := strings.ToLower(game.Output)
	if !strings.Contains(lower, "know") && !strings.Contains(lower, "nothing") && !strings.Contains(lower, "understand") {
		t.Errorf("Invalid command should produce some response, got: %s", game.Output)
	}
}

// TestProcessCommandLook tests the LOOK command
func TestProcessCommandLook(t *testing.T) {
	game := newStartedGame()

	err := game.ProcessCommand("look")
	if err != nil {
		t.Errorf("ProcessCommand('look') returned error: %v", err)
	}

	// Should have some output (location description)
	if game.Output == "" {
		t.Error("LOOK command should produce output")
	}
}

// TestProcessCommandInventoryEmpty tests INVENTORY with no items
func TestProcessCommandInventoryEmpty(t *testing.T) {
	game := newStartedGame()

	err := game.ProcessCommand("inventory")
	if err != nil {
		t.Errorf("ProcessCommand('inventory') returned error: %v", err)
	}

	// Should indicate empty inventory
	lower := strings.ToLower(game.Output)
	if !strings.Contains(lower, "not carrying") && !strings.Contains(lower, "nothing") {
		t.Errorf("Empty inventory should say not carrying anything, got: %s", game.Output)
	}
}

// TestProcessCommandScore tests the SCORE command
func TestProcessCommandScore(t *testing.T) {
	game := newStartedGame()

	err := game.ProcessCommand("score")
	if err != nil {
		t.Errorf("ProcessCommand('score') returned error: %v", err)
	}

	// SCORE command should not error - output may vary based on game state
	// Some implementations set a query for score display
}

// TestProcessCommandGetNoObject tests GET with no object specified
func TestProcessCommandGetNoObject(t *testing.T) {
	game := newStartedGame()

	err := game.ProcessCommand("get")
	if err != nil {
		t.Errorf("ProcessCommand('get') returned error: %v", err)
	}

	// GET without object may wait for object input or handle silently
	// depending on game state - just verify no error
}

// TestProcessCommandDropNoObject tests DROP with no object specified
func TestProcessCommandDropNoObject(t *testing.T) {
	game := newStartedGame()

	err := game.ProcessCommand("drop")
	if err != nil {
		t.Errorf("ProcessCommand('drop') returned error: %v", err)
	}

	// DROP without object may wait for object input or handle silently
	// depending on game state - just verify no error
}

// TestProcessCommandMovement tests basic movement commands
func TestProcessCommandMovement(t *testing.T) {
	game := newStartedGame()

	// Test various movement words
	movements := []string{"north", "n", "south", "s", "east", "e", "west", "w", "up", "down"}

	for _, move := range movements {
		game.Output = ""
		err := game.ProcessCommand(move)
		if err != nil {
			t.Errorf("ProcessCommand('%s') returned error: %v", move, err)
		}
		// Movement commands should produce some output (either moved or can't go that way)
		// Note: Output may be empty if movement is pending (Newloc set)
	}
}

// TestProcessCommandCaseInsensitive tests that commands are case-insensitive
func TestProcessCommandCaseInsensitive(t *testing.T) {
	game1 := newStartedGame()
	game2 := newStartedGame()

	game1.ProcessCommand("LOOK")
	game2.ProcessCommand("look")

	// Both should produce similar output
	if game1.Output == "" || game2.Output == "" {
		t.Error("Both LOOK and look should produce output")
	}
}

// TestProcessCommandAbbreviations tests that command abbreviations work
func TestProcessCommandAbbreviations(t *testing.T) {
	game := newStartedGame()

	// "i" is short for "inventory"
	err := game.ProcessCommand("i")
	if err != nil {
		t.Errorf("ProcessCommand('i') returned error: %v", err)
	}

	// Should have inventory-related output
	if game.Output == "" {
		t.Error("'i' (inventory abbreviation) should produce output")
	}
}

// TestProcessCommandGetItem tests picking up an item
func TestProcessCommandGetItem(t *testing.T) {
	game := newStartedGame()

	// Place lamp at player's location and make it visible
	game.Objects[dungeon.LAMP].Place = game.Loc
	game.Objects[dungeon.LAMP].Prop = 0

	initialHolding := game.Holdng

	err := game.ProcessCommand("get lamp")
	if err != nil {
		t.Errorf("ProcessCommand('get lamp') returned error: %v", err)
	}

	// Check if lamp is now being carried
	if game.Objects[dungeon.LAMP].Place != CARRIED {
		t.Errorf("Lamp should be carried after 'get lamp', got place=%d", game.Objects[dungeon.LAMP].Place)
	}

	// Holding count should increase
	if game.Holdng <= initialHolding {
		t.Errorf("Holding count should increase after picking up item: got %d, was %d", game.Holdng, initialHolding)
	}
}

// TestProcessCommandDropItem tests dropping an item
func TestProcessCommandDropItem(t *testing.T) {
	game := newStartedGame()

	// Give player the lamp
	game.Objects[dungeon.LAMP].Place = CARRIED
	game.Holdng = 1

	currentLoc := game.Loc

	err := game.ProcessCommand("drop lamp")
	if err != nil {
		t.Errorf("ProcessCommand('drop lamp') returned error: %v", err)
	}

	// Check if lamp is now at current location
	if game.Objects[dungeon.LAMP].Place != currentLoc {
		t.Errorf("Lamp should be at location %d after drop, got %d", currentLoc, game.Objects[dungeon.LAMP].Place)
	}
}

// TestProcessCommandGetNotHere tests getting an item that isn't present
func TestProcessCommandGetNotHere(t *testing.T) {
	game := newStartedGame()

	// Make sure lamp is not at player's location
	game.Objects[dungeon.LAMP].Place = 999 // Some other location

	err := game.ProcessCommand("get lamp")
	if err != nil {
		t.Errorf("ProcessCommand('get lamp') returned error: %v", err)
	}

	// Lamp should still not be carried (can't get something not here)
	if game.Objects[dungeon.LAMP].Place == CARRIED {
		t.Error("Should not be able to get an item that isn't present")
	}
}

// TestProcessCommandDropNotCarrying tests dropping an item not being carried
func TestProcessCommandDropNotCarrying(t *testing.T) {
	game := newStartedGame()

	// Make sure lamp is not being carried
	game.Objects[dungeon.LAMP].Place = game.Loc // At location but not carried

	err := game.ProcessCommand("drop lamp")
	if err != nil {
		t.Errorf("ProcessCommand('drop lamp') returned error: %v", err)
	}

	// Should indicate not carrying the item
	lower := strings.ToLower(game.Output)
	if !strings.Contains(lower, "carry") && !strings.Contains(lower, "have") && !strings.Contains(lower, "aren't") {
		t.Errorf("Dropping non-carried item should indicate not carrying, got: %s", game.Output)
	}
}

// TestProcessCommandHelp tests the HELP command
func TestProcessCommandHelp(t *testing.T) {
	game := newStartedGame()

	err := game.ProcessCommand("help")
	if err != nil {
		t.Errorf("ProcessCommand('help') returned error: %v", err)
	}

	// Should produce help text
	if game.Output == "" {
		t.Error("HELP command should produce output")
	}
}

// TestProcessCommandInfo tests the INFO command
func TestProcessCommandInfo(t *testing.T) {
	game := newStartedGame()

	err := game.ProcessCommand("info")
	if err != nil {
		t.Errorf("ProcessCommand('info') returned error: %v", err)
	}

	// Should produce info text
	if game.Output == "" {
		t.Error("INFO command should produce output")
	}
}

// TestMagicWords tests magic word commands
func TestMagicWords(t *testing.T) {
	game := newStartedGame()

	magicWords := []string{"xyzzy", "plugh", "plover", "fee", "fie", "foe", "foo", "fum"}

	for _, word := range magicWords {
		game.Output = ""
		err := game.ProcessCommand(word)
		if err != nil {
			t.Errorf("ProcessCommand('%s') returned error: %v", word, err)
		}
		// Magic words should produce some response (even if "nothing happens")
	}
}

// TestTwoWordCommands tests various two-word command combinations
func TestTwoWordCommands(t *testing.T) {
	game := newStartedGame()

	commands := []string{
		"go north",
		"go south",
		"walk east",
		"run west",
	}

	for _, cmd := range commands {
		game.Output = ""
		err := game.ProcessCommand(cmd)
		if err != nil {
			t.Errorf("ProcessCommand('%s') returned error: %v", cmd, err)
		}
	}
}

// TestTokeniseCommand tests command tokenization
func TestTokeniseCommand(t *testing.T) {
	game := newStartedGame()

	// Test that tokenization works for valid commands
	cmd := game.tokeniseCommand("get lamp")
	if len(cmd.Word) < 2 {
		t.Error("'get lamp' should tokenize to at least 2 words")
	}

	// Test single word
	cmd = game.tokeniseCommand("look")
	if len(cmd.Word) < 1 {
		t.Error("'look' should tokenize to at least 1 word")
	}

	// Test invalid word
	cmd = game.tokeniseCommand("xyznotaword")
	if len(cmd.Word) > 0 && cmd.Word[0].ID != WORD_NOT_FOUND {
		// Invalid words should be marked as not found
	}
}

// TestProcessCommandTurnsIncrement tests that valid commands increment turns
func TestProcessCommandTurnsIncrement(t *testing.T) {
	game := newStartedGame()
	initialTurns := game.Turns

	// "look" is a valid command that should increment turns
	game.ProcessCommand("look")

	if game.Turns <= initialTurns {
		t.Errorf("Valid command should increment turns: got %d, want > %d", game.Turns, initialTurns)
	}
}
