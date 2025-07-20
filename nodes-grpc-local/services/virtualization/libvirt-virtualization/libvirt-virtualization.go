package libvirt_virtualization

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	virtualization_model "nodes-grpc-local/services/model/virtualization-model"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"libvirt.org/go/libvirt"
	"libvirt.org/go/libvirtxml"
)

const (
	// timeout in second when waiting the cloud init operations
	CLOUD_INIT_TIMEOUT = 30

	SHUTDOWN_RETRIES = 3
)

type LibvirtVirtualization struct {
	libvirtConnection *libvirt.Connect
}

func NewLibvirtVirtualization(
	libvirtConnection *libvirt.Connect,
) *LibvirtVirtualization {
	return &LibvirtVirtualization{
		libvirtConnection: libvirtConnection,
	}
}

// wrapper for create master and create worker
// this function will be used in queue
func (c *LibvirtVirtualization) CreateInstance(
	ctx context.Context,
	virtRequest *virtualization_model.CreateInstanceRequest,
) (*virtualization_model.VirtCreateInstanceResponse, error) {
	if virtRequest.IsMaster {
		return c.createMaster(ctx, virtRequest)
	} else {
		return c.createWorker(ctx, virtRequest)
	}
}

func (c *LibvirtVirtualization) createMaster(
	ctx context.Context,
	virtRequest *virtualization_model.CreateInstanceRequest,
) (*virtualization_model.VirtCreateInstanceResponse, error) {
	createRes := new(virtualization_model.VirtCreateInstanceResponse)
	createRes.CreationStatus = false

	thisInstanceName := virtRequest.Name

	// defering vm cleanup in case
	// error happens or deadline exceeded
	defer func() {
		if createRes.CreationStatus == false {
			slogFunction(virtRequest.ClusterName, thisInstanceName, "failed to create instance, deleting...", nil)
			err := c.DeleteInstance(thisInstanceName)
			if err != nil {
				slogFunction(virtRequest.ClusterName, thisInstanceName, "could not delete instance", err)
			}
		}
	}()

	slog.Info(fmt.Sprintf("master node name is %s", thisInstanceName))
	slogFunction(virtRequest.ClusterName, thisInstanceName, "creating master instance", nil)

	// create network
	slogFunction(virtRequest.ClusterName, thisInstanceName, "creating node network", nil)
	err := createNetwork()
	if err != nil {
		slogFunction(virtRequest.ClusterName, thisInstanceName, "could not create node network", err)

		return createRes, err
	}

	// create cloudinit configuration
	slogFunction(virtRequest.ClusterName, thisInstanceName, "creating node cloud-init configuration", nil)
	err = createCloudInitMaster(thisInstanceName, virtRequest.Token)
	if err != nil {
		slogFunction(virtRequest.ClusterName, thisInstanceName, "could not create node cloud-init configuration", err)

		return createRes, err
	}

	// creating the image
	slogFunction(virtRequest.ClusterName, thisInstanceName, "creating base image for instance", nil)
	err = copyImage(thisInstanceName, virtRequest)
	if err != nil {
		slogFunction(virtRequest.ClusterName, thisInstanceName, "could not create node image", err)

		return createRes, err
	}

	// base xml for the vm
	slogFunction(virtRequest.ClusterName, thisInstanceName, "creating instance base xml", nil)
	domainXmlConfig, err := createBaseXml(thisInstanceName, virtRequest)
	if err != nil {
		slogFunction(virtRequest.ClusterName, thisInstanceName, "could not create node base xml", err)

		return createRes, err
	}

	// spawning
	slogFunction(virtRequest.ClusterName, thisInstanceName, "spawning node", nil)
	dom, err := c.libvirtConnection.DomainCreateXML(domainXmlConfig, libvirt.DomainCreateFlags(0))
	if err != nil {
		slogFunction(virtRequest.ClusterName, thisInstanceName, "could not spawn node", err)

		return createRes, err
	}

	go sendLogs(thisInstanceName, virtRequest.MainServerIpAddress, virtRequest.ClusterId)

	slogFunction(virtRequest.ClusterName, thisInstanceName, "waiting until the vm is ready..", nil)
	time.Sleep(CLOUD_INIT_TIMEOUT * time.Second)
	waitCloudInitCmd := "cloud-init status --wait"
	_, err = guestAgentExecStatus(dom, waitCloudInitCmd)
	if err != nil {
		slogFunction(virtRequest.ClusterName, thisInstanceName, "error waiting cloud-init process", err)

		return createRes, err
	}

	// creating token
	kubeCreateTokenCmd := "k3s kubectl -n kubernetes-dashboard create token admin-user"
	createTokenStatus, err := guestAgentExecStatus(dom, kubeCreateTokenCmd)
	if err != nil {
		slogFunction(virtRequest.ClusterName, thisInstanceName, "error creating kubernetes dashboard token", err)
		c.DeleteInstance(thisInstanceName)

		return createRes, err
	}
	decodedTokenBytes, err := base64.StdEncoding.DecodeString(createTokenStatus.Return.OutData)
	if err != nil {
		slogFunction(virtRequest.ClusterName, thisInstanceName, "error decoding kubernetes dashboard bytes", err)
		c.DeleteInstance(thisInstanceName)

		return createRes, err
	}

	// getting the IP address
	// DO NOT TOUCH THE sed SEQUENCE
	ipAddressCmd := `ip -f inet addr show enp1s0 | awk '/inet / {print $2}' | cut -d'/' -f1 | tr -d '\n'`
	ipAddressStatus, err := guestAgentExecStatus(dom, ipAddressCmd)
	if err != nil {
		slogFunction(virtRequest.ClusterName, thisInstanceName, "error getting master instance ip address", err)

		return createRes, err
	}
	decodedIpAddressBytes, err := base64.StdEncoding.DecodeString(ipAddressStatus.Return.OutData)
	if err != nil {
		slogFunction(virtRequest.ClusterName, thisInstanceName, "error decoding kubernetes dashboard bytes", err)

		return createRes, err
	}

	// getting the kubeconfig
	getKubeConfigContentCmd := fmt.Sprintf(`cat /etc/rancher/k3s/k3s.yaml | sed 's/127.0.0.1/%s/g' | head -c -1`, string(decodedIpAddressBytes))
	getKubeConfigContentStatus, err := guestAgentExecStatus(dom, getKubeConfigContentCmd)
	if err != nil {
		slogFunction(virtRequest.ClusterName, thisInstanceName, "error getting kubeconfig contents", err)

		return createRes, err
	}
	decodedKubeconfigContentsBytes, err := base64.StdEncoding.DecodeString(getKubeConfigContentStatus.Return.OutData)
	if err != nil {
		slogFunction(virtRequest.ClusterName, thisInstanceName, "error decoding kubernetes dashboard bytes", err)

		return createRes, err
	}

	createRes.CreationStatus = true
	createRes.DashboardToken = string(decodedTokenBytes)
	createRes.MasterIpAddress = string(decodedIpAddressBytes)
	createRes.KubeconfigContents = decodedKubeconfigContentsBytes

	slogFunction(virtRequest.ClusterName, thisInstanceName, "vm provision done", err)

	return createRes, nil
}

func (c *LibvirtVirtualization) createWorker(
	ctx context.Context,
	virtRequest *virtualization_model.CreateInstanceRequest,
) (*virtualization_model.VirtCreateInstanceResponse, error) {
	createRes := new(virtualization_model.VirtCreateInstanceResponse)
	createRes.CreationStatus = false

	thisInstanceName := virtRequest.Name

	defer func() {
		if createRes.CreationStatus == false {
			slogFunction(virtRequest.ClusterName, thisInstanceName, "failed to create instance, deleting...", nil)
			err := c.DeleteInstance(thisInstanceName)
			if err != nil {
				slogFunction(virtRequest.ClusterName, thisInstanceName, "could not delete instance", err)
			}
		}
	}()

	slog.Info(fmt.Sprintf("worker node name is %s", thisInstanceName))
	slogFunction(virtRequest.ClusterName, thisInstanceName, "creating worker instance", nil)

	// create network
	slogFunction(virtRequest.ClusterName, thisInstanceName, "creating instance network", nil)
	err := createNetwork()
	if err != nil {
		slogFunction(virtRequest.ClusterName, thisInstanceName, "could not create node network", err)

		return createRes, err
	}

	// create cloudinit configuration
	slogFunction(virtRequest.ClusterName, thisInstanceName, "creating node cloud-init configuration", nil)
	err = createCloudInitWorker(thisInstanceName, virtRequest.MasterIpAddress, virtRequest.Token)
	if err != nil {
		slogFunction(virtRequest.ClusterName, thisInstanceName, "could not create node cloud-init configuration", err)

		return createRes, err
	}

	// creating the image
	slogFunction(virtRequest.ClusterName, thisInstanceName, "creating node image", nil)
	err = copyImage(thisInstanceName, virtRequest)
	if err != nil {
		slogFunction(virtRequest.ClusterName, thisInstanceName, "could not create node image", err)

		return createRes, err
	}

	// base xml for the vm
	slogFunction(virtRequest.ClusterName, thisInstanceName, "creating node base xml", nil)
	domainXmlConfig, err := createBaseXml(thisInstanceName, virtRequest)
	if err != nil {
		slogFunction(virtRequest.ClusterName, thisInstanceName, "could not create node base xml", err)

		return createRes, err
	}

	// spawning
	slogFunction(virtRequest.ClusterName, thisInstanceName, "spawning node", nil)
	dom, err := c.libvirtConnection.DomainCreateXML(domainXmlConfig, libvirt.DOMAIN_NONE)
	if err != nil {
		slogFunction(virtRequest.Name, thisInstanceName, "could not spawn node", err)

		return createRes, err
	}

	slogFunction(virtRequest.ClusterName, thisInstanceName, "waiting until the vm is ready..", nil)
	time.Sleep(CLOUD_INIT_TIMEOUT * time.Second)
	waitCloudInitCmd := "cloud-init status --wait"
	_, err = guestAgentExecStatus(dom, waitCloudInitCmd)
	if err != nil {
		slogFunction(virtRequest.ClusterName, thisInstanceName, "could not spawn node", err)

		return createRes, err
	}

	createRes.CreationStatus = true

	slogFunction(virtRequest.ClusterName, thisInstanceName, "vm provision done", err)

	return createRes, nil
}

func createBaseXml(
	instanceName string,
	instanceConfig *virtualization_model.CreateInstanceRequest,
) (string, error) {
	instanceStorage := POOL_DIR + "/" + instanceName + ".qcow2"
	seedFile := POOL_DIR + "/" + instanceName + ".iso"
	logSocket := INSTANCE_LOGS_DIR + "/" + instanceName + ".sock"

	domConfig := &libvirtxml.Domain{
		Type: "kvm",
		Name: instanceName,
		Metadata: &libvirtxml.DomainMetadata{
			XML: `
<libosinfo:libosinfo xmlns:libosinfo="http://libosinfo.org/xmlns/libvirt/domain/1.0">
  <libosinfo:os id="http://ubuntu.com/ubuntu/24.10"/>
</libosinfo:libosinfo>`,
		},
		Memory: &libvirtxml.DomainMemory{
			Value: uint(instanceConfig.Memory),
			Unit:  "GB",
		},
		VCPU: &libvirtxml.DomainVCPU{
			Value: uint(instanceConfig.Cpu),
		},
		OS: &libvirtxml.DomainOS{
			Firmware: "efi",
			Type: &libvirtxml.DomainOSType{
				Arch:    "x86_64",
				Machine: "q35",
				Type:    "hvm",
			},
			FirmwareInfo: &libvirtxml.DomainOSFirmwareInfo{
				Features: []libvirtxml.DomainOSFirmwareFeature{
					{
						Enabled: "no",
						Name:    "enrolled-keys",
					},
					{
						Enabled: "no",
						Name:    "secure-boot",
					},
				},
			},
			Loader: &libvirtxml.DomainLoader{
				Readonly: "yes",
				Type:     "pflash",
				Format:   "raw",
				// Path:     LOADER_LOCAL, // NOTE: works on local only
				Path: LOADER,
			},
			NVRam: &libvirtxml.DomainNVRam{
				NVRam: fmt.Sprintf("/var/lib/libvirt/qemu/nvram/%s_VARS.fd", instanceName),
				// Template: NVRAM_TEMPLATE_LOCAL, // NOTE: only works in local
				Template:       NVRAM_TEMPLATE, // NOTE: directory according to ubuntu 24.04
				TemplateFormat: "raw",
				Format:         "raw",
			},
			BootDevices: []libvirtxml.DomainBootDevice{
				{
					Dev: "hd",
				},
			},
			BIOS: &libvirtxml.DomainBIOS{
				UseSerial: "yes",
			},
		},
		Features: &libvirtxml.DomainFeatureList{
			KVM: &libvirtxml.DomainFeatureKVM{
				Hidden: &libvirtxml.DomainFeatureState{
					State: "on",
				},
			},
			ACPI: &libvirtxml.DomainFeature{},
			APIC: &libvirtxml.DomainFeatureAPIC{},
			VMPort: &libvirtxml.DomainFeatureState{
				State: "off",
			},
		},
		CPU: &libvirtxml.DomainCPU{
			Mode: "host-passthrough",
		},
		OnPoweroff: "destroy",
		OnCrash:    "restart",
		Devices: &libvirtxml.DomainDeviceList{
			Emulator: "/usr/bin/qemu-system-x86_64",
			Disks: []libvirtxml.DomainDisk{
				{
					Device: "disk",
					Driver: &libvirtxml.DomainDiskDriver{
						Name: "qemu",
						Type: "qcow2",
					},
					Source: &libvirtxml.DomainDiskSource{
						File: &libvirtxml.DomainDiskSourceFile{
							File: instanceStorage,
						},
					},
					Target: &libvirtxml.DomainDiskTarget{
						Dev: "vda",
						Bus: "virtio",
					},
				},
				{
					Device: "cdrom",
					Driver: &libvirtxml.DomainDiskDriver{
						Name: "qemu",
						Type: "raw",
					},
					Source: &libvirtxml.DomainDiskSource{
						File: &libvirtxml.DomainDiskSourceFile{
							File: seedFile,
						},
					},
					Target: &libvirtxml.DomainDiskTarget{
						Dev: "sda",
						Bus: "sata",
					},
					ReadOnly: &libvirtxml.DomainDiskReadOnly{},
				},
			},
			Interfaces: []libvirtxml.DomainInterface{
				{
					Source: &libvirtxml.DomainInterfaceSource{
						Bridge: &libvirtxml.DomainInterfaceSourceBridge{
							Bridge: BRIDGE_NAME,
						},
						// Network: &libvirtxml.DomainInterfaceSourceNetwork{
						// 	Network: DEFAULT_BRIDGE_NAME,
						// 	Bridge:  DEFAULT_BRIDGE_NAME,
						// },
					},
					Model: &libvirtxml.DomainInterfaceModel{
						Type: "virtio",
					},
				},
			},
			Serials: []libvirtxml.DomainSerial{
				{
					Protocol: &libvirtxml.DomainChardevProtocol{
						Type: "pty",
					},
					Target: &libvirtxml.DomainSerialTarget{
						Type: "isa-serial",
						Port: func() *uint {
							temp := uint(0)
							return &temp
						}(),
						Model: &libvirtxml.DomainSerialTargetModel{
							Name: "isa-serial",
						},
					},
				},
				{
					Source: &libvirtxml.DomainChardevSource{
						UNIX: &libvirtxml.DomainChardevSourceUNIX{
							Mode: "bind",
							Path: logSocket,
						},
					},
					Target: &libvirtxml.DomainSerialTarget{
						Type: "isa-serial",
						Port: func() *uint {
							temp := uint(1)
							return &temp
						}(),
						Model: &libvirtxml.DomainSerialTargetModel{
							Name: "isa-serial",
						},
					},
				},
			},
			Consoles: []libvirtxml.DomainConsole{
				{
					Protocol: &libvirtxml.DomainChardevProtocol{
						Type: "pty",
					},
					Target: &libvirtxml.DomainConsoleTarget{
						Type: "serial",
						Port: func() *uint {
							temp := uint(1)
							return &temp
						}(),
					},
				},
				// {
				// 	Source: &libvirtxml.DomainChardevSource{
				// 		UNIX: &libvirtxml.DomainChardevSourceUNIX{
				// 			Mode: "bind",
				// 			Path: logSocket,
				// 		},
				// 	},
				// 	Target: &libvirtxml.DomainConsoleTarget{
				// 		Type: "serial",
				// 		Port: func() *uint {
				// 			temp := uint(1)
				// 			return &temp
				// 		}(),
				// 	},
				// },
			},
			Channels: []libvirtxml.DomainChannel{
				{
					Target: &libvirtxml.DomainChannelTarget{
						VirtIO: &libvirtxml.DomainChannelTargetVirtIO{
							Name: "org.qemu.guest_agent.0",
						},
					},
				},
			},
		},
	}
	xmlConfig, err := domConfig.Marshal()
	if err != nil {
		return "", err
	}

	// hacky way to handle <channel>
	// why, libvirtxml, why?
	res := strings.Replace(xmlConfig, "<channel>", `<channel type="unix">`, 1)

	return res, nil
}

func copyImage(
	instanceName string,
	virtRequest *virtualization_model.CreateInstanceRequest,
) error {
	imageMut.Lock()
	defer imageMut.Unlock()

	destinationPath := POOL_DIR + "/" + instanceName + ".qcow2"

	data, err := os.ReadFile(BASE_IMAGE_NAME_FULL_PATH)
	if err != nil {
		return err
	}

	err = os.WriteFile(destinationPath, data, 0644)
	if err != nil {
		return err
	}

	// resize d := exec.Command("qemu-img", "resize", destinationPath, "+10G")
	resizeCmd := exec.Command("qemu-img", "resize", destinationPath, fmt.Sprintf("+%dG", virtRequest.Storage))
	err = resizeCmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func createCloudInitMaster(instanceName string, clusterToken string) error {
	cloudInitMut.Lock()
	defer cloudInitMut.Unlock()

	userDataContent := fmt.Sprintf(`#cloud-config
hostname: %s
locale: en_US.UTF-8
timezone: Asia/Jakarta
users:
- default
- name: user
  groups: sudo
  sudo: ALL=(ALL:ALL) ALL
  plain_text_passwd: user
  lock_passwd: false
  shell: /bin/bash

write_files:
- path: /root/service-account.yaml
  content: |
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: admin-user
      namespace: kubernetes-dashboard
- path: /root/role-binding.yaml
  content: |
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRoleBinding
    metadata:
      name: admin-user
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: ClusterRole
      name: cluster-admin
    subjects:
    - kind: ServiceAccount
      name: admin-user
      namespace: kubernetes-dashboard
- path: /etc/systemd/system/kube-dashboard.service
  content: |
    [Unit]
    Description=Kubernetes dashboard
    Wants=network-online.target
    After=k3s.service

    [Install]
    WantedBy=multi-user.target

    [Service]
    Type=simple
    User=root
    Restart=always
    RestartSec=5s
    ExecStart=/usr/local/bin/k3s \
        kubectl -n kubernetes-dashboard \
        port-forward svc/kubernetes-dashboard-kong-proxy \
        8443:443 --address 0.0.0.0 \

runcmd:
- |
  echo "wait until network is connected"
  nm-online -s

  echo "running k3s"
  curl -sfL https://get.k3s.io | INSTALL_K3S_SKIP_DOWNLOAD=true INSTALL_K3S_EXEC="server --token %s" sh -s -

  export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
  echo "export KUBECONFIG=/etc/rancher/k3s/k3s.yaml" >> /etc/profile

  echo "creating kubernetes dashboard"
  helm repo add kubernetes-dashboard https://kubernetes.github.io/dashboard/
  helm upgrade --install kubernetes-dashboard kubernetes-dashboard/kubernetes-dashboard --create-namespace --namespace kubernetes-dashboard

  echo "setting up user for kubernetes dashboard"
  k3s kubectl apply -f /root/service-account.yaml -f /root/role-binding.yaml

  echo "waiting until all pods in the kubernetes-dashboard namespaces is running"
  k3s kubectl wait pod --all --for=condition=Ready --namespace=kubernetes-dashboard --timeout=-1s
  systemctl start kube-dashboard.service

  echo "done"
`, instanceName, clusterToken)

	err := os.WriteFile(USER_DATA_CONFIG_PATH, []byte(userDataContent), 0644)
	if err != nil {
		return err
	}

	// create the iso
	cmd := exec.Command(
		"cloud-localds",
		"-N", NETWORK_CONFIG_PATH, // network-config
		POOL_DIR+"/"+instanceName+".iso", // iso path
		USER_DATA_CONFIG_PATH,            // user-data path
	)
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func createCloudInitWorker(instanceName string, masterIp string, clusterToken string) error {
	cloudInitMut.Lock()
	defer cloudInitMut.Unlock()

	userDataContent := fmt.Sprintf(`#cloud-config
hostname: %s
locale: en_US.UTF-8
timezone: Asia/Jakarta
users:
- default
- name: user
  groups: sudo
  sudo: ALL=(ALL:ALL) ALL
  plain_text_passwd: user
  lock_passwd: false
  shell: /bin/bash

runcmd:
- echo "installing k3s"
- curl -sfL https://get.k3s.io | INSTALL_K3S_SKIP_DOWNLOAD=true INSTALL_K3S_EXEC="agent --server https://%s:6443 --token %s" sh -s -
- echo "done"
`, instanceName, masterIp, clusterToken)

	err := os.WriteFile(USER_DATA_CONFIG_PATH, []byte(userDataContent), 0644)
	if err != nil {
		return err
	}

	// create the iso
	cmd := exec.Command(
		"cloud-localds",
		"-N", NETWORK_CONFIG_PATH, // network-config
		POOL_DIR+"/"+instanceName+".iso", // iso path
		USER_DATA_CONFIG_PATH,            // user-data path
	)
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// CLEAR, no problem
func createNetwork() error {
	networkMut.Lock()
	defer networkMut.Unlock()

	filePath := BASE_POOL_DIR + "/" + "network-config"
	userDataContent := `network:
  version: 2
  renderer: NetworkManager
  ethernets:
    enp1s0:
      dhcp4: true`

	err := os.WriteFile(filePath, []byte(userDataContent), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c *LibvirtVirtualization) DeleteInstance(
	domainName string,
) error {
	slog.Info(fmt.Sprintf("deleting instance %s", domainName))

	domain, err := c.libvirtConnection.LookupDomainByName(domainName)
	if err != nil {
		// if the err value is other than VIR_ERR_NO_DOMAIN, return err
		// otherwise, just continue
		if !errors.Is(err, libvirt.ERR_NO_DOMAIN) {
			slog.Error("error getting the domain",
				"error", err,
			)

			return err
		}
	}

	// every domain is set to be deleted when shutting down
	// make sure the domain is not NULL
	if domain != nil {
		var err error

		for i := 1; i <= SHUTDOWN_RETRIES; i++ {
			err = domain.Shutdown()
			if err != nil {
				slog.Error("error occurred when shutting down the domain, retrying...",
					"error", err,
				)
			} else {
				break
			}

			time.Sleep(5 * time.Second)
		}

		if err != nil {
			slog.Info("could not normally shut down the domain, forcing the domain to shutdown...")
			err = domain.Destroy()
			if err != nil {
				slog.Error("could not destroy the domain",
					"error", err,
				)

				return err
			}
		}
	}

	// domain files cleanup
	deleteFilesCommand := fmt.Sprintf("rm %s/%s.*", POOL_DIR, domainName)
	cmd := exec.Command("/bin/bash", "-c", deleteFilesCommand)
	err = cmd.Run()
	if err != nil {
		slog.Error("could not clean domain files",
			"error", err,
		)
		return err
	}
	deleteFilesCommand = fmt.Sprintf("rm %s/%s.*", NVRAM_DIR, domainName)
	cmd = exec.Command("/bin/bash", "-c", deleteFilesCommand)
    out, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("could not clean domain files",
			"error", err,
            "output", string(out),
		)
		return err
	}

	return nil
}

func guestAgentExecStatus(
	dom *libvirt.Domain,
	execString string,
) (*ExecStatusQemuGuestAgent, error) {
	res := new(ExecStatusQemuGuestAgent)

	execCmd := fmt.Sprintf(`{"execute":"guest-exec","arguments":{"path":"/bin/bash","arg":["-c", "%s"],"capture-output":true}}`,
		execString)
	cmdPid, err := dom.QemuAgentCommand(execCmd, libvirt.DOMAIN_QEMU_AGENT_COMMAND_BLOCK, 0)
	if err != nil {
		return res, err
	}

	// handle the pid
	pidStruct := new(PidQemuGuestAgent)
	err = json.Unmarshal([]byte(cmdPid), pidStruct)
	if err != nil {
		return res, err
	}

	for {
		cmd := fmt.Sprintf(`{"execute":"guest-exec-status","arguments":{"pid":%d}}`, pidStruct.Return.PID)
		status, err := dom.QemuAgentCommand(cmd, libvirt.DOMAIN_QEMU_AGENT_COMMAND_BLOCK, 0)
		if err != nil {
			return res, err
		}

		err = json.Unmarshal([]byte(status), res)
		if err != nil {
			slog.Error("gues-exec error",
				"error", err,
			)
			return res, err
		}

		if res.Return.Exited {
			break
		}

		time.Sleep(5 * time.Second)
	}

	return res, nil
}

func sendLogs(
	instanceName string,
	mainServerIpAddress string,
	clusterId string,
) {
	var sock net.Conn
	var err error

	logSocketFile := INSTANCE_LOGS_DIR + "/" + instanceName + ".sock"

	for {
		sock, err = net.Dial("unix", logSocketFile)
		if err != nil {
			slog.Error(fmt.Sprintf("%s | error accessing socket file, retrying...", instanceName),
				"error", err,
			)
			time.Sleep(1 * time.Second)

			continue
		}
		break
	}
	defer sock.Close()

	u := url.URL{
		Scheme: "ws",
		Host:   mainServerIpAddress + ":8000", // directly send to the backend
		Path:   fmt.Sprintf("/api/logs/receive/%s", clusterId),
	}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		slog.Error(fmt.Sprintf("%s | error dialing websocket", instanceName),
			"error", err,
		)
		return
	}
	defer c.Close()

	scanner := bufio.NewScanner(sock)
	for scanner.Scan() {
		line := scanner.Text()
		if err := c.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
			slog.Error("Send error:", "error", err)
			break
		}
	}
}
