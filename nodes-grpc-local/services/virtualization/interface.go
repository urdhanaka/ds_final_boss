package virtualization

type VirtualizationInterface interface {
	CreateMaster() error
	CreateWorker() error
}
