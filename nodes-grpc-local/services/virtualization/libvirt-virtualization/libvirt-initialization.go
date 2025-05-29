package libvirt_virtualization

import (
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"sync"

	"libvirt.org/go/libvirt"
)

var (
	cloudInitMut sync.Mutex
	imageMut     sync.Mutex
	networkMut   sync.Mutex
	efiMut       sync.Mutex
	spawnMut     sync.Mutex
)

func init() {
	slog.Info("checking libvirt requirements")

	slog.Info("checking if pool directory exists")
	_, err := os.Stat(POOL_DIR)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		slog.Info("pool directory doesn't exist, creating...")

		err = os.Mkdir(POOL_DIR, 711)
		if err != nil {
			slog.Error("could not create pool directory",
				"error", err.Error())
			os.Exit(1)
		}
	}

	// slog.Info("checking if virtual bridge interface exists")
	// bridgeIface, err := net.InterfaceByName(BRIDGE_NAME)
	// if err != nil && (err.Error() == "no such network interface") {
	//        fmt.Println(err)
	// 	slog.Error("could not get virtual bridge interface",
	// 		"error", err.Error(),
	// 	)
	// 	os.Exit(1)
	// }
	// if bridgeIface == nil {
	// 	slog.Info("virtual bridge interface doesn't exists, creating...")
	//
	// 	la := netlink.NewLinkAttrs()
	// 	la.Name = BRIDGE_NAME
	// 	bridge := &netlink.Bridge{LinkAttrs: la}
	// 	err = netlink.LinkAdd(bridge)
	// 	if err != nil && errors.Is(err, netlink.LinkNotFoundError{}) {
	// 		slog.Error("could not create virtual bridge interface",
	// 			"error", err,
	// 		)
	// 		os.Exit(1)
	// 	}
	//
	// 	// get ethernet interfaces
	// 	interfaces, err := net.Interfaces()
	// 	if err != nil {
	// 		slog.Error("could not get network interfaces",
	// 			"error", err.Error(),
	// 		)
	// 		os.Exit(1)
	// 	}
	// 	for _, iface := range interfaces {
	// 		// if interface is up and NOT loopback
	// 		if (iface.Flags&net.FlagUp) != 0 && (iface.Flags&net.FlagLoopback) == 0 {
	// 			addrs, err := iface.Addrs()
	// 			if err != nil {
	// 				slog.Error("could not get interface addresses",
	// 					"error", err.Error(),
	// 				)
	// 				os.Exit(1)
	// 			}
	//
	// 			for _, addr := range addrs {
	// 				ipnet, ok := addr.(*net.IPNet)
	// 				if ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
	// 					if strings.Contains(iface.Name, "en") || strings.Contains(iface.Name, "eth") {
	// 						// add the interface to the bridge
	// 						ethernetInterface, _ := netlink.LinkByName(iface.Name)
	// 						netlink.LinkSetMaster(ethernetInterface, bridge)
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	// }

	slog.Info("checking libvirt requirements completed")
}

func InitLibvirtConnection() *libvirt.Connect {
	c, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		slog.Error("error connecting to QEMU system",
			"error", err.Error(),
		)
		os.Exit(1)
	}

	return c
}
