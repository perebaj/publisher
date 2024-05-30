// Package scheduler provides a simple scheduler implementation to trigger a function every time.
package scheduler

import (
	"log/slog"
	"time"
)

// Scheduler gather all the necessary fields to run a scheduler
type Scheduler struct {
	// Ticker is a time.Ticker object
	Ticker *time.Ticker
	// Done is a channel to stop the scheduler
	Done chan struct{}
}

// NewScheduler creates a new scheduler
func NewScheduler(duration time.Duration) *Scheduler {
	return &Scheduler{
		Ticker: time.NewTicker(duration),
		Done:   make(chan struct{}),
	}
}

// Run starts the scheduler
func (s *Scheduler) Run(f func()) {
	go func() {
		for {
			select {
			case <-s.Done:
				slog.Info("The scheduler has been stopped")
				return
			case t := <-s.Ticker.C:
				f()
				slog.Info("Tick at", "time", t)
			}
		}
	}()
}
