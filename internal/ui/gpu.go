package ui

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/benaskins/panopticon/internal/hw"
	"github.com/charmbracelet/lipgloss"
)

// GPUPanel renders a GPU group grid with workload distribution.
type GPUPanel struct {
	Width      int
	Height     int
	Topology   hw.Topology
	Data       hw.GPUSample
	ActiveProcs []hw.GPUClientActivity // delta-based, from collector
}

func NewGPUPanel() GPUPanel {
	return GPUPanel{}
}

func (g GPUPanel) View() string {
	topo := g.Topology
	if topo.GPUCores == 0 {
		return RenderPanel("GPU", LabelStyle.Render("  no data"), g.Width)
	}

	gpuGroups := topo.GPUGroups
	coresPerGroup := topo.GPUCores / gpuGroups
	if coresPerGroup < 1 {
		coresPerGroup = 1
	}

	util := g.Data.Util

	// How many groups active for each workload
	tilerGroups := int(float64(gpuGroups) * util.Tiler / 100.0 + 0.5)
	rendererGroups := int(float64(gpuGroups) * util.Renderer / 100.0 + 0.5)
	computeGroups := int(float64(gpuGroups) * util.Compute / 100.0 + 0.5)
	activeGroups := tilerGroups + rendererGroups + computeGroups
	if activeGroups > gpuGroups {
		scale := float64(gpuGroups) / float64(activeGroups)
		tilerGroups = int(float64(tilerGroups)*scale + 0.5)
		rendererGroups = int(float64(rendererGroups)*scale + 0.5)
		computeGroups = gpuGroups - tilerGroups - rendererGroups
	}
	idleGroups := gpuGroups - tilerGroups - rendererGroups - computeGroups

	// Build group assignments with stable seed
	slots := make([]string, 0, gpuGroups)
	for i := 0; i < computeGroups; i++ {
		slots = append(slots, "compute")
	}
	for i := 0; i < tilerGroups; i++ {
		slots = append(slots, "tiler")
	}
	for i := 0; i < rendererGroups; i++ {
		slots = append(slots, "renderer")
	}
	for i := 0; i < idleGroups; i++ {
		slots = append(slots, "idle")
	}

	seed := int64(util.Tiler*100) + int64(util.Renderer*10000) + int64(util.Compute*1000000)
	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(slots), func(i, j int) { slots[i], slots[j] = slots[j], slots[i] })

	colorMap := map[string]lipgloss.Color{
		"compute":  GPUCompute,
		"tiler":    GPUTiler,
		"renderer": GPURenderer,
		"idle":     GPUIdle,
	}

	half := gpuGroups / 2
	var rows []string

	for gi := 0; gi < len(slots); gi++ {
		slot := slots[gi]
		color := colorMap[slot]
		block := "██"
		if slot == "idle" {
			block = "░░"
		}
		style := lipgloss.NewStyle().Foreground(color)

		var line strings.Builder
		for i := 0; i < coresPerGroup; i++ {
			if i > 0 {
				line.WriteString(" ")
			}
			line.WriteString(style.Render(block))
		}
		line.WriteString(LabelStyle.Render(fmt.Sprintf("  G%d", gi)))

		rows = append(rows, line.String())

		// Die gap
		if gi == half-1 && half > 0 {
			rows = append(rows, "")
		}
	}

	grid := strings.Join(rows, "\n")

	// Legend
	legend := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Foreground(GPUCompute).Render("██"),
		LabelStyle.Render(" COMPUTE  "),
		lipgloss.NewStyle().Foreground(GPUTiler).Render("██"),
		LabelStyle.Render(" TILER  "),
		lipgloss.NewStyle().Foreground(GPURenderer).Render("██"),
		LabelStyle.Render(" RENDER  "),
		lipgloss.NewStyle().Foreground(GPUIdle).Render("░░"),
		LabelStyle.Render(" IDLE"),
	)

	// Left side: grid + legend
	left := lipgloss.JoinVertical(lipgloss.Left, grid, "", legend)

	// Right side: top GPU clients
	clientList := g.renderClients()

	var content string
	if clientList != "" {
		content = lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", clientList)
	} else {
		content = left
	}

	return RenderPanel("GPU", content, g.Width)
}

func (g GPUPanel) renderClients() string {
	if len(g.ActiveProcs) == 0 {
		return ""
	}

	// Already sorted by rate descending from collector; show top 5
	n := 5
	if len(g.ActiveProcs) < n {
		n = len(g.ActiveProcs)
	}

	procStyle := lipgloss.NewStyle().Foreground(LabelColor)
	rateStyle := lipgloss.NewStyle().Foreground(ValueColor)

	var lines []string
	for i := 0; i < n; i++ {
		c := g.ActiveProcs[i]
		name := c.Name
		if len(name) > 20 {
			name = name[:20]
		}
		// Rate as percentage of one GPU core-equivalent
		pct := c.Rate * 100
		lines = append(lines, fmt.Sprintf(" %s %s",
			procStyle.Render(fmt.Sprintf("%-20s", name)),
			rateStyle.Render(fmt.Sprintf("%5.0f%%", pct)),
		))
	}

	return strings.Join(lines, "\n")
}
