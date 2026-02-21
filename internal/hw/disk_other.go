//go:build !darwin

package hw

type DiskIO struct {
	ReadBytes  uint64
	WriteBytes uint64
}

func SampleDisk() DiskIO { return DiskIO{} }

func DiskRate(prev, curr DiskIO, dtSeconds float64) (readMBs, writeMBs float64) {
	return 0, 0
}
