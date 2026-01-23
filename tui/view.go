package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View function to render the UI
func (m model) View() string {

	// Sort out terminal dimensions and widths for layout
	totalWidth := m.gameOutput.Width - 4
	if totalWidth <= 0 {
		totalWidth = 80
	}

	gapWidth := 1
	inventoryWidth := totalWidth / 3
	if inventoryWidth < 18 {
		inventoryWidth = 18
	}
	if inventoryWidth > 32 {
		inventoryWidth = 32
	}

	if remaining := totalWidth - inventoryWidth - gapWidth; remaining < 24 {
		inventoryWidth = totalWidth - 24 - gapWidth
		if inventoryWidth < 12 {
			inventoryWidth = 12
		}
	}

	mainWidth := totalWidth - inventoryWidth - gapWidth
	if mainWidth < 20 {
		mainWidth = totalWidth
		inventoryWidth = 0
	}

	innerWidth := mainWidth - 4
	if innerWidth < 10 {
		innerWidth = 10
	}

	// Style for the UI elements
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("10")).
		Padding(0, 1, 0, 1)

	inputViewport := m.input
	inputViewport.Width = innerWidth

	outputViewport := m.gameOutput
	outputViewport.Width = innerWidth

	inputStyle := boxStyle.
		AlignHorizontal(lipgloss.Left).
		Width(mainWidth)

	outputStyle := boxStyle.
		AlignHorizontal(lipgloss.Left).
		Width(mainWidth)

	inputBox := inputStyle.Render(inputViewport.View())
	outputBox := outputStyle.Render(outputViewport.View())

	mainColumn := fmt.Sprintf("%s\n%s", outputBox, inputBox)

	/*
		if inventoryWidth <= 0 || m.game == nil {
			return mainColumn
		} */

	inventoryItems := m.game.InventoryDescriptions()

	inventoryLines := make([]string, 0, len(inventoryItems))
	for _, line := range inventoryItems {
		inventoryLines = append(inventoryLines, "• "+line)
	}

	inventoryBody := "Nothing carried."
	if len(inventoryLines) > 0 {
		inventoryBody = strings.Join(inventoryLines, "\n")
	}

	inventTitleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("186"))
	inventoryContent := lipgloss.JoinVertical(lipgloss.Left,
		inventTitleStyle.Render("Inventory\n"),
		inventoryBody,
	)

	inventoryStyle := boxStyle.
		AlignHorizontal(lipgloss.Left).
		AlignVertical(lipgloss.Top).
		Width(inventoryWidth)

	inventoryBox := inventoryStyle.Render(inventoryContent)

	// Movement history pane
	historyTitleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("186"))
	historyBody := "No moves yet."
	if len(m.moveHistory) > 0 {
		historyLines := make([]string, 0, len(m.moveHistory))
		for i := len(m.moveHistory) - 1; i >= 0; i-- {
			historyLines = append(historyLines, "→ "+m.moveHistory[i])
		}
		historyBody = strings.Join(historyLines, "\n")
	}

	historyContent := lipgloss.JoinVertical(lipgloss.Left,
		historyTitleStyle.Render("Recent Moves\n"),
		historyBody,
	)

	historyStyle := boxStyle.
		AlignHorizontal(lipgloss.Left).
		AlignVertical(lipgloss.Top).
		Width(inventoryWidth)

	historyBox := historyStyle.Render(historyContent)

	// Combine inventory and history into right column
	rightColumn := lipgloss.JoinVertical(lipgloss.Top, inventoryBox, historyBox)

	gap := lipgloss.NewStyle().Width(gapWidth).Render("")

	// Make the header and footer
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		PaddingLeft(1)

	footerText := footerStyle.Render(fmt.Sprintf("Score: %d  |  Turns: %d", m.game.GetScore(), m.game.Turns))

	titleString := strings.Join([]string{
		"█▀█ █▀▄ █ █ █▀▀ █▄ █ ▀█▀ █ █ █▀█ █▀▀                  ▀█   █▀",
		"█▀█ █▄▀ ▀▄▀ ██▄ █ ▀█  █  █▄█ █▀▄ ██▄              ▀▄▀ █▄ ▄ ▄█ ",
	}, "\n")

	// 105 = purple
	// 184 = light yellow
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("186"))
	title := titleStyle.Render(titleString)
	headerText := fmt.Sprintf(title)

	header := fmt.Sprintf("%s", headerText)

	mainScreen := lipgloss.JoinHorizontal(lipgloss.Top, mainColumn, gap, rightColumn)

	return lipgloss.JoinVertical(lipgloss.Top, header, mainScreen, footerText)

}
