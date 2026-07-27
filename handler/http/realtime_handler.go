package http

import (
	"go-api/handler/http/context"
	"go-api/infrastructure/centrifugo"
	"go-api/infrastructure/config"

	"github.com/gofiber/fiber/v3"
)

type RealtimeHandler struct {
	env *config.Config
}

func NewRealtimeHandler(env *config.Config) *RealtimeHandler {
	return &RealtimeHandler{env: env}
}

func (h *RealtimeHandler) GetConnection(c fiber.Ctx) error {
	user, err := context.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "unauthorized",
		})
	}

	info, err := centrifugo.NewConnectionInfo(h.env, user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to create realtime connection",
		})
	}

	return c.JSON(info)
}
