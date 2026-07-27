package wire

import (
	"go-api/cmd/shared/pipeline"
	"go-api/handler/worker"
	"go-api/infrastructure/aimodel"
	"go-api/infrastructure/config"
	"go-api/infrastructure/sightengine"
	"go-api/infrastructure/storage"
	aimodelUseCase "go-api/usecase/aimodel"
	"go-api/usecase/signal"
	"log"

	"gorm.io/gorm"
)

type Container struct {
	AiModelHandler *worker.AiModelHandler
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

	sightengineClient := sightengine.NewClient(env)
	analyzer := aimodel.NewAnalyzer(sightengineClient)
	analyzeMediaAiModelUseCase := aimodelUseCase.NewAnalyzeMediaAiModelUseCase(storageClient, analyzer)
	createSignalUseCase := signal.NewCreateSignalUseCase(shared.SignalRepo)

	return &Container{
		AiModelHandler: worker.NewAiModelHandler(
			shared.Parser,
			shared.SecurityValidator,
			shared.Dispatcher,
			analyzeMediaAiModelUseCase,
			createSignalUseCase,
		),
	}
}
