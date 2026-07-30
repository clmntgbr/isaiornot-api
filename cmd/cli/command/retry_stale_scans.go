package command

import (
	"fmt"

	"go-api/cmd/cli/di"
	"go-api/infrastructure/config"

	"github.com/spf13/cobra"
)

func NewRetryStaleScansCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "retry-stale-scans",
		Short: "Retry in-progress stale scans (fail after 3 retries)",
		RunE: func(cmd *cobra.Command, args []string) error {
			env := config.Load()
			db := config.ConnectDatabase(env)
			container, err := di.NewContainer(db, env)
			if err != nil {
				return err
			}

			result, err := container.RetryStaleScansUseCase.Execute(cmd.Context())
			if err != nil {
				return err
			}

			fmt.Printf("Retried %d scan(s) (%d media), failed %d\n", result.Retried, result.Medias, result.Failed)
			return nil
		},
	}
}
