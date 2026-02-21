//go:build !darwin

package hw

type GPUUtil struct {
	Tiler    float64
	Renderer float64
	Compute  float64
	Device   float64
}

type GPUClientInfo struct {
	PID       int
	Name      string
	GPUTimeNS uint64
}

type GPUSample struct {
	Util    GPUUtil
	Clients []GPUClientInfo
}

func SampleGPU() GPUSample { return GPUSample{} }
