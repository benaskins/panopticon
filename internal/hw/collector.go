package hw

import "time"

// EMA smoothing factor. At 5Hz polling, 0.3 gives ~0.5s settling time.
const emaAlpha = 0.3

// GPUClientActivity represents a process's GPU usage rate over one poll interval.
type GPUClientActivity struct {
	PID       int
	Name      string
	GPUTimeNS uint64  // delta GPU time in nanoseconds since last poll
	Rate      float64 // GPU-seconds per wall-second (0.0 to ~N for N GPU cores)
}

// Snapshot holds all hardware metrics from a single poll.
type Snapshot struct {
	// CPU per-core usage percentages (P-cores first, then E-cores)
	CPUUsage []float64

	// Memory
	Memory MemoryInfo

	// GPU (with smoothed utilization)
	GPU GPUSample

	// GPU client activity (delta, sorted by rate descending)
	GPUClients []GPUClientActivity

	// Disk I/O rates (MB/s)
	DiskReadMBs  float64
	DiskWriteMBs float64

	// Thermal
	Thermal string

	// Timestamp of this snapshot
	Time time.Time
}

// Collector aggregates hardware metrics with delta computation.
type Collector struct {
	prevCPU        CPUTicks
	prevDisk       DiskIO
	prevGPUClients map[int]uint64 // pid -> cumulative GPU time ns
	prevTime       time.Time

	// EMA state for GPU utilization
	emaTiler    float64
	emaRenderer float64
	emaDevice   float64
	emaInit     bool
}

// NewCollector creates a collector and takes initial samples for delta computation.
func NewCollector() *Collector {
	gpuSample := SampleGPU()
	clientMap := make(map[int]uint64, len(gpuSample.Clients))
	for _, c := range gpuSample.Clients {
		clientMap[c.PID] = c.GPUTimeNS
	}

	return &Collector{
		prevCPU:        SampleCPU(),
		prevDisk:       SampleDisk(),
		prevGPUClients: clientMap,
		prevTime:       time.Now(),
	}
}

// Poll reads all hardware metrics and computes deltas from the previous poll.
func (c *Collector) Poll() Snapshot {
	now := time.Now()
	dt := now.Sub(c.prevTime).Seconds()

	// CPU
	cpuNow := SampleCPU()
	cpuUsage := CPUUsage(c.prevCPU, cpuNow)
	c.prevCPU = cpuNow

	// Memory
	mem := SampleMemory()

	// GPU — smooth utilization with EMA
	gpu := SampleGPU()
	gpu.Util = c.smoothGPUUtil(gpu.Util)

	// GPU client deltas
	gpuClients := c.computeGPUClientDeltas(gpu.Clients, dt)

	// Disk
	diskNow := SampleDisk()
	diskR, diskW := DiskRate(c.prevDisk, diskNow, dt)
	c.prevDisk = diskNow

	// Thermal
	thermal := ThermalState()

	c.prevTime = now

	return Snapshot{
		CPUUsage:     cpuUsage,
		Memory:       mem,
		GPU:          gpu,
		GPUClients:   gpuClients,
		DiskReadMBs:  diskR,
		DiskWriteMBs: diskW,
		Thermal:      thermal,
		Time:         now,
	}
}

func (c *Collector) smoothGPUUtil(raw GPUUtil) GPUUtil {
	if !c.emaInit {
		// Seed with first sample
		c.emaTiler = raw.Tiler
		c.emaRenderer = raw.Renderer
		c.emaDevice = raw.Device
		c.emaInit = true
	} else {
		c.emaTiler = ema(c.emaTiler, raw.Tiler)
		c.emaRenderer = ema(c.emaRenderer, raw.Renderer)
		c.emaDevice = ema(c.emaDevice, raw.Device)
	}

	compute := c.emaDevice - emaMax(c.emaTiler, c.emaRenderer)
	if compute < 0 {
		compute = 0
	}

	return GPUUtil{
		Tiler:    c.emaTiler,
		Renderer: c.emaRenderer,
		Compute:  compute,
		Device:   c.emaDevice,
	}
}

func ema(prev, cur float64) float64 {
	return prev + emaAlpha*(cur-prev)
}

func emaMax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func (c *Collector) computeGPUClientDeltas(clients []GPUClientInfo, dt float64) []GPUClientActivity {
	newMap := make(map[int]uint64, len(clients))
	nameMap := make(map[int]string, len(clients))
	for _, cl := range clients {
		newMap[cl.PID] = cl.GPUTimeNS
		nameMap[cl.PID] = cl.Name
	}

	var active []GPUClientActivity
	for pid, currTime := range newMap {
		prevTime := c.prevGPUClients[pid]
		if currTime > prevTime {
			deltaNS := currTime - prevTime
			rate := 0.0
			if dt > 0 {
				rate = float64(deltaNS) / 1e9 / dt
			}
			if rate > 0.001 { // filter out noise
				active = append(active, GPUClientActivity{
					PID:       pid,
					Name:      nameMap[pid],
					GPUTimeNS: deltaNS,
					Rate:      rate,
				})
			}
		}
	}

	// Sort by rate descending
	for i := 0; i < len(active); i++ {
		for j := i + 1; j < len(active); j++ {
			if active[j].Rate > active[i].Rate {
				active[i], active[j] = active[j], active[i]
			}
		}
	}

	c.prevGPUClients = newMap
	return active
}
