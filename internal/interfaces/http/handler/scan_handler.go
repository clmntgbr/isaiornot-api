package handler

import (
	"errors"

	cmdscan "go-api/internal/application/command/scan"
	cmdsubscription "go-api/internal/application/command/subscription"
	queryscan "go-api/internal/application/query/scan"
	querysubscription "go-api/internal/application/query/subscription"
	"go-api/internal/domain/paginate"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/validation"

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
	if err := validation.BindAndValidateBody(c, &req); err != nil {
		return err
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
		if errors.Is(err, querysubscription.ErrSubscriptionNotFound) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "No active subscription found",
			})
		}
		if errors.Is(err, cmdsubscription.ErrImageQuotaExceeded) ||
			errors.Is(err, cmdsubscription.ErrVideoQuotaExceeded) ||
			errors.Is(err, cmdsubscription.ErrImagesNotAllowed) ||
			errors.Is(err, cmdsubscription.ErrVideosNotAllowed) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": err.Error(),
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
	if err := validation.BindAndValidateQuery(c, &query); err != nil {
		return err
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
		if errors.Is(err, querysubscription.ErrSubscriptionNotFound) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "No active subscription found",
			})
		}
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
		if errors.Is(err, queryscan.ErrScanNotFound) ||
			errors.Is(err, queryscan.ErrHistoryOutsideRetention) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Scan not found",
			})
		}
		if errors.Is(err, querysubscription.ErrSubscriptionNotFound) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "No active subscription found",
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
		if errors.Is(err, querysubscription.ErrSubscriptionNotFound) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "No active subscription found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get scan statistics",
		})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewScanStatisticsResponse(stats))
}
