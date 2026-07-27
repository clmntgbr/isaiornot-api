package di

import (
	"go-api/cmd/shared/pipeline"
	"go-api/handler/worker"
	"go-api/infrastructure/config"
	metadatainfra "go-api/infrastructure/metadata"
	"go-api/infrastructure/storage"
	metadataUseCase "go-api/usecase/metadata"
	"go-api/usecase/signal"
	"log"

	"gorm.io/gorm"
)

type Container struct {
	MetadataHandler *worker.MetadataHandler
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	shared, err := pipeline.New(db, env)
	if err != nil {
		log.Fatalf("failed to create pipeline: %v", err)
	}

	storageClient, err := storage.NewMinIOStorage(env)
	if err != nil {
		log.Fatalf("failed to create storage client: %v", err)
	}

	analyzer := metadatainfra.NewAnalyzer()
	analyzeMediaMetadataUseCase := metadataUseCase.NewAnalyzeMediaMetadataUseCase(storageClient, analyzer)
	createSignalUseCase := signal.NewCreateSignalUseCase(shared.SignalRepo)

	return &Container{
		MetadataHandler: worker.NewMetadataHandler(
			shared.Parser,
			shared.SecurityValidator,
			shared.Dispatcher,
			analyzeMediaMetadataUseCase,
			createSignalUseCase,
		),
	}
}
