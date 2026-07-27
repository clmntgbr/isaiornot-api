package di

import (
	"go-api/domain/port"
	"go-api/domain/repository"
	httphandler "go-api/handler/http"
	"go-api/handler/http/middleware"
	"go-api/infrastructure/centrifugo"
	infraClerk "go-api/infrastructure/clerk"
	"go-api/infrastructure/config"
	"go-api/infrastructure/messaging/rabbitmq"
	"go-api/infrastructure/storage"
	repoGorm "go-api/repository/gorm"
	"log"

	"gorm.io/gorm"
)

type Container struct {
	AuthenticateMiddleware       *middleware.AuthenticateMiddleware
	UserWebhookMiddleware        *middleware.UserWebhookMiddleware
	MediaUploadWebhookMiddleware *middleware.MediaUploadWebhookMiddleware
	BillingWebhookMiddleware     *middleware.BillingWebhookMiddleware
	UserWebhookHandler           *httphandler.UserWebhookHandler
	MediaUploadWebhookHandler    *httphandler.MediaUploadWebhookHandler
	UserHandler                  *httphandler.UserHandler
	ScanHandler                  *httphandler.ScanHandler
	MediaHandler                 *httphandler.MediaHandler
	RealtimeHandler              *httphandler.RealtimeHandler
	PlanHandler                  *httphandler.PlanHandler
	BillingWebhookHandler        *httphandler.BillingWebhookHandler
	SubscriptionHandler          *httphandler.SubscriptionHandler
}

type apiDeps struct {
	env                 *config.Config
	db                  *gorm.DB
	storage             *storage.MinIOStorage
	publisher           port.MessagePublisher
	centrifugoPublisher port.RealtimePublisher
	userRepo            repository.UserRepository
	mediaRepo           repository.MediaRepository
	scanRepo            repository.ScanRepository
	planRepo            repository.PlanRepository
	subscriptionRepo    repository.SubscriptionRepository
	jwksProvider        port.TokenKeyProvider
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	jwksProvider, err := infraClerk.NewJWKSProvider(env)
	if err != nil {
		log.Fatalf("failed to create JWKS provider: %v", err)
	}

	storageClient, err := storage.NewMinIOStorage(env)
	if err != nil {
		log.Fatalf("failed to create storage client: %v", err)
	}

	d := &apiDeps{
		env:                 env,
		db:                  db,
		storage:             storageClient,
		publisher:           rabbitmq.NewLazyPublisherFromEnv(env),
		centrifugoPublisher: centrifugo.NewPublisher(env),
		userRepo:            repoGorm.NewUserRepository(db),
		mediaRepo:           repoGorm.NewMediaRepository(db),
		scanRepo:            repoGorm.NewScanRepository(db),
		planRepo:            repoGorm.NewPlanRepository(db),
		subscriptionRepo:    repoGorm.NewSubscriptionRepository(db),
		jwksProvider:        jwksProvider,
	}

	authBundle := wireAuth(d)
	scanBundle := wireScan(d)
	mediaBundle := wireMedia(d, scanBundle)
	subscriptionBundle := wireSubscription(d, authBundle)

	return &Container{
		AuthenticateMiddleware:       authBundle.authenticateMiddleware,
		UserWebhookMiddleware:        middleware.NewUserWebhookMiddleware(env.ClerkWebhookSecret),
		MediaUploadWebhookMiddleware: middleware.NewMediaUploadWebhookMiddleware(env.MinIOWebhookSecret),
		BillingWebhookMiddleware:     middleware.NewBillingWebhookMiddleware(env.StripeWebhookSecret),
		BillingWebhookHandler:        subscriptionBundle.billingWebhookHandler,
		UserWebhookHandler:           authBundle.userWebhookHandler,
		MediaUploadWebhookHandler:    mediaBundle.mediaUploadWebhookHandler,
		UserHandler:                  httphandler.NewUserHandler(),
		ScanHandler:                  scanBundle.scanHandler,
		MediaHandler:                 mediaBundle.mediaHandler,
		RealtimeHandler:              httphandler.NewRealtimeHandler(env),
		PlanHandler:                  subscriptionBundle.planHandler,
		SubscriptionHandler:          subscriptionBundle.subscriptionHandler,
	}
}
