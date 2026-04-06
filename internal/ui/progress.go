package ui

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// Spinner shows animated progress for long-running operations.
type Spinner struct {
	w     io.Writer
	label string
	done  chan struct{}
	once  sync.Once
}

var spinChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// NewSpinner creates a spinner writing to w with the given label.
func NewSpinner(w io.Writer, label string) *Spinner {
	return &Spinner{w: w, label: label, done: make(chan struct{})}
}

// Start begins spinning in a background goroutine.
func (s *Spinner) Start() {
	go func() {
		i := 0
		for {
			select {
			case <-s.done:
				fmt.Fprintf(s.w, "\r%s\r", strings.Repeat(" ", len(s.label)+4))
				return
			case <-time.After(80 * time.Millisecond):
				fmt.Fprintf(s.w, "\r  %s %s", spinChars[i%len(spinChars)], s.label)
				i++
			}
		}
	}()
}

// Stop halts the spinner and clears the line.
func (s *Spinner) Stop() {
	s.once.Do(func() { close(s.done) })
	time.Sleep(90 * time.Millisecond) // let goroutine clear line
}
