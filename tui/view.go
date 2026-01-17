package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// View function to render the UI
func (m model) View() string {
	// Calculate input box width - account for border (2) and padding (2)
	inputWidth := m.gameOutput.Width - 4
	if inputWidth < 10 {
		inputWidth = 10
	}

	// Style for the input box
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		AlignHorizontal(lipgloss.Left).
		BorderForeground(lipgloss.Color("10")).
		Padding(0, 1, 0, 1).
		Width(inputWidth)

	// Render the game output area (using viewport for scrolling)
	gameView := m.gameOutput.View()

	// Render the input box
	inputBox := inputStyle.Render(m.input.View())

	// Combine: game output on top, input at bottom
	return fmt.Sprintf("%s\n%s", gameView, inputBox)
}
