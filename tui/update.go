package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg: // Handle keyboard input
		switch msg.String() {
		case "enter": // When the user presses Enter
			if m.input.Value() == "exit" {

				// TODO: Add prompt to save
				return m, tea.Quit // Exit the program
			}

			// TODO: Send this to the game
			gameOutput := m.game.ProcessCommand(m.input.Value())
			m.output = gameOutput

			m.debug = fmt.Sprintf("CMD: %s\n", m.input.Value())
			m.input.SetValue("") // Clear the input field

		case "ctrl+c": // Handle Ctrl+C to quit
			return m, tea.Quit

		}

	case tea.WindowSizeMsg: // Handle window resize
		m.input.Width = msg.Width // Adjust input width

	}

	m.input, cmd = m.input.Update(msg)

	return m, cmd
}
