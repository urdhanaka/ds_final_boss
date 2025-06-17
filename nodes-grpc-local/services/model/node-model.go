package model

type MemoryStat struct {
	Memory           uint64
	MaxMemory        uint64
	MemoryPercentage float64
}

type StorageStat struct {
	Storage           uint64
	MaxStorage        uint64
	StoragePercentage float64
}
