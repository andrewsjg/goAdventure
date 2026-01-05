package advent

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/andrewsjg/goAdventure/dungeon"
)

// SaveGame represents the complete game state for serialization
type SaveGame struct {
	Magic   string `json:"magic"`   // Magic string to identify save files
	Version int    `json:"version"` // Save file version
	Game    Game   `json:"game"`    // The actual game state
}

// SaveToFile saves the current game state to a file
func (g *Game) SaveToFile(filename string) error {
	// Create the save file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create save file: %w", err)
	}
	defer file.Close()

	// Create the save structure
	saveData := SaveGame{
		Magic:   ADVENT_MAGIC,
		Version: SAVE_VERSION,
		Game:    *g,
	}

	// Encode the game state as JSON with indentation for readability
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(saveData); err != nil {
		return fmt.Errorf("failed to encode game state: %w", err)
	}

	if g.Settings.EnableDebug {
		fmt.Printf("DEBUG: Game saved to %s\n", filename)
	}

	return nil
}

// LoadFromFile loads a game state from a file
func (g *Game) LoadFromFile(filename string) error {
	// Open the save file
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open save file: %w", err)
	}
	defer file.Close()

	// Decode the save file
	var saveData SaveGame
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&saveData); err != nil {
		return fmt.Errorf("failed to decode save file: %w", err)
	}

	// Validate the save file
	if saveData.Magic != ADVENT_MAGIC {
		return fmt.Errorf("invalid save file: bad magic number")
	}

	if saveData.Version != SAVE_VERSION {
		return fmt.Errorf("save file version mismatch: file is version %d, expected %d",
			saveData.Version, SAVE_VERSION)
	}

	// Validate the game state
	if !isValidGameState(&saveData.Game) {
		return fmt.Errorf("save file contains invalid game state (possible tampering)")
	}

	// Restore the game state
	// Preserve settings and callback from current game
	settings := g.Settings
	callback := g.OnQueryResponse

	*g = saveData.Game

	// Restore the preserved fields
	g.Settings = settings
	g.OnQueryResponse = callback

	if g.Settings.EnableDebug {
		fmt.Printf("DEBUG: Game loaded from %s\n", filename)
	}

	return nil
}

// isValidGameState validates that a loaded game state has valid values
func isValidGameState(g *Game) bool {
	// Check for division by zero
	if g.Abbnum == 0 {
		return false
	}

	// Check RNG overflow
	if g.LcgX >= LCG_M {
		return false
	}

	// Bounds check for locations
	if g.Chloc < -1 || g.Chloc > dungeon.NLOCATIONS ||
		g.Chloc2 < -1 || g.Chloc2 > dungeon.NLOCATIONS ||
		g.Loc < 0 || g.Loc > dungeon.NLOCATIONS ||
		g.Newloc < 0 || g.Newloc > dungeon.NLOCATIONS ||
		g.Oldloc < 0 || g.Oldloc > dungeon.NLOCATIONS ||
		g.Oldlc2 < 0 || g.Oldlc2 > dungeon.NLOCATIONS {
		return false
	}

	// Bounds check for dwarves
	for i := 0; i <= dungeon.NDWARVES; i++ {
		if g.Dwarves[i].Loc < -1 || g.Dwarves[i].Loc > dungeon.NLOCATIONS ||
			g.Dwarves[i].Oldloc < -1 || g.Dwarves[i].Oldloc > dungeon.NLOCATIONS {
			return false
		}
	}

	// Bounds check for objects
	for i := 0; i <= dungeon.NOBJECTS; i++ {
		if g.Objects[i].Place < -1 || g.Objects[i].Place > dungeon.NLOCATIONS ||
			g.Objects[i].Fixed < -1 || g.Objects[i].Fixed > dungeon.NLOCATIONS {
			return false
		}
	}

	// Bounds check for dwarf counts
	if g.Dtotal < 0 || g.Dtotal > dungeon.NDWARVES ||
		g.Dkill < 0 || g.Dkill > dungeon.NDWARVES {
		return false
	}

	// Validate death count
	if g.Numdie >= dungeon.NDEATHS {
		return false
	}

	// Recalculate and verify tally
	tempTally := int32(0)
	for i := 1; i <= dungeon.NOBJECTS; i++ {
		if dungeon.Objects[i].Is_Treasure {
			if g.Objects[i].Prop < 0 { // OBJECT_IS_NOTFOUND2
				tempTally++
			}
		}
	}
	if tempTally != g.Tally {
		return false
	}

	// Validate object properties
	for i := 0; i <= dungeon.NOBJECTS; i++ {
		// Properties should be within reasonable bounds
		// Most properties are 0-3, but some can be negative for stashed items
		if g.Objects[i].Prop < -10 || g.Objects[i].Prop > 10 {
			return false
		}
	}

	// Validate linked lists for objects
	for i := int32(0); i <= dungeon.NLOCATIONS; i++ {
		if g.Locs[i].Atloc < -1 || g.Locs[i].Atloc > dungeon.NOBJECTS*2 {
			return false
		}
	}

	for i := 0; i <= dungeon.NOBJECTS*2; i++ {
		if g.Link[i] < -1 || g.Link[i] > dungeon.NOBJECTS*2 {
			return false
		}
	}

	return true
}

// AutoSave saves the game to the autosave file if autosave is enabled
func (g *Game) AutoSave() error {
	if g.Settings.Autosave && g.Settings.AutoSaveFileName != "" {
		return g.SaveToFile(g.Settings.AutoSaveFileName)
	}
	return nil
}
