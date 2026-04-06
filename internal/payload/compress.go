package payload

import (
	"bytes"
	"compress/gzip"
	"fmt"
)

// Compress compresses data with gzip (BestSpeed).
func Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		return nil, fmt.Errorf("payload: compress: %w", err)
	}
	if _, err = w.Write(data); err != nil {
		return nil, fmt.Errorf("payload: compress write: %w", err)
	}
	if err = w.Close(); err != nil {
		return nil, fmt.Errorf("payload: compress close: %w", err)
	}
	return buf.Bytes(), nil
}
