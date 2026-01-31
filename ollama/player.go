package ollama

import (
	"fmt"
	"regexp"
	"strings"
)

const defaultSystemPrompt = `You are playing the classic text adventure game Colossal Cave Adventure, also known as ADVENT or Adventure.

Your goal is to explore the cave, collect treasures, and achieve the highest score.


=== RULES ===
- Respond with ONLY the command. Commands are 1 or 2 words ONLY
- The game only understand the 1-2 word commands as input
- Examples: "NORTH", "GET LAMP", "UNLOCK GRATE", "LIGHT LAMP", "Y", "N"
- If the game asks a yes/no question, respond with "Y" or "N"
- ALWAYS follow the SUGGESTED ACTION if one is provided.
- Prioritize actions that previously gave positive rewards.

=== EARLY GAME WALKTHROUGH ===
Follow these steps at the start of the game:
1. You start at the well house (building). GET LAMP and GET KEYS first.
2. Also GET FOOD and GET BOTTLE (water) - you'll need these later.
3. Exit the building.
4. Go SOUTH, then SOUTH again to reach the grate.
5. UNLOCK GRATE (requires keys), then OPEN GRATE.
6. DOWN to enter the cave.
7. LIGHT LAMP before moving in dark areas.
8. LOOK 
9. Explore the cave, collecting treasures (gold, diamonds, etc.)
10. Bring treasures back to the well house for points.

=== STRATEGY TIPS ===
- The LAMP is essential - get it first, always keep it lit in caves
- KEYS are needed to unlock the grate at the cave entrance
- FOOD can feed animals you encounter (like a bear or snake)
- The BOTTLE with water may be needed for certain puzzles
- Magic words XYZZY and PLUGH teleport you between locations
- If you die, you may be resurrected but lose points
- Snakes don't like birds
- A rod dropped on the ground near the bird allows you to catch it if you have the cage
- Pay attention to the cardinal directions that the game says there are paths to go on
- Favour following one direction (N,S,E,W,U,D) until there is an alternative path to explore or item to pick up or puzzle to solve
- If the game offers no direction to go, continue in the direction you were going before
- If a puzzle requires multiple steps, do them in two word command blocks. i.e DROP ROD, then GET BIRD
- For more context about the game issue the HELP command

=== USING VALID COMMANDS ===
- You will see a "VALID COMMANDS" section showing exactly what you can do
- ONLY use commands from the DIRECTIONS, ACTIONS, and OBJECTS lists provided
- Combine an ACTION with an OBJECT: "GET LAMP", "DROP KEYS", "UNLOCK GRATE"
- Use DIRECTIONS alone: "NORTH", "SOUTH", "UP", "DOWN"
- If an object is listed under "OBJECTS YOU CAN USE", you can interact with it

=== LEARNING FROM REWARDS ===
- You will see "RECENT ACTION REWARDS" showing your past actions and their point outcomes
- Actions marked "+X points! Good move!" are GOOD - repeat similar actions
- Actions marked "DIED!" are VERY BAD - never repeat these
- Actions with "no score change" are neutral - focus on finding rewarding actions
- Your goal is to MAXIMIZE SCORE by learning from these rewards

`

// GameContext provides rich context about the current game state.
type GameContext struct {
	GameOutput      string   // The latest output from the game
	LocationDesc    string   // Description of current location
	VisibleObjects  []string // Objects at current location
	Inventory       []string // Items being carried
	Score           int      // Current score
	Turns           int      // Number of turns taken
	Hints           []string // State-aware hints/suggested actions
	RewardFeedback  []string // Recent actions with their score changes
	ValidActions    []string // Valid action verbs (GET, DROP, LOOK, etc.)
	ValidDirections []string // Valid movement directions (NORTH, SOUTH, etc.)
	ValidObjects    []string // Objects player can interact with
}

// ActionReward represents the outcome of a single action.
type ActionReward struct {
	Action      string
	ScoreBefore int
	ScoreAfter  int
	Outcome     string // "positive", "negative", "neutral", "death"
}

// RewardTracker tracks action outcomes for reinforcement feedback.
type RewardTracker struct {
	History    []ActionReward
	MaxHistory int
	LastScore  int
}

// NewRewardTracker creates a new reward tracker.
func NewRewardTracker() *RewardTracker {
	return &RewardTracker{
		History:    make([]ActionReward, 0),
		MaxHistory: 10,
		LastScore:  0,
	}
}

// RecordAction records an action and its outcome.
func (rt *RewardTracker) RecordAction(action string, scoreBefore, scoreAfter int, died bool) {
	outcome := "neutral"
	if died {
		outcome = "death"
	} else if scoreAfter > scoreBefore {
		outcome = "positive"
	} else if scoreAfter < scoreBefore {
		outcome = "negative"
	}

	rt.History = append(rt.History, ActionReward{
		Action:      action,
		ScoreBefore: scoreBefore,
		ScoreAfter:  scoreAfter,
		Outcome:     outcome,
	})

	// Trim history if needed
	if len(rt.History) > rt.MaxHistory {
		rt.History = rt.History[1:]
	}

	rt.LastScore = scoreAfter
}

// GetFeedback returns formatted feedback strings for recent actions.
func (rt *RewardTracker) GetFeedback() []string {
	var feedback []string
	for _, ar := range rt.History {
		change := ar.ScoreAfter - ar.ScoreBefore
		var msg string
		switch ar.Outcome {
		case "positive":
			msg = fmt.Sprintf("%s: +%d points! Good move!", ar.Action, change)
		case "negative":
			msg = fmt.Sprintf("%s: %d points. Avoid this action.", ar.Action, change)
		case "death":
			msg = fmt.Sprintf("%s: DIED! Never do this again!", ar.Action)
		default:
			msg = fmt.Sprintf("%s: no score change", ar.Action)
		}
		feedback = append(feedback, msg)
	}
	return feedback
}

// GetPositiveActions returns actions that resulted in score gains.
func (rt *RewardTracker) GetPositiveActions() []string {
	var positive []string
	for _, ar := range rt.History {
		if ar.Outcome == "positive" {
			positive = append(positive, ar.Action)
		}
	}
	return positive
}

// GetNegativeActions returns actions that resulted in score loss or death.
func (rt *RewardTracker) GetNegativeActions() []string {
	var negative []string
	for _, ar := range rt.History {
		if ar.Outcome == "negative" || ar.Outcome == "death" {
			negative = append(negative, ar.Action)
		}
	}
	return negative
}

// FormatContext builds a formatted string from the game context.
func (gc *GameContext) FormatContext() string {
	var sb strings.Builder

	sb.WriteString("=== CURRENT STATE ===\n")

	if gc.LocationDesc != "" {
		fmt.Fprintf(&sb, "LOCATION: %s\n", gc.LocationDesc)
	}

	if len(gc.VisibleObjects) > 0 {
		fmt.Fprintf(&sb, "OBJECTS HERE: %s\n", strings.Join(gc.VisibleObjects, ", "))
	} else {
		sb.WriteString("OBJECTS HERE: None visible\n")
	}

	if len(gc.Inventory) > 0 {
		fmt.Fprintf(&sb, "CARRYING: %s\n", strings.Join(gc.Inventory, ", "))
	} else {
		sb.WriteString("CARRYING: Nothing\n")
	}

	fmt.Fprintf(&sb, "SCORE: %d | TURNS: %d\n", gc.Score, gc.Turns)

	// Show valid commands the player can use
	sb.WriteString("\n=== VALID COMMANDS ===\n")
	if len(gc.ValidDirections) > 0 {
		fmt.Fprintf(&sb, "DIRECTIONS: %s\n", strings.ToUpper(strings.Join(gc.ValidDirections, ", ")))
	}
	if len(gc.ValidActions) > 0 {
		// Show just a subset of the most useful actions
		usefulActions := []string{}
		for _, a := range gc.ValidActions {
			switch strings.ToLower(a) {
			case "get", "drop", "look", "inventory", "open", "unlock", "light", "on", "off", "read", "eat", "drink", "take":
				usefulActions = append(usefulActions, a)
			}
		}
		if len(usefulActions) > 0 {
			fmt.Fprintf(&sb, "ACTIONS: %s\n", strings.ToUpper(strings.Join(usefulActions, ", ")))
		}
	}
	if len(gc.ValidObjects) > 0 {
		fmt.Fprintf(&sb, "OBJECTS YOU CAN USE: %s\n", strings.ToUpper(strings.Join(gc.ValidObjects, ", ")))
	}

	// Include reward feedback from recent actions
	if len(gc.RewardFeedback) > 0 {
		sb.WriteString("\n=== RECENT ACTION REWARDS ===\n")
		// Show last few rewards (most recent last)
		start := 0
		if len(gc.RewardFeedback) > 5 {
			start = len(gc.RewardFeedback) - 5
		}
		for _, feedback := range gc.RewardFeedback[start:] {
			fmt.Fprintf(&sb, "  %s\n", feedback)
		}
	}

	// Include state-aware hints if available
	if len(gc.Hints) > 0 {
		sb.WriteString("\n=== SUGGESTED ACTION ===\n")
		sb.WriteString(gc.Hints[0]) // Show the most important hint
		sb.WriteString("\n")
	}

	sb.WriteString("\n=== GAME OUTPUT ===\n")
	sb.WriteString(gc.GameOutput)

	return sb.String()
}

// Player manages AI interactions with the game.
type Player struct {
	client       *Client
	history      []Message
	maxHistory   int
	showThinking bool
	systemPrompt string
}

// NewPlayer creates a new AI player.
func NewPlayer(client *Client, showThinking bool) *Player {
	return &Player{
		client:       client,
		history:      make([]Message, 0),
		maxHistory:   20,
		showThinking: showThinking,
		systemPrompt: defaultSystemPrompt,
	}
}

// GetCommand generates a command based on the game context.
// Returns (command, thinking, error).
func (p *Player) GetCommand(ctx *GameContext) (string, string, error) {
	// Format the context into a rich prompt
	contextStr := ctx.FormatContext()

	// Add the game context as a user message
	p.history = append(p.history, Message{
		Role:    "user",
		Content: contextStr,
	})

	// Trim history if needed (keep pairs to maintain context)
	if len(p.history) > p.maxHistory*2 {
		p.history = p.history[2:]
	}

	// Build messages with system prompt
	messages := []Message{
		{Role: "system", Content: p.systemPrompt},
	}
	messages = append(messages, p.history...)

	// Get response from Ollama
	response, err := p.client.Chat(messages)
	if err != nil {
		return "", "", err
	}

	// Parse the response
	command, thinking := parseResponse(response)

	// Add the assistant's response to history
	p.history = append(p.history, Message{
		Role:    "assistant",
		Content: response,
	})

	return command, thinking, nil
}

// Reset clears the conversation history for a new game.
func (p *Player) Reset() {
	p.history = make([]Message, 0)
}

// parseResponse extracts the command and any thinking from the model's response.
func parseResponse(response string) (command string, thinking string) {
	response = strings.TrimSpace(response)

	// Handle thinking tags: <think>...</think> or <thinking>...</thinking>
	thinkRegex := regexp.MustCompile(`(?is)<think(?:ing)?>(.*?)</think(?:ing)?>`)
	matches := thinkRegex.FindStringSubmatch(response)
	if len(matches) > 1 {
		thinking = strings.TrimSpace(matches[1])
		response = thinkRegex.ReplaceAllString(response, "")
		response = strings.TrimSpace(response)
	}

	// If empty after removing thinking, return first line of thinking as command
	if response == "" && thinking != "" {
		lines := strings.Split(thinking, "\n")
		return extractCommand(lines[len(lines)-1]), thinking
	}

	return extractCommand(response), thinking
}

// extractCommand finds a valid command in the response.
func extractCommand(response string) string {
	response = strings.TrimSpace(response)

	// Handle empty response
	if response == "" {
		return "LOOK"
	}

	// Split into lines and find the first non-empty one
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Remove common prefixes that models sometimes add
		line = strings.TrimPrefix(line, "> ")
		line = strings.TrimPrefix(line, "Command: ")
		line = strings.TrimPrefix(line, "command: ")

		// If it's a short response (likely a command), return it
		if len(line) <= 50 {
			return strings.ToUpper(line)
		}

		// Try to extract just the first word or two if the line is too long
		words := strings.Fields(line)
		if len(words) >= 1 && len(words) <= 3 {
			return strings.ToUpper(strings.Join(words, " "))
		}
		if len(words) > 3 {
			// Take first two words as command
			return strings.ToUpper(strings.Join(words[:2], " "))
		}
	}

	// Fallback: return the whole response uppercased if short enough
	if len(response) <= 30 {
		return strings.ToUpper(response)
	}

	// Last resort: first two words
	words := strings.Fields(response)
	if len(words) >= 2 {
		return strings.ToUpper(strings.Join(words[:2], " "))
	}
	if len(words) == 1 {
		return strings.ToUpper(words[0])
	}

	return "LOOK"
}
