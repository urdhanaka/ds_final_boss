package services

import (
	"fmt"
	libvirt_virtualization "nodes-grpc-local/services/virtualization/libvirt-virtualization"
	"nodes-grpc-local/services/websocket"
	"os"
)

type InitStruct struct {
	VirtualizationService *libvirt_virtualization.LibvirtVirtualization
	WebsocketService      *websocket.Websocket
	// QueueService          *queue.Queue
	Lab          string
	MainServerIp string
}

func NewInitStruct() *InitStruct {
	// redisClient := queue.InitRedisConnection()

	// queueService := queue.NewQueue(redisClient)

	websocketConnection := websocket.NewWebsocket()

	libvirtConnection := libvirt_virtualization.InitLibvirtConnection()
	libvirtService := libvirt_virtualization.NewLibvirtVirtualization(libvirtConnection)

	input := handleInput()

	return &InitStruct{
		VirtualizationService: libvirtService,
		// QueueService:          queueService,
		WebsocketService: websocketConnection,
		Lab:              input[1],
		MainServerIp:     input[2],
	}
}

// handle which lab the computer belongs to
func handleInput() []string {
	labMapping := map[string]string{
		"AJK":  "AJK",
		"RPL":  "RPL",
		"KBJ":  "KBJ",
		"GIGA": "GIGA",
		"KCV":  "KCV",
		"MCI":  "MCI",
		"PKT":  "PKT",
		"AP":   "AP",
	}

	args := os.Args
	if len(args) != 3 {
		fmt.Printf("usage: %s <nama_lab> <main server ip address>\n", args[0])
		fmt.Println("available lab (case sensitive): AJK, RPL, KBJ, GIGA, KCV, MCI, PKT, AP")
		os.Exit(1)
	}

	_, ok := labMapping[args[1]]
	if !ok {
		fmt.Println("lab is not available")
		fmt.Println("available lab (case sensitive): AJK, RPL, KBJ, GIGA, KCV, MCI, PKT, AP")
		fmt.Printf("usage: %s <nama_lab>\n", args[0])
		os.Exit(1)
	}

	return args
}
