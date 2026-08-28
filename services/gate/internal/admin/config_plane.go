package admin

import (
	"context"
	"sync"
	"time"
)

// configWatcherState holds config watcher state.
type configWatcherState struct {
	mu       sync.RWMutex
	lastHash string
	lastLoad time.Time
	stopCh   chan struct{}
}

// startConfigPlane starts the config control plane
// (hot-reload, watcher, apply).
func startConfigPlane(ctx context.Context, onChange func() error) *configWatcherState {
	s := &configWatcherState{
		stopCh: make(chan struct{}),
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-s.stopCh:
				return
			case <-ticker.C:
				if err := onChange(); err != nil {
					// log error
					continue
				}
				s.mu.Lock()
				s.lastLoad = time.Now()
				s.mu.Unlock()
			}
		}
	}()

	return s
}

func (s *configWatcherState) Stop() {
	close(s.stopCh)
}
