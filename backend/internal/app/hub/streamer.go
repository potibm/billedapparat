package hub

import (
	"log/slog"
	"sync"
)

type StreamEvent string

const (
	EventCreate StreamEvent = "CREATE"
	EventUpdate StreamEvent = "UPDATE"
	EventDelete StreamEvent = "DELETE"
)

type SSEMessage struct {
	Event   string      `json:"event"`
	Payload interface{} `json:"payload"`
}

type Streamer struct {
	clients map[chan SSEMessage]bool
	mu      sync.RWMutex
	logger  *slog.Logger
}

func NewStreamer(logger *slog.Logger) *Streamer {
	return &Streamer{
		clients: make(map[chan SSEMessage]bool),
		logger:  logger,
	}
}

func (s *Streamer) Broadcast(event StreamEvent, payload interface{}) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	msg := SSEMessage{
		Event:   string(event),
		Payload: payload,
	}

	for ch := range s.clients {
		select {
		case ch <- msg:
		default:
			s.logger.Warn("Client channel full, dropping message")
		}
	}
}

func (s *Streamer) addClient() chan SSEMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	const clientChanBufferSize = 10

	ch := make(chan SSEMessage, clientChanBufferSize)
	s.clients[ch] = true
	s.logger.Info("Client connected", "active_clients", len(s.clients))

	return ch
}

func (s *Streamer) removeClient(ch chan SSEMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.clients[ch]; ok {
		delete(s.clients, ch)
		close(ch)
		s.logger.Info("Client disconnected", "active_clients", len(s.clients))
	}
}
