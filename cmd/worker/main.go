package main

import (
	"os"

	"github.com/cryptellation/version"
	"github.com/spf13/cobra"
)

// rootCmd is the worker root command.
var rootCmd = &cobra.Command{
	Use:     "worker",
	Version: version.FullVersion(),
	Short:   "worker - a worker executing cryptellation sma temporal workflows",
}

func main() {
	// Set commands
	rootCmd.AddCommand(serveCmd)
	addDatabaseCommands(rootCmd)

	// Execute command
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
