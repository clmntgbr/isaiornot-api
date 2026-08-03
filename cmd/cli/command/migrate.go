package command

import (
	"embed"
	"fmt"
	"go-api/domain/entity"
	"go-api/infrastructure/config"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

//go:embed sql/*.sql
var sqlFiles embed.FS

func NewMigrateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Migrate the database",
		Long:  "Migrate the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			env := config.Load()
			db := config.ConnectDatabase(env)

			err := db.Transaction(func(tx *gorm.DB) error {
				if err := tx.AutoMigrate(
					&entity.Quota{},
					&entity.Plan{},
					&entity.Subscription{},
					&entity.User{},
					&entity.Insight{},
					&entity.Scan{},
					&entity.Media{},
					&entity.Signal{},
					&entity.Invoice{},
				); err != nil {
					return err
				}

				plansSQL, err := sqlFiles.ReadFile("sql/plans.sql")
				if err != nil {
					return fmt.Errorf("failed to read plans.sql: %w", err)
				}

				if err := tx.Exec(string(plansSQL)).Error; err != nil {
					return fmt.Errorf("failed to seed plans: %w", err)
				}

				return nil
			})

			if err != nil {
				return fmt.Errorf(
					"🚨 failed to migrate database: %w",
					err,
				)
			}

			fmt.Println("🎉 Database migrations completed successfully")
			return nil
		},
	}
}
