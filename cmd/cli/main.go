package main

import (
	"fmt"
	cliCommand "go-api/cmd/cli/command"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "cli",
		Short: "IsAIorNot CLI",
		Long:  "Operational commands for IsAIorNot API",
	}

	rootCmd.AddCommand(
		cliCommand.NewMigrateCommand(),
		cliCommand.NewRetryStaleScansCommand(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}
