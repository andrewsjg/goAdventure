package tui

import (
	"fmt"
	"time"

	"github.com/andrewsjg/goAdventure/advent"
	"github.com/andrewsjg/goAdventure/ollama"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	// Check if game is over - show final output and quit
	if m.game.GameOver {
		if m.game.Output != "" {
			highlighted := highlightOutput(m.game.Output, m.game)
			m.content += "\n" + highlighted + "\n"
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
		case "up": // Browse command history (older)
			if len(m.commandHistory) > 0 {
				if m.historyIndex < len(m.commandHistory)-1 {
					m.historyIndex++
				}
				idx := len(m.commandHistory) - 1 - m.historyIndex
				m.input.SetValue(m.commandHistory[idx])
				m.input.CursorEnd()
			}
			// Clear any tab completion state
			m.completions = nil
			m.completionIdx = 0
			return m, nil

		case "down": // Browse command history (newer)
			if m.historyIndex > 0 {
				m.historyIndex--
				idx := len(m.commandHistory) - 1 - m.historyIndex
				m.input.SetValue(m.commandHistory[idx])
				m.input.CursorEnd()
			} else if m.historyIndex == 0 {
				m.historyIndex = -1
				m.input.SetValue("")
			}
			// Clear any tab completion state
			m.completions = nil
			m.completionIdx = 0
			return m, nil

		case "tab": // Tab completion
			currentInput := m.input.Value()

			// If we have completions, cycle through them
			if len(m.completions) > 0 {
				m.completionIdx = (m.completionIdx + 1) % len(m.completions)
				m.input.SetValue(m.completions[m.completionIdx])
				m.input.CursorEnd()
				return m, nil
			}

			// Generate new completions
			if currentInput != "" {
				m.completionBase = currentInput
				m.completions = m.game.GetCompletions(currentInput)
				if len(m.completions) > 0 {
					m.completionIdx = 0
					m.input.SetValue(m.completions[0])
					m.input.CursorEnd()
				}
			}
			return m, nil

		case "enter": // When the user presses Enter
			m.previousOutput = m.game.Output

			// Clear tab completion state
			m.completions = nil
			m.completionIdx = 0

			// Add to command history (if non-empty and not a duplicate of last)
			userCmd := m.input.Value()
			if userCmd != "" {
				if len(m.commandHistory) == 0 || m.commandHistory[len(m.commandHistory)-1] != userCmd {
					m.commandHistory = append(m.commandHistory, userCmd)
					if len(m.commandHistory) > maxCommandHistory {
						m.commandHistory = m.commandHistory[1:]
					}
				}
			}
			m.historyIndex = -1 // Reset history browsing

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
						highlighted := highlightOutput(m.game.Output, m.game)
						m.content += highlighted + "\n"
						m.gameOutput.SetContent(m.content)
						m.gameOutput.GotoBottom() // auto-scroll to show latest
						// Only clear output if no new question was set up
						// (nested AskQuestion calls set QueryFlag back to true)
						if !m.game.QueryFlag {
							m.game.Output = ""
						}
					}

					// Ensure no movement happens on next frame to preserve query response output
					m.game.Newloc = m.game.Loc

				} else {
					m.debug = "No OnQueryResponse function set\n"
				}

			} else {
				// Track location before command to detect movement
				locBefore := m.game.Loc

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

		default:
			// Any other key clears the completion state (user is typing new text)
			if len(m.completions) > 0 {
				m.completions = nil
				m.completionIdx = 0
			}
		}

	case temporaryMessageExpiredMsg:
		m.debug = fmt.Sprintf("Timer expired, restoring previous output: %s\n", m.previousOutput)

		// Restore the previous output when the timer expires
		m.game.Output = m.previousOutput
		m.previousOutput = "" // Clear the saved output

		m.game.OutputType = 0 // Reset the output type

	case scriptTickMsg:
		// Execute next script command if available
		if scriptCmd, ok := m.game.NextScriptCommand(); ok {
			// Show the command being executed
			m.content += "\n> " + scriptCmd + "\n"

			if m.game.QueryFlag {
				// Handle query response from script
				m.game.QueryResponse = scriptCmd
				m.game.QueryFlag = false
				if m.game.OnQueryResponse != nil {
					m.game.OnQueryResponse(m.game.QueryResponse, m.game)
				}
			} else {
				// Process regular command
				_ = m.game.ProcessCommand(scriptCmd)
			}

			// Add output and continue script execution
			if m.game.Output != "" {
				highlighted := highlightOutput(m.game.Output, m.game)
				m.content += highlighted + "\n"
				m.gameOutput.SetContent(m.content)
				m.gameOutput.GotoBottom()
				m.game.Output = ""
			}

			// Schedule next script command if more available
			if m.game.HasScriptCommands() && !m.game.GameOver {
				return m, scriptTick()
			}
		}

	case aiTickMsg:
		// AI tick - generate a command asynchronously
		if m.aiEnabled && m.aiPlayer != nil && !m.game.GameOver {
			// Set thinking state
			m.aiIsThinking = true

			// Track score before AI action for reward feedback
			m.lastScore = m.game.GetScore()

			// Build rich context for AI including reward feedback
			ctx := &ollama.GameContext{
				GameOutput:      m.game.Output,
				LocationDesc:    m.game.GetLocationDescription(),
				VisibleObjects:  m.game.GetVisibleObjects(),
				Inventory:       m.game.InventoryDescriptions(),
				Score:           m.game.GetScore(),
				Turns:           int(m.game.Turns),
				Hints:           m.game.GenerateHints(),
				ValidActions:    m.game.GetAllVerbs(),
				ValidDirections: m.game.GetAllDirections(),
				ValidObjects:    m.game.GetInteractableObjects(),
			}
			if m.rewardTracker != nil {
				ctx.RewardFeedback = m.rewardTracker.GetFeedback()
			}

			// Return both the AI call and spinner tick to keep spinner animating
			aiCmd := func() tea.Msg {
				cmd, thinking, err := m.aiPlayer.GetCommand(ctx)
				return aiCommandMsg{command: cmd, thinking: thinking, err: err}
			}
			return m, tea.Batch(aiCmd, m.aiSpinner.Tick)
		}

	case spinner.TickMsg:
		// Update the spinner animation
		if m.aiIsThinking {
			var cmd tea.Cmd
			m.aiSpinner, cmd = m.aiSpinner.Update(msg)
			return m, cmd
		}

	case aiCommandMsg:
		// AI command response received - stop thinking indicator
		m.aiIsThinking = false

		if msg.err != nil {
			m.content += fmt.Sprintf("\n[AI Error: %v]\n", msg.err)
			m.gameOutput.SetContent(m.content)
			m.gameOutput.GotoBottom()
			return m, nil
		}

		// Show AI thinking if enabled
		if m.showThinking && msg.thinking != "" {
			m.aiThinking = msg.thinking
			m.content += fmt.Sprintf("\n[AI Thinking: %s]\n", msg.thinking)
		}

		// Show the command being executed
		m.content += "\n> " + msg.command + "\n"

		// Track location before command to detect movement
		locBefore := m.game.Loc

		if m.game.QueryFlag {
			// Handle query response from AI
			m.game.QueryResponse = msg.command
			m.game.QueryFlag = false
			if m.game.OnQueryResponse != nil {
				m.game.OnQueryResponse(m.game.QueryResponse, m.game)
			}
		} else {
			// Process regular command
			_ = m.game.ProcessCommand(msg.command)
		}

		// If location changed, record the command in move history
		if m.game.Newloc != locBefore || m.game.Loc != locBefore {
			m.moveHistory = append(m.moveHistory, msg.command)
			if len(m.moveHistory) > maxMoveHistory {
				m.moveHistory = m.moveHistory[1:]
			}
		}

		// Add output
		if m.game.Output != "" {
			highlighted := highlightOutput(m.game.Output, m.game)
			m.content += highlighted + "\n"
			m.gameOutput.SetContent(m.content)
			m.gameOutput.GotoBottom()
			m.game.Output = ""
		}

		// Record reward for AI action
		if m.rewardTracker != nil {
			scoreAfter := m.game.GetScore()
			died := m.game.GameOver
			m.rewardTracker.RecordAction(msg.command, m.lastScore, scoreAfter, died)
		}

		// Schedule next AI command if game not over
		if m.aiEnabled && !m.game.GameOver {
			return m, aiTick(m.aiDelay)
		}

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
		highlighted := highlightOutput(m.game.Output, m.game)
		m.content += "\n" + highlighted + "\n"
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

// Message type for script command execution
type scriptTickMsg struct{}

// scriptTick returns a command that triggers script execution after a short delay
func scriptTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return scriptTickMsg{}
	})
}

// Message types for AI player
type aiTickMsg struct{}

type aiCommandMsg struct {
	command  string
	thinking string
	err      error
}

// aiTick returns a command that triggers AI command generation after a delay
func aiTick(delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		return aiTickMsg{}
	})
}
