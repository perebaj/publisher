// Package main from cmd/metrics/root.go is the CLI implementation to get metrics from GCP.
package main

import (
	"context"
	"log/slog"
	"sort"

	"github.com/perebaj/publisher"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "metrics",
	Long:    "CLI to get pubsub metrics from GCP",
	Aliases: []string{"m"},
	Example: "metrics --projectID=jojo-is-awesome-project --subscriptionID=jojoisawesome.push.dlq.pull",
	Run: func(cmd *cobra.Command, _ []string) {
		subscriptionID := cmd.Flag("subscriptionID").Value.String()
		projectID := cmd.Flag("projectID").Value.String()
		ctx := context.Background()
		undeliveredMessMean, err := publisher.NumUndeliveredMessagesMean(ctx, projectID, subscriptionID)
		if err != nil {
			slog.Error("Error getting the number of undelivered messages", "error", err)
		}

		slog.Info("The number of undelivered messages is", "undelivedMessagesMean", *undeliveredMessMean, "projectID", projectID, "subscriptionID", subscriptionID)
	},
}

var list = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "List all the dead-letter-queue subscriptions",
	Example: "metrics list --projectID=jojo-is-awesome-project",
	Run: func(cmd *cobra.Command, _ []string) {
		projectID := cmd.Flag("projectID").Value.String()
		ctx := context.Background()
		subscriptions, err := publisher.ListDLQSubscriptions(ctx, projectID)
		if err != nil {
			slog.Error("Error listing the DLQ subscriptions", "error", err)
			return
		}

		for _, sub := range subscriptions {
			slog.Info("DLQ subscription", "subscription", sub)
		}
	},
}

var metricsAllDLQSubscriptionsCmd = &cobra.Command{
	Use:     "metrics-dlq",
	Aliases: []string{"mdlq"},
	Short:   "Show the number of undelivered messages for the top 5 dead-letter-queue subscriptions",
	Example: "metrics metrics-dlq --projectID=jojo-is-awesome-project --numMostOffenders=5",
	Run: func(cmd *cobra.Command, _ []string) {
		projectID := cmd.Flag("projectID").Value.String()
		numMostOffenders, err := cmd.Flags().GetInt("numMostOffenders")
		if err != nil {
			slog.Error("Error getting the number of most offenders", "error", err)
			return
		}

		if numMostOffenders < 0 {
			slog.Error("The number of most offenders must be greater than 0")
			return
		}
		ctx := context.Background()
		subscriptions, err := publisher.ListDLQSubscriptions(ctx, projectID)
		if err != nil {
			slog.Error("Error listing the DLQ subscriptions", "error", err)
			return
		}

		slog.Info("DLQ subscriptions", "len", len(subscriptions))

		type SubMetric struct {
			Subscription          string
			UndeliveredMessagMean float64
		}

		var subMetrics []SubMetric

		//TODO(@perebaj): Create worker pool to process the subscriptions concurrently
		bar := progressbar.Default(int64(len(subscriptions)))
		ctx = context.Background()
		for _, sub := range subscriptions {
			undeliveredMessMean, err := publisher.NumUndeliveredMessagesMean(ctx, projectID, sub)
			if err != nil {
				slog.Error("Error getting the number of undelivered messages", "error", err)
			}

			subMetrics = append(subMetrics, SubMetric{
				Subscription:          sub,
				UndeliveredMessagMean: *undeliveredMessMean,
			})
			err = bar.Add(1)
			if err != nil {
				slog.Error("Error adding a new tick to the progress bar", "error", err)
			}
		}

		sort.Slice(subMetrics, func(i, j int) bool {
			return subMetrics[i].UndeliveredMessagMean > subMetrics[j].UndeliveredMessagMean
		})

		for _, subMetric := range subMetrics[:numMostOffenders] {
			slog.Info("DLQ subscription", "subscription", subMetric.Subscription, "undeliveredMessagesMean", subMetric.UndeliveredMessagMean)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringP("projectID", "p", "", "The project ID")
	rootCmd.Flags().StringP("subscriptionID", "s", "", "The subscription ID")
	rootCmd.AddCommand(list)
	rootCmd.AddCommand(metricsAllDLQSubscriptionsCmd)
	metricsAllDLQSubscriptionsCmd.Flags().IntP("numMostOffenders", "n", 5, "The number of most offenders")
}

func execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Error executing the CLI", "error", err)
	}
}

func main() {
	execute()
}
