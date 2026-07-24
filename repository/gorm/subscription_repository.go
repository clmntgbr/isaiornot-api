package gorm

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type subscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) repository.SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

func (r *subscriptionRepository) Create(ctx context.Context, subscription *entity.Subscription) error {
	return dbWithContext(ctx, r.db).Omit(clause.Associations).Create(subscription).Error
}

func (r *subscriptionRepository) Update(ctx context.Context, subscription *entity.Subscription) error {
	return dbWithContext(ctx, r.db).Omit(clause.Associations).Save(subscription).Error
}

func (r *subscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error) {
	var subscription entity.Subscription
	err := dbWithContext(ctx, r.db).
		Preload("Plan").
		Preload("Plan.Quota").
		First(&subscription, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &subscription, nil
}

func (r *subscriptionRepository) GetByStripeSubscriptionID(ctx context.Context, stripeSubscriptionID string) (*entity.Subscription, error) {
	var subscription entity.Subscription
	err := dbWithContext(ctx, r.db).
		Preload("Plan").
		Preload("Plan.Quota").
		Where("stripe_subscription_id = ?", stripeSubscriptionID).
		First(&subscription).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &subscription, nil
}

func (r *subscriptionRepository) GetByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*entity.Subscription, error) {
	var subscription entity.Subscription
	err := dbWithContext(ctx, r.db).
		Preload("Plan").
		Preload("Plan.Quota").
		Where("stripe_customer_id = ?", stripeCustomerID).
		First(&subscription).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &subscription, nil
}
