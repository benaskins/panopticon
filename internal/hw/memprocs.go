package hw

import (
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// MemProc describes a process's resident memory usage.
type MemProc struct {
	PID  int
	Name string
	RSS  uint64 // resident set size in bytes
}

// SampleMemProcs returns the top memory-consuming processes via ps.
func SampleMemProcs(topN int) []MemProc {
	// ps -axo pid=,rss=,comm= — rss is in kilobytes
	out, err := exec.Command("ps", "-axo", "pid=,rss=,comm=").Output()
	if err != nil {
		return nil
	}

	var procs []MemProc
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		rssKB, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		// comm may contain spaces; rejoin remaining fields
		name := strings.Join(fields[2:], " ")
		// Use just the binary name (last path component)
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			name = name[idx+1:]
		}

		procs = append(procs, MemProc{
			PID:  pid,
			Name: name,
			RSS:  rssKB * 1024,
		})
	}

	sort.Slice(procs, func(i, j int) bool {
		return procs[i].RSS > procs[j].RSS
	})

	if len(procs) > topN {
		procs = procs[:topN]
	}
	return procs
}
