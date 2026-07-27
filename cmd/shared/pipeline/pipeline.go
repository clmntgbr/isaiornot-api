package pipeline

import (
	"go-api/domain/port"
	"go-api/domain/repository"
	"go-api/infrastructure/centrifugo"
	"go-api/infrastructure/config"
	"go-api/infrastructure/messaging/rabbitmq"
	"go-api/infrastructure/messaging/security"
	repoGorm "go-api/repository/gorm"
	pipelineuc "go-api/usecase/pipeline"
	scanuc "go-api/usecase/scan"

	"gorm.io/gorm"
)

type Shared struct {
	Publisher           port.MessagePublisher
	CentrifugoPublisher port.RealtimePublisher
	MediaRepo           repository.MediaRepository
	ScanRepo            repository.ScanRepository
	SignalRepo          repository.SignalRepository
	Dispatcher          *pipelineuc.Dispatcher
	Parser              *security.WorkerParser
	SecurityValidator   *security.WorkerSecurityValidator
}

func New(db *gorm.DB, env *config.Config) (*Shared, error) {
	publisher, err := rabbitmq.NewPublisherFromEnv(env)
	if err != nil {
		return nil, err
	}

	centrifugoPublisher := centrifugo.NewPublisher(env)

	mediaRepo := repoGorm.NewMediaRepository(db)
	scanRepo := repoGorm.NewScanRepository(db)
	signalRepo := repoGorm.NewSignalRepository(db)

	updateScanStatusUseCase := scanuc.NewUpdateScanStatusUseCase(scanRepo)
	aggregateScanUseCase := pipelineuc.NewAggregateScanUseCase(
		mediaRepo,
		scanRepo,
		signalRepo,
		updateScanStatusUseCase,
		centrifugoPublisher,
	)
	dispatcher := pipelineuc.NewDispatcher(env.AnalyzeQueues(), mediaRepo, publisher, aggregateScanUseCase)

	return &Shared{
		Publisher:           publisher,
		CentrifugoPublisher: centrifugoPublisher,
		MediaRepo:           mediaRepo,
		ScanRepo:            scanRepo,
		SignalRepo:          signalRepo,
		Dispatcher:          dispatcher,
		Parser:              security.NewWorkerParser(env),
		SecurityValidator:   security.NewWorkerSecurityValidator(env),
	}, nil
}
