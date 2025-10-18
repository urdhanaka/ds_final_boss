package handlers

import (
	"nodes-grpc-frontend-local/src/services"

	"github.com/gofiber/contrib/websocket"
)

type WebsocketHandler struct {
	streamManager *services.StreamManager
}

func NewWebsocketHandler(streamManager *services.StreamManager) *WebsocketHandler {
	return &WebsocketHandler{
		streamManager,
	}
}

func (h *WebsocketHandler) ReceiveLogs(c *websocket.Conn) {
	clusterName := c.Params("cluster_name")
	cluster := h.streamManager.GetOrCreateCluster(clusterName)

	if !h.streamManager.SetWriter(clusterName, c) {
		c.Close()
		return
	}

	defer func() {
		h.streamManager.RemoveWriter(clusterName)
		c.Close()
	}()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			if msg != nil {
				cluster.LogChan <- msg
			}
			break
		}
		cluster.LogChan <- msg
	}
}

func (h *WebsocketHandler) StreamLogs(c *websocket.Conn) {
	clusterName := c.Params("cluster_name")
	cluster := h.streamManager.GetOrCreateCluster(clusterName)

	cluster.Frontend = c
	defer func() {
		cluster.Frontend = nil
		c.Close()
	}()

	for msg := range cluster.LogChan {
		if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}
