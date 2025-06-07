package cmd

import (
	"context"
	"fmt"

	"xks/azure"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the AKS cluster",
	Long:  `Start the AKS cluster defined in your environment configuration.`,
	RunE:  runStartCommand,
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the AKS cluster",
	Long:  `Stop the AKS cluster defined in your environment configuration.`,
	RunE:  runStopCommand,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get AKS cluster status",
	Long:  `Get the current status of the AKS cluster defined in your environment configuration.`,
	RunE:  runStatusCommand,
}

func runStartCommand(cmd *cobra.Command, args []string) error {
	return runClusterCommand("start")
}

func runStopCommand(cmd *cobra.Command, args []string) error {
	return runClusterCommand("stop")
}

func runStatusCommand(cmd *cobra.Command, args []string) error {
	return runClusterCommand("status")
}

func runClusterCommand(action string) error {
	ctx := context.Background()

	// Configuration
	config, err := azure.NewConfig()
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	clusterManager := azure.NewClusterManager(config)

	// Authentification obligatoire
	authClient := azure.NewAuthClient(config)
	if !authClient.IsLoggedIn(ctx) {
		if err := authClient.Login(ctx); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}
	}

	// Exécution de l'action
	switch action {
	case "start":
		return clusterManager.Start(ctx, verbose)
	case "stop":
		return clusterManager.Stop(ctx, verbose)
	case "status":
		return clusterManager.Status(ctx, verbose)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statusCmd)
}
