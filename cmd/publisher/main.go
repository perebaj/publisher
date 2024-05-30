// Package main is the entry point of the publisher application.
package main

import (
	"log/slog"
	"time"

	scheduler "github.com/perebaj/publisher"
)

func main() {
	// Create a new scheduler
	s := scheduler.NewScheduler(1 * time.Second)

	go s.Run(func() {
		slog.Info("Hello, world!")
		s.Done <- struct{}{}
	})

	<-s.Done
}
