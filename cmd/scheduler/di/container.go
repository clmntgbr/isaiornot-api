package di

import (
	"go-api/domain/port"
	"go-api/infrastructure/centrifugo"
	"go-api/infrastructure/config"
	"go-api/infrastructure/messaging/rabbitmq"
	"go-api/infrastructure/storage"
	repoGorm "go-api/repository/gorm"
	"go-api/usecase/media"
	"go-api/usecase/scan"

	"gorm.io/gorm"
)

type Container struct {
	RetryStaleScansUseCase *scan.RetryStaleScansUseCase
}

func NewContainer(db *gorm.DB, env *config.Config) (*Container, error) {
	scanRepo := repoGorm.NewScanRepository(db)
	mediaRepo := repoGorm.NewMediaRepository(db)
	signalRepo := repoGorm.NewSignalRepository(db)

	var publisher port.MessagePublisher = rabbitmq.NewLazyPublisherFromEnv(env)
	centrifugoPublisher := centrifugo.NewPublisher(env)

	storageClient, err := storage.NewMinIOStorage(env)
	if err != nil {
		return nil, err
	}

	deleteOriginalsUseCase := media.NewDeleteOriginalsUseCase(storageClient)
	failScanUseCase := scan.NewFailScanUseCase(scanRepo, mediaRepo, centrifugoPublisher, deleteOriginalsUseCase)
	publishMetadataUseCase := media.NewPublishMetadataUseCase(
		mediaRepo,
		publisher,
		centrifugoPublisher,
		env.AnalyzeQueues(),
	)
	retryStaleScansUseCase := scan.NewRetryStaleScansUseCase(
		scanRepo,
		mediaRepo,
		signalRepo,
		publishMetadataUseCase,
		failScanUseCase,
	)

	return &Container{
		RetryStaleScansUseCase: retryStaleScansUseCase,
	}, nil
}
