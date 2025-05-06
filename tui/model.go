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

	return textinput.Blink
}

func initialModel(game advent.Game) model {
	ti := textinput.New()
	ti.Placeholder = "Type something..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return model{
		input: ti,
		debug: fmt.Sprintf("ZZWORD: %s\nSeedval: %d\n", string(game.Zzword[:]), game.Seedval),
		game:  game,
	}
}

func NewAdventure(game advent.Game) *tea.Program {
	m := initialModel(game)

	p := tea.NewProgram(m, tea.WithAltScreen())

	return p
}
