package azure

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

type AuthClient struct {
	config *Config
}

func NewAuthClient(config *Config) *AuthClient {
	return &AuthClient{config: config}
}

func (a *AuthClient) IsLoggedIn(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "aks", "show",
		"--name", a.config.AKSName,
		"--resource-group", a.config.ResourceName,
		"--output", "none",
	)
	return cmd.Run() == nil
}

func (a *AuthClient) Login(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	fmt.Println("🔐 Logging in with service principal...")
	cmd := exec.CommandContext(ctx, "az", "login",
		"--service-principal",
		"--username", a.config.AppID,
		"--password", a.config.SecretID,
		"--tenant", a.config.TenantID,
		"--output", "none",
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}
	return nil
}

func (a *AuthClient) SetupCluster(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	fmt.Println("🔧 Setting subscription...")
	cmd := exec.CommandContext(ctx, "az", "account", "set",
		"--subscription", a.config.Subscription,
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set subscription: %w", err)
	}

	fmt.Println("📥 Fetching AKS credentials...")
	credCmd := exec.CommandContext(ctx, "az", "aks", "get-credentials",
		"--resource-group", a.config.ResourceName,
		"--name", a.config.AKSName,
		"--overwrite-existing",
	)
	if err := credCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch AKS credentials: %w", err)
	}

	return nil
}
