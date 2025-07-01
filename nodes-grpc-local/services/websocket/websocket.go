package websocket

import (
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	libvirt_virtualization "nodes-grpc-local/services/virtualization/libvirt-virtualization"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func (ws *Websocket) logHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceName := vars["instance_name"]

	sockFile := libvirt_virtualization.INSTANCE_LOGS_DIR + "/" + instanceName + ".sock"

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket error",
			"error", err,
		)
		return
	}
	defer c.Close()

	sock, err := net.Dial("unix", sockFile)
	if err != nil {
		slog.Error("socket error",
			"error", err,
		)
		return
	}
	defer sock.Close()

	thisHostLogChan := ws.HostnameChan[instanceName]

	for log := range thisHostLogChan {
		err := c.WriteMessage(websocket.TextMessage, []byte(log))
		if err != nil {
			slog.Error("error on the websocket chan",
				"error", err,
			)
		}
	}

	// for {
	// 	// expect no message from client
	// 	select {
	// 	case log, ok := <-thisHostLogChan:
	// 		if !ok {
	// 			slog.Info("log channel is empty")
	// 			break
	// 		}
	//
	// 		err := c.WriteMessage(websocket.TextMessage, []byte(log))
	// 		if err != nil {
	// 			slog.Error("write error",
	// 				"error", err,
	// 			)
	// 		}
	// 	}
	// }
}

// func sockReader(logChan <-chan string, instanceName string) {
// 	sockFile := libvirt_virtualization.INSTANCE_LOGS_DIR + "/" + instanceName + ".sock"
//
// 	l, err := net.Listen("unix", sockFile)
// 	if err != nil {
// 		slog.Error("could not read socket",
// 			"error", err,
// 		)
// 		return
// 	}
//
// 	for {
// 		fd, err := l.Accept()
// 		slog.Error("could not read socket",
// 			"error", err,
// 		)
// 		return
//
// 	}
// }

func (ws *Websocket) Start() {
	r := mux.NewRouter()
	r.HandleFunc("/status/{instance_name}", ws.logHandler)
	http.Handle("/", r)

	slog.Info(fmt.Sprintf("starting websocket service at %s", WEBSOCKET_ADDRESS))

	log.Fatal(http.ListenAndServe(WEBSOCKET_ADDRESS, nil))
}
