//go:build darwin

package hw

/*
#include <mach/mach.h>
#include <mach/host_info.h>
#include <mach/mach_host.h>
#include <sys/sysctl.h>

typedef struct {
    unsigned long long active_bytes;
    unsigned long long wired_bytes;
    unsigned long long compressed_bytes;
    unsigned long long total_bytes;
} MemInfo;

static MemInfo get_mem_info() {
    MemInfo info = {0};

    // Total physical memory
    unsigned long long total = 0;
    size_t len = sizeof(total);
    sysctlbyname("hw.memsize", &total, &len, NULL, 0);
    info.total_bytes = total;

    // VM statistics
    vm_statistics64_data_t vm_stat;
    mach_msg_type_number_t count = HOST_VM_INFO64_COUNT;
    kern_return_t kr = host_statistics64(
        mach_host_self(),
        HOST_VM_INFO64,
        (host_info64_t)&vm_stat,
        &count
    );
    if (kr != KERN_SUCCESS) {
        return info;
    }

    vm_size_t page_size;
    host_page_size(mach_host_self(), &page_size);

    info.active_bytes = (unsigned long long)vm_stat.active_count * page_size;
    info.wired_bytes = (unsigned long long)vm_stat.wire_count * page_size;
    info.compressed_bytes = (unsigned long long)vm_stat.compressor_page_count * page_size;

    return info;
}
*/
import "C"

// MemoryInfo holds current memory usage.
type MemoryInfo struct {
	ActiveBytes     uint64
	WiredBytes      uint64
	CompressedBytes uint64
	TotalBytes      uint64
}

// UsedBytes returns active + wired memory.
func (m MemoryInfo) UsedBytes() uint64 {
	return m.ActiveBytes + m.WiredBytes
}

// UsagePercent returns the used memory percentage.
func (m MemoryInfo) UsagePercent() float64 {
	if m.TotalBytes == 0 {
		return 0
	}
	return float64(m.UsedBytes()) / float64(m.TotalBytes) * 100
}

// SampleMemory reads current memory usage via Mach API.
func SampleMemory() MemoryInfo {
	info := C.get_mem_info()
	return MemoryInfo{
		ActiveBytes:     uint64(info.active_bytes),
		WiredBytes:      uint64(info.wired_bytes),
		CompressedBytes: uint64(info.compressed_bytes),
		TotalBytes:      uint64(info.total_bytes),
	}
}
