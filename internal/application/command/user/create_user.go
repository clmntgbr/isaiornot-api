package user

import (
	"context"
	"errors"

	"go-api/internal/domain/plan"
	"go-api/internal/domain/port"
	domainsubscription "go-api/internal/domain/subscription"
	domainuser "go-api/internal/domain/user"
)

type CreateUserCommand struct {
	ClerkID   string
	FirstName string
	LastName  string
	Email     string
	Banned    bool
}

type CreateUserHandler struct {
	repo             domainuser.UserWriteRepository
	planRepo         plan.PlanWriteRepository
	subscriptionRepo domainsubscription.SubscriptionWriteRepository
	outbox           port.OutboxRepository
}

func NewCreateUserHandler(
	repo domainuser.UserWriteRepository,
	planRepo plan.PlanWriteRepository,
	subscriptionRepo domainsubscription.SubscriptionWriteRepository,
	outbox port.OutboxRepository,
) *CreateUserHandler {
	return &CreateUserHandler{
		repo:             repo,
		planRepo:         planRepo,
		subscriptionRepo: subscriptionRepo,
		outbox:           outbox,
	}
}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd CreateUserCommand) (*domainuser.User, error) {
	u := domainuser.NewUser(cmd.ClerkID, cmd.FirstName, cmd.LastName, cmd.Email, cmd.Banned)

	err := h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.repo.Save(txCtx, u); err != nil {
			return err
		}

		freePlan, err := h.planRepo.GetBySlug(txCtx, plan.FreePlanSlug)
		if err != nil {
			return err
		}
		if freePlan == nil {
			return errors.New("free plan not found")
		}

		sub := domainsubscription.NewFreeSubscription(freePlan.ID)
		if err := h.subscriptionRepo.Save(txCtx, sub); err != nil {
			return err
		}

		u.AssignSubscription(sub.ID)
		if err := h.repo.Update(txCtx, u); err != nil {
			return err
		}

		events := append(u.PullEvents(), sub.PullEvents()...)
		return h.outbox.StoreEvents(txCtx, events)
	})
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	return u, nil
}
