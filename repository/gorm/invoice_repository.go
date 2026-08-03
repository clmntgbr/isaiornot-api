package gorm

import (
	"context"
	"errors"

	"go-api/domain/entity"
	"go-api/domain/paginate"
	"go-api/domain/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type invoiceRepository struct {
	db *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) repository.InvoiceRepository {
	return &invoiceRepository{db: db}
}

func (r *invoiceRepository) GetByStripeInvoiceID(ctx context.Context, stripeInvoiceID string) (*entity.Invoice, error) {
	var invoice entity.Invoice
	err := dbWithContext(ctx, r.db).Where("stripe_invoice_id = ?", stripeInvoiceID).First(&invoice).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &invoice, nil
}

func (r *invoiceRepository) UpsertByStripeInvoiceID(ctx context.Context, invoice *entity.Invoice) error {
	existing, err := r.GetByStripeInvoiceID(ctx, invoice.StripeInvoiceID)
	if err != nil {
		return err
	}

	if existing == nil {
		return dbWithContext(ctx, r.db).Omit(clause.Associations).Create(invoice).Error
	}

	invoice.ID = existing.ID
	invoice.CreatedAt = existing.CreatedAt
	return dbWithContext(ctx, r.db).Omit(clause.Associations).Save(invoice).Error
}

func (r *invoiceRepository) GetByUserID(ctx context.Context, userID uuid.UUID, query paginate.PaginateQuery) ([]*entity.Invoice, int64, error) {
	db := dbWithContext(ctx, r.db).Model(&entity.Invoice{}).Where("user_id = ?", userID)

	if query.SortBy == "" {
		query.SortBy = "stripe_created_at"
	}
	if query.OrderBy == "" {
		query.OrderBy = paginate.OrderByDesc
	}

	db, total, err := Paginate(db, query)
	if err != nil {
		return nil, 0, err
	}

	var invoices []*entity.Invoice
	if err := db.Find(&invoices).Error; err != nil {
		return nil, 0, err
	}

	return invoices, total, nil
}
