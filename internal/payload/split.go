package payload

import "fmt"

// Split divides data into chunks of at most chunkSize bytes.
func Split(data []byte, chunkSize int) ([][]byte, error) {
	if chunkSize <= 0 {
		return nil, fmt.Errorf("payload: split: chunkSize must be > 0")
	}
	if len(data) == 0 {
		return nil, nil
	}
	var chunks [][]byte
	for len(data) > 0 {
		end := chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[:end])
		data = data[end:]
	}
	return chunks, nil
}
