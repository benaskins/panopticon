//go:build !darwin

package hw

// MemoryInfo holds current memory usage.
type MemoryInfo struct {
	ActiveBytes     uint64
	WiredBytes      uint64
	CompressedBytes uint64
	TotalBytes      uint64
}

func (m MemoryInfo) UsedBytes() uint64    { return m.ActiveBytes + m.WiredBytes }
func (m MemoryInfo) UsagePercent() float64 { return 0 }

func SampleMemory() MemoryInfo { return MemoryInfo{} }
