package ui

import (
	"fmt"
	"strings"

	"github.com/benaskins/panopticon/internal/hw"
	"github.com/charmbracelet/lipgloss"
)

// MemoryPanel renders a vertical memory bar with top processes.
type MemoryPanel struct {
	Width    int
	Height   int
	Data     hw.MemoryInfo
	TopProcs []hw.MemProc
}

func NewMemoryPanel() MemoryPanel {
	return MemoryPanel{Width: 20, Height: 10}
}

func (m MemoryPanel) View() string {
	usedGB := float64(m.Data.UsedBytes()) / (1024 * 1024 * 1024)
	totalGB := float64(m.Data.TotalBytes) / (1024 * 1024 * 1024)
	pct := m.Data.UsagePercent()

	// Narrow vertical bar (4 chars wide)
	barHeight := m.Height - 5 // room for border, summary, spacing
	if barHeight < 3 {
		barHeight = 3
	}
	filledRows := int(pct / 100 * float64(barHeight))
	if filledRows > barHeight {
		filledRows = barHeight
	}

	barWidth := 4
	color := LerpColor(pct/100, [3]int{25, 80, 60}, [3]int{180, 50, 40})

	var barRows []string
	for i := 0; i < barHeight; i++ {
		rowIdx := barHeight - 1 - i
		if rowIdx < filledRows {
			block := strings.Repeat("█", barWidth)
			barRows = append(barRows, lipgloss.NewStyle().Foreground(color).Render(block))
		} else {
			block := strings.Repeat("░", barWidth)
			barRows = append(barRows, lipgloss.NewStyle().Foreground(LabelColor).Render(block))
		}
	}

	bar := strings.Join(barRows, "\n")
	summary := fmt.Sprintf("%.0f/%.0fG %s",
		usedGB, totalGB,
		ValueStyle.Render(fmt.Sprintf("%.0f%%", pct)))

	// Top processes
	procList := m.renderProcs(barHeight)

	var content string
	if procList != "" {
		// Bar on left, procs on right
		left := lipgloss.JoinVertical(lipgloss.Center, bar)
		content = lipgloss.JoinHorizontal(lipgloss.Top, left, " ", procList)
		content += "\n" + LabelStyle.Render(summary)
	} else {
		content = lipgloss.JoinVertical(lipgloss.Center, bar, LabelStyle.Render(summary))
	}

	return RenderPanel("Memory", content, m.Width)
}

func (m MemoryPanel) renderProcs(maxLines int) string {
	if len(m.TopProcs) == 0 {
		return ""
	}

	n := maxLines
	if len(m.TopProcs) < n {
		n = len(m.TopProcs)
	}

	nameStyle := lipgloss.NewStyle().Foreground(LabelColor)
	sizeStyle := lipgloss.NewStyle().Foreground(ValueColor)

	var lines []string
	for i := 0; i < n; i++ {
		p := m.TopProcs[i]
		name := p.Name
		if len(name) > 10 {
			name = name[:10]
		}
		size := formatBytes(p.RSS)
		lines = append(lines, fmt.Sprintf("%s %s",
			nameStyle.Render(fmt.Sprintf("%-10s", name)),
			sizeStyle.Render(size),
		))
	}

	return strings.Join(lines, "\n")
}

func formatBytes(b uint64) string {
	gb := float64(b) / (1024 * 1024 * 1024)
	if gb >= 1.0 {
		return fmt.Sprintf("%.1fG", gb)
	}
	mb := float64(b) / (1024 * 1024)
	return fmt.Sprintf("%.0fM", mb)
}
