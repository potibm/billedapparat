package hub

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/potibm/billedapparat/internal/app/domain"
)

// defaultPingInterval defines how often the server sends a PING event to keep the SSE connection alive.
// NOTE: The frontend's WATCHDOG_INTERVAL is tightly coupled to this and should remain at >= 2x this value.
const defaultPingInterval = 10 * time.Second

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

func (s *Streamer) StartPingLoop(ctx context.Context) {
	ticker := time.NewTicker(defaultPingInterval)
	defer ticker.Stop()

	s.logger.Info("Starting SSE ping loop", "interval", defaultPingInterval)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping SSE ping loop")

			return
		case <-ticker.C:
			s.Broadcast(domain.EventPing, struct{}{})
		}
	}
}

func (s *Streamer) Broadcast(event domain.StreamEvent, payload interface{}) {
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
