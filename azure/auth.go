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

// IsLoggedIn vérifie si on est connecté avec az login ou SPN
func (a *AuthClient) IsLoggedIn(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Test simple avec az account show
	cmd := exec.CommandContext(ctx, "az", "account", "show", "--output", "none")
	if cmd.Run() == nil {
		// Vérifier l'accès au cluster AKS spécifique
		return a.canAccessCluster(ctx)
	}
	return false
}

// canAccessCluster teste l'accès au cluster AKS
func (a *AuthClient) canAccessCluster(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "aks", "show",
		"--name", a.config.AKSName,
		"--resource-group", a.config.ResourceName,
		"--output", "none",
	)
	return cmd.Run() == nil
}

// Login tente az login interactif puis SPN en fallback
func (a *AuthClient) Login(ctx context.Context) error {
	// Essayer az login interactif d'abord
	if err := a.tryInteractiveLogin(ctx); err == nil {
		return nil
	}

	// Fallback vers SPN si configuré
	if a.config.HasSPNConfig() {
		fmt.Println("🔄 Interactive login failed, trying service principal...")
		return a.loginWithSPN(ctx)
	}

	return fmt.Errorf("no authentication method available (interactive login failed and no SPN configured)")
}

// tryInteractiveLogin essaie la connexion interactive
func (a *AuthClient) tryInteractiveLogin(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	fmt.Println("🔐 Attempting interactive login...")
	cmd := exec.CommandContext(ctx, "az", "login")
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("interactive login failed: %w", err)
	}

	// Vérifier l'accès au cluster après login
	if !a.canAccessCluster(ctx) {
		return fmt.Errorf("login successful but no access to AKS cluster")
	}

	fmt.Println("✅ Interactive login successful")
	return nil
}

// loginWithSPN utilise l'authentification par service principal
func (a *AuthClient) loginWithSPN(ctx context.Context) error {
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
		return fmt.Errorf("SPN login failed: %w", err)
	}

	fmt.Println("✅ Service principal login successful")
	return nil
}

func (a *AuthClient) SetupCluster(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Définir la subscription si configurée
	if a.config.Subscription != "" {
		fmt.Println("🔧 Setting subscription...")
		cmd := exec.CommandContext(ctx, "az", "account", "set",
			"--subscription", a.config.Subscription,
		)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set subscription: %w", err)
		}
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