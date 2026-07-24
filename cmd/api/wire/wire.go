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
	"go-api/usecase/analysis"
	"go-api/usecase/auth"
	"go-api/usecase/clerk"
	"go-api/usecase/media"
	"go-api/usecase/plan"
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
	AnalysisHandler        *handler.AnalysisHandler
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
	log.Println("🚀 JWKS provider created")

	storageClient, err := storage.NewMinIOStorage(env)
	if err != nil {
		log.Fatalf("failed to create storage client: %v", err)
	}
	log.Println("🚀 Storage client created")

	userRepo := repoGorm.NewUserRepository(db)
	mediaRepo := repoGorm.NewMediaRepository(db)
	analysisRepo := repoGorm.NewAnalysisRepository(db)
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

	createMediaUseCase := media.NewCreateMediaUseCase(&analysisRepo, &mediaRepo)
	generatePresignedUploadUrlUseCase := analysis.NewGeneratePresignedUploadUrlUseCase(storageClient, &analysisRepo, &mediaRepo)
	getAnalysesUseCase := analysis.NewGetAnalysesUseCase(&analysisRepo)
	getAnalysisUseCase := analysis.NewGetAnalysisUseCase(&analysisRepo)
	getStatisticsUseCase := analysis.NewGetStatisticsUseCase(&analysisRepo)
	getMediaByIDUseCase := media.NewGetMediaByIDUseCase(&mediaRepo)
	generateImageThumbnailUseCase := thumbnail.NewGenerateImageThumbnailUseCase()
	generateThumbnailUseCase := media.NewGenerateThumbnailUseCase(storageClient, &mediaRepo, generateImageThumbnailUseCase)
	publishMetadataUseCase := media.NewPublishMetadataUseCase(&mediaRepo, publisher, centrifugoPublisher, env)
	updateMediaStatusUseCase := media.NewUpdateMediaStatusUseCase(&mediaRepo)
	frameExtractor := video.NewFrameExtractor()
	processUploadedMediaUseCase := media.NewProcessUploadedMediaUseCase(
		storageClient,
		&mediaRepo,
		createMediaUseCase,
		generateThumbnailUseCase,
		updateMediaStatusUseCase,
		publishMetadataUseCase,
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
	subscriptionNotifier := subscription.NewNotifier(&userRepo, &subscriptionRepo, centrifugoPublisher)
	handleCheckoutCompletedUseCase := subscription.NewHandleCheckoutCompletedUseCase(
		&userRepo,
		&planRepo,
		&subscriptionRepo,
		subscriptionGateway,
		subscriptionNotifier,
	)
	handleSubscriptionUpdatedUseCase := subscription.NewHandleSubscriptionUpdatedUseCase(
		&planRepo,
		&subscriptionRepo,
		subscriptionNotifier,
	)
	handleSubscriptionDeletedUseCase := subscription.NewHandleSubscriptionDeletedUseCase(
		&planRepo,
		&subscriptionRepo,
		subscriptionNotifier,
	)
	handleInvoicePaymentSucceededUseCase := subscription.NewHandleInvoicePaymentSucceededUseCase(
		&subscriptionRepo,
		subscriptionNotifier,
	)
	handleInvoicePaymentFailedUseCase := subscription.NewHandleInvoicePaymentFailedUseCase(
		&subscriptionRepo,
		subscriptionNotifier,
	)
	getUserSubscriptionUseCase := subscription.NewGetUserSubscriptionUseCase(&subscriptionRepo)

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
			handleCheckoutCompletedUseCase,
			handleSubscriptionUpdatedUseCase,
			handleSubscriptionDeletedUseCase,
			handleInvoicePaymentSucceededUseCase,
			handleInvoicePaymentFailedUseCase,
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
		UserHandler: handler.NewUserHandler(
			getUserSubscriptionUseCase,
		),
		AnalysisHandler: handler.NewAnalysisHandler(
			generatePresignedUploadUrlUseCase,
			getAnalysisUseCase,
			getAnalysesUseCase,
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
		),
	}
}
