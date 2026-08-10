package di

import (
	"log"

	mediacmd "go-api/internal/application/command/media"
	signalcmd "go-api/internal/application/command/signal"
	"go-api/internal/application/event/dedup"
	eventmedia "go-api/internal/application/event/media"
	"go-api/internal/application/registry"
	domainmedia "go-api/internal/domain/media"
	"go-api/internal/domain/port"
	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/messaging/rabbitmq"
	"go-api/internal/infrastructure/persistence/outbox"
	"go-api/internal/infrastructure/persistence/processed"
	"go-api/internal/infrastructure/persistence/write"
	"go-api/internal/infrastructure/sightengine"
	"go-api/internal/infrastructure/storage"

	"gorm.io/gorm"
)

type Container struct {
	Consumer *rabbitmq.Consumer
	Conn     *rabbitmq.Connection
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	topology := rabbitmq.DefaultTopology(
		env.RabbitMQExchange,
		env.AIModelAnalyzeQueue,
		env.AIModelAnalyzeRoutingKey,
		env.RabbitMQRetryTTLMS,
	)

	conn, err := rabbitmq.Connect(env.RabbitMQURL, topology)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}

	publisher := rabbitmq.NewPublisher(conn, env.RabbitMQExchange)
	outboxRepo := outbox.NewRepository(db)
	dedupRepo := processed.NewRepository(db)

	minioStorage, err := storage.NewMinIOStorage(env)
	if err != nil {
		log.Fatalf("failed to create storage client: %v", err)
	}

	mediaWriteRepo := write.NewMediaWriteRepository(db)
	scanWriteRepo := write.NewScanWriteRepository(db)
	signalWriteRepo := write.NewSignalWriteRepository(db)

	providers := []port.AIModelProvider{
		sightengine.NewAnalyzer(sightengine.NewClient(
			env.SightengineAPIURL,
			env.SightengineAPIUser,
			env.SightengineAPISecret,
		)),
	}

	upsertSignalHandler := signalcmd.NewUpsertSignalHandler(signalWriteRepo, outboxRepo)
	analyzeAIModelHandler := mediacmd.NewAnalyzeAIModelHandler(
		mediaWriteRepo,
		scanWriteRepo,
		minioStorage,
		providers,
		upsertSignalHandler,
		publisher,
	)

	reg := registry.NewHandlerRegistry()
	reg.Register(domainmedia.EventTypeMediaAnalyzeAIModel, dedup.With(
		dedupRepo,
		"analyze_ai_model_on_requested",
		eventmedia.NewAnalyzeAIModelOnRequestedHandler(analyzeAIModelHandler).Handle,
	))

	consumer := rabbitmq.NewConsumer(conn, reg, env.WorkerConcurrency, env.WorkerMaxRetries)

	return &Container{
		Consumer: consumer,
		Conn:     conn,
	}
}
