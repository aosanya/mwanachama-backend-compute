package concurrency

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type machine struct {
	Processor  string `json:"processor"`
	CPUs       int    `json:"cpus"`
	System     string `json:"system"`
	GoVersion  string `json:"go_version"`
	GOMAXPROCS int    `json:"gomaxprocs"`
}

func (m machine) key() string {
	return fmt.Sprintf("%s | %d cpus | %s | %s", m.Processor, m.CPUs, m.System, m.GoVersion)
}

func currentMachine() machine {
	return machine{
		Processor:  processorName(),
		CPUs:       runtime.NumCPU(),
		System:     runtime.GOOS + "/" + runtime.GOARCH,
		GoVersion:  runtime.Version(),
		GOMAXPROCS: runtime.GOMAXPROCS(0),
	}
}

func processorName() string {
	if out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
		if name := strings.TrimSpace(string(out)); name != "" {
			return name
		}
	}
	if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if name, found := strings.CutPrefix(line, "model name"); found {
				if _, value, cut := strings.Cut(name, ":"); cut {
					return strings.TrimSpace(value)
				}
			}
		}
	}
	return "unknown processor"
}

func oneMinuteLoad() (float64, bool) {
	if data, err := os.ReadFile("/proc/loadavg"); err == nil {
		if fields := strings.Fields(string(data)); len(fields) > 0 {
			if load, err := strconv.ParseFloat(fields[0], 64); err == nil {
				return load, true
			}
		}
	}
	out, err := exec.Command("sysctl", "-n", "vm.loadavg").Output()
	if err != nil {
		return 0, false
	}
	fields := strings.Fields(strings.Trim(strings.TrimSpace(string(out)), "{}"))
	if len(fields) == 0 {
		return 0, false
	}
	load, err := strconv.ParseFloat(fields[0], 64)
	return load, err == nil
}

func quietLimit(cpus int) float64 {
	return float64(cpus) / 2
}
