package ui

import (
	"fmt"

	"github.com/benaskins/panopticon/internal/aurelia"
	"github.com/charmbracelet/lipgloss"
)

// ServicesPanel shows Aurelia service states.
type ServicesPanel struct {
	Width    int
	Height   int
	Services []aurelia.ServiceState
	Selected int
	Focused  bool
}

func NewServicesPanel() ServicesPanel {
	return ServicesPanel{}
}

func (s ServicesPanel) View() string {
	if len(s.Services) == 0 {
		return RenderPanel("Services", LabelStyle.Render("  no services"), s.Width)
	}

	detailStyle := lipgloss.NewStyle().Foreground(LabelColor)
	selectedBg := lipgloss.NewStyle().Background(lipgloss.Color("#1a3a2a"))

	var rows []string
	for i, svc := range s.Services {
		indicator, color := stateIndicator(svc.State, svc.Health)
		nameStyle := lipgloss.NewStyle().Foreground(color)

		// Format: <icon> <name><:port> <type>/<pid>
		name := svc.Name
		portStr := ""
		if svc.Port > 0 {
			portStr = fmt.Sprintf(":%d", svc.Port)
		}

		meta := ""
		if svc.Type != "" || svc.PID > 0 {
			t := svc.Type
			if t == "" {
				t = "-"
			}
			if svc.PID > 0 {
				meta = fmt.Sprintf(" %s/%d", t, svc.PID)
			} else {
				meta = " " + t
			}
		}

		maxLen := s.Width - 6
		full := name + portStr + meta
		if maxLen > 0 && len(full) > maxLen {
			trim := maxLen - len(portStr) - len(meta)
			if trim > 0 && trim < len(name) {
				name = name[:trim]
			}
		}

		indicatorStyle := lipgloss.NewStyle().Foreground(color)
		line := fmt.Sprintf(" %s %s%s%s",
			indicatorStyle.Render(indicator),
			nameStyle.Render(name),
			detailStyle.Render(portStr),
			detailStyle.Render(meta),
		)

		if i == s.Selected && s.Focused {
			line = selectedBg.Width(s.Width - 4).Render(line)
		}

		rows = append(rows, line)
	}

	// Legend
	legend := fmt.Sprintf(" %s %s  %s %s  %s %s  %s %s",
		lipgloss.NewStyle().Foreground(RunningColor).Render("●"), detailStyle.Render("run"),
		lipgloss.NewStyle().Foreground(StartingColor).Render("~"), detailStyle.Render("start"),
		lipgloss.NewStyle().Foreground(UnhealthyColor).Render("!"), detailStyle.Render("sick"),
		lipgloss.NewStyle().Foreground(StoppedColor).Render("○"), detailStyle.Render("stop"),
	)

	// Limit to available height (reserve 2 for legend + blank line)
	maxRows := s.Height - 5
	if maxRows < 1 {
		maxRows = 1
	}
	if len(rows) > maxRows {
		rows = rows[:maxRows]
	}

	list := ""
	for i, r := range rows {
		if i > 0 {
			list += "\n"
		}
		list += r
	}

	return RenderPanel("Services", list+"\n\n"+legend, s.Width)
}

func stateIndicator(state, health string) (string, lipgloss.Color) {
	switch {
	case state == "running" && (health == "healthy" || health == "" || health == "-"):
		return "●", RunningColor
	case state == "running" && health == "unhealthy":
		return "!", UnhealthyColor
	case state == "starting":
		return "~", StartingColor
	case state == "stopped" || state == "failed":
		return "○", StoppedColor
	default:
		return "○", StoppedColor
	}
}
