// Package main from cmd/metrics/root.go is the CLI implementation to get metrics from GCP.
package main

import (
	"log/slog"

	"github.com/perebaj/publisher"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "metrics",
	Long:    "CLI to get pubsub metrics from GCP",
	Aliases: []string{"m"},
	Example: "metrics --projectID=jojo-is-awesome-project --subscriptionID=jojoisawesome.push.dlq.pull",
	Run: func(cmd *cobra.Command, args []string) {
		subscriptionID := cmd.Flag("subscriptionID").Value.String()
		projectID := cmd.Flag("projectID").Value.String()
		m := publisher.NewMetrics(projectID, subscriptionID)
		undeliveredMessMean, err := m.NumUndeliveredMessagesMean()
		if err != nil {
			slog.Error("Error getting the number of undelivered messages", "error", err)
		}

		slog.Info("The number of undelivered messages is", "undelivedMessagesMean", *undeliveredMessMean, "projectID", projectID, "subscriptionID", subscriptionID)
	},
}

func init() {
	rootCmd.Flags().String("subscriptionID", "", "The subscription ID")
	rootCmd.Flags().String("projectID", "", "The project ID")
}

func execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Error executing the CLI", "error", err)
	}
}

func main() {
	execute()
}
