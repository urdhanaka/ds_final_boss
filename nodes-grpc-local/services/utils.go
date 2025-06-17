package services

import (
	"math/rand"
	"net"
	"nodes-grpc-local/services/model"
	"os"
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

func getCpuUsage() float64 {
	res, err := cpu.Percent(CPU_USAGE_INTERVAL*time.Second, false)
	if err != nil {
		return 0.0
	}

	return res[0]
}

func getMemoryUsage() (*model.MemoryStat, error) {
	stat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	return &model.MemoryStat{
		Memory:           stat.Available,
		MaxMemory:        stat.Total,
		MemoryPercentage: stat.UsedPercent,
	}, nil
}

func getStorageUsage() (*model.StorageStat, error) {
	stat, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	return &model.StorageStat{
		Storage:           stat.Free,
		MaxStorage:        stat.Total,
		StoragePercentage: stat.UsedPercent,
	}, nil
}
