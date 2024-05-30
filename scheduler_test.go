package publisher

import (
	"testing"
	"time"
)

func TestScheduler_Run(t *testing.T) {
	s := NewScheduler(500 * time.Millisecond)

	go s.Run(func() {
		s.Done <- struct{}{}
	})

	select {
	case <-s.Done:
	case <-time.After(1 * time.Second):
		t.Error("The scheduler timed out")
	}
}
