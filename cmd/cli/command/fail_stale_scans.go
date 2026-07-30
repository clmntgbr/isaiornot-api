package command

import (
	"fmt"
	"time"

	"go-api/cmd/cli/di"
	"go-api/infrastructure/config"

	"github.com/spf13/cobra"
)

func NewFailStaleScansCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "fail-stale-scans",
		Short: "Fail in-progress scans older than 3h",
		RunE: func(cmd *cobra.Command, args []string) error {
			env := config.Load()
			db := config.ConnectDatabase(env)
			container, err := di.NewContainer(db, env)
			if err != nil {
				return err
			}

			result, err := container.FailStaleScansUseCase.Execute(cmd.Context(), 3*time.Hour)
			if err != nil {
				return err
			}

			fmt.Printf("Failed %d stale scan(s) older than 3h\n", result.Failed)
			return nil
		},
	}
}
