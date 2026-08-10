package handler

import (
	"errors"
	"io"
	"strconv"

	querymedia "go-api/internal/application/query/media"
	domainmedia "go-api/internal/domain/media"
	"go-api/internal/domain/port"
	httpctx "go-api/internal/interfaces/http/context"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type MediaHandler struct {
	getOwnedMediaHandler *querymedia.GetOwnedMediaHandler
	storage              port.Storage
}

func NewMediaHandler(
	getOwnedMediaHandler *querymedia.GetOwnedMediaHandler,
	storage port.Storage,
) *MediaHandler {
	return &MediaHandler{
		getOwnedMediaHandler: getOwnedMediaHandler,
		storage:              storage,
	}
}

func (h *MediaHandler) GetThumbnail(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	mediaID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	mediaView, err := h.getOwnedMediaHandler.Handle(c.Context(), querymedia.GetOwnedMediaQuery{
		UserID:  user.ID,
		MediaID: mediaID,
	})
	if err != nil {
		if errors.Is(err, querymedia.ErrThumbnailMediaNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	if mediaView.Thumbnail == "" {
		return c.SendStatus(fiber.StatusNotFound)
	}

	reader, err := h.storage.GetThumbnail(
		c.Context(),
		domainmedia.NewThumbnailObjectKey(user.ID, mediaView.ScanID, mediaView.ID),
	)
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	c.Set("Content-Type", "image/jpeg")
	c.Set("Cache-Control", "public, max-age=86400")
	c.Set("Content-Length", strconv.Itoa(len(body)))

	return c.Send(body)
}
