//go:build darwin

package hw

/*
#include <mach/mach.h>
#include <mach/host_info.h>
#include <mach/mach_host.h>
#include <stdlib.h>

typedef struct {
    int count;
    int *user_ticks;
    int *system_ticks;
    int *idle_ticks;
    int *nice_ticks;
} CPUSample;

// Fetch per-CPU tick counts via host_processor_info.
// Caller must free the tick arrays.
static CPUSample get_cpu_ticks() {
    CPUSample s = {0};
    natural_t cpu_count = 0;
    processor_info_array_t info;
    mach_msg_type_number_t info_count;

    kern_return_t kr = host_processor_info(
        mach_host_self(),
        PROCESSOR_CPU_LOAD_INFO,
        &cpu_count,
        &info,
        &info_count
    );
    if (kr != KERN_SUCCESS) {
        return s;
    }

    s.count = (int)cpu_count;
    s.user_ticks = (int *)malloc(cpu_count * sizeof(int));
    s.system_ticks = (int *)malloc(cpu_count * sizeof(int));
    s.idle_ticks = (int *)malloc(cpu_count * sizeof(int));
    s.nice_ticks = (int *)malloc(cpu_count * sizeof(int));

    for (unsigned int i = 0; i < cpu_count; i++) {
        int base = i * CPU_STATE_MAX;
        s.user_ticks[i] = info[base + CPU_STATE_USER];
        s.system_ticks[i] = info[base + CPU_STATE_SYSTEM];
        s.idle_ticks[i] = info[base + CPU_STATE_IDLE];
        s.nice_ticks[i] = info[base + CPU_STATE_NICE];
    }

    vm_deallocate(mach_task_self(), (vm_address_t)info, info_count * sizeof(int));
    return s;
}
*/
import "C"

import "unsafe"

// CPUTicks holds raw per-CPU tick counts from a single sample.
type CPUTicks struct {
	User   []int
	System []int
	Idle   []int
	Nice   []int
}

// SampleCPU reads current per-CPU tick counts via Mach API.
func SampleCPU() CPUTicks {
	s := C.get_cpu_ticks()
	if s.count == 0 {
		return CPUTicks{}
	}
	defer C.free(unsafe.Pointer(s.user_ticks))
	defer C.free(unsafe.Pointer(s.system_ticks))
	defer C.free(unsafe.Pointer(s.idle_ticks))
	defer C.free(unsafe.Pointer(s.nice_ticks))

	n := int(s.count)
	t := CPUTicks{
		User:   make([]int, n),
		System: make([]int, n),
		Idle:   make([]int, n),
		Nice:   make([]int, n),
	}

	userSlice := unsafe.Slice(s.user_ticks, n)
	sysSlice := unsafe.Slice(s.system_ticks, n)
	idleSlice := unsafe.Slice(s.idle_ticks, n)
	niceSlice := unsafe.Slice(s.nice_ticks, n)

	for i := 0; i < n; i++ {
		t.User[i] = int(userSlice[i])
		t.System[i] = int(sysSlice[i])
		t.Idle[i] = int(idleSlice[i])
		t.Nice[i] = int(niceSlice[i])
	}
	return t
}

// CPUUsage computes per-core usage percentages from two tick samples.
// Returns a float64 slice with values in [0, 100].
func CPUUsage(prev, curr CPUTicks) []float64 {
	n := len(curr.User)
	if len(prev.User) != n {
		return nil
	}
	usage := make([]float64, n)
	for i := 0; i < n; i++ {
		dUser := curr.User[i] - prev.User[i]
		dSys := curr.System[i] - prev.System[i]
		dNice := curr.Nice[i] - prev.Nice[i]
		dIdle := curr.Idle[i] - prev.Idle[i]
		total := dUser + dSys + dNice + dIdle
		if total > 0 {
			usage[i] = float64(dUser+dSys+dNice) / float64(total) * 100.0
		}
	}
	return usage
}
