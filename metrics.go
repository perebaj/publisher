package publisher

import (
	"context"
	"fmt"
	"strings"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	monitoringpb "cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	pubsub "cloud.google.com/go/pubsub/apiv1"
	pubsubpb "cloud.google.com/go/pubsub/apiv1/pubsubpb"
	"google.golang.org/api/iterator"
)

// Metrics gather all the necessary fields to get metrics from GCP
type Metrics struct {
	// Project is the GCP project ID
	Project string
	// Subscription is the GCP Pub/Sub subscription name
	Subscription string
}

// NewMetrics creates a new metrics
func NewMetrics(project, subscription string) *Metrics {
	return &Metrics{
		Project:      project,
		Subscription: subscription,
	}
}

// ListDLQSubscriptions returns a list of all the dead-letter-queue subscriptions
func (m *Metrics) ListDLQSubscriptions() ([]string, error) {
	ctx := context.Background()
	c, err := pubsub.NewSubscriberClient(ctx)
	if err != nil {
		return nil, err
	}

	req := &pubsubpb.ListSubscriptionsRequest{
		Project: fmt.Sprintf("projects/%s", m.Project),
	}
	it := c.ListSubscriptions(ctx, req)

	var subscriptions []string
	for {
		sub, err := it.Next()
		if err == iterator.Done {
			break
		}

		if strings.Contains(sub.Name, ".push.dlq.pull") {
			// The sub.Name comes in the format projects/{projectID}/subscriptions/{subscriptionID}
			// We need to extract the subscriptionID
			splitedSubs := strings.Split(sub.Name, "/")
			subscriptions = append(subscriptions, splitedSubs[len(splitedSubs)-1])
		}

		if err != nil {
			return nil, err
		}
	}

	return subscriptions, nil
}

// NumUndeliveredMessagesMean returns the number of undelivered messages in the subscription
func (m *Metrics) NumUndeliveredMessagesMean() (*float64, error) {
	ctx := context.Background()
	c, err := monitoring.NewQueryClient(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = c.Close()
	}()

	req := &monitoringpb.QueryTimeSeriesRequest{
		Name: fmt.Sprintf("projects/%s", m.Project), // optional
		Query: fmt.Sprintf(`fetch pubsub_subscription
		| metric 'pubsub.googleapis.com/subscription/num_undelivered_messages'
		| filter (resource.subscription_id == '%s')
		| group_by 1m,
			[value_num_undelivered_messages_mean:
			   mean(value.num_undelivered_messages)]
		| within 5m`, m.Subscription),
	}
	var numUndeliveredMessagesMean = 0.0
	it := c.QueryTimeSeries(ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		numUndeliveredMessagesMean = resp.GetPointData()[0].GetValues()[0].GetDoubleValue()
	}

	return &numUndeliveredMessagesMean, nil
}
