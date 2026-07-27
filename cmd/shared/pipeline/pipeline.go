package pipeline

import (
	"go-api/domain/port"
	"go-api/domain/repository"
	"go-api/infrastructure/centrifugo"
	"go-api/infrastructure/config"
	"go-api/infrastructure/messaging/rabbitmq"
	"go-api/infrastructure/messaging/security"
	repoGorm "go-api/repository/gorm"
	pipelineUseCase "go-api/usecase/pipeline"
	scanUseCase "go-api/usecase/scan"
	"go-api/usecase/subscription"

	"gorm.io/gorm"
)

type Shared struct {
	Publisher           port.MessagePublisher
	CentrifugoPublisher port.RealtimePublisher
	MediaRepo           repository.MediaRepository
	ScanRepo            repository.ScanRepository
	SignalRepo          repository.SignalRepository
	Dispatcher          *pipelineUseCase.Dispatcher
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
	userRepo := repoGorm.NewUserRepository(db)
	subscriptionRepo := repoGorm.NewSubscriptionRepository(db)
	planRepo := repoGorm.NewPlanRepository(db)

	resolveEffectivePlanUseCase := subscription.NewResolveEffectivePlanUseCase(planRepo)
	resolvePipelineAccessUseCase := subscription.NewResolvePipelineAccessUseCase(
		userRepo,
		subscriptionRepo,
		resolveEffectivePlanUseCase,
	)

	updateScanStatusUseCase := scanUseCase.NewUpdateScanStatusUseCase(scanRepo)
	aggregateScanUseCase := pipelineUseCase.NewAggregateScanUseCase(
		mediaRepo,
		scanRepo,
		signalRepo,
		updateScanStatusUseCase,
		centrifugoPublisher,
		resolvePipelineAccessUseCase,
	)
	dispatcher := pipelineUseCase.NewDispatcher(
		env.AnalyzeQueues(),
		mediaRepo,
		publisher,
		aggregateScanUseCase,
		resolvePipelineAccessUseCase,
	)

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
