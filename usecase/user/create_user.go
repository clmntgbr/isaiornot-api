package user

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"
	"go-api/usecase/subscription"
)

type CreateUserUseCase struct {
	userRepo                      *repository.UserRepository
	createFreeSubscriptionUseCase *subscription.CreateFreeSubscriptionUseCase
}

func NewCreateUserUseCase(
	userRepo *repository.UserRepository,
	createFreeSubscriptionUseCase *subscription.CreateFreeSubscriptionUseCase,
) *CreateUserUseCase {
	return &CreateUserUseCase{
		userRepo:                      userRepo,
		createFreeSubscriptionUseCase: createFreeSubscriptionUseCase,
	}
}

func (u *CreateUserUseCase) Execute(ctx context.Context, clerkID string, firstName string, lastName string, banned bool, email string) (*entity.User, error) {
	user := entity.User{
		ClerkID:   clerkID,
		FirstName: firstName,
		LastName:  lastName,
		Banned:    banned,
		Email:     email,
	}

	err := (*u.userRepo).Create(ctx, &user)
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	if _, err := u.createFreeSubscriptionUseCase.Execute(ctx, &user); err != nil {
		return nil, err
	}

	return &user, nil
}
