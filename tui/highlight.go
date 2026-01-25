package tui

import (
	"regexp"
	"strings"

	"github.com/andrewsjg/goAdventure/advent"
	"github.com/andrewsjg/goAdventure/dungeon"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Styles for different content types
	objectStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214")) // Orange for objects
	warnStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red for warnings
	actionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))  // Green for successful actions
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // Dim for less important text

	// Warning patterns to highlight in red
	warningPatterns = []string{
		"dwarf",
		"knife",
		"threatening",
		"attack",
		"dead",
		"killed",
		"dying",
		"dark",
		"pit",
		"fell",
		"broken",
		"snake",
		"dragon",
		"bear",
		"troll",
		"pirate",
		"lamp.*dim",
		"lamp.*out",
		"batteries",
	}

	// Action success patterns
	actionPatterns = []string{
		"^OK$",
		"^Taken\\.",
		"^Dropped\\.",
		"^Done\\.",
	}
)

// highlightOutput applies syntax highlighting to game output
func highlightOutput(output string, game *advent.Game) string {
	if output == "" {
		return output
	}

	lines := strings.Split(output, "\n")
	var result []string

	for _, line := range lines {
		highlighted := highlightLine(line, game)
		result = append(result, highlighted)
	}

	return strings.Join(result, "\n")
}

// highlightLine applies highlighting to a single line
func highlightLine(line string, game *advent.Game) string {
	if line == "" {
		return line
	}

	// Check for warning patterns first (highest priority)
	lowerLine := strings.ToLower(line)
	for _, pattern := range warningPatterns {
		matched, _ := regexp.MatchString(pattern, lowerLine)
		if matched {
			return warnStyle.Render(line)
		}
	}

	// Check for action success patterns
	for _, pattern := range actionPatterns {
		matched, _ := regexp.MatchString("(?i)"+pattern, line)
		if matched {
			return actionStyle.Render(line)
		}
	}

	// Highlight object names within the line
	return highlightObjects(line)
}

// highlightObjects highlights known object names in the text
func highlightObjects(line string) string {
	// Build a map of object words to highlight
	objectWords := make(map[string]bool)
	for i := 1; i < len(dungeon.Objects); i++ {
		for _, word := range dungeon.Objects[i].Words.Strs {
			if word != "" && len(word) > 2 {
				objectWords[strings.ToLower(word)] = true
			}
		}
	}

	// Find and highlight object words
	words := strings.Fields(line)
	for i, word := range words {
		// Clean punctuation for matching
		cleanWord := strings.ToLower(strings.Trim(word, ".,!?;:\"'()"))
		if objectWords[cleanWord] {
			// Preserve original punctuation but highlight the word
			prefix := ""
			suffix := ""
			for _, r := range word {
				if r == '.' || r == ',' || r == '!' || r == '?' || r == ';' || r == ':' || r == '"' || r == '\'' || r == '(' {
					prefix += string(r)
				} else {
					break
				}
			}
			for j := len(word) - 1; j >= 0; j-- {
				r := rune(word[j])
				if r == '.' || r == ',' || r == '!' || r == '?' || r == ';' || r == ':' || r == '"' || r == '\'' || r == ')' {
					suffix = string(r) + suffix
				} else {
					break
				}
			}
			middle := word[len(prefix) : len(word)-len(suffix)]
			words[i] = prefix + objectStyle.Render(middle) + suffix
		}
	}

	return strings.Join(words, " ")
}
