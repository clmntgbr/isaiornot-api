package http

import (
	mediadto "go-api/domain/media"
	"go-api/handler/http/context"
	"go-api/infrastructure/storage"
	"go-api/usecase/media"
	"io"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type MediaHandler struct {
	storage             *storage.MinIOStorage
	getMediaByIDUseCase *media.GetMediaByIDUseCase
}

func NewMediaHandler(
	storage *storage.MinIOStorage,
	getMediaByIDUseCase *media.GetMediaByIDUseCase,
) *MediaHandler {
	return &MediaHandler{
		storage:             storage,
		getMediaByIDUseCase: getMediaByIDUseCase,
	}
}

func (h *MediaHandler) GetThumbnail(c fiber.Ctx) error {
	user, err := context.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	mediaID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	mediaEntity, err := h.getMediaByIDUseCase.Execute(c.Context(), user.ID, mediaID)
	if err != nil || mediaEntity.Thumbnail == "" {
		return c.SendStatus(fiber.StatusNotFound)
	}

	reader, err := h.storage.GetThumbnail(c.Context(), mediadto.NewThumbnailObjectKey(user.ID, mediaEntity.ID))
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	body, err := io.ReadAll(reader)
	reader.Close()
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	c.Set("Content-Type", "image/jpeg")
	c.Set("Cache-Control", "public, max-age=86400")
	c.Set("Content-Length", strconv.Itoa(len(body)))

	return c.Send(body)
}
