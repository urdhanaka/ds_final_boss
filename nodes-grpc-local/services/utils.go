package services

import (
	"net"
	"nodes-grpc-local/services/model"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}

	return hostname
}

func getIpAddress() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().String()

	return strings.Split(localAddr, ":")[0]
}

func getMemoryStatus() (*model.MemoryStat, error) {
	stat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	// the value returned by the stat is in bytes,
	// convert it first to gigabytes
	gigabytes := 1024 * 1024 * 1024

	return &model.MemoryStat{
		Memory:           int(stat.Available / uint64(gigabytes)),
		MaxMemory:        int(stat.Total / uint64(gigabytes)),
		MemoryPercentage: stat.UsedPercent,
	}, nil
}

func getStorageStatus() (*model.StorageStat, error) {
	stat, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	gigabytes := 1024 * 1024 * 1024

	return &model.StorageStat{
		Storage:           int(stat.Free / uint64(gigabytes)),
		MaxStorage:        int(stat.Total / uint64(gigabytes)),
		StoragePercentage: stat.UsedPercent,
	}, nil
}

func getCpuStatus() (*model.CpuStat, error) {
	logicalCpuCounts, err := cpu.Counts(true)
	if err != nil {
		return nil, err
	}

	// executing shell script
	output, err := exec.Command("/bin/bash", "./scripts/get-used-vcpu.sh").Output()
	if err != nil {
		return nil, err
	}
	outputInt, _ := strconv.Atoi(string(output))

	// cpu usage
	usage, err := cpu.Percent(time.Second*1, false)
	if err != nil {
		return nil, err
	}

	return &model.CpuStat{
		LogicalCounts: logicalCpuCounts,
		FreeLogical:   outputInt,
		CurrentUsage:  usage[0],
	}, nil
}
