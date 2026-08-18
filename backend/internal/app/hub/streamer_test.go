package hub

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/potibm/billedapparat/internal/app/domain"
	"github.com/stretchr/testify/assert"
)

func newTestStreamer() *Streamer {
	return NewStreamer(slog.New(slog.DiscardHandler))
}

func TestStreamer_AddClient(t *testing.T) {
	s := newTestStreamer()

	assert.Equal(t, 0, len(s.clients))

	ch := s.addClient()

	assert.NotNil(t, ch)
	assert.Equal(t, 1, len(s.clients))
}

func TestStreamer_RemoveClient(t *testing.T) {
	s := newTestStreamer()

	ch := s.addClient()
	assert.Equal(t, 1, len(s.clients))

	s.removeClient(ch)

	assert.Equal(t, 0, len(s.clients))
}

func TestStreamer_RemoveUnknownClient(t *testing.T) {
	s := newTestStreamer()

	ch := make(chan SSEMessage)
	s.removeClient(ch) // should not panic

	assert.Equal(t, 0, len(s.clients))
}

func TestStreamer_BroadcastNoClients(t *testing.T) {
	s := newTestStreamer()

	// Should not panic with no clients
	s.Broadcast(domain.EventCreate, "test")
}

func TestStreamer_BroadcastOneClient(t *testing.T) {
	s := newTestStreamer()

	ch := s.addClient()

	s.Broadcast(domain.EventCreate, "hello")

	select {
	case msg := <-ch:
		assert.Equal(t, string(domain.EventCreate), msg.Event)
		assert.Equal(t, "hello", msg.Payload)
	default:
		t.Fatal("expected message on channel")
	}
}

func TestStreamer_BroadcastMultipleClients(t *testing.T) {
	s := newTestStreamer()

	ch1 := s.addClient()
	ch2 := s.addClient()
	ch3 := s.addClient()

	s.Broadcast(domain.EventUpdate, "update")

	for i, ch := range []chan SSEMessage{ch1, ch2, ch3} {
		select {
		case msg := <-ch:
			assert.Equal(t, string(domain.EventUpdate), msg.Event, "client %d", i)
			assert.Equal(t, "update", msg.Payload, "client %d", i)
		default:
			t.Fatalf("client %d did not receive message", i)
		}
	}
}

func TestStreamer_BroadcastFullChannel(t *testing.T) {
	s := newTestStreamer()

	ch := s.addClient()

	// Fill the channel to capacity (buffer size is 10)
	for i := 0; i < 10; i++ {
		s.Broadcast(domain.EventCreate, i)
	}

	// Channel should be full now; next broadcast should be dropped, not deadlock
	s.Broadcast(domain.EventCreate, "overflow")

	// Drain channel - should have 10 messages
	count := 0

	done := false
	for !done {
		select {
		case <-ch:
			count++
		default:
			done = true
		}
	}

	assert.Equal(t, 10, count, "only first 10 messages should be buffered, overflow should be dropped")
}

func TestStreamer_ConcurrentOperations(t *testing.T) {
	s := newTestStreamer()

	var wg sync.WaitGroup

	numGoroutines := 20

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ch := s.addClient()
			s.Broadcast(domain.EventCreate, "msg")
			s.removeClient(ch)
		}()
	}

	wg.Wait()
	// No race detector errors expected
}

func TestStreamer_RunPingLoop_BroadcastsOnTick(t *testing.T) {
	// Drives runPingLoop with a real ticker at a short interval and asserts
	// that at least one PING event is broadcast to a connected client.
	s := newTestStreamer()

	ch := s.addClient()

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)

	done := make(chan struct{})

	go func() {
		s.runPingLoop(ctx, ticker)
		close(done)
	}()

	select {
	case msg := <-ch:
		assert.Equal(t, string(domain.EventPing), msg.Event)
	case <-time.After(time.Second):
		t.Fatal("expected PING broadcast within 1s")
	}

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("runPingLoop did not exit on ctx.Done()")
	}
}

func TestStreamer_RunPingLoop_ExitsOnCtxCancel(t *testing.T) {
	// Even when the ticker never fires, runPingLoop must exit promptly on
	// ctx cancellation (no goroutine leak).
	s := newTestStreamer()

	ctx, cancel := context.WithCancel(context.Background())

	ticker := time.NewTicker(time.Hour) // intentionally far in the future

	done := make(chan struct{})

	go func() {
		s.runPingLoop(ctx, ticker)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("runPingLoop did not exit on ctx.Done()")
	}
}
