package virtualization

type Virt interface {
	Spawn() error
	Stop() error
	List() error
}
