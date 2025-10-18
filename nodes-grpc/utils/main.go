package utils

import (
	"log/slog"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	// threshold for detecting if a node is used or not
	//
	// value > 40.0 = node is used
	//
	// otherwise, node is not used
	CPU_PERCENTAGE_THRESHOLD float64 = 40.0

	// time (in second) for detecting cpu usage
	//
	// checking cpu usage now and then checking it again 5 seconds later
	TIME_BETWEEN_CPU_USAGE_CHECK = 5
)

type CpuStats struct {
	User, Nice, System, Idle, Iowait, Irq, Softirq, Steal uint64
}

func GetNodeIP() (string, error) {
	nodeInterfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, nodeInterface := range nodeInterfaces {
		interfaceName := nodeInterface.Name

		// skip loopback interface
		if strings.Index(interfaceName, "lo") == 0 {
			continue
		}

		// handling ethernet interface
		// common ethernet interface name: "en", "eth"
		// for TC purpose, all the node (lab computer)
		// will always use ethernet.
		// So handle the ethernet first
		if strings.Index(nodeInterface.Name, "en") == 0 ||
			strings.Index(nodeInterface.Name, "eth") == 0 {
			addrs, err := nodeInterface.Addrs()
			if err != nil {
				return "", err
			}

			for _, addr := range addrs {
				// handle ipv4 only for now
				if addr.Network() == "ip+net" && strings.Contains(addr.String(), ".") {
					ipOnly := strings.Split(addr.String(), "/")[0]

					return ipOnly, nil
				}
			}
		}

		if strings.Index(nodeInterface.Name, "en") == 0 ||
			strings.Index(nodeInterface.Name, "eth") == 0 {
			addrs, err := nodeInterface.Addrs()
			if err != nil {
				return "", err
			}

			for _, addr := range addrs {
				// handle ipv4 only for now
				if addr.Network() == "ip+net" && strings.Contains(addr.String(), ".") {
					ipOnly := strings.Split(addr.String(), "/")[0]

					return ipOnly, nil
				}
			}
		}
	}

	return "", nil
}

// Get node (physical machine) used network interface,
// could be ethernet or wireless, or nothing at all.
// Returns node interface and error
func GetNodeUsedInterface() (string, error) {
	nodeInterfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, nodeInterface := range nodeInterfaces {
		interfaceName := nodeInterface.Name

		if strings.Index(interfaceName, "lo") == 0 {
			continue
		}

		if strings.Index(nodeInterface.Name, "en") == 0 ||
			strings.Index(nodeInterface.Name, "eth") == 0 {
			addrs, err := nodeInterface.Addrs()
			if err != nil {
				return "", err
			}

			for _, addr := range addrs {
				if addr.Network() == "ip+net" && strings.Contains(addr.String(), ".") {
					return nodeInterface.Name, nil
				}
			}
		}
	}

	return "", nil
}

func IsNodeAvailable() bool {
	nodeAvailability := false

	prevStats, err := getCurrentCpuUsage()
	if err != nil {
		slog.Error("could not get cpu usage",
			"error", err.Error(),
		)

		return nodeAvailability
	}

	time.Sleep(TIME_BETWEEN_CPU_USAGE_CHECK * time.Second)

	currStats, err := getCurrentCpuUsage()
	if err != nil {
		slog.Error("could not get cpu usage",
			"error", err.Error(),
		)

		return nodeAvailability
	}

	prevTotal := prevStats.User + prevStats.Nice + prevStats.System +
		prevStats.Idle + prevStats.Iowait + prevStats.Irq + prevStats.Softirq + prevStats.Steal
	prevIdle := prevStats.Idle + prevStats.Iowait

	currTotal := currStats.User + currStats.Nice + currStats.System +
		currStats.Idle + currStats.Iowait + currStats.Irq + currStats.Softirq + currStats.Steal
	currIdle := currStats.Idle + currStats.Iowait

	totalDiff := currTotal - prevTotal
	idleDiff := currIdle - prevIdle

	if totalDiff == 0 {
		return nodeAvailability
	}

	usagePercentage := (float64(totalDiff-idleDiff) / float64(totalDiff)) * 100.0

	if usagePercentage < CPU_PERCENTAGE_THRESHOLD {
		nodeAvailability = true
	}

	return nodeAvailability
}

func getCurrentCpuUsage() (*CpuStats, error) {
	var cpuStats CpuStats

	procStatFile, err := os.ReadFile("/proc/stat")
	if err != nil {
		slog.Error("could not read /proc/stat file",
			"error", err.Error(),
		)

		return &cpuStats, err
	}

	lines := strings.Split(string(procStatFile), "\n")

	// get the first lines containing "cpu"
	// only use "cpu" for all the cpu cores usage
	fields := strings.Fields(lines[0])

	user, _ := strconv.ParseUint(fields[1], 10, 64)
	nice, _ := strconv.ParseUint(fields[2], 10, 64)
	system, _ := strconv.ParseUint(fields[3], 10, 64)
	idle, _ := strconv.ParseUint(fields[4], 10, 64)
	iowait, _ := strconv.ParseUint(fields[5], 10, 64)
	irq, _ := strconv.ParseUint(fields[6], 10, 64)
	softirq, _ := strconv.ParseUint(fields[7], 10, 64)
	steal, _ := strconv.ParseUint(fields[8], 10, 64)

	cpuStats.User = user
	cpuStats.Nice = nice
	cpuStats.System = system
	cpuStats.Idle = idle
	cpuStats.Iowait = iowait
	cpuStats.Irq = irq
	cpuStats.Softirq = softirq
	cpuStats.Steal = steal

	return &cpuStats, nil
}

func CleanK3SProcess(runningProcess *exec.Cmd) error {
	// if runningProcess is not allocated, return
	if runningProcess == nil {
		return nil
	}

	// kill the running process
	err := runningProcess.Process.Kill()
	if err != nil {
		return err
	}

	runningProcess = nil

	return nil
}

func RandomIntegerPicker(bottom, upper int) (int, error) {
	return 0, nil
}
