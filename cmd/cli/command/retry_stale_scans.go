package command

import (
	"fmt"
	"time"

	"go-api/cmd/cli/di"
	"go-api/infrastructure/config"

	"github.com/spf13/cobra"
)

func NewRetryStaleScansCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "retry-stale-scans",
		Short: "Republish in-progress scans older than 1h",
		RunE: func(cmd *cobra.Command, args []string) error {
			env := config.Load()
			db := config.ConnectDatabase(env)
			container, err := di.NewContainer(db, env)
			if err != nil {
				return err
			}

			result, err := container.RetryStaleScansUseCase.Execute(cmd.Context(), time.Hour)
			if err != nil {
				return err
			}

			fmt.Printf("Retried %d scan(s) (%d media) older than 1h\n", result.Retried, result.Medias)
			return nil
		},
	}
}
