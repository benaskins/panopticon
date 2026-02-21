package ui

import (
	"fmt"
	"strings"

	"github.com/benaskins/panopticon/internal/hw"
	"github.com/charmbracelet/lipgloss"
)

// MemoryPanel renders a vertical memory bar.
type MemoryPanel struct {
	Width  int
	Height int
	Data   hw.MemoryInfo
}

func NewMemoryPanel() MemoryPanel {
	return MemoryPanel{Width: 20, Height: 10}
}

func (m MemoryPanel) View() string {
	usedGB := float64(m.Data.UsedBytes()) / (1024 * 1024 * 1024)
	totalGB := float64(m.Data.TotalBytes) / (1024 * 1024 * 1024)
	pct := m.Data.UsagePercent()

	// Vertical bar: filled from bottom
	barHeight := m.Height - 3 // leave room for label + numbers
	if barHeight < 1 {
		barHeight = 1
	}
	filledRows := int(pct / 100 * float64(barHeight))
	if filledRows > barHeight {
		filledRows = barHeight
	}

	barWidth := m.Width - 4 // padding inside border
	if barWidth < 4 {
		barWidth = 4
	}

	var rows []string
	for i := 0; i < barHeight; i++ {
		rowIdx := barHeight - 1 - i // top to bottom
		if rowIdx < filledRows {
			// Filled
			block := strings.Repeat("█", barWidth)
			color := LerpColor(pct/100, [3]int{25, 80, 60}, [3]int{180, 50, 40})
			rows = append(rows, lipgloss.NewStyle().Foreground(color).Render(block))
		} else {
			// Empty
			block := strings.Repeat("░", barWidth)
			rows = append(rows, lipgloss.NewStyle().Foreground(LabelColor).Render(block))
		}
	}

	bar := strings.Join(rows, "\n")
	label := fmt.Sprintf("%.1f/%.0fGB", usedGB, totalGB)
	pctLabel := fmt.Sprintf("%.0f%%", pct)

	content := lipgloss.JoinVertical(lipgloss.Center,
		bar,
		LabelStyle.Render(label),
		ValueStyle.Render(pctLabel),
	)

	return RenderPanel("Memory", content, m.Width)
}
