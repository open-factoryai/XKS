package azure

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ClusterManager gère les opérations sur le cluster AKS
type ClusterManager struct {
	config *Config
}

// NewClusterManager crée une nouvelle instance de ClusterManager
func NewClusterManager(config *Config) *ClusterManager {
	return &ClusterManager{config: config}
}

func (c *ClusterManager) Start(ctx context.Context, verbose bool) error {
	// Vérifier d'abord si le cluster est déjà démarré
	status, err := c.getClusterStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to check cluster status: %w", err)
	}

	if status == "Running" {
		if verbose {
			fmt.Printf("✅ AKS cluster '%s' is already running\n", c.config.AKSName)
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	if verbose {
		fmt.Printf("🚀 Starting AKS cluster '%s' in resource group '%s'...\n", c.config.AKSName, c.config.ResourceName)
	}

	cmd := exec.CommandContext(ctx, "az", "aks", "start",
		"--name", c.config.AKSName,
		"--resource-group", c.config.ResourceName,
		"--subscription", c.config.Subscription,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) > 0 {
			fmt.Fprintf(os.Stderr, "Azure CLI output: %s\n", string(output))
		}
		return fmt.Errorf("failed to start AKS cluster: %w", err)
	}

	if verbose {
		fmt.Println("✅ AKS cluster started successfully")
		if len(output) > 0 {
			fmt.Printf("Output: %s\n", string(output))
		}
	}
	return nil
}

func (c *ClusterManager) Stop(ctx context.Context, verbose bool) error {
	// Vérifier d'abord si le cluster est déjà arrêté
	status, err := c.getClusterStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to check cluster status: %w", err)
	}

	if status == "Stopped" || status == "Deallocated" {
		if verbose {
			fmt.Printf("✅ AKS cluster '%s' is already stopped\n", c.config.AKSName)
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	if verbose {
		fmt.Printf("🛑 Stopping AKS cluster '%s' in resource group '%s'...\n", c.config.AKSName, c.config.ResourceName)
	}

	cmd := exec.CommandContext(ctx, "az", "aks", "stop",
		"--name", c.config.AKSName,
		"--resource-group", c.config.ResourceName,
		"--subscription", c.config.Subscription,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) > 0 {
			fmt.Fprintf(os.Stderr, "Azure CLI output: %s\n", string(output))
		}
		return fmt.Errorf("failed to stop AKS cluster: %w", err)
	}

	if verbose {
		fmt.Println("✅ AKS cluster stopped successfully")
		if len(output) > 0 {
			fmt.Printf("Output: %s\n", string(output))
		}
	}
	return nil
}

// Status affiche le statut actuel du cluster
func (c *ClusterManager) Status(ctx context.Context, verbose bool) error {
	status, err := c.getClusterStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to get cluster status: %w", err)
	}

	if verbose {
		fmt.Printf("🔍 AKS cluster '%s' status: %s\n", c.config.AKSName, status)
	} else {
		fmt.Println(status)
	}
	return nil
}

// getClusterStatus obtient le statut du cluster
func (c *ClusterManager) getClusterStatus(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "aks", "show",
		"--name", c.config.AKSName,
		"--resource-group", c.config.ResourceName,
		"--subscription", c.config.Subscription,
		"--query", "powerState.code",
		"--output", "tsv",
	)

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	status := strings.TrimSpace(string(output))
	return status, nil
}
