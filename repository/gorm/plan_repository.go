package gorm

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type planRepository struct {
	db *gorm.DB
}

func NewPlanRepository(db *gorm.DB) repository.PlanRepository {
	return &planRepository{db: db}
}

func (r *planRepository) Create(ctx context.Context, plan *entity.Plan) error {
	return dbWithContext(ctx, r.db).Create(plan).Error
}

func (r *planRepository) Update(ctx context.Context, plan *entity.Plan) error {
	return dbWithContext(ctx, r.db).Save(plan).Error
}

func (r *planRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return dbWithContext(ctx, r.db).Delete(&entity.Plan{}, id).Error
}

func (r *planRepository) GetAll(ctx context.Context) ([]*entity.Plan, error) {
	var plans []*entity.Plan
	err := dbWithContext(ctx, r.db).
		Preload("Quota").
		Find(&plans).Error
	if err != nil {
		return nil, err
	}
	return plans, nil
}

func (r *planRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Plan, error) {
	var plan entity.Plan
	err := dbWithContext(ctx, r.db).
		Preload("Quota").
		First(&plan, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &plan, nil
}

func (r *planRepository) GetByStripePriceID(ctx context.Context, stripePriceID string) (*entity.Plan, error) {
	var plan entity.Plan
	err := dbWithContext(ctx, r.db).
		Preload("Quota").
		Where("stripe_price_id = ?", stripePriceID).
		First(&plan).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &plan, nil
}

func (r *planRepository) GetBySlug(ctx context.Context, slug string) (*entity.Plan, error) {
	var plan entity.Plan
	err := dbWithContext(ctx, r.db).
		Preload("Quota").
		Where("slug = ?", slug).
		First(&plan).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &plan, nil
}
