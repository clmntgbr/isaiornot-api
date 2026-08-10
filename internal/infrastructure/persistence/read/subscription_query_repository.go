package read

import (
	"context"
	"errors"
	"time"

	domainplan "go-api/internal/domain/plan"
	domainsubscription "go-api/internal/domain/subscription"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type subscriptionRow struct {
	ID                   uuid.UUID
	PlanID               uuid.UUID
	StripeCustomerID     string
	StripeSubscriptionID string
	Status               string
	StartDate            time.Time
	EndDate              time.Time
	CancelAtPeriodEnd    bool
	QuotaPeriodStart     time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (subscriptionRow) TableName() string { return "subscriptions" }

type subscriptionReadRepository struct {
	db       *gorm.DB
	planRepo domainplan.PlanReadRepository
}

func NewSubscriptionReadRepository(
	db *gorm.DB,
	planRepo domainplan.PlanReadRepository,
) domainsubscription.SubscriptionReadRepository {
	return &subscriptionReadRepository{db: db, planRepo: planRepo}
}

func (r *subscriptionReadRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainsubscription.SubscriptionView, error) {
	var row subscriptionRow
	err := r.subscriptionQuery(ctx).First(&row, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.toSubscriptionView(ctx, row)
}

func (r *subscriptionReadRepository) FindByStripeSubscriptionID(
	ctx context.Context,
	stripeSubscriptionID string,
) (*domainsubscription.SubscriptionView, error) {
	var row subscriptionRow
	err := r.subscriptionQuery(ctx).
		Where("stripe_subscription_id = ?", stripeSubscriptionID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.toSubscriptionView(ctx, row)
}

func (r *subscriptionReadRepository) FindByStripeCustomerID(
	ctx context.Context,
	stripeCustomerID string,
) (*domainsubscription.SubscriptionView, error) {
	var row subscriptionRow
	err := r.subscriptionQuery(ctx).
		Where("stripe_customer_id = ?", stripeCustomerID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.toSubscriptionView(ctx, row)
}

func (r *subscriptionReadRepository) subscriptionQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Select(
		"id",
		"plan_id",
		"stripe_customer_id",
		"stripe_subscription_id",
		"status",
		"start_date",
		"end_date",
		"cancel_at_period_end",
		"quota_period_start",
		"created_at",
		"updated_at",
	)
}

func (r *subscriptionReadRepository) toSubscriptionView(
	ctx context.Context,
	row subscriptionRow,
) (*domainsubscription.SubscriptionView, error) {
	plan, err := r.planRepo.FindByID(ctx, row.PlanID)
	if err != nil {
		return nil, err
	}
	return &domainsubscription.SubscriptionView{
		ID:                   row.ID,
		PlanID:               row.PlanID,
		Plan:                 plan,
		StripeCustomerID:     row.StripeCustomerID,
		StripeSubscriptionID: row.StripeSubscriptionID,
		Status:               domainsubscription.Status(row.Status),
		StartDate:            row.StartDate,
		EndDate:              row.EndDate,
		CancelAtPeriodEnd:    row.CancelAtPeriodEnd,
		QuotaPeriodStart:     row.QuotaPeriodStart,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
	}, nil
}
