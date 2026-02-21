package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// StatusBar renders disk I/O, thermal state, and Aurelia connection status.
type StatusBar struct {
	Width        int
	DiskReadMBs  float64
	DiskWriteMBs float64
	Thermal      string
	AureliaUp    bool
}

func NewStatusBar() StatusBar {
	return StatusBar{}
}

func (s StatusBar) View() string {
	sep := LabelStyle.Render("  │  ")

	// Disk I/O
	disk := LabelStyle.Render(" DISK  ") +
		ValueStyle.Render(fmt.Sprintf("R: %5.1f MB/s  W: %5.1f MB/s", s.DiskReadMBs, s.DiskWriteMBs))

	// Thermal
	thermalColor := thermalStateColor(s.Thermal)
	thermal := LabelStyle.Render("THERMAL  ") +
		lipgloss.NewStyle().Foreground(thermalColor).Render(s.Thermal)

	// Aurelia connection
	aureliaStatus := LabelStyle.Render("aurelia: ")
	if s.AureliaUp {
		aureliaStatus += lipgloss.NewStyle().Foreground(RunningColor).Render("connected")
	} else {
		aureliaStatus += lipgloss.NewStyle().Foreground(StoppedColor).Render("disconnected")
	}

	bar := disk + sep + thermal + sep + aureliaStatus

	return lipgloss.NewStyle().
		Width(s.Width).
		Render(bar)
}

func thermalStateColor(state string) lipgloss.Color {
	switch state {
	case "Nominal":
		return ThermalNominal
	case "Fair":
		return ThermalFair
	case "Serious":
		return ThermalSerious
	case "Critical":
		return ThermalCritical
	default:
		return LabelColor
	}
}
