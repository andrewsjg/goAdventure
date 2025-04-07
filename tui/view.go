package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// View function to render the UI
func (m model) View() string {

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(1, 2).
		BorderForeground(lipgloss.Color("63")).
		Width(m.input.Width - 4)

	contentBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(1, 2).
		BorderForeground(lipgloss.Color("63")).
		Width(m.input.Width - 4).Height(10)

	commandBox := boxStyle.Render(fmt.Sprintf(" %s", m.input.View()))
	contentBox := contentBoxStyle.Render(m.output)

	if m.game.Settings.EnableDebug {
		debugBox := boxStyle.Render(fmt.Sprintf("DEBUG:\n------\n\n%s", m.debug))
		contentBox = fmt.Sprintf("%s\n%s", contentBox, debugBox)
	}

	return fmt.Sprintf(

		"%s\n%s",
		contentBox,
		commandBox)
}
