package websocket

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func (ws *Websocket) addOrUseMap(hostname string) {
	if ws.HostnameChan[hostname] != nil {
		return
	}

	logChan := make(chan string)
	ws.HostnameChan[hostname] = logChan
}

func (ws *Websocket) addLogToMap(hostname string, log string) {
	if ws.HostnameChan[hostname] != nil {
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
				slog.Info("log channel down")
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

func (ws *Websocket) start() {
	r := mux.NewRouter()
	http.HandleFunc("/{hostname}/status", ws.logHandler)
	http.Handle("/", r)
}
