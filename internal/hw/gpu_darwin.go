//go:build darwin

package hw

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation -framework Foundation
#include <stdlib.h>

typedef struct {
    int tiler;      // 0-100
    int renderer;   // 0-100
    int device;     // 0-100
} GPUUtilization;

typedef struct {
    int pid;
    const char *name;
    unsigned long long gpu_time_ns;
} GPUClient;

typedef struct {
    GPUUtilization util;
    GPUClient *clients;
    int client_count;
} GPUInfo;

extern GPUInfo pollGPUInfo();
extern void freeGPUInfo(GPUInfo info);
*/
import "C"

import "unsafe"

// GPUUtil holds GPU utilization percentages.
type GPUUtil struct {
	Tiler    float64 // tiler pipeline utilization %
	Renderer float64 // renderer pipeline utilization %
	Compute  float64 // compute (device - max(tiler, renderer)) %
	Device   float64 // overall device utilization %
}

// GPUClientInfo describes a process using the GPU.
type GPUClientInfo struct {
	PID       int
	Name      string
	GPUTimeNS uint64
}

// GPUSample holds utilization and per-client info.
type GPUSample struct {
	Util    GPUUtil
	Clients []GPUClientInfo
}

// SampleGPU reads GPU utilization and client info from IOKit.
func SampleGPU() GPUSample {
	info := C.pollGPUInfo()
	defer C.freeGPUInfo(info)

	tiler := float64(info.util.tiler)
	renderer := float64(info.util.renderer)
	device := float64(info.util.device)
	compute := device - max(tiler, renderer)
	if compute < 0 {
		compute = 0
	}

	sample := GPUSample{
		Util: GPUUtil{
			Tiler:    tiler,
			Renderer: renderer,
			Compute:  compute,
			Device:   device,
		},
	}

	if info.client_count > 0 && info.clients != nil {
		clients := unsafe.Slice(info.clients, int(info.client_count))
		for _, c := range clients {
			name := ""
			if c.name != nil {
				name = C.GoString(c.name)
			}
			sample.Clients = append(sample.Clients, GPUClientInfo{
				PID:       int(c.pid),
				Name:      name,
				GPUTimeNS: uint64(c.gpu_time_ns),
			})
		}
	}

	return sample
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
