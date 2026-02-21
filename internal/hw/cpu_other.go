//go:build !darwin

package hw

// CPUTicks holds raw per-CPU tick counts from a single sample.
type CPUTicks struct {
	User   []int
	System []int
	Idle   []int
	Nice   []int
}

func SampleCPU() CPUTicks { return CPUTicks{} }

func CPUUsage(prev, curr CPUTicks) []float64 { return nil }
