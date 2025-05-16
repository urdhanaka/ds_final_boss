package libvirt_virtualization

const (
	BASE_POOL_DIR        = "/var/lib/libvirt/images"
	POOL_DIR             = "/var/lib/libvirt/k3s-virt"
	NVRAM_DIR            = "/var/lib/libvirt/qemu/nvram"
	LOADER_LOCAL         = "/usr/share/edk2/ovmf/OVMF_CODE.fd"
	NVRAM_TEMPLATE_LOCAL = "/usr/share/edk2/ovmf/OVMF_VARS.fd"
	NVRAM_TEMPLATE       = "/usr/share/OVMF/OVMF_VARS_4M.fd"
	BASE_IMAGE_NAME      = "oracular-server-cloudimg-amd64.img"
	// BASE_IMAGE_NAME = "ubuntu-24.10-minimal-cloudimg-amd64.img"
	// BASE_IMAGE_NAME = "debian-12-nocloud-amd64-20250428-2096.qcow2"
	// BASE_IMAGE_NAME = "debian-12-generic-amd64-20250428-2096.qcow2"

	BRIDGE_NAME = "k3s-br0"
	// BASE_IMAGE_NAME        = "k3s-virt.qcow2"
)
