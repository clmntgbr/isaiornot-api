package di

import (
	"go-api/cmd/shared/pipeline"
	"go-api/handler/worker"
	"go-api/infrastructure/config"
	"go-api/infrastructure/messaging/rabbitmq"
	"log"

	"gorm.io/gorm"
)

type Container struct {
	AnalyzeRequestHandler *worker.AnalyzeRequestHandler
	StageDoneHandler      *worker.StageDoneHandler
	WorkerPool            *rabbitmq.WorkerPool
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	shared, err := pipeline.New(db, env)
	if err != nil {
		log.Fatalf("failed to create pipeline: %v", err)
	}

	analyzeRequestHandler := worker.NewAnalyzeRequestHandler(shared.Parser, shared.SecurityValidator, shared.Dispatcher)
	stageDoneHandler := worker.NewStageDoneHandler(shared.Parser, shared.SecurityValidator, shared.Dispatcher)
	workerPool := rabbitmq.NewDispatcherWorkers(env, analyzeRequestHandler, stageDoneHandler)

	return &Container{
		AnalyzeRequestHandler: analyzeRequestHandler,
		StageDoneHandler:      stageDoneHandler,
		WorkerPool:            workerPool,
	}
}
