package tui

import (
	"fmt"
	"time"

	"github.com/andrewsjg/goAdventure/advent"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	// Check if game is over - show final output and quit
	if m.game.GameOver {
		if m.game.Output != "" {
			m.content += "\n" + m.game.Output + "\n"
			m.gameOutput.SetContent(m.content)
			m.game.Output = ""
		}
		return m, tea.Quit
	}

	// Perform the move if newloc has been set (but not if waiting for query response)
	if m.game.Newloc != m.game.Loc && !m.game.QueryFlag {
		m.game.DoMove()
		m.game.DescribeLocation()
		m.game.ListObjects()
	}

	var cmd tea.Cmd
	var vpCmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg: // Handle keyboard input
		switch msg.String() {
		case "enter": // When the user presses Enter

			m.previousOutput = m.game.Output

			//			}

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
					m.game.OnQueryResponse(m.game.QueryResponse, m.game)

					m.output = m.game.Output

					if m.game.Output != "" {
						m.content += m.game.Output + "\n"
						m.gameOutput.SetContent(m.content)
						m.gameOutput.GotoBottom() // auto-scroll to show latest
						m.game.Output = ""        // clear to prevent duplicate add at end of Update
					}

					// Ensure no movement happens on next frame to preserve query response output
					m.game.Newloc = m.game.Loc

				} else {
					m.debug = "No OnQueryResponse function set\n"
				}

			} else {
				// Track location before command to detect movement
				locBefore := m.game.Loc
				userCmd := m.input.Value()

				err := m.game.ProcessCommand(userCmd)

				if err != nil {
					m.output = fmt.Sprintf("Error: %s", err.Error())
				} else {
					m.output = m.game.Output

					// If location changed or newloc is set, record the direction
					if m.game.Newloc != locBefore || m.game.Loc != locBefore {
						m.moveHistory = append(m.moveHistory, userCmd)
						if len(m.moveHistory) > maxMoveHistory {
							m.moveHistory = m.moveHistory[1:]
						}
					}

					m.debug = fmt.Sprintf("CMD: %s LOC: %d\nOutput: %s\n", userCmd, m.game.Loc, m.output)
					m.input.SetValue("") // Clear the input field

				}
			}

			// TODO: I dont know if I like this. Maybe we need another message panel?
			// If the output type is 1, set a timer to clear the message
			if m.game.OutputType == advent.MSG_TEMP {
				m.debug = fmt.Sprintf("OutputType: %d Previous Output: %s \n", m.game.OutputType, m.previousOutput)
				m.game.OutputType = advent.MSG_REG // Reset the output type

				return m, temporaryMessageTimer(2 * time.Second)
			}

		case "ctrl+c": // Handle Ctrl+C to quit
			return m, tea.Quit

		}

	case temporaryMessageExpiredMsg:
		m.debug = fmt.Sprintf("Timer expired, restoring previous output: %s\n", m.previousOutput)

		// Restore the previous output when the timer expires
		m.game.Output = m.previousOutput
		m.previousOutput = "" // Clear the saved output

		m.game.OutputType = 0 // Reset the output type

	default:
		// No command to process yet
		// Check for forced moves (but not during a query)
		if m.game.LocForced() && !m.game.QueryFlag {
			m.game.MoveHere()
		}

	case tea.WindowSizeMsg: // Handle window resize
		m.input.Width = msg.Width // Adjust input width

		m.gameOutput.Width = msg.Width
		m.gameOutput.Height = 40           //msg.Height - 3 // leave room for input + prompt
		m.gameOutput.SetContent(m.content) // Re-apply content after resize

	}

	if m.game.Output != "" {
		m.content += "\n" + m.game.Output + "\n"
		m.gameOutput.SetContent(m.content)
		m.gameOutput.GotoBottom()
		m.game.Output = "" // clear it so we don't re-add
	}

	m.input, cmd = m.input.Update(msg)

	// Only pass non-key messages to viewport (prevents typing from scrolling)
	switch msg.(type) {
	case tea.KeyMsg:
		// Don't pass key events to viewport
	default:
		m.gameOutput, vpCmd = m.gameOutput.Update(msg)
	}

	return m, tea.Batch(cmd, vpCmd)
}

func temporaryMessageTimer(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(t time.Time) tea.Msg {
		return temporaryMessageExpiredMsg{}
	})
}

// Message type for when the timer expires
type temporaryMessageExpiredMsg struct{}
