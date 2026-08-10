package handler

import (
	"errors"

	cmdscan "go-api/internal/application/command/scan"
	queryscan "go-api/internal/application/query/scan"
	"go-api/internal/domain/paginate"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/presenter"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type ScanHandler struct {
	listScansHandler     *queryscan.ListScansHandler
	getScanByIDHandler   *queryscan.GetScanByIDHandler
	getStatisticsHandler *queryscan.GetStatisticsHandler
	presignUploadHandler *cmdscan.PresignUploadHandler
}

func NewScanHandler(
	listScansHandler *queryscan.ListScansHandler,
	getScanByIDHandler *queryscan.GetScanByIDHandler,
	getStatisticsHandler *queryscan.GetStatisticsHandler,
	presignUploadHandler *cmdscan.PresignUploadHandler,
) *ScanHandler {
	return &ScanHandler{
		listScansHandler:     listScansHandler,
		getScanByIDHandler:   getScanByIDHandler,
		getStatisticsHandler: getStatisticsHandler,
		presignUploadHandler: presignUploadHandler,
	}
}

func (h *ScanHandler) PresignUpload(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	var req dto.PresignUploadRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"errors":  err.Error(),
		})
	}

	result, err := h.presignUploadHandler.Handle(c.Context(), cmdscan.PresignUploadCommand{
		UserID:      user.ID,
		Filename:    req.Filename,
		ContentType: req.ContentType,
	})
	if err != nil {
		if errors.Is(err, cmdscan.ErrUnsupportedMediaType) {
			return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
				"message": "Unsupported media type",
				"errors":  err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to generate upload url",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(presenter.NewPresignUploadResponse(result))
}

func (h *ScanHandler) GetScans(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
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

	orderBy := query.OrderBy
	sortBy := query.SortBy
	query.Normalize()
	if sortBy == "" {
		query.SortBy = "created_at"
	}
	if orderBy == "" {
		query.OrderBy = paginate.OrderByDesc
	}

	scans, total, err := h.listScansHandler.Handle(c.Context(), queryscan.ListScansQuery{
		UserID: user.ID,
		Query:  query,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to list scans",
		})
	}

	return c.Status(fiber.StatusOK).JSON(paginate.NewPaginateResponse(
		presenter.NewScanResponses(scans),
		int(total),
		query,
	))
}

func (h *ScanHandler) GetScan(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	scanID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid scan id",
		})
	}

	view, err := h.getScanByIDHandler.Handle(c.Context(), queryscan.GetScanByIDQuery{
		UserID: user.ID,
		ScanID: scanID,
	})

	if err != nil {
		if errors.Is(err, queryscan.ErrScanNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Scan not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get scan",
		})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewScanResponse(view))
}

func (h *ScanHandler) GetStatistics(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	stats, err := h.getStatisticsHandler.Handle(c.Context(), queryscan.GetStatisticsQuery{
		UserID: user.ID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get scan statistics",
		})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewScanStatisticsResponse(stats))
}
