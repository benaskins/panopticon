package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// LogsPanel shows a tail of log lines for the selected service.
type LogsPanel struct {
	Width       int
	Height      int
	ServiceName string
	Lines       []string
	ScrollPos   int
}

func NewLogsPanel() LogsPanel {
	return LogsPanel{}
}

func (l LogsPanel) View() string {
	title := "Logs"
	if l.ServiceName != "" {
		title = "Logs: " + l.ServiceName
	}

	if len(l.Lines) == 0 {
		return RenderPanel(title, LabelStyle.Render("  no logs"), l.Width)
	}

	// Show lines that fit in the available height
	maxLines := l.Height - 3
	if maxLines < 1 {
		maxLines = 1
	}

	start := len(l.Lines) - maxLines - l.ScrollPos
	if start < 0 {
		start = 0
	}
	end := start + maxLines
	if end > len(l.Lines) {
		end = len(l.Lines)
	}

	visible := l.Lines[start:end]

	// Truncate long lines
	maxWidth := l.Width - 4
	if maxWidth < 10 {
		maxWidth = 10
	}
	var trimmed []string
	for _, line := range visible {
		if len(line) > maxWidth {
			line = line[:maxWidth]
		}
		trimmed = append(trimmed, line)
	}

	content := strings.Join(trimmed, "\n")
	logStyle := lipgloss.NewStyle().Foreground(LabelColor)

	return RenderPanel(title, logStyle.Render(content), l.Width)
}
