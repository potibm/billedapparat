package hub

import (
	"log/slog"
	"sync"
)

type SSEMessage struct {
	Event   string      `json:"event"`
	Payload interface{} `json:"payload"`
}

type Streamer struct {
	clients map[chan SSEMessage]bool
	mu      sync.RWMutex
}

func NewStreamer() *Streamer {
	return &Streamer{
		clients: make(map[chan SSEMessage]bool),
	}
}

func (s *Streamer) Broadcast(event string, payload interface{}) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	msg := SSEMessage{
		Event:   event,
		Payload: payload,
	}

	for ch := range s.clients {
		select {
		case ch <- msg:
		default:
			slog.Warn("[SSE] Client channel full, dropping message")
		}
	}
}

func (s *Streamer) addClient() chan SSEMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	const clientChanBufferSize = 10

	ch := make(chan SSEMessage, clientChanBufferSize)
	s.clients[ch] = true
	slog.Info("[SSE] Client connected", "active_clients", len(s.clients))

	return ch
}

func (s *Streamer) removeClient(ch chan SSEMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.clients[ch]; ok {
		delete(s.clients, ch)
		close(ch)
		slog.Info("[SSE] Client disconnected", "active_clients", len(s.clients))
	}
}
