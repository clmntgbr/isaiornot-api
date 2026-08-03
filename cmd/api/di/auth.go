package di

import (
	httphandler "go-api/handler/http"
	"go-api/handler/http/middleware"
	infraClerk "go-api/infrastructure/clerk"
	"go-api/usecase/auth"
	"go-api/usecase/identity"
	"go-api/usecase/user"
)

type authBundle struct {
	fetchUserUseCase       *identity.FetchUserUseCase
	authenticateMiddleware *middleware.AuthenticateMiddleware
	userWebhookHandler     *httphandler.UserWebhookHandler
}

func wireAuth(d *apiDeps) authBundle {
	validateTokenUseCase := auth.NewValidateTokenUseCase(d.jwksProvider, d.userRepo)
	fetchUserUseCase := identity.NewFetchUserUseCase(infraClerk.NewUserGateway(d.env.ClerkSecretKey))
	getUserByExternalIDUseCase := user.NewGetUserByExternalIDUseCase(d.userRepo)
	createFreeSubscriptionUseCase := subscriptionFree(d)
	createUserUseCase := user.NewCreateUserUseCase(d.userRepo, createFreeSubscriptionUseCase, d.centrifugoPublisher)
	updateUserUseCase := user.NewUpdateUserUseCase(d.userRepo, d.centrifugoPublisher)
	deleteUserByExternalIDUseCase := user.NewDeleteUserByExternalIDUseCase(d.userRepo)

	return authBundle{
		fetchUserUseCase: fetchUserUseCase,
		authenticateMiddleware: middleware.NewAuthenticateMiddleware(
			validateTokenUseCase,
			fetchUserUseCase,
			createUserUseCase,
			updateUserUseCase,
		),
		userWebhookHandler: httphandler.NewUserWebhookHandler(
			getUserByExternalIDUseCase,
			createUserUseCase,
			updateUserUseCase,
			deleteUserByExternalIDUseCase,
		),
	}
}
