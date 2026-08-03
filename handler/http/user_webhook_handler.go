package http

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"go-api/handler/http/dto"
	"go-api/usecase/user"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type UserWebhookHandler struct {
	getUserByExternalIDUseCase    *user.GetUserByExternalIDUseCase
	createUserUseCase             *user.CreateUserUseCase
	updateUserUseCase             *user.UpdateUserUseCase
	deleteUserByExternalIDUseCase *user.DeleteUserByExternalIDUseCase
}

func NewUserWebhookHandler(
	getUserByExternalIDUseCase *user.GetUserByExternalIDUseCase,
	createUserUseCase *user.CreateUserUseCase,
	updateUserUseCase *user.UpdateUserUseCase,
	deleteUserByExternalIDUseCase *user.DeleteUserByExternalIDUseCase,
) *UserWebhookHandler {
	return &UserWebhookHandler{
		getUserByExternalIDUseCase:    getUserByExternalIDUseCase,
		createUserUseCase:             createUserUseCase,
		updateUserUseCase:             updateUserUseCase,
		deleteUserByExternalIDUseCase: deleteUserByExternalIDUseCase,
	}
}

func (h *UserWebhookHandler) Execute(c fiber.Ctx) error {
	clerkEvent := c.Locals("payload").(dto.ClerkEvent)
	validate := validator.New()

	switch clerkEvent.Type {
	case "user.created":
		var data dto.ClerkUserCreated
		if err := json.Unmarshal(clerkEvent.Data, &data); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid request body",
			})
		}

		if err := validate.Struct(data); err != nil {
			return err
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := h.createUser(ctx, data); err != nil {
				log.Printf("failed to handle clerk user.created %s: %v", data.ID, err)
			}
		}()

		return c.SendStatus(fiber.StatusOK)

	case "user.updated":
		var data dto.ClerkUserUpdated
		if err := json.Unmarshal(clerkEvent.Data, &data); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid request body",
			})
		}

		if err := validate.Struct(data); err != nil {
			return err
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := h.updateUser(ctx, data); err != nil {
				log.Printf("failed to handle clerk user.updated %s: %v", data.ID, err)
			}
		}()

		return c.SendStatus(fiber.StatusOK)

	case "user.deleted":
		var data dto.ClerkUserDeleted
		if err := json.Unmarshal(clerkEvent.Data, &data); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid request body",
			})
		}

		if err := validate.Struct(data); err != nil {
			return err
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := h.deleteUser(ctx, data); err != nil {
				log.Printf("failed to handle clerk user.deleted %s: %v", data.ID, err)
			}
		}()

		return c.SendStatus(fiber.StatusOK)

	default:
		return c.SendStatus(fiber.StatusOK)
	}
}

func (h *UserWebhookHandler) createUser(ctx context.Context, data dto.ClerkUserCreated) error {
	existing, err := h.getUserByExternalIDUseCase.Execute(ctx, data.ID)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}

	banned := false
	if data.Banned != nil {
		banned = *data.Banned
	}

	_, err = h.createUserUseCase.Execute(ctx, data.ID, data.FirstName, data.LastName, banned, data.Email)
	return err
}

func (h *UserWebhookHandler) updateUser(ctx context.Context, data dto.ClerkUserUpdated) error {
	user, err := h.getUserByExternalIDUseCase.Execute(ctx, data.ID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	user.FirstName = data.FirstName
	user.LastName = data.LastName
	if data.Banned != nil {
		user.Banned = *data.Banned
	}
	if data.Email != "" {
		user.Email = data.Email
	}

	return h.updateUserUseCase.Execute(ctx, user)
}

func (h *UserWebhookHandler) deleteUser(ctx context.Context, data dto.ClerkUserDeleted) error {
	return h.deleteUserByExternalIDUseCase.Execute(ctx, data.ID)
}
