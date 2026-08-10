package di

import (
	"log"

	scancmd "go-api/internal/application/command/scan"
	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/persistence/outbox"
	"go-api/internal/infrastructure/persistence/write"

	"gorm.io/gorm"
)

type Container struct {
	RetryStaleScansHandler *scancmd.RetryStaleScansHandler
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	_ = env

	scanWriteRepo := write.NewScanWriteRepository(db)
	mediaWriteRepo := write.NewMediaWriteRepository(db)
	signalWriteRepo := write.NewSignalWriteRepository(db)
	outboxRepo := outbox.NewRepository(db)

	failScanHandler := scancmd.NewFailScanHandler(scanWriteRepo, mediaWriteRepo, outboxRepo)
	retryStaleScansHandler := scancmd.NewRetryStaleScansHandler(
		scanWriteRepo,
		mediaWriteRepo,
		signalWriteRepo,
		outboxRepo,
		failScanHandler,
	)

	log.Println("scheduler container ready")

	return &Container{
		RetryStaleScansHandler: retryStaleScansHandler,
	}
}
