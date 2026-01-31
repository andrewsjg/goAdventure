package advent

import (
	"strings"

	"github.com/andrewsjg/goAdventure/dungeon"
)

// GetCompletions returns possible completions for a partial command.
// It returns verbs, directions, and visible/carried objects that match the prefix.
func (g *Game) GetCompletions(partial string) []string {
	partial = strings.ToUpper(partial)
	if partial == "" {
		return nil
	}

	seen := make(map[string]bool)
	var completions []string

	addCompletion := func(word string) {
		word = strings.ToLower(word)
		if word != "" && strings.HasPrefix(strings.ToUpper(word), partial) && !seen[word] {
			seen[word] = true
			completions = append(completions, word)
		}
	}

	// Add action verbs (GET, DROP, LOOK, etc.)
	for _, action := range dungeon.Actions {
		for _, word := range action.Words.Strs {
			if word != "" {
				addCompletion(word)
			}
		}
	}

	// Add motion words (NORTH, SOUTH, UP, DOWN, etc.)
	for _, motion := range dungeon.Motions {
		for _, word := range motion.Words.Strs {
			if word != "" {
				addCompletion(word)
			}
		}
	}

	// Add visible objects at current location
	for i := 1; i <= dungeon.NOBJECTS; i++ {
		if g.Objects[i].Place == g.Loc || g.Objects[i].Place == -1 {
			// Object is here or being carried
			for _, word := range dungeon.Objects[i].Words.Strs {
				if word != "" {
					addCompletion(word)
				}
			}
		}
	}

	return completions
}

// GetAllVerbs returns all known action verbs for help display
func (g *Game) GetAllVerbs() []string {
	seen := make(map[string]bool)
	var verbs []string

	for _, action := range dungeon.Actions {
		if len(action.Words.Strs) > 0 && action.Words.Strs[0] != "" {
			word := strings.ToLower(action.Words.Strs[0])
			if !seen[word] {
				seen[word] = true
				verbs = append(verbs, word)
			}
		}
	}

	return verbs
}

// GetAllDirections returns common movement directions
func (g *Game) GetAllDirections() []string {
	// Return the most common/useful directions
	return []string{"north", "south", "east", "west", "up", "down", "in", "out"}
}

// GetInteractableObjects returns names of objects the player can currently interact with
// (objects at current location or being carried)
func (g *Game) GetInteractableObjects() []string {
	seen := make(map[string]bool)
	var objects []string

	for i := 1; i <= dungeon.NOBJECTS; i++ {
		if g.Objects[i].Place == g.Loc || g.Objects[i].Place == -1 {
			// Object is here or being carried - get its primary name
			if len(dungeon.Objects[i].Words.Strs) > 0 && dungeon.Objects[i].Words.Strs[0] != "" {
				word := strings.ToLower(dungeon.Objects[i].Words.Strs[0])
				if !seen[word] {
					seen[word] = true
					objects = append(objects, word)
				}
			}
		}
	}

	return objects
}

// GetLocationDescription returns the current location description without side effects
func (g *Game) GetLocationDescription() string {
	if g.Loc <= 0 || int(g.Loc) >= len(dungeon.Locations) {
		return ""
	}

	// Check if dark
	if g.dark() {
		return dungeon.Arbitrary_Messages[dungeon.PITCH_DARK]
	}

	// Get appropriate description (short or long based on visit count)
	desc := dungeon.Locations[g.Loc].Description.Small
	if desc == "" {
		desc = dungeon.Locations[g.Loc].Description.Big
	}

	return desc
}

// GetVisibleObjects returns descriptions of objects at the current location
func (g *Game) GetVisibleObjects() []string {
	if g.dark() {
		return nil
	}

	var objects []string
	for i := 1; i <= dungeon.NOBJECTS; i++ {
		if g.Objects[i].Place == g.Loc {
			// Get object description based on state
			prop := g.Objects[i].Prop
			if prop < 0 {
				prop = 0 // Default state for unfound objects
			}
			if int(prop) < len(dungeon.Objects[i].Descriptions) {
				desc := dungeon.Objects[i].Descriptions[prop]
				if desc != "" {
					objects = append(objects, desc)
				}
			}
		}
	}
	return objects
}

// IsDark returns true if the current location is dark
func (g *Game) IsDark() bool {
	return g.dark()
}

// IsLampOn returns true if the lamp is lit
func (g *Game) IsLampOn() bool {
	return g.Objects[dungeon.LAMP].Prop == dungeon.LAMP_BRIGHT
}

// HasLamp returns true if player is carrying the lamp
func (g *Game) HasLamp() bool {
	return g.toting(int(dungeon.LAMP))
}

// HasKeys returns true if player is carrying the keys
func (g *Game) HasKeys() bool {
	return g.toting(int(dungeon.KEYS))
}

// IsGrateOpen returns true if the grate is open
func (g *Game) IsGrateOpen() bool {
	return g.Objects[dungeon.GRATE].Prop == dungeon.GRATE_OPEN
}

// IsAtGrate returns true if player is at the grate location
func (g *Game) IsAtGrate() bool {
	return g.Loc == int32(dungeon.LOC_GRATE)
}

// IsWaitingForInstructions returns true if the game is waiting for Y/N answer
// about whether to show instructions (at very start of game)
func (g *Game) IsWaitingForInstructions() bool {
	return g.Settings.NewGame
}

// IsAtStart returns true if player is at the starting location (road)
func (g *Game) IsAtStart() bool {
	return g.Loc == int32(dungeon.LOC_START)
}

// IsAtBuilding returns true if player is at the starting building
func (g *Game) IsAtBuilding() bool {
	return g.Loc == int32(dungeon.LOC_BUILDING)
}

// IsInCave returns true if player is inside the cave (below grate)
func (g *Game) IsInCave() bool {
	return inside(g.Loc) && g.Loc != int32(dungeon.LOC_BUILDING)
}

// CanSeeLamp returns true if the lamp is visible at current location
func (g *Game) CanSeeLamp() bool {
	return g.Objects[dungeon.LAMP].Place == g.Loc
}

// CanSeeKeys returns true if keys are visible at current location
func (g *Game) CanSeeKeys() bool {
	return g.Objects[dungeon.KEYS].Place == g.Loc
}

// CanSeeFood returns true if food is visible at current location
func (g *Game) CanSeeFood() bool {
	return g.Objects[dungeon.FOOD].Place == g.Loc
}

// CanSeeBottle returns true if bottle is visible at current location
func (g *Game) CanSeeBottle() bool {
	return g.Objects[dungeon.BOTTLE].Place == g.Loc
}

// GenerateHints creates state-aware hints based on the current game situation.
// These hints help guide the AI player through the game.
func (g *Game) GenerateHints() []string {
	var hints []string

	// Very first: check if waiting for instructions Y/N prompt
	if g.IsWaitingForInstructions() {
		hints = append(hints, "N - Answer NO to the instructions question")
		return hints // This is the only valid action right now
	}

	// At starting location - go to the building first
	if g.IsAtStart() {
		if !g.HasLamp() || !g.HasKeys() {
			hints = append(hints, "EAST - Enter the building to get the lamp and keys")
		} else {
			hints = append(hints, "SOUTH - Head toward the grate with your supplies")
		}
	}

	// At building - pick up items
	if g.IsAtBuilding() {
		if g.CanSeeLamp() && !g.HasLamp() {
			hints = append(hints, "GET LAMP - You need the lamp to explore dark caves")
		}
		if g.CanSeeKeys() && !g.HasKeys() {
			hints = append(hints, "GET KEYS - You need keys to unlock the grate")
		}
		if g.CanSeeFood() {
			hints = append(hints, "GET FOOD - Food is useful for feeding animals")
		}
		if g.CanSeeBottle() {
			hints = append(hints, "GET BOTTLE - The water may be useful")
		}
		// If we have lamp and keys, suggest leaving
		if g.HasLamp() && g.HasKeys() && len(hints) == 0 {
			hints = append(hints, "WEST - Leave the building and head toward the cave")
		}
	}

	// At the grate
	if g.IsAtGrate() {
		if !g.IsGrateOpen() {
			if g.HasKeys() {
				hints = append(hints, "UNLOCK GRATE - Use your keys to unlock the grate")
			} else {
				hints = append(hints, "You need KEYS to open the grate. Go back to the building.")
			}
		} else {
			hints = append(hints, "DOWN - Enter the cave through the open grate")
		}
	}

	// In a dark place without lamp lit
	if g.IsDark() {
		if g.HasLamp() && !g.IsLampOn() {
			hints = append(hints, "LIGHT LAMP - It's dark here! Light your lamp before moving")
		} else if !g.HasLamp() {
			hints = append(hints, "DANGER! It's dark and you don't have a lamp. Go back!")
		}
	}

	// In the cave with lamp
	if g.IsInCave() && g.HasLamp() && g.IsLampOn() && len(hints) == 0 {
		hints = append(hints, "Explore! Look for treasures and try different directions.")
	}

	return hints
}
