package handler

import (
	"encoding/json"
	"fmt"
	clerkdto "go-api/infrastructure/clerk"
	"go-api/usecase/user"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type ClerkHandler struct {
	getUserByClerkIDUseCase    *user.GetUserByClerkIDUseCase
	createUserUseCase          *user.CreateUserUseCase
	updateUserUseCase          *user.UpdateUserUseCase
	deleteUserByClerkIDUseCase *user.DeleteUserByClerkIDUseCase
}

func NewClerkHandler(
	getUserByClerkIDUseCase *user.GetUserByClerkIDUseCase,
	createUserUseCase *user.CreateUserUseCase,
	updateUserUseCase *user.UpdateUserUseCase,
	deleteUserByClerkIDUseCase *user.DeleteUserByClerkIDUseCase,
) *ClerkHandler {
	return &ClerkHandler{
		getUserByClerkIDUseCase:    getUserByClerkIDUseCase,
		createUserUseCase:          createUserUseCase,
		updateUserUseCase:          updateUserUseCase,
		deleteUserByClerkIDUseCase: deleteUserByClerkIDUseCase,
	}
}

func (h *ClerkHandler) Execute(c fiber.Ctx) error {
	clerkEvent := c.Locals("payload").(clerkdto.ClerkEvent)
	validate := validator.New()

	fmt.Printf("Clerk event type: %s", clerkEvent.Type)

	switch clerkEvent.Type {
	case "user.created":
		var data clerkdto.ClerkUserCreated
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
		var data clerkdto.ClerkUserUpdated
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
		var data clerkdto.ClerkUserDeleted
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
		log.Printf("Unhandled event type: %s", clerkEvent.Type)
		return c.SendStatus(fiber.StatusOK)
	}
}

func (h *ClerkHandler) CreateUser(c fiber.Ctx, data clerkdto.ClerkUserCreated) error {
	user, err := h.getUserByClerkIDUseCase.Execute(c.Context(), data.ID)
	if err != nil {
		log.Printf("Error finding user by Clerk ID %s: %v", data.ID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create user",
		})
	}

	if user != nil {
		log.Printf("User with Clerk ID %s already exists, skipping creation", data.ID)
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

		log.Printf("Successfully created user with Clerk ID %s", data.ID)
		return nil
	}

	return txFunc(nil)
}

func (h *ClerkHandler) UpdateUser(c fiber.Ctx, data clerkdto.ClerkUserUpdated) error {
	user, err := h.getUserByClerkIDUseCase.Execute(c.Context(), data.ID)
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

	log.Printf("Successfully updated user with Clerk ID %s", data.ID)
	return nil
}

func (h *ClerkHandler) DeleteUser(c fiber.Ctx, data clerkdto.ClerkUserDeleted) error {
	if err := h.deleteUserByClerkIDUseCase.Execute(c.Context(), data.ID); err != nil {
		log.Printf("Error deleting user with Clerk ID %s: %v", data.ID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to delete user",
		})
	}

	log.Printf("Successfully deleted user with Clerk ID %s", data.ID)
	return nil
}
