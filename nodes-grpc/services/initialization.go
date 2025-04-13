package services

import (
	"log/slog"
	"nodes-grpc/services/virtualization"
	"nodes-grpc/utils"

	"github.com/vishvananda/netlink"
)

type Connection struct {
	VirtualizationService virtualization.VirtualizationInterface
}

func NewConnection() *Connection {
	// connection
	incusConnection := virtualization.InitIncusConnection()

	// services
	incusService := virtualization.NewIncusVirtualization(incusConnection)

	return &Connection{
		VirtualizationService: incusService,
	}
}

func CreateBridge() error {
	// check if bridge already created
	bridge, err := netlink.LinkByName(BRIDGE_NAME)
	if err != nil {
		return err
	}

	if bridge != nil {
		return nil
	}

	// get used ethernet interface
	ethInterface, err := utils.GetNodeUsedInterface()
	if err != nil {
		return err
	}

	eth, err := netlink.LinkByName(ethInterface)
	if err != nil {
		return err
	}

	bridgeLinkAttrs := netlink.NewLinkAttrs()
	bridgeLinkAttrs.Name = BRIDGE_NAME

	newBridge := &netlink.Bridge{LinkAttrs: bridgeLinkAttrs}
	err = netlink.LinkAdd(newBridge)
	if err != nil {
		slog.Error("CreateBridge(): could not create network bridge",
			"error", err.Error(),
		)

		return err
	}

	// add interfaces to bridge
	err = netlink.LinkSetMaster(eth, newBridge)
	if err != nil {
		slog.Error("CreateBridge(): could not set network interfaces master, removing bridge...",
			"error", err.Error(),
		)

		// remove the bridge
		err := netlink.LinkDel(newBridge)
		if err != nil {
			slog.Error("CreateBridge(): could not remove the bridge",
				"error", err.Error(),
			)

			return err
		}

		return err
	}

	return nil
}
