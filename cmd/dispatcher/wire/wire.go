package wire

import (
	"go-api/handler"
	"go-api/infrastructure/centrifugo"
	"go-api/infrastructure/config"
	"go-api/infrastructure/messaging/rabbitmq"
	"go-api/infrastructure/messaging/security"
	repoGorm "go-api/repository/gorm"
	analysisuc "go-api/usecase/analysis"
	pipelineuc "go-api/usecase/pipeline"
	"log"

	"gorm.io/gorm"
)

type Container struct {
	AnalyzeRequestHandler *handler.AnalyzeRequestHandler
	StageDoneHandler      *handler.StageDoneHandler
	WorkerPool            *rabbitmq.WorkerPool
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	publisher, err := rabbitmq.NewPublisherFromEnv(env)
	if err != nil {
		log.Fatalf("failed to create publisher: %v", err)
	}

	centrifugoPublisher := centrifugo.NewPublisher(env)

	mediaRepo := repoGorm.NewMediaRepository(db)
	analysisRepo := repoGorm.NewAnalysisRepository(db)
	signalRepo := repoGorm.NewSignalRepository(db)

	updateAnalysisStatusUseCase := analysisuc.NewUpdateAnalysisStatusUseCase(&analysisRepo)
	aggregateAnalysisUseCase := pipelineuc.NewAggregateAnalysisUseCase(
		&mediaRepo,
		&analysisRepo,
		&signalRepo,
		updateAnalysisStatusUseCase,
		centrifugoPublisher,
	)
	dispatcher := pipelineuc.NewDispatcher(env, &mediaRepo, publisher, aggregateAnalysisUseCase)

	parser := security.NewWorkerParser(env)
	securityValidator := security.NewWorkerSecurityValidator(env)

	analyzeRequestHandler := handler.NewAnalyzeRequestHandler(parser, securityValidator, dispatcher)
	stageDoneHandler := handler.NewStageDoneHandler(parser, securityValidator, dispatcher)

	workerPool := rabbitmq.NewDispatcherWorkers(env, analyzeRequestHandler, stageDoneHandler)

	return &Container{
		AnalyzeRequestHandler: analyzeRequestHandler,
		StageDoneHandler:      stageDoneHandler,
		WorkerPool:            workerPool,
	}
}
