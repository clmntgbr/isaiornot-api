package read

import (
	"context"
	"errors"
	"time"

	domainplan "go-api/internal/domain/plan"
	domainquota "go-api/internal/domain/quota"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type planRow struct {
	ID              uuid.UUID
	Name            string
	Description     string
	Slug            string
	StripePriceID   string
	IsActive        bool
	BillingInterval string
	Price           float64
	Currency        string
	QuotaID         uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (planRow) TableName() string { return "plans" }

type planReadRepository struct {
	db        *gorm.DB
	quotaRepo domainquota.QuotaReadRepository
}

func NewPlanReadRepository(db *gorm.DB, quotaRepo domainquota.QuotaReadRepository) domainplan.PlanReadRepository {
	return &planReadRepository{db: db, quotaRepo: quotaRepo}
}

func (r *planReadRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainplan.PlanView, error) {
	var row planRow
	err := r.planQuery(ctx).First(&row, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.toPlanView(ctx, row)
}

func (r *planReadRepository) FindBySlug(ctx context.Context, slug string) (*domainplan.PlanView, error) {
	var row planRow
	err := r.planQuery(ctx).Where("slug = ?", slug).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.toPlanView(ctx, row)
}

func (r *planReadRepository) FindByStripePriceID(ctx context.Context, stripePriceID string) (*domainplan.PlanView, error) {
	var row planRow
	err := r.planQuery(ctx).Where("stripe_price_id = ?", stripePriceID).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.toPlanView(ctx, row)
}

func (r *planReadRepository) FindAll(ctx context.Context) ([]*domainplan.PlanView, error) {
	var rows []planRow
	err := r.planQuery(ctx).Order("created_at ASC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return r.toPlanViews(ctx, rows)
}

func (r *planReadRepository) FindActive(ctx context.Context) ([]*domainplan.PlanView, error) {
	var rows []planRow
	err := r.planQuery(ctx).
		Where("is_active = ?", true).
		Order("price ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return r.toPlanViews(ctx, rows)
}

func (r *planReadRepository) planQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Select(
		"id",
		"name",
		"description",
		"slug",
		"stripe_price_id",
		"is_active",
		"billing_interval",
		"price",
		"currency",
		"quota_id",
		"created_at",
		"updated_at",
	)
}

func (r *planReadRepository) toPlanViews(ctx context.Context, rows []planRow) ([]*domainplan.PlanView, error) {
	quotaByID, err := r.loadQuotas(ctx, rows)
	if err != nil {
		return nil, err
	}
	views := make([]*domainplan.PlanView, 0, len(rows))
	for _, row := range rows {
		views = append(views, planViewFromRow(row, quotaByID[row.QuotaID]))
	}
	return views, nil
}

func (r *planReadRepository) toPlanView(ctx context.Context, row planRow) (*domainplan.PlanView, error) {
	quota, err := r.quotaRepo.FindByID(ctx, row.QuotaID)
	if err != nil {
		return nil, err
	}
	return planViewFromRow(row, quota), nil
}

func (r *planReadRepository) loadQuotas(ctx context.Context, rows []planRow) (map[uuid.UUID]*domainquota.QuotaView, error) {
	quotas, err := r.quotaRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	quotaByID := make(map[uuid.UUID]*domainquota.QuotaView, len(quotas))
	for _, q := range quotas {
		quotaByID[q.ID] = q
	}
	needed := make(map[uuid.UUID]struct{}, len(rows))
	for _, row := range rows {
		needed[row.QuotaID] = struct{}{}
	}
	result := make(map[uuid.UUID]*domainquota.QuotaView, len(needed))
	for id := range needed {
		if q, ok := quotaByID[id]; ok {
			result[id] = q
		}
	}
	return result, nil
}

func planViewFromRow(row planRow, quota *domainquota.QuotaView) *domainplan.PlanView {
	return &domainplan.PlanView{
		ID:              row.ID,
		Name:            row.Name,
		Description:     row.Description,
		Slug:            row.Slug,
		StripePriceID:   row.StripePriceID,
		IsActive:        row.IsActive,
		BillingInterval: domainplan.BillingInterval(row.BillingInterval),
		Price:           row.Price,
		Currency:        domainplan.Currency(row.Currency),
		QuotaID:         row.QuotaID,
		Quota:           quota,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}
