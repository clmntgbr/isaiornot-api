package write

import (
	"context"
	"errors"

	domainplan "go-api/internal/domain/plan"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type planWriteRepository struct {
	db *gorm.DB
}

func NewPlanWriteRepository(db *gorm.DB) domainplan.PlanWriteRepository {
	return &planWriteRepository{db: db}
}

func (r *planWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *planWriteRepository) Save(ctx context.Context, plan *domainplan.Plan) error {
	return DBWithContext(ctx, r.db).Create(planModelFromDomain(plan)).Error
}

func (r *planWriteRepository) Update(ctx context.Context, plan *domainplan.Plan) error {
	return DBWithContext(ctx, r.db).Save(planModelFromDomain(plan)).Error
}

func (r *planWriteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return DBWithContext(ctx, r.db).Delete(&PlanModel{}, "id = ?", id).Error
}

func (r *planWriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainplan.Plan, error) {
	var model PlanModel
	err := DBWithContext(ctx, r.db).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return planDomainFromModel(&model), nil
}

func (r *planWriteRepository) GetBySlug(ctx context.Context, slug string) (*domainplan.Plan, error) {
	var model PlanModel
	err := DBWithContext(ctx, r.db).Where("slug = ?", slug).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return planDomainFromModel(&model), nil
}

func (r *planWriteRepository) GetByStripePriceID(ctx context.Context, stripePriceID string) (*domainplan.Plan, error) {
	var model PlanModel
	err := DBWithContext(ctx, r.db).Where("stripe_price_id = ?", stripePriceID).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return planDomainFromModel(&model), nil
}

func planModelFromDomain(p *domainplan.Plan) *PlanModel {
	return &PlanModel{
		ID:              p.ID,
		Name:            p.Name,
		Description:     p.Description,
		Slug:            p.Slug,
		StripePriceID:   p.StripePriceID,
		IsActive:        p.IsActive,
		BillingInterval: string(p.BillingInterval),
		Price:           p.Price,
		Currency:        string(p.Currency),
		QuotaID:         p.QuotaID,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

func planDomainFromModel(model *PlanModel) *domainplan.Plan {
	return &domainplan.Plan{
		ID:              model.ID,
		Name:            model.Name,
		Description:     model.Description,
		Slug:            model.Slug,
		StripePriceID:   model.StripePriceID,
		IsActive:        model.IsActive,
		BillingInterval: domainplan.BillingInterval(model.BillingInterval),
		Price:           model.Price,
		Currency:        domainplan.Currency(model.Currency),
		QuotaID:         model.QuotaID,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
	}
}
