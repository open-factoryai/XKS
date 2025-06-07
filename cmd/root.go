package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"xks/azure"

	"github.com/spf13/cobra"
)

var (
	verbose      bool
	file         string
	noWait       bool
	output       string
	subscription string
	debug        bool
	onlyErrors   bool
	query        string
	commandStr   string
	commandID    string
	getResult    bool
)

var rootCmd = &cobra.Command{
	Use:                "xks [command]",
	Short:              "XKS wraps az aks command invoke for kubectl and helm",
	Long:               `XKS is a CLI for running kubectl or helm commands securely inside a private AKS cluster. It supports all az aks command invoke options and provides an enhanced experience.`,
	RunE:               runCommand,
	DisableFlagParsing: false, // On garde le parsing pour nos flags
}

func runCommand(cmd *cobra.Command, args []string) error {
	if len(args) == 0 && commandStr == "" && !getResult {
		return fmt.Errorf("usage: xks <kubectl|helm> ... OR xks --command 'your command' OR xks --get-result [--command-id ID] OR xks <start|stop|status>")
	}

	ctx := context.Background()

	// Initialisation de la configuration
	config, err := azure.NewConfig()
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	commandRunner := azure.NewCommandRunner(config)

	// Mode récupération de résultat
	if getResult {
		options := &azure.CommandOptions{
			Output:       output,
			Subscription: subscription,
			Debug:        debug,
			OnlyErrors:   onlyErrors,
			Query:        query,
			CommandID:    commandID,
		}
		return commandRunner.GetCommandResult(ctx, commandID, options, verbose)
	}

	// Mode exécution normale
	authClient := azure.NewAuthClient(config)

	// Authentification si nécessaire
	if !authClient.IsLoggedIn(ctx) {
		if err := authClient.Login(ctx); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		if err := authClient.SetupCluster(ctx); err != nil {
			return fmt.Errorf("cluster setup failed: %w", err)
		}
	} else if verbose {
		fmt.Println("✅ Already authenticated.")
	}

	// Préparation des options
	options := &azure.CommandOptions{
		Command:      commandStr,
		File:         file,
		NoWait:       noWait,
		Output:       output,
		Subscription: subscription,
		Debug:        debug,
		OnlyErrors:   onlyErrors,
		Query:        query,
	}

	// Exécution de la commande
	return commandRunner.RunRemoteCommand(ctx, args, options, verbose)
}

func Execute() {
	// Parse manuel des arguments pour séparer les flags xks des commandes kubectl/helm
	originalArgs := os.Args[1:]

	// Declare all variables at the beginning to avoid goto issues
	var xksArgs []string
	var cmdArgs []string
	kubectlIndex := -1
	helmIndex := -1

	// Vérifier si c'est une commande cluster
	if len(originalArgs) > 0 {
		firstArg := originalArgs[0]
		if firstArg == "start" || firstArg == "stop" || firstArg == "status" {
			// Laisser Cobra gérer ces commandes
			goto executeCommand
		}
	}

	// Si les arguments contiennent kubectl ou helm, on sépare

	for i, arg := range originalArgs {
		if arg == "kubectl" {
			kubectlIndex = i
			break
		}
		if arg == "helm" {
			helmIndex = i
			break
		}
	}

	if kubectlIndex >= 0 {
		xksArgs = originalArgs[:kubectlIndex]
		cmdArgs = originalArgs[kubectlIndex:]
	} else if helmIndex >= 0 {
		xksArgs = originalArgs[:helmIndex]
		cmdArgs = originalArgs[helmIndex:]
	} else {
		xksArgs = originalArgs
		cmdArgs = []string{}
	}

	// Reconstruction des arguments pour Cobra
	if len(cmdArgs) > 0 {
		// Convertir la commande en string pour éviter le parsing des flags
		commandStr = strings.Join(cmdArgs, " ")
		os.Args = append([]string{os.Args[0]}, append(xksArgs, "--command", commandStr)...)
	}

executeCommand:
	// Flags principaux
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVar(&commandStr, "command", "", "Command or shell script to run (alternative to args)")
	rootCmd.PersistentFlags().StringVar(&file, "file", "", "Files to attach (use '.' for current directory)")
	rootCmd.PersistentFlags().BoolVar(&noWait, "no-wait", false, "Don't wait for operation to finish")

	// Flags de sortie Azure CLI
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "Output format (json, table, yaml, tsv)")
	rootCmd.PersistentFlags().StringVar(&subscription, "subscription", "", "Subscription name or ID")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Increase logging verbosity")
	rootCmd.PersistentFlags().BoolVar(&onlyErrors, "only-show-errors", false, "Only show errors")
	rootCmd.PersistentFlags().StringVar(&query, "query", "", "JMESPath query string")

	// Flags pour récupération de résultat
	rootCmd.PersistentFlags().BoolVar(&getResult, "get-result", false, "Get result from previous command")
	rootCmd.PersistentFlags().StringVar(&commandID, "command-id", "", "Command ID to get result for")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
