package di

import (
	"log"

	authcmd "go-api/internal/application/command/auth"
	identitycmd "go-api/internal/application/command/identity"
	mediacmd "go-api/internal/application/command/media"
	scancmd "go-api/internal/application/command/scan"
	cmdsubscription "go-api/internal/application/command/subscription"
	usercmd "go-api/internal/application/command/user"
	querymedia "go-api/internal/application/query/media"
	queryplan "go-api/internal/application/query/plan"
	queryscan "go-api/internal/application/query/scan"
	querysubscription "go-api/internal/application/query/subscription"
	queryuser "go-api/internal/application/query/user"
	queryinvoice "go-api/internal/application/query/invoice"
	infraClerk "go-api/internal/infrastructure/clerk"
	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/imaging"
	"go-api/internal/infrastructure/persistence/outbox"
	"go-api/internal/infrastructure/persistence/read"
	"go-api/internal/infrastructure/persistence/write"
	"go-api/internal/infrastructure/storage"
	infraStripe "go-api/internal/infrastructure/stripe"
	"go-api/internal/infrastructure/video"
	httphandler "go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/middleware"

	"gorm.io/gorm"
)

type Container struct {
	AuthenticateMiddleware       *middleware.AuthenticateMiddleware
	UserWebhookMiddleware        *middleware.UserWebhookMiddleware
	UserWebhookHandler           *httphandler.UserWebhookHandler
	MediaUploadWebhookMiddleware *middleware.MediaUploadWebhookMiddleware
	MediaUploadWebhookHandler    *httphandler.MediaUploadWebhookHandler
	BillingWebhookMiddleware     *middleware.BillingWebhookMiddleware
	BillingWebhookHandler        *httphandler.BillingWebhookHandler
	UserHandler                  *httphandler.UserHandler
	ScanHandler                  *httphandler.ScanHandler
	MediaHandler                 *httphandler.MediaHandler
	PlanHandler                  *httphandler.PlanHandler
	SubscriptionHandler          *httphandler.SubscriptionHandler
	InvoiceHandler               *httphandler.InvoiceHandler
	RealtimeHandler              *httphandler.RealtimeHandler
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	jwksProvider, err := infraClerk.NewJWKSProvider(env)
	if err != nil {
		log.Fatalf("failed to create JWKS provider: %v", err)
	}

	userWriteRepo := write.NewUserWriteRepository(db)
	userReadRepo := read.NewUserReadRepository(db)
	scanWriteRepo := write.NewScanWriteRepository(db)
	scanReadRepo := read.NewScanReadRepository(db)
	mediaWriteRepo := write.NewMediaWriteRepository(db)
	mediaReadRepo := read.NewMediaReadRepository(db)
	signalReadRepo := read.NewSignalReadRepository(db)
	quotaReadRepo := read.NewQuotaReadRepository(db)
	planWriteRepo := write.NewPlanWriteRepository(db)
	planReadRepo := read.NewPlanReadRepository(db, quotaReadRepo)
	subscriptionWriteRepo := write.NewSubscriptionWriteRepository(db)
	subscriptionReadRepo := read.NewSubscriptionReadRepository(db, planReadRepo)
	invoiceWriteRepo := write.NewInvoiceWriteRepository(db)
	invoiceReadRepo := read.NewInvoiceReadRepository(db)
	outboxRepo := outbox.NewRepository(db)

	minioStorage, err := storage.NewMinIOStorage(env)
	if err != nil {
		log.Fatalf("failed to create storage client: %v", err)
	}

	createUserHandler := usercmd.NewCreateUserHandler(
		userWriteRepo,
		planWriteRepo,
		subscriptionWriteRepo,
		outboxRepo,
	)
	updateUserHandler := usercmd.NewUpdateUserHandler(userWriteRepo, outboxRepo)
	getUserByExternalIDHandler := usercmd.NewGetUserByExternalIDHandler(userWriteRepo)
	deleteUserByExternalIDHandler := usercmd.NewDeleteUserByExternalIDHandler(userWriteRepo, outboxRepo)
	validateTokenHandler := authcmd.NewValidateTokenHandler(jwksProvider, userWriteRepo)
	fetchUserHandler := identitycmd.NewFetchUserHandler(infraClerk.NewUserGateway(env.ClerkSecretKey))
	getUserByIDHandler := queryuser.NewGetUserByIDHandler(userReadRepo)
	getCurrentSubscriptionHandler := querysubscription.NewGetCurrentSubscriptionHandler(
		userReadRepo,
		subscriptionReadRepo,
	)
	historyCutoff := queryscan.NewHistoryCutoffResolver(getCurrentSubscriptionHandler)
	listScansHandler := queryscan.NewListScansHandler(scanReadRepo, mediaReadRepo, signalReadRepo, historyCutoff)
	getScanByIDHandler := queryscan.NewGetScanByIDHandler(scanReadRepo, mediaReadRepo, signalReadRepo, historyCutoff)
	getStatisticsHandler := queryscan.NewGetStatisticsHandler(scanReadRepo, historyCutoff)
	getOwnedMediaHandler := querymedia.NewGetOwnedMediaHandler(mediaReadRepo, scanReadRepo, historyCutoff)
	listPlansHandler := queryplan.NewListPlansHandler(planReadRepo)
	getQuotaUsageHandler := querysubscription.NewGetQuotaUsageHandler(
		userReadRepo,
		subscriptionReadRepo,
		scanReadRepo,
	)
	subscriptionGateway := infraStripe.NewSubscriptionGateway(env)
	checkoutSessionGateway := infraStripe.NewCheckoutSessionGateway(env)
	billingPortalGateway := infraStripe.NewBillingPortalGateway(env)
	previewPlanChangeHandler := querysubscription.NewPreviewPlanChangeHandler(
		userReadRepo,
		planReadRepo,
		subscriptionReadRepo,
		subscriptionGateway,
	)
	createSubscriptionHandler := cmdsubscription.NewCreateSubscriptionHandler(
		userReadRepo,
		planReadRepo,
		subscriptionReadRepo,
		subscriptionWriteRepo,
		outboxRepo,
		fetchUserHandler,
		checkoutSessionGateway,
		subscriptionGateway,
	)
	createBillingPortalHandler := cmdsubscription.NewCreateBillingPortalHandler(
		userReadRepo,
		subscriptionReadRepo,
		billingPortalGateway,
	)
	checkoutCompletedHandler := cmdsubscription.NewCheckoutCompletedHandler(
		userWriteRepo,
		planWriteRepo,
		subscriptionWriteRepo,
		outboxRepo,
		subscriptionGateway,
	)
	subscriptionUpdatedHandler := cmdsubscription.NewSubscriptionUpdatedHandler(
		planWriteRepo,
		subscriptionWriteRepo,
		outboxRepo,
	)
	subscriptionDeletedHandler := cmdsubscription.NewSubscriptionDeletedHandler(
		planWriteRepo,
		subscriptionWriteRepo,
		outboxRepo,
	)
	invoicePaymentSucceededHandler := cmdsubscription.NewInvoicePaymentSucceededHandler(
		subscriptionWriteRepo,
		outboxRepo,
	)
	invoicePaymentFailedHandler := cmdsubscription.NewInvoicePaymentFailedHandler(
		subscriptionWriteRepo,
		outboxRepo,
	)
	upsertInvoiceHandler := cmdsubscription.NewUpsertInvoiceHandler(
		invoiceWriteRepo,
		subscriptionWriteRepo,
		userWriteRepo,
		outboxRepo,
	)
	listInvoicesHandler := queryinvoice.NewListInvoicesHandler(invoiceReadRepo)
	assertUploadAllowedHandler := cmdsubscription.NewAssertUploadAllowedHandler(getQuotaUsageHandler)
	generateThumbnailHandler := mediacmd.NewGenerateThumbnailHandler(
		mediaWriteRepo,
		scanWriteRepo,
		outboxRepo,
		minioStorage,
		imaging.NewThumbnailer(),
	)
	presignUploadHandler := scancmd.NewPresignUploadHandler(
		scanWriteRepo,
		mediaWriteRepo,
		outboxRepo,
		minioStorage,
		assertUploadAllowedHandler,
	)
	processUploadedMediaHandler := mediacmd.NewProcessUploadedMediaHandler(
		scanWriteRepo,
		mediaWriteRepo,
		outboxRepo,
		minioStorage,
		video.NewFrameExtractor(),
		imaging.NewThumbnailer(),
		generateThumbnailHandler,
		assertUploadAllowedHandler,
	)

	return &Container{
		AuthenticateMiddleware: middleware.NewAuthenticateMiddleware(
			validateTokenHandler,
			fetchUserHandler,
			createUserHandler,
		),
		UserWebhookMiddleware: middleware.NewUserWebhookMiddleware(env.ClerkWebhookSecret),
		UserWebhookHandler: httphandler.NewUserWebhookHandler(
			getUserByExternalIDHandler,
			createUserHandler,
			updateUserHandler,
			deleteUserByExternalIDHandler,
		),
		MediaUploadWebhookMiddleware: middleware.NewMediaUploadWebhookMiddleware(env.MinIOWebhookSecret),
		MediaUploadWebhookHandler: httphandler.NewMediaUploadWebhookHandler(
			env.StorageBucket,
			processUploadedMediaHandler,
		),
		BillingWebhookMiddleware: middleware.NewBillingWebhookMiddleware(env.StripeWebhookSecret),
		BillingWebhookHandler: httphandler.NewBillingWebhookHandler(
			checkoutCompletedHandler,
			subscriptionUpdatedHandler,
			subscriptionDeletedHandler,
			invoicePaymentSucceededHandler,
			invoicePaymentFailedHandler,
			upsertInvoiceHandler,
		),
		UserHandler: httphandler.NewUserHandler(getUserByIDHandler),
		ScanHandler: httphandler.NewScanHandler(
			listScansHandler,
			getScanByIDHandler,
			getStatisticsHandler,
			presignUploadHandler,
		),
		MediaHandler: httphandler.NewMediaHandler(getOwnedMediaHandler, minioStorage),
		PlanHandler:  httphandler.NewPlanHandler(listPlansHandler),
		SubscriptionHandler: httphandler.NewSubscriptionHandler(
			getCurrentSubscriptionHandler,
			getQuotaUsageHandler,
			previewPlanChangeHandler,
			createSubscriptionHandler,
			createBillingPortalHandler,
		),
		InvoiceHandler:  httphandler.NewInvoiceHandler(listInvoicesHandler),
		RealtimeHandler: httphandler.NewRealtimeHandler(env),
	}
}
