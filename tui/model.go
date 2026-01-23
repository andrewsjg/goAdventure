package tui

import (
	"fmt"

	"github.com/andrewsjg/goAdventure/advent"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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
}

const maxMoveHistory = 4

func (m model) Init() tea.Cmd {

	return textinput.Blink
}

func initialModel(game *advent.Game) model {
	ti := textinput.New()
	ti.Placeholder = "Type something..."
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

	return model{
		input:       ti,
		gameOutput:  vp,
		content:     content,
		debug:       fmt.Sprintf("ZZWORD: %s\nSeedval: %d\nOutput:%s", string(game.Zzword[:]), game.Seedval, game.Output),
		game:        game,
		moveHistory: make([]string, 0, maxMoveHistory),
	}
}

func NewAdventure(game *advent.Game) *tea.Program {
	m := initialModel(game)
	p := tea.NewProgram(m, tea.WithAltScreen())

	return p
}
