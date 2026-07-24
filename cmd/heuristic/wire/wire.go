package wire

import (
	"go-api/handler"
	"go-api/infrastructure/centrifugo"
	"go-api/infrastructure/config"
	heuristicsinfra "go-api/infrastructure/heuristics"
	"go-api/infrastructure/messaging/rabbitmq"
	"go-api/infrastructure/messaging/security"
	"go-api/infrastructure/storage"
	repoGorm "go-api/repository/gorm"
	analysisuc "go-api/usecase/analysis"
	heuristicuc "go-api/usecase/heuristic"
	insightuc "go-api/usecase/insight"
	pipelineuc "go-api/usecase/pipeline"
	"go-api/usecase/signal"
	"log"

	"gorm.io/gorm"
)

type Container struct {
	HeuristicHandler *handler.HeuristicHandler
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	storageClient, err := storage.NewMinIOStorage(env)
	if err != nil {
		log.Fatalf("failed to create storage client: %v", err)
	}

	publisher, err := rabbitmq.NewPublisherFromEnv(env)
	if err != nil {
		log.Fatalf("failed to create publisher: %v", err)
	}

	centrifugoPublisher := centrifugo.NewPublisher(env)

	mediaRepo := repoGorm.NewMediaRepository(db)
	analysisRepo := repoGorm.NewAnalysisRepository(db)
	signalRepo := repoGorm.NewSignalRepository(db)
	insightRepo := repoGorm.NewInsightRepository(db)

	updateAnalysisStatusUseCase := analysisuc.NewUpdateAnalysisStatusUseCase(&analysisRepo)
	aggregateAnalysisUseCase := pipelineuc.NewAggregateAnalysisUseCase(
		&mediaRepo,
		&analysisRepo,
		&signalRepo,
		updateAnalysisStatusUseCase,
		centrifugoPublisher,
	)
	dispatcher := pipelineuc.NewDispatcher(env, &mediaRepo, publisher, aggregateAnalysisUseCase)

	analyzer := heuristicsinfra.NewAnalyzer()
	analyzeMediaHeuristicsUseCase := heuristicuc.NewAnalyzeMediaHeuristicsUseCase(storageClient, analyzer)
	createSignalUseCase := signal.NewCreateSignalUseCase(&signalRepo)
	createInsightUseCase := insightuc.NewCreateInsightUseCase(&insightRepo, &mediaRepo)

	parser := security.NewWorkerParser(env)
	securityValidator := security.NewWorkerSecurityValidator(env)

	heuristicHandler := handler.NewHeuristicHandler(
		parser,
		securityValidator,
		dispatcher,
		analyzeMediaHeuristicsUseCase,
		createSignalUseCase,
		createInsightUseCase,
	)

	return &Container{
		HeuristicHandler: heuristicHandler,
	}
}
