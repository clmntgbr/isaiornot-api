package repository

import (
	"context"

	"go-api/domain/entity"
	"go-api/domain/paginate"

	"github.com/google/uuid"
)

type InvoiceRepository interface {
	UpsertByStripeInvoiceID(ctx context.Context, invoice *entity.Invoice) error
	GetByStripeInvoiceID(ctx context.Context, stripeInvoiceID string) (*entity.Invoice, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, query paginate.PaginateQuery) ([]*entity.Invoice, int64, error)
}
