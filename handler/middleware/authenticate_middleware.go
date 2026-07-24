package middleware

import (
	"go-api/handler/context"
	authdto "go-api/infrastructure/auth"
	"go-api/usecase/auth"
	"go-api/usecase/clerk"
	"go-api/usecase/user"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type AuthenticateMiddleware struct {
	validateTokenUseCase *auth.ValidateTokenUseCase
	fetchUserUseCase     *clerk.FetchUserUseCase
	createUserUseCase    *user.CreateUserUseCase
	updateUserUseCase    *user.UpdateUserUseCase
}

func NewAuthenticateMiddleware(
	validateTokenUseCase *auth.ValidateTokenUseCase,
	fetchUserUseCase *clerk.FetchUserUseCase,
	createUserUseCase *user.CreateUserUseCase,
	updateUserUseCase *user.UpdateUserUseCase,
) *AuthenticateMiddleware {
	return &AuthenticateMiddleware{
		validateTokenUseCase: validateTokenUseCase,
		fetchUserUseCase:     fetchUserUseCase,
		createUserUseCase:    createUserUseCase,
		updateUserUseCase:    updateUserUseCase,
	}
}

func (m *AuthenticateMiddleware) Protected() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Missing authorization header",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid authorization header format",
			})
		}

		if parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Authorization scheme must be Bearer",
			})
		}

		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Token cannot be empty",
			})
		}

		output, err := m.validateTokenUseCase.Execute(c.Context(), authdto.ValidateTokenInput{
			Token: tokenString,
		})

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid token",
			})
		}

		if output.User == nil {
			clerkUser, err := m.fetchUserUseCase.Execute(c.Context(), output.Claims.Subject)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"message": "Failed to get user",
				})
			}

			user, err := m.createUserUseCase.Execute(c.Context(), output.Claims.Subject, clerkUser.FirstName, clerkUser.LastName, clerkUser.Banned, clerkUser.Email)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"message": "Failed to create user",
				})
			}

			output.User = user
		}

		if output.User.Banned {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "User is banned",
			})
		}

		context.SetUser(c, *output.User)

		return c.Next()
	}
}
