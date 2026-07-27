package wire

import (
	"go-api/cmd/shared/pipeline"
	"go-api/handler/worker"
	"go-api/infrastructure/config"
	heuristicsinfra "go-api/infrastructure/heuristics"
	"go-api/infrastructure/storage"
	repoGorm "go-api/repository/gorm"
	heuristicuc "go-api/usecase/heuristic"
	insightuc "go-api/usecase/insight"
	"go-api/usecase/signal"
	"log"

	"gorm.io/gorm"
)

type Container struct {
	HeuristicHandler *worker.HeuristicHandler
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

	insightRepo := repoGorm.NewInsightRepository(db)
	analyzer := heuristicsinfra.NewAnalyzer()
	analyzeMediaHeuristicsUseCase := heuristicuc.NewAnalyzeMediaHeuristicsUseCase(storageClient, analyzer)
	createSignalUseCase := signal.NewCreateSignalUseCase(shared.SignalRepo)
	createInsightUseCase := insightuc.NewCreateInsightUseCase(insightRepo, shared.MediaRepo)

	return &Container{
		HeuristicHandler: worker.NewHeuristicHandler(
			shared.Parser,
			shared.SecurityValidator,
			shared.Dispatcher,
			analyzeMediaHeuristicsUseCase,
			createSignalUseCase,
			createInsightUseCase,
		),
	}
}
