package handlers

import (
	"fmt"
	"net/http"
	"nodes-grpc-be/services"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
	// CheckOrigin: func(r *http.Request) bool { // prod
	// 	origin := r.Header.Get("origin")
	//
	//        return origin == "http://localhost:8080"
	// },
}

type WebsocketHandler struct {
	streamManager *services.StreamManager
}

func NewWebsocketHandler(
	streamManager *services.StreamManager,
) *WebsocketHandler {
	return &WebsocketHandler{
		streamManager,
	}
}

func (h *WebsocketHandler) ReceiveLogs(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("receive: ", err)
		return
	}
	clusterId := c.Params.ByName("cluster_id")

	// setup stream logs struct
	cluster := h.streamManager.GetOrCreateClusterLogs(clusterId)
	if err := h.streamManager.SetReader(clusterId, conn); err != nil {
		return
	}
	defer func() {
		h.streamManager.RemoveReader(clusterId)
	}()

	// main loop
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if msg != nil {
				cluster.LogChan <- msg
			}
		}

		cluster.LogChan <- msg
	}
}

func (h *WebsocketHandler) StreamLogs(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("stream: ", err)
		return
	}

	clusterId := c.Params.ByName("cluster_id")
	clusterLogs := h.streamManager.GetOrCreateClusterLogs(clusterId)
	clusterLogs.Writer = conn
	defer func() {
		h.streamManager.RemoveWriter(clusterId)
		h.streamManager.CleanClusterLogs(clusterId)
		conn.Close()
	}()

	for msg := range clusterLogs.LogChan {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}
