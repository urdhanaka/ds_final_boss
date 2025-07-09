package services

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type ClusterLogsStream struct {
	LogChan     chan []byte
	Reader      *websocket.Conn
	Writer      *websocket.Conn
	WriterAlive bool
}

type StreamManager struct {
	mu          sync.Mutex
	clusterLogs map[string]*ClusterLogsStream
}

func NewStreamManager() *StreamManager {
	return &StreamManager{
		clusterLogs: make(map[string]*ClusterLogsStream),
	}
}

func (s *StreamManager) GetOrCreateClusterLogs(name string) *ClusterLogsStream {
	s.mu.Lock()
	defer s.mu.Unlock()

	if c, ok := s.clusterLogs[name]; ok {
		return c
	}

	c := &ClusterLogsStream{
		LogChan: make(chan []byte, 256),
	}
	s.clusterLogs[name] = c

	return c
}

func (s *StreamManager) SetWriter(
	name string,
	conn *websocket.Conn,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c := s.clusterLogs[name]
	if !c.WriterAlive {
		c.Writer = conn
		c.WriterAlive = true
	}

	return nil
}

func (s *StreamManager) RemoveWriter(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if c, ok := s.clusterLogs[name]; ok {
		c.WriterAlive = false
		c.Writer = nil
	}

	return nil
}

func (s *StreamManager) SetReader(
	name string,
	conn *websocket.Conn,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.clusterLogs[name]
	if !ok {
		return nil
	}

	if c.Reader != nil {
		return fmt.Errorf("reader exists already")
	}

	c.Reader = conn

	return nil
}

func (s *StreamManager) RemoveReader(
	name string,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.clusterLogs[name]

	if !ok {
		return nil
	}

	c.Reader = nil

	return nil
}

func (s *StreamManager) CleanClusterLogs(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if c, ok := s.clusterLogs[name]; ok {
		if c.WriterAlive {
			return fmt.Errorf("writer is still alive")
		}
	}

	delete(s.clusterLogs, name)

    return nil
}
