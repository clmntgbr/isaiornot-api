package command

import (
	"fmt"
	"log"
	"time"

	"go-api/cmd/cli/di"
	"go-api/infrastructure/config"

	"github.com/spf13/cobra"
)

func NewRetryStaleScansCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "retry-stale-scans",
		Short: "Fail stuck uploaded scans; retry processing scans (fail after 3 retries)",
		RunE: func(cmd *cobra.Command, args []string) error {
			startedAt := time.Now().UTC()
			log.Printf("retry-stale-scans started")

			env := config.Load()
			db := config.ConnectDatabase(env)
			container, err := di.NewContainer(db, env)
			if err != nil {
				return err
			}

			result, err := container.RetryStaleScansUseCase.Execute(cmd.Context())
			if err != nil {
				log.Printf("retry-stale-scans failed: %v", err)
				return err
			}

			log.Printf(
				"retry-stale-scans finished in %s: %d retried, %d failed, %d media",
				time.Since(startedAt).Round(time.Millisecond),
				result.Retried,
				result.Failed,
				result.Medias,
			)
			fmt.Printf("Retried %d scan(s) (%d media), failed %d\n", result.Retried, result.Medias, result.Failed)
			return nil
		},
	}
}
