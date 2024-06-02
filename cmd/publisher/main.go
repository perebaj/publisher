// Package main is the entry point of the publisher application.
package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/perebaj/publisher"
	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/gcppubsub" // gcp driver. (for more drivers see: https://gocloud.dev/howto/pubsub/subscribe/#gcp)
)

func main() {

	messages2Process := 1
	slog.Info("Starting publisher", "messages2Process", messages2Process)
	project := "registry-receivables-prd"
	subscriptionID := "jojo-tmp-sub"
	// Create a new publisher
	subName := fmt.Sprintf("gcppubsub://projects/%s/subscriptions/%s", project, subscriptionID)
	ctx := context.Background()

	// Open a connection to the subscription
	subscription, err := pubsub.OpenSubscription(ctx, subName)
	if err != nil {
		slog.Error("error opening subscription", "error", err)
		return
	}

	publisher := publisher.NewSubscription(subscription, messages2Process)

	// Define the handler function
	handler := func(msg *pubsub.Message) error {
		slog.Info("message received", "message", string(msg.Body))
		return nil
	}
	err = publisher.Receive(ctx, handler)
	if err != nil {
		slog.Error("error receiving messages", "error", err)
		return
	}
	//select using context to keep the main function alive
}
