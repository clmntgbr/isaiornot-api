package subscription

import (
	"context"
	"errors"

	"go-api/domain/entity"
	"go-api/domain/paginate"
	"go-api/domain/repository"

	"github.com/google/uuid"
)

type GetInvoicesUseCase struct {
	invoiceRepo repository.InvoiceRepository
}

func NewGetInvoicesUseCase(invoiceRepo repository.InvoiceRepository) *GetInvoicesUseCase {
	return &GetInvoicesUseCase{invoiceRepo: invoiceRepo}
}

func (u *GetInvoicesUseCase) Execute(
	ctx context.Context,
	userID uuid.UUID,
	query paginate.PaginateQuery,
) ([]*entity.Invoice, int64, error) {
	orderBy := query.OrderBy
	sortBy := query.SortBy
	query.Normalize()

	if sortBy == "" {
		query.SortBy = "stripe_created_at"
	}
	if orderBy == "" {
		query.OrderBy = paginate.OrderByDesc
	}

	invoices, total, err := u.invoiceRepo.GetByUserID(ctx, userID, query)
	if err != nil {
		return nil, 0, errors.New("failed to list invoices")
	}

	return invoices, total, nil
}
