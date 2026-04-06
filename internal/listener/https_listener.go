package listener

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPSListener listens for incoming POST callbacks on an HTTPS/HTTP address.
type HTTPSListener struct {
	Addr     string
	received chan ReceivedPayload
}

// NewHTTPSListener creates an HTTPS callback listener on the given address (e.g. ":8443").
func NewHTTPSListener(addr string) *HTTPSListener {
	return &HTTPSListener{
		Addr:     addr,
		received: make(chan ReceivedPayload, 64),
	}
}

// Received returns the channel of received payloads.
func (l *HTTPSListener) Received() <-chan ReceivedPayload {
	return l.received
}

// Start begins the HTTP server and blocks until ctx is cancelled.
func (l *HTTPSListener) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		data, err := io.ReadAll(io.LimitReader(r.Body, 10*1024*1024))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		l.received <- ReceivedPayload{
			Channel:   r.Header.Get("X-Channel"),
			Data:      data,
			SourceIP:  r.RemoteAddr,
			Timestamp: time.Now().Unix(),
		}
		// Echo back the correlation ID so the channel can confirm receipt.
		if corrID := r.Header.Get("X-Correlation-ID"); corrID != "" {
			w.Header().Set("X-Correlation-ID", corrID)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	srv := &http.Server{
		Addr:         l.Addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("https listener: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutCtx)
	case err := <-errCh:
		return err
	}
}
