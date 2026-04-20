package hub

import (
	"io"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func (s *Server) streamSlides(c *gin.Context) {
	// 1. Set HTTP headers for SSE (Server-Sent Events)
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	// 2. Connect client to streamer
	clientChan := s.streamer.addClient()
	defer s.streamer.removeClient(clientChan)

	// 3. INITIAL LOAD: Once the client connects fetch active slides from DB and send as INIT-Event
	activeSlides, err := s.slideRepo.GetActive(c.Request.Context())
	if err != nil {
		slog.Error("[SSE] Failed to fetch initial slides", "error", err)
	} else {
		c.SSEvent("INIT", activeSlides)
		c.Writer.Flush()
	}

	// 4. Wait for updates from the streamer
	clientGone := c.Request.Context().Done()

	// 5. c.Stream blocks and keeps the connection open
	c.Stream(func(w io.Writer) bool {
		select {
		case <-clientGone:
			// browser closed connection, clean up and stop streaming
			return false
		case msg, ok := <-clientChan:
			if !ok {
				return false
			}

			c.SSEvent(msg.Event, msg.Payload)

			return true
		}
	})
}
