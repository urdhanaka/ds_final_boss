package services

import (
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type ClusterStream struct {
	LogChan     chan []byte
	Frontend    *websocket.Conn
	Writer      *websocket.Conn
	WriterAlive bool
}

type StreamManager struct {
	mu       sync.Mutex
	clusters map[string]*ClusterStream
}

func NewStreamManager() *StreamManager {
	return &StreamManager{
		clusters: make(map[string]*ClusterStream),
	}
}

func (sm *StreamManager) GetOrCreateCluster(name string) *ClusterStream {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if c, ok := sm.clusters[name]; ok {
		return c
	}

	c := &ClusterStream{
		LogChan: make(chan []byte, 256),
	}
	sm.clusters[name] = c

	return c
}

func (sm *StreamManager) SetWriter(
	clusterName string,
	conn *websocket.Conn,
) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	c := sm.clusters[clusterName]
	if !c.WriterAlive {
		c.Writer = conn
		c.WriterAlive = true

		return true
	}

	return false
}

func (sm *StreamManager) RemoveWriter(clusterName string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	c := sm.clusters[clusterName]
	c.WriterAlive = false
	c.Writer = nil
}
