package di

import (
	"log"

	authcmd "go-api/internal/application/command/auth"
	identitycmd "go-api/internal/application/command/identity"
	mediacmd "go-api/internal/application/command/media"
	scancmd "go-api/internal/application/command/scan"
	usercmd "go-api/internal/application/command/user"
	queryscan "go-api/internal/application/query/scan"
	queryuser "go-api/internal/application/query/user"
	infraClerk "go-api/internal/infrastructure/clerk"
	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/persistence/outbox"
	"go-api/internal/infrastructure/persistence/read"
	"go-api/internal/infrastructure/persistence/write"
	"go-api/internal/infrastructure/storage"
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
	UserHandler                  *httphandler.UserHandler
	ScanHandler                  *httphandler.ScanHandler
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
	outboxRepo := outbox.NewRepository(db)

	minioStorage, err := storage.NewMinIOStorage(env)
	if err != nil {
		log.Fatalf("failed to create storage client: %v", err)
	}

	createUserHandler := usercmd.NewCreateUserHandler(userWriteRepo, outboxRepo)
	updateUserHandler := usercmd.NewUpdateUserHandler(userWriteRepo, outboxRepo)
	getUserByExternalIDHandler := usercmd.NewGetUserByExternalIDHandler(userWriteRepo)
	deleteUserByExternalIDHandler := usercmd.NewDeleteUserByExternalIDHandler(userWriteRepo, outboxRepo)
	validateTokenHandler := authcmd.NewValidateTokenHandler(jwksProvider, userWriteRepo)
	fetchUserHandler := identitycmd.NewFetchUserHandler(infraClerk.NewUserGateway(env.ClerkSecretKey))
	getUserByIDHandler := queryuser.NewGetUserByIDHandler(userReadRepo)
	listScansHandler := queryscan.NewListScansHandler(scanReadRepo, mediaReadRepo, signalReadRepo)
	getScanByIDHandler := queryscan.NewGetScanByIDHandler(scanReadRepo, mediaReadRepo, signalReadRepo)
	presignUploadHandler := scancmd.NewPresignUploadHandler(
		scanWriteRepo,
		mediaWriteRepo,
		outboxRepo,
		minioStorage,
	)
	processUploadedMediaHandler := mediacmd.NewProcessUploadedMediaHandler(
		scanWriteRepo,
		mediaWriteRepo,
		outboxRepo,
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
		UserHandler:     httphandler.NewUserHandler(getUserByIDHandler),
		ScanHandler:     httphandler.NewScanHandler(listScansHandler, getScanByIDHandler, presignUploadHandler),
		RealtimeHandler: httphandler.NewRealtimeHandler(env),
	}
}
