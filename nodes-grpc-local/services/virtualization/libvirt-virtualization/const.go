package libvirt_virtualization

// development and production
const (
	// bridge interface name
	BRIDGE_NAME = "k3s-br0"

	// libvirt related
	BASE_POOL_DIR = "/var/lib/libvirt/images"
	POOL_DIR      = "/var/lib/libvirt/k3s-virt"
	NVRAM_DIR     = "/var/lib/libvirt/qemu/nvram"

	// image that will be used for VM
	BASE_IMAGE_NAME = "oracular-server-cloudimg-amd64.img"
)

// production
const (
	LOADER         = "/usr/share/OVMF/OVMF_CODE_4M.fd"
	NVRAM_TEMPLATE = "/usr/share/OVMF/OVMF_VARS_4M.fd"
)

// development
const (
	LOADER_LOCAL         = "/usr/share/edk2/ovmf/OVMF_CODE.fd"
	NVRAM_TEMPLATE_LOCAL = "/usr/share/edk2/ovmf/OVMF_VARS.fd"
)
