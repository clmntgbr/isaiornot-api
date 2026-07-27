package wire

import (
	"go-api/handler"
	"go-api/handler/middleware"
	"go-api/infrastructure/centrifugo"
	infraClerk "go-api/infrastructure/clerk"
	"go-api/infrastructure/config"
	"go-api/infrastructure/messaging/rabbitmq"
	"go-api/infrastructure/storage"
	infraStripe "go-api/infrastructure/stripe"
	"go-api/infrastructure/video"
	repoGorm "go-api/repository/gorm"
	"go-api/usecase/auth"
	"go-api/usecase/clerk"
	"go-api/usecase/media"
	"go-api/usecase/plan"
	"go-api/usecase/scan"
	"go-api/usecase/subscription"
	"go-api/usecase/thumbnail"
	"go-api/usecase/user"
	"log"

	"gorm.io/gorm"
)

type Container struct {
	AuthenticateMiddleware *middleware.AuthenticateMiddleware
	ClerkMiddleware        *middleware.ClerkMiddleware
	MinIOMiddleware        *middleware.MinIOMiddleware
	StripeMiddleware       *middleware.StripeMiddleware
	ClerkHandler           *handler.ClerkHandler
	MinIOHandler           *handler.MinIOHandler
	UserHandler            *handler.UserHandler
	ScanHandler            *handler.ScanHandler
	MediaHandler           *handler.MediaHandler
	RealtimeHandler        *handler.RealtimeHandler
	PlanHandler            *handler.PlanHandler
	StripeHandler          *handler.StripeHandler
	SubscriptionHandler    *handler.SubscriptionHandler
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

	userRepo := repoGorm.NewUserRepository(db)
	mediaRepo := repoGorm.NewMediaRepository(db)
	scanRepo := repoGorm.NewScanRepository(db)
	planRepo := repoGorm.NewPlanRepository(db)
	subscriptionRepo := repoGorm.NewSubscriptionRepository(db)

	publisher := rabbitmq.NewLazyPublisherFromEnv(env)
	centrifugoPublisher := centrifugo.NewPublisher(env)

	validateTokenUseCase := auth.NewValidateTokenUseCase(jwksProvider, &userRepo)
	fetchUserUseCase := clerk.NewFetchUserUseCase(env)
	getUserByClerkIDUseCase := user.NewGetUserByClerkIDUseCase(&userRepo)
	createFreeSubscriptionUseCase := subscription.NewCreateFreeSubscriptionUseCase(&planRepo, &subscriptionRepo, &userRepo)
	createUserUseCase := user.NewCreateUserUseCase(&userRepo, createFreeSubscriptionUseCase)
	updateUserUseCase := user.NewUpdateUserUseCase(&userRepo)
	deleteUserByClerkIDUseCase := user.NewDeleteUserByClerkIDUseCase(&userRepo)

	createMediaUseCase := media.NewCreateMediaUseCase(&scanRepo, &mediaRepo)
	resolveEffectivePlanUseCase := subscription.NewResolveEffectivePlanUseCase(&planRepo)
	getQuotaUsageUseCase := subscription.NewGetQuotaUsageUseCase(&mediaRepo)
	assertUploadAllowedUseCase := subscription.NewAssertUploadAllowedUseCase(
		&userRepo,
		&subscriptionRepo,
		resolveEffectivePlanUseCase,
		getQuotaUsageUseCase,
	)
	failScanUseCase := scan.NewFailScanUseCase(&scanRepo, centrifugoPublisher)
	generatePresignedUploadUrlUseCase := scan.NewGeneratePresignedUploadUrlUseCase(
		storageClient,
		&scanRepo,
		&mediaRepo,
	)
	getScansUseCase := scan.NewGetScansUseCase(&scanRepo)
	getScanUseCase := scan.NewGetScanUseCase(&scanRepo)
	getStatisticsUseCase := scan.NewGetStatisticsUseCase(&scanRepo)
	getMediaByIDUseCase := media.NewGetMediaByIDUseCase(&mediaRepo)
	generateImageThumbnailUseCase := thumbnail.NewGenerateImageThumbnailUseCase()
	generateThumbnailUseCase := media.NewGenerateThumbnailUseCase(storageClient, &mediaRepo, generateImageThumbnailUseCase)
	publishMetadataUseCase := media.NewPublishMetadataUseCase(&mediaRepo, publisher, centrifugoPublisher, env)
	updateScanStatusUseCase := scan.NewUpdateScanStatusUseCase(&scanRepo)
	updateMediaStatusUseCase := media.NewUpdateMediaStatusUseCase(&mediaRepo, updateScanStatusUseCase)
	frameExtractor := video.NewFrameExtractor()
	processUploadedMediaUseCase := media.NewProcessUploadedMediaUseCase(
		storageClient,
		&mediaRepo,
		createMediaUseCase,
		generateThumbnailUseCase,
		updateMediaStatusUseCase,
		publishMetadataUseCase,
		assertUploadAllowedUseCase,
		failScanUseCase,
		frameExtractor,
		generateImageThumbnailUseCase,
	)

	getPlansUseCase := plan.NewGetPlansUseCase(&planRepo)
	checkoutSessionGateway := infraStripe.NewCheckoutSessionGateway(env)
	subscriptionGateway := infraStripe.NewSubscriptionGateway(env)
	createSubscriptionUseCase := subscription.NewCreateSubscriptionUseCase(
		&planRepo,
		fetchUserUseCase,
		checkoutSessionGateway,
	)
	billingPortalGateway := infraStripe.NewBillingPortalGateway(env)
	createBillingPortalUseCase := subscription.NewCreateBillingPortalUseCase(
		&subscriptionRepo,
		billingPortalGateway,
	)
	subscriptionNotifier := subscription.NewNotifier(&userRepo, &subscriptionRepo, centrifugoPublisher)
	checkoutCompletedUseCase := subscription.NewCheckoutCompletedUseCase(
		&userRepo,
		&planRepo,
		&subscriptionRepo,
		subscriptionGateway,
		subscriptionNotifier,
	)
	subscriptionUpdatedUseCase := subscription.NewSubscriptionUpdatedUseCase(
		&planRepo,
		&subscriptionRepo,
		subscriptionNotifier,
	)
	subscriptionDeletedUseCase := subscription.NewSubscriptionDeletedUseCase(
		&planRepo,
		&subscriptionRepo,
		subscriptionNotifier,
	)
	invoicePaymentSucceededUseCase := subscription.NewInvoicePaymentSucceededUseCase(
		&subscriptionRepo,
		subscriptionNotifier,
	)
	invoicePaymentFailedUseCase := subscription.NewInvoicePaymentFailedUseCase(
		&subscriptionRepo,
		subscriptionNotifier,
	)
	getUserSubscriptionUseCase := subscription.NewGetUserSubscriptionUseCase(
		&subscriptionRepo,
		resolveEffectivePlanUseCase,
		getQuotaUsageUseCase,
	)

	clerkMiddleware := middleware.NewClerkMiddleware(env.ClerkWebhookSecret)
	minIOMiddleware := middleware.NewMinIOMiddleware(env.MinIOWebhookSecret)
	stripeMiddleware := middleware.NewStripeMiddleware(env.StripeWebhookSecret)
	authenticateMiddleware := middleware.NewAuthenticateMiddleware(
		validateTokenUseCase,
		fetchUserUseCase,
		createUserUseCase,
		updateUserUseCase,
	)

	return &Container{
		AuthenticateMiddleware: authenticateMiddleware,
		ClerkMiddleware:        clerkMiddleware,
		MinIOMiddleware:        minIOMiddleware,
		StripeMiddleware:       stripeMiddleware,
		StripeHandler: handler.NewStripeHandler(
			checkoutCompletedUseCase,
			subscriptionUpdatedUseCase,
			subscriptionDeletedUseCase,
			invoicePaymentSucceededUseCase,
			invoicePaymentFailedUseCase,
		),
		ClerkHandler: handler.NewClerkHandler(
			getUserByClerkIDUseCase,
			createUserUseCase,
			updateUserUseCase,
			deleteUserByClerkIDUseCase,
		),
		MinIOHandler: handler.NewMinIOHandler(
			env.StorageBucket,
			processUploadedMediaUseCase,
		),
		UserHandler: handler.NewUserHandler(),
		ScanHandler: handler.NewScanHandler(
			generatePresignedUploadUrlUseCase,
			getScanUseCase,
			getScansUseCase,
			getStatisticsUseCase,
		),
		MediaHandler: handler.NewMediaHandler(
			storageClient,
			getMediaByIDUseCase,
		),
		RealtimeHandler: handler.NewRealtimeHandler(env),
		PlanHandler: handler.NewPlanHandler(
			getPlansUseCase,
		),
		SubscriptionHandler: handler.NewSubscriptionHandler(
			createSubscriptionUseCase,
			createBillingPortalUseCase,
			getUserSubscriptionUseCase,
		),
	}
}
