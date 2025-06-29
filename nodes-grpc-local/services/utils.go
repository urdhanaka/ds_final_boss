package services

import (
	"math/rand"
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

func generateRandom(stringLength int) string {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	letters := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, stringLength)
	for i := range b {
		b[i] = letters[random.Intn(len(letters))]
	}

	return string(b)
}

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

	return &model.MemoryStat{
		Memory:           int(stat.Available),
		MaxMemory:        int(stat.Total),
		MemoryPercentage: stat.UsedPercent,
	}, nil
}

func getStorageStatus() (*model.StorageStat, error) {
	stat, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	return &model.StorageStat{
		Storage:           int(stat.Free),
		MaxStorage:        int(stat.Total),
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
