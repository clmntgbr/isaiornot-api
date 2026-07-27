package http

import (
	"encoding/json"
	"go-api/handler/http/dto"
	"go-api/usecase/user"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
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

		if err := h.CreateUser(c, data); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to create user",
			})
		}

		return c.SendStatus(fiber.StatusCreated)

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

		if err := h.UpdateUser(c, data); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to update user",
			})
		}

		return c.SendStatus(fiber.StatusNoContent)

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

		if err := h.DeleteUser(c, data); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to delete user",
			})
		}

		return c.SendStatus(fiber.StatusNoContent)

	default:
		return c.SendStatus(fiber.StatusOK)
	}
}

func (h *UserWebhookHandler) CreateUser(c fiber.Ctx, data dto.ClerkUserCreated) error {
	user, err := h.getUserByExternalIDUseCase.Execute(c.Context(), data.ID)
	if err != nil {
		log.Printf("Error finding user by Clerk ID %s: %v", data.ID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create user",
		})
	}

	if user != nil {
		return nil
	}

	txFunc := func(_ *gorm.DB) error {
		user, err = h.createUserUseCase.Execute(c.Context(), data.ID, data.FirstName, data.LastName, *data.Banned, data.Email)
		if err != nil {
			log.Printf("Error creating user with Clerk ID %s: %v", data.ID, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to create user",
			})
		}

		return nil
	}

	return txFunc(nil)
}

func (h *UserWebhookHandler) UpdateUser(c fiber.Ctx, data dto.ClerkUserUpdated) error {
	user, err := h.getUserByExternalIDUseCase.Execute(c.Context(), data.ID)
	if err != nil {
		log.Printf("Error finding user by Clerk ID %s: %v", data.ID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to find user",
		})
	}

	if user == nil {
		log.Printf("User with Clerk ID %s not found for update", data.ID)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	user.FirstName = data.FirstName
	user.LastName = data.LastName
	user.Banned = *data.Banned
	if data.Email != "" {
		user.Email = data.Email
	}

	if err := h.updateUserUseCase.Execute(c.Context(), user); err != nil {
		log.Printf("Error updating user with Clerk ID %s: %v", data.ID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update user",
		})
	}

	return nil
}

func (h *UserWebhookHandler) DeleteUser(c fiber.Ctx, data dto.ClerkUserDeleted) error {
	if err := h.deleteUserByExternalIDUseCase.Execute(c.Context(), data.ID); err != nil {
		log.Printf("Error deleting user with Clerk ID %s: %v", data.ID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to delete user",
		})
	}

	return nil
}
