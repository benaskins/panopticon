//go:build darwin

package hw

/*
#include <sys/sysctl.h>
#include <stdlib.h>
#include <string.h>

static int sysctl_int(const char *name) {
    int val = 0;
    size_t len = sizeof(val);
    if (sysctlbyname(name, &val, &len, NULL, 0) != 0) {
        return 0;
    }
    return val;
}

static unsigned long long sysctl_ull(const char *name) {
    unsigned long long val = 0;
    size_t len = sizeof(val);
    if (sysctlbyname(name, &val, &len, NULL, 0) != 0) {
        return 0;
    }
    return val;
}
*/
import "C"

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var gpuCoreRe = regexp.MustCompile(`Total Number of Cores:\s*(\d+)`)
var mgpuRe = regexp.MustCompile(`"num_mgpus"\s*=\s*(\d+)`)

func detectTopology() Topology {
	pCores := int(C.sysctl_int(C.CString("hw.perflevel0.physicalcpu")))
	eCores := int(C.sysctl_int(C.CString("hw.perflevel1.physicalcpu")))
	pPerCluster := int(C.sysctl_int(C.CString("hw.perflevel0.cpusperl2")))
	ePerCluster := int(C.sysctl_int(C.CString("hw.perflevel1.cpusperl2")))
	totalMem := uint64(C.sysctl_ull(C.CString("hw.memsize")))

	pClusters := 0
	if pPerCluster > 0 {
		pClusters = pCores / pPerCluster
	}
	eClusters := 0
	if ePerCluster > 0 {
		eClusters = eCores / ePerCluster
	}

	gpuCores := detectGPUCores()
	gpuGroups := detectGPUGroups()

	return Topology{
		PCores:      pCores,
		ECores:      eCores,
		PPerCluster: pPerCluster,
		EPerCluster: ePerCluster,
		PClusters:   pClusters,
		EClusters:   eClusters,
		GPUCores:    gpuCores,
		GPUGroups:   gpuGroups,
		TotalMemory: totalMem,
	}
}

func detectGPUCores() int {
	out, err := exec.Command("system_profiler", "SPDisplaysDataType").Output()
	if err != nil {
		return 0
	}
	m := gpuCoreRe.FindSubmatch(out)
	if m == nil {
		return 0
	}
	n, _ := strconv.Atoi(string(m[1]))
	return n
}

func detectGPUGroups() int {
	out, err := exec.Command("ioreg", "-r", "-c", "AGXAcceleratorG15X", "-d", "1").Output()
	if err != nil {
		return 1
	}
	m := mgpuRe.FindStringSubmatch(strings.TrimSpace(string(out)))
	if m == nil {
		return 1
	}
	n, _ := strconv.Atoi(m[1])
	if n < 1 {
		return 1
	}
	return n
}
