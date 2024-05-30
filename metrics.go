package publisher

import (
	"context"
	"fmt"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	monitoringpb "cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
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
