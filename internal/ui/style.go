package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Color palettes from core-monitor
var (
	// P-core colors: warm amber
	PIdle = lipgloss.Color("#461E12") // rgb(70,30,18)
	PFull = lipgloss.Color("#B45028") // rgb(180,80,40)

	// E-core colors: teal
	EIdle = lipgloss.Color("#193732") // rgb(25,55,50)
	EFull = lipgloss.Color("#37918C") // rgb(55,145,140)

	// GPU workload colors
	GPUIdle     = lipgloss.Color("#281641") // rgb(40,22,65)
	GPUTiler    = lipgloss.Color("#37918C") // rgb(55,145,140)
	GPURenderer = lipgloss.Color("#7841AF") // rgb(120,65,175)
	GPUCompute  = lipgloss.Color("#00823C") // rgb(0,130,60)

	// Border and title
	BorderColor = lipgloss.Color("#005028") // rgb(0,80,40)
	TitleColor  = lipgloss.Color("#00B450") // rgb(0,180,80)

	// Service state indicators
	RunningColor  = lipgloss.Color("#00B450")
	StoppedColor  = lipgloss.Color("#555555")
	UnhealthyColor = lipgloss.Color("#CC3322")
	StartingColor  = lipgloss.Color("#CCAA22")

	// Status bar
	LabelColor = lipgloss.Color("#505050") // rgb(80,80,80)
	ValueColor = lipgloss.Color("#007837") // rgb(0,120,55)

	// Thermal state colors
	ThermalNominal  = lipgloss.Color("#007837")
	ThermalFair     = lipgloss.Color("#B48228")
	ThermalSerious  = lipgloss.Color("#B43228")
	ThermalCritical = lipgloss.Color("#B43228")
)

// Shared styles
var (
	PanelBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor)

	PanelTitle = lipgloss.NewStyle().
			Foreground(TitleColor).
			Bold(true)

	LabelStyle = lipgloss.NewStyle().Foreground(LabelColor)
	ValueStyle = lipgloss.NewStyle().Foreground(ValueColor)
)

// RenderPanel draws a rounded border with the title inset in the top edge.
// Produces: ╭── Title ──────────────╮
func RenderPanel(title string, content string, width int) string {
	border := lipgloss.RoundedBorder()
	borderStyle := lipgloss.NewStyle().Foreground(BorderColor)
	titleStyle := lipgloss.NewStyle().Foreground(TitleColor).Bold(true)

	innerWidth := width - 2 // subtract left + right border chars
	if innerWidth < 4 {
		innerWidth = 4
	}

	// Top border: ╭── Title ───────╮
	titleStr := titleStyle.Render(title)
	titleLen := len(title) // raw length for padding calc
	dashLeft := borderStyle.Render("──")
	dashLeftLen := 2

	rightDashes := innerWidth - dashLeftLen - 1 - titleLen - 1 // 1 space each side of title
	if rightDashes < 1 {
		rightDashes = 1
	}
	dashRight := borderStyle.Render(repeat("─", rightDashes))

	topLine := borderStyle.Render(border.TopLeft) +
		dashLeft +
		borderStyle.Render(" ") + titleStr + borderStyle.Render(" ") +
		dashRight +
		borderStyle.Render(border.TopRight)

	// Content with side borders, padded to innerWidth
	contentStyle := lipgloss.NewStyle().Width(innerWidth)
	paddedContent := contentStyle.Render(content)

	var lines []string
	lines = append(lines, topLine)
	for _, line := range splitLines(paddedContent) {
		lines = append(lines, borderStyle.Render(border.Left)+line+borderStyle.Render(border.Right))
	}

	// Bottom border
	bottomLine := borderStyle.Render(border.BottomLeft) +
		borderStyle.Render(repeat("─", innerWidth)) +
		borderStyle.Render(border.BottomRight)
	lines = append(lines, bottomLine)

	result := ""
	for i, l := range lines {
		if i > 0 {
			result += "\n"
		}
		result += l
	}
	return result
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	lines = append(lines, s[start:])
	return lines
}

// Shade returns Unicode block characters based on activity percentage.
func Shade(activity float64) string {
	switch {
	case activity < 25:
		return "░░"
	case activity < 50:
		return "▒▒"
	case activity < 75:
		return "▓▓"
	default:
		return "██"
	}
}

// LerpColor interpolates between two RGB colors. t in [0, 1].
func LerpColor(t float64, idle, full [3]int) lipgloss.Color {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	r := idle[0] + int(float64(full[0]-idle[0])*t)
	g := idle[1] + int(float64(full[1]-idle[1])*t)
	b := idle[2] + int(float64(full[2]-idle[2])*t)
	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, g, b))
}

// RGB tuples for lerp
var (
	PIdleRGB = [3]int{70, 30, 18}
	PFullRGB = [3]int{180, 80, 40}
	EIdleRGB = [3]int{25, 55, 50}
	EFullRGB = [3]int{55, 145, 140}
)
