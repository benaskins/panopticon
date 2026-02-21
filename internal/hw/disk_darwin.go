//go:build darwin

package hw

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation
#include <stdlib.h>

typedef struct {
    unsigned long long read_bytes;
    unsigned long long write_bytes;
} DiskStats;

extern DiskStats pollDiskStats();
*/
import "C"

// DiskIO holds cumulative disk read/write byte counts.
type DiskIO struct {
	ReadBytes  uint64
	WriteBytes uint64
}

// SampleDisk reads cumulative disk I/O bytes from IOKit.
func SampleDisk() DiskIO {
	stats := C.pollDiskStats()
	return DiskIO{
		ReadBytes:  uint64(stats.read_bytes),
		WriteBytes: uint64(stats.write_bytes),
	}
}

// DiskRate computes MB/s read and write rates between two samples taken dt seconds apart.
func DiskRate(prev, curr DiskIO, dtSeconds float64) (readMBs, writeMBs float64) {
	if dtSeconds <= 0 {
		return 0, 0
	}
	dr := float64(curr.ReadBytes-prev.ReadBytes) / (1024 * 1024) / dtSeconds
	dw := float64(curr.WriteBytes-prev.WriteBytes) / (1024 * 1024) / dtSeconds
	if dr < 0 {
		dr = 0
	}
	if dw < 0 {
		dw = 0
	}
	return dr, dw
}
