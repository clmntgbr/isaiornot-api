package wire

import (
	httphandler "go-api/handler/http"
	"go-api/handler/http/middleware"
	infraClerk "go-api/infrastructure/clerk"
	"go-api/usecase/auth"
	"go-api/usecase/clerk"
	"go-api/usecase/user"
)

type authBundle struct {
	fetchUserUseCase       *clerk.FetchUserUseCase
	authenticateMiddleware *middleware.AuthenticateMiddleware
	clerkHandler           *httphandler.ClerkHandler
}

func wireAuth(d *apiDeps) authBundle {
	validateTokenUseCase := auth.NewValidateTokenUseCase(d.jwksProvider, d.userRepo)
	fetchUserUseCase := clerk.NewFetchUserUseCase(infraClerk.NewUserGateway(d.env.ClerkSecretKey))
	getUserByClerkIDUseCase := user.NewGetUserByClerkIDUseCase(d.userRepo)
	createFreeSubscriptionUseCase := subscriptionFree(d)
	createUserUseCase := user.NewCreateUserUseCase(d.userRepo, createFreeSubscriptionUseCase)
	updateUserUseCase := user.NewUpdateUserUseCase(d.userRepo)
	deleteUserByClerkIDUseCase := user.NewDeleteUserByClerkIDUseCase(d.userRepo)

	return authBundle{
		fetchUserUseCase: fetchUserUseCase,
		authenticateMiddleware: middleware.NewAuthenticateMiddleware(
			validateTokenUseCase,
			fetchUserUseCase,
			createUserUseCase,
			updateUserUseCase,
		),
		clerkHandler: httphandler.NewClerkHandler(
			getUserByClerkIDUseCase,
			createUserUseCase,
			updateUserUseCase,
			deleteUserByClerkIDUseCase,
		),
	}
}
