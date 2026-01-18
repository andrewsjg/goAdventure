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
	gap := lipgloss.NewStyle().Width(gapWidth).Render("")

	// Make the header and footer

	footerText := fmt.Sprintf("")

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

	mainScreen := lipgloss.JoinHorizontal(lipgloss.Top, mainColumn, gap, inventoryBox)

	return lipgloss.JoinVertical(lipgloss.Top, header, mainScreen, footerText)

}
