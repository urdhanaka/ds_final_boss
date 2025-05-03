package libvirt_virtualization

const (
	BASE_POOL_DIR   = "/var/lib/libvirt/images"
	POOL_DIR        = "/var/lib/libvirt/k3s-virt"
	NVRAM_DIR       = "/var/lib/libvirt/qemu/nvram"
	NVRAM_TEMPLATE  = "/usr/share/edk2/ovmf/OVMF_VARS.fd"
	BASE_IMAGE_NAME = "k3s-virt.qcow2"
	BRIDGE_NAME     = "k3s-br0"
)
