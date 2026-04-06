// Package payload generates, encrypts, compresses, and chunks synthetic payloads.
package payload

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
)

// Generator produces payloads for channel runs.
type Generator struct {
	SizeBytes int
	FilePath  string // optional — read from file instead of random
}

// Generate returns a payload byte slice.
// If FilePath is set, reads from that file (capped at SizeBytes if > 0).
// Otherwise generates SizeBytes of cryptographically random data.
func (g *Generator) Generate() ([]byte, error) {
	if g.FilePath != "" {
		return g.fromFile()
	}
	return g.random()
}

func (g *Generator) random() ([]byte, error) {
	size := g.SizeBytes
	if size <= 0 {
		size = 1024
	}
	buf := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return nil, fmt.Errorf("payload: generate random: %w", err)
	}
	return buf, nil
}

func (g *Generator) fromFile() ([]byte, error) {
	data, err := os.ReadFile(g.FilePath)
	if err != nil {
		return nil, fmt.Errorf("payload: read file %q: %w", g.FilePath, err)
	}
	if g.SizeBytes > 0 && len(data) > g.SizeBytes {
		data = data[:g.SizeBytes]
	}
	return data, nil
}
