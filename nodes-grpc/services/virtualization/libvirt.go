package virtualization

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"nodes-grpc/services/db"
	virtualization_model "nodes-grpc/services/model/virtualization-model"
	"nodes-grpc/utils"
	"os"
	"strings"

	"github.com/digitalocean/go-libvirt"
	"github.com/google/uuid"
)

const (
	// cpu threshold
	NODE_BUSY_THRESHOLD float64 = 60

	// libvirt-related error
	LIBVIRT_NOT_INITIALIZED string = "libvirt connection is not initialized"

	// storage pool related
	STORAGE_POOL_NAME           string = "virt-tc"
	STORAGE_POOL_DIRECTORY_PATH string = "/var/lib/libvirt/virt-tc-images"
)

type LibvirtVirtualization struct {
	libvirtConnection *libvirt.Libvirt
	dbConnection      db.DatabaseInterface
}

func NewLibvirtService(
	libvirtConnection *libvirt.Libvirt, dbConnection db.DatabaseInterface,
) *LibvirtVirtualization {
	return &LibvirtVirtualization{
		libvirtConnection: libvirtConnection,
		dbConnection:      dbConnection,
	}
}

func (v *LibvirtVirtualization) Spawn(
	virtRequest virtualization_model.NodeCreateRequest,
) error {
	slog.Info("Spawn(): creating vm instance using libvirt...")

	if v.libvirtConnection == nil {
		slog.Error("Spawn(): could not create vm instance",
			"error", errors.New(LIBVIRT_NOT_INITIALIZED))

		return errors.New(LIBVIRT_NOT_INITIALIZED)
	}

	slog.Info("Spawn(): creating storage pool if not exists...")
	err := v.CreatePool()
	if err != nil {
		slog.Error("Spawn(): could not create storage pool",
			"error", err.Error(),
		)

		return err
	}

	// err := v.CreateNetwork()
	// if err != nil {
	// 	slog.Error("Spawn(): error creating network",
	// 		"error", err.Error(),
	// 	)
	//
	// 	return err
	// }

	// _, err = v.CreateVolume(virtualization_model.VirtualizationCreateRequest{})
	// if err != nil {
	// 	slog.Error("Spawn(): error creating volume",
	// 		"error", err.Error(),
	// 	)
	//
	// 	return err
	// }
	// v.CreateVM(virtualization_model.VirtualizationCreateRequest{})

	return nil
}

func (v *LibvirtVirtualization) Stop(
	virtRequest virtualization_model.InstanceIdentification,
) error {
	return nil
}

func (v *LibvirtVirtualization) List() error {
	return nil
}

func (v LibvirtVirtualization) CreateVM(
	virtRequest virtualization_model.VirtualizationCreateRequest,
) error {
	if v.libvirtConnection == nil {
		return errors.New(LIBVIRT_NOT_INITIALIZED)
	}

	template, err := os.ReadFile("./common/templates/domain-template.xml")
	if err != nil {
		return err
	}

	domainDefinition, err := v.libvirtConnection.DomainDefineXML(string(template))
	if err != nil {
		return err
	}

	err = v.libvirtConnection.DomainCreate(domainDefinition)
	if err != nil {
		return err
	}

	return nil
}

func (v LibvirtVirtualization) CreateVolume(
	virtRequest virtualization_model.VirtualizationCreateRequest,
) (string, error) {
	if v.libvirtConnection == nil {
		return "", errors.New(LIBVIRT_NOT_INITIALIZED)
	}

	storagePool, err := v.libvirtConnection.StoragePoolLookupByName(STORAGE_POOL_NAME)
	if err != nil {
		return "", err
	}

	volumeName := uuid.New().String()

	storageVolCreateFlags := libvirt.StorageVolCreateFlags(1)
	_, err = v.libvirtConnection.StorageVolCreateXML(storagePool,
		fmt.Sprintf(`
            <volume type='file'>
                <name>test.qcow2</name>
                <key>/var/lib/libvirt/images/test.qcow2</key>
                <capacity unit='bytes'>4294967296</capacity>
                <allocation unit='bytes'>856064</allocation>
                <physical unit='bytes'>4295884800</physical>
                <target>
                    <path>/var/lib/libvirt/images/test.qcow2</path>
                    <format type='qcow2'/>
                    <permissions>
                        <mode>0600</mode>
                        <owner>107</owner>
                        <group>107</group>
                        <label>system_u:object_r:svirt_image_t:s0:c723,c1015</label>
                    </permissions>
                    <compat>1.1</compat>
                    <clusterSize unit='B'>65536</clusterSize>
                    <features>
                        <lazy_refcounts/>
                    </features>
                </target>
            </volume>
        `), storageVolCreateFlags)
	if err != nil {
		return "", err
	}

	return volumeName, nil
}

func (v LibvirtVirtualization) CreatePool() error {
	slog.Info("CreatePool(): checking if storage pool is already created...")

	if v.libvirtConnection == nil {
		return errors.New(LIBVIRT_NOT_INITIALIZED)
	}

	storagePool, err := v.libvirtConnection.StoragePoolLookupByName(STORAGE_POOL_NAME)
	if err != nil {
		// check if storage pool is not available
		if strings.Contains(err.Error(), "no storage pool with matching name") {
			slog.Info("CreatePool(): storage pool doesn't exists, creating...")
			// create the directory
			_, err := os.Stat(STORAGE_POOL_DIRECTORY_PATH)
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					err = os.Mkdir(STORAGE_POOL_DIRECTORY_PATH, 0711)
					if err != nil {
						return err
					}
				}
			}

			_, err = v.libvirtConnection.StoragePoolCreateXML(
				fmt.Sprintf(`
                    <pool type="dir">
                        <name>%s</name>
                        <target>
                            <path>%s</path>
                        </target>
                    </pool>
                    `, STORAGE_POOL_NAME, STORAGE_POOL_DIRECTORY_PATH), libvirt.StoragePoolCreateWithBuild)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	isActive, err := v.libvirtConnection.StoragePoolIsActive(storagePool)
	if err != nil {
		return err
	}

	if isActive == 0 { // not running
		err := v.libvirtConnection.StoragePoolCreate(storagePool, libvirt.StoragePoolCreateNormal)
		if err != nil {
			return err
		}
	}

	slog.Info("CreatePool(): storage pool is ready")

	return nil
}

func (v LibvirtVirtualization) CreateNetwork() error {
	if v.libvirtConnection == nil {
		return errors.New(LIBVIRT_NOT_INITIALIZED)
	}

	thisNodeInterface, err := utils.GetNodeUsedInterface()
	if err != nil || thisNodeInterface == "" {
		return err
	}

	if strings.Contains(thisNodeInterface, "wl") {
		// wireless interface (wlan, wl)
	} else {
		// wired interface (eth, en)
	}

	_, err = os.ReadFile("./common/templates/network-template.xml")
	if err != nil {
		return err
	}

	// _, err = v.libvirtConnection.NetworkCreateXML(fmt.Sprintf(string(template), thisNodeInterface))
	// if err != nil {
	// 	slog.Error("CreateNetwork(): could not create network",
	// 		"error", err.Error(),
	// 	)
	//
	// 	return err
	// }

	return nil
}
