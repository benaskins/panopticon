package ui

import (
	"sort"
	"time"

	"github.com/benaskins/panopticon/internal/aurelia"
	"github.com/benaskins/panopticon/internal/hw"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Tick message types
type hwTickMsg time.Time
type aureliaTickMsg time.Time

// HW tick: 200ms (5Hz)
func hwTickCmd() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return hwTickMsg(t)
	})
}

// Aurelia tick: 1s
func aureliaTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return aureliaTickMsg(t)
	})
}

// Focus targets
type focusTarget int

const (
	focusHardware focusTarget = iota
	focusServices
)

// Model is the root bubbletea model.
type Model struct {
	// Dimensions
	width  int
	height int

	// Hardware
	topology  hw.Topology
	collector *hw.Collector
	snapshot  hw.Snapshot

	// Aurelia
	aureliaClient *aurelia.Client
	aureliaUp     bool
	services      []aurelia.ServiceState
	logLines      []string
	selectedSvc   int
	showLogs      bool

	// Focus
	focus focusTarget

	// Panels
	memoryPanel   MemoryPanel
	cpuPanel      CPUPanel
	gpuPanel      GPUPanel
	servicesPanel ServicesPanel
	logsPanel     LogsPanel
	statusBar     StatusBar

	// Help overlay
	showHelp bool
}

// NewModel creates the root model.
func NewModel() Model {
	topo := hw.DetectTopology()
	collector := hw.NewCollector()
	// Take an initial sample immediately so the first frame has data
	snap := collector.Poll()

	client := aurelia.NewClient()

	return Model{
		topology:      topo,
		collector:     collector,
		snapshot:      snap,
		aureliaClient: client,
		aureliaUp:     client.Available(),
		memoryPanel:   NewMemoryPanel(),
		cpuPanel:      NewCPUPanel(),
		gpuPanel:      NewGPUPanel(),
		servicesPanel: NewServicesPanel(),
		logsPanel:     NewLogsPanel(),
		statusBar:     NewStatusBar(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(hwTickCmd(), aureliaTickCmd())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layoutPanels()
		return m, nil

	case hwTickMsg:
		m.snapshot = m.collector.Poll()
		m.updateHWPanels()
		return m, hwTickCmd()

	case aureliaTickMsg:
		m.pollAurelia()
		m.layoutPanels()
		return m, aureliaTickCmd()
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "?":
		m.showHelp = !m.showHelp

	case "tab":
		if m.aureliaUp && len(m.services) > 0 {
			if m.focus == focusHardware {
				m.focus = focusServices
			} else {
				m.focus = focusHardware
			}
			m.servicesPanel.Focused = m.focus == focusServices
		}

	case "j", "down":
		if m.focus == focusServices && len(m.services) > 0 {
			m.selectedSvc++
			if m.selectedSvc >= len(m.services) {
				m.selectedSvc = len(m.services) - 1
			}
			m.servicesPanel.Selected = m.selectedSvc
			if m.showLogs {
				m.fetchLogs()
			}
		}

	case "k", "up":
		if m.focus == focusServices {
			m.selectedSvc--
			if m.selectedSvc < 0 {
				m.selectedSvc = 0
			}
			m.servicesPanel.Selected = m.selectedSvc
			if m.showLogs {
				m.fetchLogs()
			}
		}

	case "enter":
		if m.focus == focusServices && len(m.services) > 0 {
			m.showLogs = !m.showLogs
			m.layoutPanels()
			if m.showLogs {
				m.fetchLogs()
			} else {
				m.logLines = nil
				m.logsPanel.Lines = nil
				m.logsPanel.ServiceName = ""
			}
		}

	case "s":
		if m.focus == focusServices && len(m.services) > 0 {
			name := m.services[m.selectedSvc].Name
			_ = m.aureliaClient.ServiceAction(name, "start")
		}

	case "x":
		if m.focus == focusServices && len(m.services) > 0 {
			name := m.services[m.selectedSvc].Name
			_ = m.aureliaClient.ServiceAction(name, "stop")
		}

	case "r":
		if m.focus == focusServices && len(m.services) > 0 {
			name := m.services[m.selectedSvc].Name
			_ = m.aureliaClient.ServiceAction(name, "restart")
		}
	}

	return m, nil
}

func (m *Model) pollAurelia() {
	m.aureliaUp = m.aureliaClient.Available()
	if !m.aureliaUp {
		m.services = nil
		m.servicesPanel.Services = nil
		return
	}

	svcs, err := m.aureliaClient.Services()
	if err != nil {
		m.aureliaUp = false
		m.services = nil
		m.servicesPanel.Services = nil
		return
	}
	sort.Slice(svcs, func(i, j int) bool {
		return svcs[i].Name < svcs[j].Name
	})
	m.services = svcs
	m.servicesPanel.Services = svcs

	if m.showLogs && len(m.services) > 0 {
		m.fetchLogs()
	}
}

func (m *Model) fetchLogs() {
	if m.selectedSvc >= len(m.services) {
		return
	}
	name := m.services[m.selectedSvc].Name
	lines, err := m.aureliaClient.Logs(name, 50)
	if err != nil {
		return
	}
	m.logLines = lines
	m.logsPanel.Lines = lines
	m.logsPanel.ServiceName = name
}

func (m *Model) updateHWPanels() {
	m.memoryPanel.Data = m.snapshot.Memory
	m.memoryPanel.TopProcs = m.snapshot.MemoryProcs
	m.cpuPanel.Usage = m.snapshot.CPUUsage
	m.cpuPanel.Topology = m.topology
	m.gpuPanel.Data = m.snapshot.GPU
	m.gpuPanel.ActiveProcs = m.snapshot.GPUClients
	m.gpuPanel.Topology = m.topology
	m.statusBar.DiskReadMBs = m.snapshot.DiskReadMBs
	m.statusBar.DiskWriteMBs = m.snapshot.DiskWriteMBs
	m.statusBar.Thermal = m.snapshot.Thermal
	m.statusBar.AureliaUp = m.aureliaUp
}

// clamp returns v if v >= min, otherwise min.
func clamp(v, min int) int {
	if v < min {
		return min
	}
	return v
}

func (m *Model) layoutPanels() {
	// Left column: memory (fixed width 20)
	memWidth := 28
	m.memoryPanel.Width = memWidth
	m.memoryPanel.Height = clamp(m.height-2, 4)

	// Status bar at bottom: 1 row
	m.statusBar.Width = m.width

	// Remaining width for middle + right columns
	remaining := clamp(m.width-memWidth-2, 10) // 2 for gap

	showServices := m.aureliaUp && len(m.services) > 0

	// CPU: fixed height based on cluster count
	cpuHeight := m.topology.PClusters
	if m.topology.EClusters > cpuHeight {
		cpuHeight = m.topology.EClusters
	}
	cpuHeight += 3 // borders + title

	if showServices {
		// Right column: ~35% of remaining, min 25
		rightWidth := remaining * 35 / 100
		if rightWidth < 25 {
			rightWidth = 25
		}
		middleWidth := clamp(remaining-rightWidth-1, 10)

		m.cpuPanel.Width = middleWidth
		m.cpuPanel.Height = cpuHeight
		m.gpuPanel.Width = middleWidth
		m.gpuPanel.Height = clamp(m.height-cpuHeight-3, 4)

		fullHeight := clamp(m.height-2, 4)
		if m.showLogs {
			// Split: services top half, logs bottom half
			svcHeight := clamp(fullHeight/2, 4)
			m.servicesPanel.Width = rightWidth
			m.servicesPanel.Height = svcHeight
			m.logsPanel.Width = rightWidth
			m.logsPanel.Height = clamp(fullHeight-svcHeight-1, 4)
		} else {
			// Services take full height
			m.servicesPanel.Width = rightWidth
			m.servicesPanel.Height = fullHeight
			m.logsPanel.Width = rightWidth
			m.logsPanel.Height = 0
		}
	} else {
		// No services — middle takes all remaining
		m.cpuPanel.Width = remaining
		m.cpuPanel.Height = cpuHeight
		m.gpuPanel.Width = remaining
		m.gpuPanel.Height = clamp(m.height-cpuHeight-3, 4)
	}
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	if m.showHelp {
		return m.helpView()
	}

	// Left column: memory
	left := m.memoryPanel.View()

	// Middle column: CPU over GPU
	middle := lipgloss.JoinVertical(lipgloss.Left,
		m.cpuPanel.View(),
		m.gpuPanel.View(),
	)

	showServices := m.aureliaUp && len(m.services) > 0

	var body string
	if showServices {
		// Right column: services, with optional logs below
		var right string
		if m.showLogs {
			right = lipgloss.JoinVertical(lipgloss.Left,
				m.servicesPanel.View(),
				m.logsPanel.View(),
			)
		} else {
			right = m.servicesPanel.View()
		}
		body = lipgloss.JoinHorizontal(lipgloss.Top, left, " ", middle, " ", right)
	} else {
		body = lipgloss.JoinHorizontal(lipgloss.Top, left, " ", middle)
	}

	// Status bar at bottom
	return lipgloss.JoinVertical(lipgloss.Left,
		body,
		m.statusBar.View(),
	)
}

func (m Model) helpView() string {
	help := `
  Panopticon — Apple Silicon Cockpit

  tab     toggle focus (hardware / services)
  j/k     select service
  enter   toggle log tail
  s       start service
  x       stop service
  r       restart service
  ?       toggle help
  q       quit
`
	style := PanelBorder.
		Width(50).
		Padding(1, 2).
		BorderForeground(TitleColor)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		style.Render(PanelTitle.Render(" HELP ")+help),
	)
}
