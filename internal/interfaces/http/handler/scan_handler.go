package handler

import (
	"errors"

	queryscan "go-api/internal/application/query/scan"
	"go-api/internal/domain/paginate"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/presenter"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type ScanHandler struct {
	listScansHandler   *queryscan.ListScansHandler
	getScanByIDHandler *queryscan.GetScanByIDHandler
}

func NewScanHandler(
	listScansHandler *queryscan.ListScansHandler,
	getScanByIDHandler *queryscan.GetScanByIDHandler,
) *ScanHandler {
	return &ScanHandler{
		listScansHandler:   listScansHandler,
		getScanByIDHandler: getScanByIDHandler,
	}
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
