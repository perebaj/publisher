// Package main cmd/pool/main.go is the entry point of the pool application.
// The pool application is the main application that will be used to process the metrics of the subscriptions in the DLQ.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/perebaj/publisher"
)

// subMetric is a struct that contains the metrics of a subscription
type subMetric struct {
	// projectID is the project where the subscription is located
	projectID string
	// susbcriptionID is the ID of the subscription
	susbcriptionID string
	// nUndeliveredMessagesMean is the number of undelivered messages in the subscription
	// This value is calculated based on the mean of the number of undelivered messages in the last 5 minutes.
	nUndeliveredMessagesMean float64
}

func main() {
	ctx := context.Background()
	projectID := "registry-contracts-prd-001"

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo, // TODO(@perebaj) create a structure to handle the log level as an environment variable
		AddSource: true,
	}))
	slog.SetDefault(logger)

	DLQSubs, err := publisher.ListDLQSubscriptions(ctx, projectID)
	if len(DLQSubs) == 0 {
		slog.Info("no DLQ subscriptions found")
		return
	}

	slog.Info("DLQ subscriptions found", "subscriptions", DLQSubs, "total", len(DLQSubs))

	if err != nil {
		slog.Error("error listing DLQ subscriptions", "error", err)
		return
	}

	inputCh := make(chan string)
	// Populate the input channel with the DLQ subscriptions
	go func() {
		for _, sub := range DLQSubs {
			inputCh <- sub
		}

		close(inputCh)
	}()

	errorCh := make(chan error)
	// Initialize the workers to process the metrics of the subscriptions
	const nWorkers = 2
	metricsOutputCh := make(chan subMetric)
	var wg sync.WaitGroup
	for i := 0; i < nWorkers; i++ {
		go func(i int) {
			err = dlqWorkerMetrics(ctx, i, inputCh, metricsOutputCh, &wg, projectID)
			if err != nil {
				slog.Error("error processing metrics", "error", err)
				errorCh <- err
			}
		}(i)
	}

	wg.Add(nWorkers)

	// Waiting for the workers to finish and close the output channel to avoid deadlocks
	go func() {
		wg.Wait()
		close(metricsOutputCh)
	}()

	// Reprocess all DLQ subscriptions that have more than 0 undelivered messages to the default topic
	for _ = range metricsOutputCh {
		// if metric.nUndeliveredMessagesMean > 0 {
		// 	slog.Info("subscription with undelivered messages", "subscription", metric.susbcriptionID, "projectID", metric.projectID, "nUndeliveredMessagesMean", metric.nUndeliveredMessagesMean)

		// 	go func() {
		// 		err := publisher.Reprocess(ctx, metric.projectID, metric.susbcriptionID)
		// 		if err != nil {
		// 			slog.Error("error reprocessing DLQ", "error", err)
		// 			errorCh <- err
		// 		}
		// 	}()
		// }
	}

	<-errorCh
}

// dlqWorkerMetrics is a worker that processes the metrics of the subscriptions and sends the results to the output channel
// enabling the main goroutine to process the results concurrently
func dlqWorkerMetrics(ctx context.Context, id int, inputCh <-chan string, outputCh chan<- subMetric, wg *sync.WaitGroup, projectID string) error {
	for sub := range inputCh {
		slog.Info(fmt.Sprintf("worker with id: %d processing the value: %v", id, sub), "projectID", projectID)
		nUndeliveredMess, err := publisher.NumUndeliveredMessagesMean(ctx, projectID, sub)
		if err != nil {
			return fmt.Errorf("error getting number of undelivered messages: %v", err)
		}

		outputCh <- subMetric{
			projectID:                projectID,
			susbcriptionID:           sub,
			nUndeliveredMessagesMean: *nUndeliveredMess,
		}
	}

	wg.Done()
	return nil
}
