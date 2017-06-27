package util

import "sync"

// ShutdownGuard facilitates coordinating the shutdown of multiple components.
type ShutdownGuard struct {
	sync.Mutex
	sync.WaitGroup
	ShuttingDown chan struct{}
}

// NewShutdownGuard creates a new ShutdownGuard.
func NewShutdownGuard() *ShutdownGuard {
	return &ShutdownGuard{
		ShuttingDown: make(chan struct{}),
	}
}

// InitiateShutdown signals to all components that they should begin shutting down.
func (s *ShutdownGuard) InitiateShutdown() {
	s.Lock()
	defer s.Unlock()

	select {
	case <-s.ShuttingDown:
		// already closed
	default:
		close(s.ShuttingDown)
	}
}

// ShutdownAndWait initiates a shutdown, and waits for all components to finish.
func (s *ShutdownGuard) ShutdownAndWait() {
	s.InitiateShutdown()
	s.Wait()
}

// ShutdownAndDone initiates a shutdown and signals completion.
func (s *ShutdownGuard) ShutdownAndDone() {
	s.InitiateShutdown()
	s.Done()
}
