package model

type MemoryStat struct {
	Memory           int
	MaxMemory        int
	MemoryPercentage float64
}

type StorageStat struct {
	Storage           int
	MaxStorage        int
	StoragePercentage float64
}

type CpuStat struct {
	LogicalCounts int
	FreeLogical   int
	CurrentUsage  float64
}
