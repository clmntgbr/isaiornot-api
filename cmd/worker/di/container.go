package di

import (
	"log"

	mediacmd "go-api/internal/application/command/media"
	"go-api/internal/application/event/dedup"
	eventmedia "go-api/internal/application/event/media"
	eventscan "go-api/internal/application/event/scan"
	eventuser "go-api/internal/application/event/user"
	"go-api/internal/application/registry"
	domainmedia "go-api/internal/domain/media"
	domainscan "go-api/internal/domain/scan"
	domainuser "go-api/internal/domain/user"
	"go-api/internal/infrastructure/centrifugo"
	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/imaging"
	"go-api/internal/infrastructure/messaging/rabbitmq"
	"go-api/internal/infrastructure/notification"
	"go-api/internal/infrastructure/persistence/outbox"
	"go-api/internal/infrastructure/persistence/processed"
	"go-api/internal/infrastructure/persistence/write"
	"go-api/internal/infrastructure/storage"

	"gorm.io/gorm"
)

type Container struct {
	Relay    *outbox.Relay
	Consumer *rabbitmq.Consumer
	Conn     *rabbitmq.Connection
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	topology := rabbitmq.DefaultTopology(
		env.RabbitMQExchange,
		env.RabbitMQQueue,
		env.RabbitMQRoutingKey,
		env.RabbitMQRetryTTLMS,
	)

	conn, err := rabbitmq.Connect(env.RabbitMQURL, topology)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}

	publisher := rabbitmq.NewPublisher(conn, env.RabbitMQExchange)
	outboxRepo := outbox.NewRepository(db)
	relay := outbox.NewRelay(outboxRepo, publisher, env.OutboxPollInterval, 50)

	dedupRepo := processed.NewRepository(db)
	notifier := notification.NewLogNotifier()
	realtimePublisher := centrifugo.NewPublisher(env)
	reg := registry.NewHandlerRegistry()

	minioStorage, err := storage.NewMinIOStorage(env)
	if err != nil {
		log.Fatalf("failed to create storage client: %v", err)
	}

	mediaWriteRepo := write.NewMediaWriteRepository(db)
	scanWriteRepo := write.NewScanWriteRepository(db)
	generateThumbnailHandler := mediacmd.NewGenerateThumbnailHandler(
		mediaWriteRepo,
		scanWriteRepo,
		outboxRepo,
		minioStorage,
		imaging.NewThumbnailer(),
	)

	publishUserRealtime := eventuser.NewPublishRealtimeHandler(realtimePublisher)
	publishScanRealtime := eventscan.NewPublishRealtimeHandler(realtimePublisher)
	publishMediaRealtime := eventmedia.NewPublishRealtimeHandler(realtimePublisher, scanWriteRepo)

	reg.Register(domainuser.EventTypeUserCreated, dedup.With(
		dedupRepo,
		"user_created",
		eventuser.NewUserCreatedHandler().Handle,
	))
	reg.Register(domainuser.EventTypeUserCreated, dedup.With(
		dedupRepo,
		"notify_user_on_created",
		eventuser.NewNotifyUserOnCreatedHandler(notifier).Handle,
	))
	reg.Register(domainuser.EventTypeUserCreated, dedup.With(
		dedupRepo,
		"publish_user_created_realtime",
		publishUserRealtime.OnCreated,
	))
	reg.Register(domainuser.EventTypeUserUpdated, dedup.With(
		dedupRepo,
		"user_updated",
		eventuser.NewUserUpdatedHandler().Handle,
	))
	reg.Register(domainuser.EventTypeUserUpdated, dedup.With(
		dedupRepo,
		"publish_user_updated_realtime",
		publishUserRealtime.OnUpdated,
	))
	reg.Register(domainuser.EventTypeUserDeleted, dedup.With(
		dedupRepo,
		"user_deleted",
		eventuser.NewUserDeletedHandler().Handle,
	))
	reg.Register(domainuser.EventTypeUserDeleted, dedup.With(
		dedupRepo,
		"publish_user_deleted_realtime",
		publishUserRealtime.OnDeleted,
	))

	reg.Register(domainscan.EventTypeScanCreated, dedup.With(
		dedupRepo,
		"publish_scan_created_realtime",
		publishScanRealtime.OnCreated,
	))
	reg.Register(domainscan.EventTypeScanUpdated, dedup.With(
		dedupRepo,
		"publish_scan_updated_realtime",
		publishScanRealtime.OnUpdated,
	))
	reg.Register(domainscan.EventTypeScanCompleted, dedup.With(
		dedupRepo,
		"publish_scan_completed_realtime",
		publishScanRealtime.OnCompleted,
	))
	reg.Register(domainscan.EventTypeScanFailed, dedup.With(
		dedupRepo,
		"publish_scan_failed_realtime",
		publishScanRealtime.OnFailed,
	))

	reg.Register(domainmedia.EventTypeMediaCreated, dedup.With(
		dedupRepo,
		"publish_media_created_realtime",
		publishMediaRealtime.OnCreated,
	))
	reg.Register(domainmedia.EventTypeMediaUpdated, dedup.With(
		dedupRepo,
		"publish_media_updated_realtime",
		publishMediaRealtime.OnUpdated,
	))
	reg.Register(domainmedia.EventTypeMediaUploaded, dedup.With(
		dedupRepo,
		"generate_thumbnail_on_media_uploaded",
		eventmedia.NewGenerateThumbnailOnUploadedHandler(generateThumbnailHandler).Handle,
	))
	reg.Register(domainmedia.EventTypeMediaUploaded, dedup.With(
		dedupRepo,
		"publish_media_uploaded_realtime",
		publishMediaRealtime.OnUploaded,
	))
	reg.Register(domainmedia.EventTypeMediaCompleted, dedup.With(
		dedupRepo,
		"publish_media_completed_realtime",
		publishMediaRealtime.OnCompleted,
	))
	reg.Register(domainmedia.EventTypeMediaFailed, dedup.With(
		dedupRepo,
		"publish_media_failed_realtime",
		publishMediaRealtime.OnFailed,
	))

	consumer := rabbitmq.NewConsumer(conn, reg, env.WorkerConcurrency, env.WorkerMaxRetries)

	return &Container{
		Relay:    relay,
		Consumer: consumer,
		Conn:     conn,
	}
}
