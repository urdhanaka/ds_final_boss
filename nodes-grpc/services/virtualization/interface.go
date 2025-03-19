package virtualization

type Virt interface {
	Spawn() error
}
