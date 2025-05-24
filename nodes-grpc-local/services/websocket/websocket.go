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
		fmt.Println("map is present")
		return
	}
	if ws.HostnameChan == nil {
		fmt.Println("map is nil")
		return
	}

	logChan := make(chan string, WEBSOCKET_CHAN_MAX_SIZE)
	ws.HostnameChan[hostname] = logChan
}

func (ws *Websocket) AddLogToMap(hostname string, log string) {
	hostnameChan := ws.HostnameChan[hostname]

	hostnameChan <- log
}

func (ws *Websocket) logHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hostname := vars["hostname"]

	fmt.Println("this hostname is", hostname)

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket error",
			"error", err,
		)
	}
	defer c.Close()

	fmt.Println("accessing websocket")

	thisHostLogChan := ws.HostnameChan[hostname]

	fmt.Println(hostname)

	for {
		// expect no message from client
		select {
		case log, ok := <-thisHostLogChan:
			if !ok {
				slog.Info("log channel is empty")
				break
			}
			fmt.Println(log)

			err := c.WriteMessage(websocket.TextMessage, []byte(log))
			if err != nil {
				slog.Error("write error",
					"error", err,
				)
				break
			}
		}
	}
}

func (ws *Websocket) Start() {
	r := mux.NewRouter()
	r.HandleFunc("/status/{hostname}", ws.logHandler)
	http.Handle("/", r)

	slog.Info(fmt.Sprintf("starting websocket service at %s", WEBSOCKET_ADDRESS))

	log.Fatal(http.ListenAndServe(WEBSOCKET_ADDRESS, nil))
}
