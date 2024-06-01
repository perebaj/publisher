// Package publisher pubsub.go is the implementation of the function that deal with pubsub messages.
package publisher

import (
	"context"
	"fmt"
	"log/slog"

	"gocloud.dev/pubsub"
)

// Subscription represents a subscription to a pubsub topic.
type Subscription struct {
	// sub is the subscription object that will be used to receive messages.
	sub *pubsub.Subscription
	// messages2Pull is the number of messages to pull from the subscription.
	messages2Pull int
}

// NewSubscription creates a new Subscription.
func NewSubscription(subscription *pubsub.Subscription, messages2Pull int) *Subscription {
	return &Subscription{sub: subscription,
		messages2Pull: messages2Pull}
}

// Handler is a generic function to handle messages.
type Handler func(msg *pubsub.Message) error

// Receive starts receiving messages from the subscription.
func (s *Subscription) Receive(ctx context.Context, handler Handler) error {
	for i := 0; i < s.messages2Pull; i++ {

		msg, err := s.sub.Receive(ctx)
		if err != nil {
			return fmt.Errorf("error receiving message: %v", err)
		}

		go func() {
			defer msg.Ack()
			err := handler(msg)
			if err != nil {
				slog.Error("error handling message", "error", err)
				if msg.Nackable() {
					msg.Nack()
				}
				return
			}
		}()
	}

	return nil
}

