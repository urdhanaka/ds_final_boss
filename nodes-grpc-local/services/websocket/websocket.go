package websocket

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func (ws *Websocket) AddMap(hostname string) {
	_, isPresent := ws.HostnameChan[hostname]
	if isPresent {
		return
	}
	if ws.HostnameChan == nil {
		return
	}

	logChan := make(chan string, WEBSOCKET_CHAN_MAX_SIZE)
	ws.HostnameChan[hostname] = logChan
}

func (ws *Websocket) AddLogToMap(hostname string, log string) {
	_, isPresent := ws.HostnameChan[hostname]
	if isPresent {
		return
	}
	if ws.HostnameChan == nil {
		return
	}

	hostnameChan := ws.HostnameChan[hostname]

	hostnameChan <- log
}

func (ws *Websocket) logHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hostname := vars["hostname"]

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket error",
			"error", err,
		)
	}
	defer c.Close()

	thisHostLogChan := ws.HostnameChan[hostname]

	for {
		// expect no message from client
		select {
		case log, ok := <-thisHostLogChan:
			if !ok {
				slog.Info("log channel is empty")
				return
			}

			err := c.WriteMessage(websocket.TextMessage, []byte(log))
			if err != nil {
				slog.Error("write error",
					"error", err,
				)
			}
		}
	}
}

func (ws *Websocket) Start() {
	r := mux.NewRouter()
	http.HandleFunc("/{hostname}/status", ws.logHandler)
	http.Handle("/", r)

    slog.Info(fmt.Sprintf("starting websocket service at %s", WEBSOCKET_ADDRESS))

	log.Fatal(http.ListenAndServe(WEBSOCKET_ADDRESS, nil))
}
