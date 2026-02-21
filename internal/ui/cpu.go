package ui

import (
	"fmt"
	"strings"

	"github.com/benaskins/panopticon/internal/hw"
	"github.com/charmbracelet/lipgloss"
)

// CPUPanel renders a clustered CPU heatmap.
type CPUPanel struct {
	Width    int
	Height   int
	Topology hw.Topology
	Usage    []float64 // per-core usage percentages
}

func NewCPUPanel() CPUPanel {
	return CPUPanel{}
}

func (p CPUPanel) View() string {
	topo := p.Topology
	if topo.PCores == 0 && topo.ECores == 0 {
		return RenderPanel("CPU", LabelStyle.Render("  no data"), p.Width)
	}

	// Split usage into P and E cores
	// macOS reports P-cores first, then E-cores
	pUsage := make([]float64, topo.PCores)
	eUsage := make([]float64, topo.ECores)
	for i := 0; i < topo.PCores && i < len(p.Usage); i++ {
		pUsage[i] = p.Usage[i]
	}
	for i := 0; i < topo.ECores && topo.PCores+i < len(p.Usage); i++ {
		eUsage[i] = p.Usage[topo.PCores+i]
	}

	maxRows := topo.PClusters
	if topo.EClusters > maxRows {
		maxRows = topo.EClusters
	}

	var rows []string
	for row := 0; row < maxRows; row++ {
		var line strings.Builder

		// P-cluster column
		if row < topo.PClusters {
			for i := 0; i < topo.PPerCluster; i++ {
				coreIdx := row*topo.PPerCluster + i
				activity := 0.0
				if coreIdx < len(pUsage) {
					activity = pUsage[coreIdx]
				}
				shade := Shade(activity)
				color := LerpColor(activity/100.0, PIdleRGB, PFullRGB)
				if i > 0 {
					line.WriteString(" ")
				}
				line.WriteString(lipgloss.NewStyle().Foreground(color).Render(shade))
			}
			line.WriteString(LabelStyle.Render(fmt.Sprintf("  P%d", row)))
		} else {
			// Empty space for alignment
			w := topo.PPerCluster*2 + (topo.PPerCluster - 1) + 4
			line.WriteString(strings.Repeat(" ", w))
		}

		line.WriteString("     ")

		// E-cluster column
		if row < topo.EClusters {
			for i := 0; i < topo.EPerCluster; i++ {
				coreIdx := row*topo.EPerCluster + i
				activity := 0.0
				if coreIdx < len(eUsage) {
					activity = eUsage[coreIdx]
				}
				shade := Shade(activity)
				color := LerpColor(activity/100.0, EIdleRGB, EFullRGB)
				if i > 0 {
					line.WriteString(" ")
				}
				line.WriteString(lipgloss.NewStyle().Foreground(color).Render(shade))
			}
			line.WriteString(LabelStyle.Render(fmt.Sprintf("  E%d", row)))
		}

		rows = append(rows, line.String())
	}

	grid := strings.Join(rows, "\n")

	return RenderPanel("CPU", grid, p.Width)
}
