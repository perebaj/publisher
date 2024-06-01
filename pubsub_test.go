package publisher

import (
	"context"
	"testing"

	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/mempubsub"
)

func TestSubscription_Receive(t *testing.T) {
	ctx := context.Background()
	topicName := "mem://test-topic"
	topic, err := pubsub.OpenTopic(ctx, topicName)
	if err != nil {
		t.Fatalf("error opening topic: %v", err)
	}

	t.Cleanup(func() {
		_ = topic.Shutdown(ctx)
	})

	msgCh := make(chan *pubsub.Message)
	handler := func(msg *pubsub.Message) error {
		msgCh <- msg
		return nil
	}

	subs, err := pubsub.OpenSubscription(ctx, topicName)
	if err != nil {
		t.Fatalf("error opening subscription: %v", err)
	}

	s := NewSubscription(subs, 1)

	go func() {
		err := s.Receive(ctx, handler)
		if err != nil {
			t.Errorf("error receiving message: %v", err)
		}
	}()

	err = topic.Send(ctx, &pubsub.Message{
		Body: []byte("test message"),
	})

	if err != nil {
		t.Errorf("error sending message: %v", err)
	}

	if msg := <-msgCh; string(msg.Body) != "test message" {
		t.Errorf("expected message body to be 'test message', got %s", msg.Body)
	}
}
