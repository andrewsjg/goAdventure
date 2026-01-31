package tui

import (
	"fmt"
	"time"

	"github.com/andrewsjg/goAdventure/advent"
	"github.com/andrewsjg/goAdventure/ollama"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Define the model
type model struct {
	input          textinput.Model
	gameOutput     viewport.Model
	content        string
	output         string
	debug          string
	previousOutput string
	game           *advent.Game
	moveHistory    []string // Last N directions moved

	// Command history for up/down navigation
	commandHistory []string
	historyIndex   int // -1 means not browsing history, 0+ means index from end

	// Tab completion state
	completions    []string
	completionIdx  int
	completionBase string // The partial text being completed

	// Pinned location description
	locationDesc string

	// AI player fields
	aiPlayer      *ollama.Player
	aiEnabled     bool
	aiDelay       time.Duration
	showThinking  bool
	aiThinking    string // Last AI reasoning (for display)
	aiIsThinking  bool   // True when waiting for AI response
	aiSpinner     spinner.Model
	rewardTracker *ollama.RewardTracker // Tracks action rewards for AI feedback
	lastScore     int                   // Score before last AI action
}

const maxMoveHistory = 4
const maxCommandHistory = 50

func (m model) Init() tea.Cmd {
	// Start script execution if there are script commands
	if m.game.HasScriptCommands() {
		return tea.Batch(textinput.Blink, scriptTick())
	}
	// Start AI execution if AI is enabled
	if m.aiEnabled && m.aiPlayer != nil {
		return tea.Batch(textinput.Blink, m.aiSpinner.Tick, aiTick(m.aiDelay))
	}
	return textinput.Blink
}

func initialModel(game *advent.Game, aiPlayer *ollama.Player, rewardTracker *ollama.RewardTracker, showThinking bool, aiDelay time.Duration) model {
	ti := textinput.New()
	ti.Placeholder = "What would you like to do?"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	vp := viewport.New(80, 20) // will be resized on WindowSizeMsg

	content := ""
	if game.Output != "" {
		content = game.Output + "\n"
		game.Output = "" // clear after consuming
	}
	vp.SetContent(content)

	// Create spinner for AI thinking indicator
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		input:          ti,
		gameOutput:     vp,
		content:        content,
		debug:          fmt.Sprintf("ZZWORD: %s\nSeedval: %d\nOutput:%s", string(game.Zzword[:]), game.Seedval, game.Output),
		game:           game,
		moveHistory:    make([]string, 0, maxMoveHistory),
		commandHistory: make([]string, 0, maxCommandHistory),
		historyIndex:   -1,
		completions:    nil,
		completionIdx:  0,
		completionBase: "",
		locationDesc:   "",
		aiPlayer:       aiPlayer,
		aiEnabled:      aiPlayer != nil,
		aiDelay:        aiDelay,
		showThinking:   showThinking,
		aiThinking:     "",
		aiIsThinking:   false,
		aiSpinner:      sp,
		rewardTracker:  rewardTracker,
		lastScore:      0,
	}
}

func NewAdventure(game *advent.Game, aiPlayer *ollama.Player, rewardTracker *ollama.RewardTracker, showThinking bool, aiDelay time.Duration) *tea.Program {
	m := initialModel(game, aiPlayer, rewardTracker, showThinking, aiDelay)
	p := tea.NewProgram(m, tea.WithAltScreen())

	return p
}
