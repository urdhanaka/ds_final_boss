package libvirt_virtualization

// development and production
const (
	// bridge interface name
	// BRIDGE_NAME = "k3s-br0"
	BRIDGE_NAME = "default"

	// libvirt related
	BASE_POOL_DIR     = "/var/lib/libvirt/images"
	POOL_DIR          = "/var/lib/libvirt/k3s-virt"
	NVRAM_DIR         = "/var/lib/libvirt/qemu/nvram"
	INSTANCE_LOGS_DIR = "/var/lib/libvirt/instance-logs"

	// image that will be used for VM
	BASE_IMAGE_NAME = "oracular-server-cloudimg-amd64.img"
)

// production environment
// software engineering lab computer, ubuntu 24.04
const (
	LOADER         = "/usr/share/OVMF/OVMF_CODE_4M.fd"
	NVRAM_TEMPLATE = "/usr/share/OVMF/OVMF_VARS_4M.fd"
)

// development environment
// my own laptop, fedora 42 immutable
const (
	LOADER_LOCAL         = "/usr/share/edk2/ovmf/OVMF_CODE.fd"
	NVRAM_TEMPLATE_LOCAL = "/usr/share/edk2/ovmf/OVMF_VARS.fd"
)
