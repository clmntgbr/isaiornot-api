package http

import (
	"go-api/domain/paginate"
	"go-api/handler/http/context"
	"go-api/presenter"
	"go-api/usecase/subscription"

	"github.com/gofiber/fiber/v3"
)

type InvoiceHandler struct {
	getInvoicesUseCase *subscription.GetInvoicesUseCase
}

func NewInvoiceHandler(getInvoicesUseCase *subscription.GetInvoicesUseCase) *InvoiceHandler {
	return &InvoiceHandler{getInvoicesUseCase: getInvoicesUseCase}
}

func (h *InvoiceHandler) GetInvoices(c fiber.Ctx) error {
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

	orderBy := query.OrderBy
	sortBy := query.SortBy
	query.Normalize()
	if sortBy == "" {
		query.SortBy = "stripe_created_at"
	}
	if orderBy == "" {
		query.OrderBy = paginate.OrderByDesc
	}

	invoices, total, err := h.getInvoicesUseCase.Execute(c.Context(), user.ID, query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
			"errors":  err.Error(),
		})
	}

	return c.JSON(paginate.NewPaginateResponse(
		presenter.NewInvoiceResponses(invoices),
		int(total),
		query,
	))
}
