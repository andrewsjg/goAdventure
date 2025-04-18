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

			// TODO: Maybe the game should handle the exit
			// condition in the processcommand function?
			if m.input.Value() == "exit" {

				// TODO: Add prompt to save
				return m, tea.Quit // Exit the program
			}

			if m.game.QueryFlag {
				m.game.QueryResponse = m.input.Value()
				m.game.QueryFlag = false

				m.debug = fmt.Sprintf("QueryFlag Set. Query Response: %s\n", m.game.QueryResponse)

				m.input.SetValue("")

				if m.game.OnQueryResponse != nil {
					m.debug = fmt.Sprintf("Query Response: %s Calling OnQueryResponse \n", m.game.QueryResponse)

					m.output = m.game.OnQueryResponse(m.game.QueryResponse)

				} else {
					m.debug = fmt.Sprintln("No OnQueryResponse function set")
				}

			} else {
				err := m.game.ProcessCommand(m.input.Value())

				if err != nil {
					m.output = fmt.Sprintf("Error: %s", err.Error())
				} else {
					m.output = m.game.Output

					m.debug = fmt.Sprintf("CMD: %s\n", m.input.Value())
					m.input.SetValue("") // Clear the input field
				}
			}

		case "ctrl+c": // Handle Ctrl+C to quit
			return m, tea.Quit

		}

	case tea.WindowSizeMsg: // Handle window resize
		m.input.Width = msg.Width // Adjust input width

	}

	m.input, cmd = m.input.Update(msg)

	return m, cmd
}
