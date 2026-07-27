package http

import (
	"errors"
	mediadto "go-api/domain/media"
	"go-api/domain/paginate"
	"go-api/handler/http/context"
	"go-api/presenter"
	"go-api/usecase/scan"
	"go-api/usecase/subscription"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type ScanHandler struct {
	generatePresignedUploadUrlUseCase *scan.GeneratePresignedUploadUrlUseCase
	getScanUseCase                    *scan.GetScanUseCase
	getScansUseCase                   *scan.GetScansUseCase
	getMediaStatisticsUseCase         *scan.GetMediaStatisticsUseCase
}

func NewScanHandler(
	generatePresignedUploadUrlUseCase *scan.GeneratePresignedUploadUrlUseCase,
	getScanUseCase *scan.GetScanUseCase,
	getScansUseCase *scan.GetScansUseCase,
	getMediaStatisticsUseCase *scan.GetMediaStatisticsUseCase,
) *ScanHandler {
	return &ScanHandler{
		generatePresignedUploadUrlUseCase: generatePresignedUploadUrlUseCase,
		getScanUseCase:                    getScanUseCase,
		getScansUseCase:                   getScansUseCase,
		getMediaStatisticsUseCase:         getMediaStatisticsUseCase,
	}
}

func (h *ScanHandler) GeneratePresignedUploadUrl(c fiber.Ctx) error {
	user, err := context.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	var request mediadto.PresignUploadInput
	if err := c.Bind().JSON(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"errors":  err.Error(),
		})
	}

	result, err := h.generatePresignedUploadUrlUseCase.Execute(c.Context(), user.ID, request)
	if err != nil {
		if errors.Is(err, scan.ErrUnsupportedMediaType) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Unsupported media type",
				"errors":  err.Error(),
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
			"errors":  err.Error(),
		})
	}

	return c.JSON(presenter.NewGeneratePresignedUploadUrlDetailResponse(result.URL, result.ScanID.String()))
}

func (h *ScanHandler) GetScans(c fiber.Ctx) error {
	user, err := context.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	var query paginate.PaginateQuery
	if err := c.Bind().Query(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid query parameters",
			"errors":  err.Error(),
		})
	}
	query.Normalize()

	scans, total, err := h.getScansUseCase.Execute(c.Context(), user.ID, query)
	if err != nil {
		if errors.Is(err, subscription.ErrSubscriptionNotFound) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "No active subscription found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
			"errors":  err.Error(),
		})
	}

	return c.JSON(paginate.NewPaginateResponse(presenter.NewScanListResponses(scans), int(total), query))
}

func (h *ScanHandler) GetScan(c fiber.Ctx) error {
	user, err := context.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	scanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	result, err := h.getScanUseCase.Execute(c.Context(), user.ID, scanID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Scan not found",
		})
	}

	return c.JSON(presenter.NewScanDetailResponse(result))
}

func (h *ScanHandler) GetStatistics(c fiber.Ctx) error {
	user, err := context.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	stats, err := h.getMediaStatisticsUseCase.Execute(c.Context(), user.ID)
	if err != nil {
		if errors.Is(err, subscription.ErrSubscriptionNotFound) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "No active subscription found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
			"errors":  err.Error(),
		})
	}

	return c.JSON(presenter.NewMediaStatisticsResponse(stats))
}
