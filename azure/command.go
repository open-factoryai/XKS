package azure

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type CommandOptions struct {
	Command      string
	File         string
	NoWait       bool
	Output       string
	Subscription string
	Debug        bool
	OnlyErrors   bool
	Query        string
	CommandID    string // Pour az aks command result
}

type CommandRunner struct {
	config *Config
}

func NewCommandRunner(config *Config) *CommandRunner {
	return &CommandRunner{config: config}
}

func (c *CommandRunner) RunRemoteCommand(ctx context.Context, args []string, options *CommandOptions, verbose bool) error {
	if len(args) == 0 && options.Command == "" {
		return fmt.Errorf("no command provided")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	var command string
	if options.Command != "" {
		command = options.Command
	} else {
		command = strings.Join(args, " ")
	}

	baseArgs := c.buildCommandArgs(command, options)

	if verbose {
		fmt.Println("➡️ Running command inside AKS cluster:")
		fmt.Printf("az %s\n", strings.Join(baseArgs, " "))
	}

	cmd := exec.CommandContext(ctx, "az", baseArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

func (c *CommandRunner) buildCommandArgs(command string, options *CommandOptions) []string {
	baseArgs := []string{
		"aks", "command", "invoke",
		"--resource-group", c.config.ResourceName,
		"--name", c.config.AKSName,
	}

	// Command obligatoire
	if command != "" {
		baseArgs = append(baseArgs, "--command", command)
	}

	// Gestion des fichiers
	if options.File != "" {
		// Fichier explicitement spécifié
		baseArgs = append(baseArgs, "--file", options.File)
	} else {
		// Détection automatique des fichiers dans la commande
		files := c.extractFilesFromCommand(command)
		if len(files) > 0 {
			// Uploader seulement les fichiers détectés
			baseArgs = append(baseArgs, "--file", strings.Join(files, ","))
		}
	}

	if options.NoWait {
		baseArgs = append(baseArgs, "--no-wait")
	}

	if options.Output != "" {
		baseArgs = append(baseArgs, "--output", options.Output)
	}

	if options.Subscription != "" {
		baseArgs = append(baseArgs, "--subscription", options.Subscription)
	}

	if options.Debug {
		baseArgs = append(baseArgs, "--debug")
	}

	if options.OnlyErrors {
		baseArgs = append(baseArgs, "--only-show-errors")
	}

	if options.Query != "" {
		baseArgs = append(baseArgs, "--query", options.Query)
	}

	return baseArgs
}

func (c *CommandRunner) GetCommandResult(ctx context.Context, commandID string, options *CommandOptions, verbose bool) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	baseArgs := []string{
		"aks", "command", "result",
		"--resource-group", c.config.ResourceName,
		"--name", c.config.AKSName,
	}

	if commandID != "" {
		baseArgs = append(baseArgs, "--command-id", commandID)
	}

	if options.Output != "" {
		baseArgs = append(baseArgs, "--output", options.Output)
	}

	if options.Subscription != "" {
		baseArgs = append(baseArgs, "--subscription", options.Subscription)
	}

	if options.Debug {
		baseArgs = append(baseArgs, "--debug")
	}

	if options.OnlyErrors {
		baseArgs = append(baseArgs, "--only-show-errors")
	}

	if options.Query != "" {
		baseArgs = append(baseArgs, "--query", options.Query)
	}

	if verbose {
		fmt.Println("➡️ Fetching command result:")
		fmt.Printf("az %s\n", strings.Join(baseArgs, " "))
	}

	cmd := exec.CommandContext(ctx, "az", baseArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to get command result: %w", err)
	}

	return nil
}

// extractFilesFromCommand extrait les noms de fichiers des commandes kubectl/helm
func (c *CommandRunner) extractFilesFromCommand(command string) []string {
	var files []string

	// Patterns pour détecter les fichiers dans les commandes
	patterns := []string{
		`-f\s+([^\s]+)`,          // -f file.yaml
		`--filename\s+([^\s]+)`,  // --filename file.yaml
		`-k\s+([^\s]+)`,          // -k directory (kustomize)
		`--kustomize\s+([^\s]+)`, // --kustomize directory
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(command, -1)

		for _, match := range matches {
			if len(match) > 1 {
				filename := strings.TrimSpace(match[1])
				// Vérifier que le fichier existe
				if c.fileExists(filename) {
					files = append(files, filename)
				}
			}
		}
	}

	// Détecter les commandes helm avec des chemins de charts
	if strings.Contains(command, "helm install") || strings.Contains(command, "helm upgrade") {
		parts := strings.Fields(command)
		for i, part := range parts {
			// Chart path est généralement après le nom de release
			if (part == "install" || part == "upgrade") && i+2 < len(parts) {
				chartPath := parts[i+2]
				// Si c'est un répertoire local ou un fichier
				if strings.HasPrefix(chartPath, "./") || strings.HasPrefix(chartPath, "/") || !strings.Contains(chartPath, "/") {
					if c.fileExists(chartPath) {
						files = append(files, chartPath)
					}
				}
				break
			}
		}
	}

	return c.removeDuplicates(files)
}

// fileExists vérifie si un fichier ou répertoire existe
func (c *CommandRunner) fileExists(path string) bool {
	// Convertir les chemins relatifs
	if !filepath.IsAbs(path) {
		wd, err := os.Getwd()
		if err != nil {
			return false
		}
		path = filepath.Join(wd, path)
	}

	_, err := os.Stat(path)
	return err == nil
}

// removeDuplicates supprime les doublons du slice
func (c *CommandRunner) removeDuplicates(files []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, file := range files {
		if !keys[file] {
			keys[file] = true
			result = append(result, file)
		}
	}

	return result
}
