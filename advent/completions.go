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
