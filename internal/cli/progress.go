package cli

import (
	"fmt"
	"sync"
	"time"
)

// Spinner shows a lightweight progress indicator.
type Spinner struct {
	message string
	frames  []string
	stop    chan struct{}
	wg      sync.WaitGroup
}

// NewSpinner creates a spinner with ASCII-safe frames.
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		frames:  []string{"-", "\\", "|", "/"},
		stop:    make(chan struct{}),
	}
}

// Start begins the spinner animation.
func (s *Spinner) Start() {
	if !IsColorEnabled() {
		fmt.Printf("%s...\n", s.message)
		return
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		index := 0
		for {
			select {
			case <-s.stop:
				fmt.Print("\r\033[K")
				return
			default:
				fmt.Printf("\r%s %s", colorize(colorBlue, s.frames[index%len(s.frames)]), s.message)
				index++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

// Stop ends the spinner animation.
func (s *Spinner) Stop() {
	select {
	case <-s.stop:
	default:
		close(s.stop)
	}
	s.wg.Wait()
}

// StopWithSuccess ends the spinner and prints a success message.
func (s *Spinner) StopWithSuccess(message string) {
	s.Stop()
	Success("%s", message)
}

// StopWithError ends the spinner and prints an error message.
func (s *Spinner) StopWithError(message string) {
	s.Stop()
	Errorf("%s", message)
}
