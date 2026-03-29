package worker

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

// SSEHandler serves SSE progress events for batch generation.
type SSEHandler struct {
	rdb *redis.Client
}

// NewSSEHandler creates a new SSE handler.
func NewSSEHandler(redisURL string) (*SSEHandler, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	rdb := redis.NewClient(opt)
	return &SSEHandler{rdb: rdb}, nil
}

// ServeProgress handles GET /api/v1/staff/batches/{id}/progress
func (h *SSEHandler) ServeProgress(w http.ResponseWriter, r *http.Request) {
	batchID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid batch id", http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	channel := progressChannel(batchID)
	ctx := r.Context()

	sub := h.rdb.Subscribe(ctx, channel)
	defer sub.Close()

	ch := sub.Channel()

	// Send initial keepalive
	fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", msg.Payload)
			flusher.Flush()

			// Check if this is a completion event
			if isCompletionEvent(msg.Payload) {
				return
			}
		}
	}
}

func isCompletionEvent(payload string) bool {
	// Quick check for "complete" status
	return len(payload) > 0 &&
		(contains(payload, `"status":"complete"`) ||
			contains(payload, `"status": "complete"`))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// Close cleans up Redis connection.
func (h *SSEHandler) Close() error {
	return h.rdb.Close()
}

// LogInfo logs that SSE handler is ready.
func (h *SSEHandler) LogInfo() {
	slog.Info("SSE handler ready for batch progress streaming")
}
