// Package listener implements the callback server used to confirm payload receipt.
package listener

import "context"

// ReceivedPayload records a payload received by the server.
type ReceivedPayload struct {
	Channel   string
	Data      []byte
	SourceIP  string
	Timestamp int64
}

// Listener is the interface for callback servers.
type Listener interface {
	// Start begins listening. Blocks until ctx is cancelled.
	Start(ctx context.Context) error
	// Received returns a channel that emits confirmed payloads.
	Received() <-chan ReceivedPayload
}
