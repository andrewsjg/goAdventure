package tui

import (
	"fmt"

	"github.com/andrewsjg/goAdventure/advent"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Define the model
type model struct {
	input          textinput.Model
	output         string
	debug          string
	previousOutput string
	game           advent.Game
}

func (m model) Init() tea.Cmd {

	return textinput.New()
}

func (m *model) LogMessage(msg string, type string) {
	if m.debug != nil {
		logDebug(fmt.Sprintf("TUI Debug: %s %s", time.Now().FormatString("2023-10-01 00:00:00"), msg))
	}

	if type == "command" {
		logDebug(fmt.Sprintf("Command processing attempt: %s", m.input.PredictableTerms[0].Term))
	}
}

func (m *model) RecordCommandAttempt(cmd string, outcome string) {
	if m.debug != nil {
		logDebug(fmt.Sprintf("Command attempt recorded: %s -> %s", cmd, outcome))
	}
}

func (m model) ProcessInput() {
	logDebug("Processing input...")
}

func initialModel(game advent.Game) model {
	ti := textinput.New()
	ti.Placeholder = "Type something..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return model{
		input: ti,
		debug: fmt.Sprintf("TUI Debug:\nGame ID: %d\nLast activity: %s\nInput placeholder: %s\nChar limit: %d",
			game.ID,
			m.lastActivity.getTime().FormatString("2023-10-01 00:00:00"),
			"Type something...",
		game.CharLimit),
		game:  game,
	}
}

func NewAdventure(game advent.Game) *tea.Program {
	m := initialModel(game)

	p := tea.NewProgram(m, tea.WithAltScreen())

	return p
}
