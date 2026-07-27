package user

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"
)

type GetUserByClerkIDUseCase struct {
	userRepo repository.UserRepository
}

func NewGetUserByClerkIDUseCase(userRepo repository.UserRepository) *GetUserByClerkIDUseCase {
	return &GetUserByClerkIDUseCase{userRepo: userRepo}
}

func (u *GetUserByClerkIDUseCase) Execute(ctx context.Context, clerkID string) (*entity.User, error) {
	user, err := u.userRepo.GetByClerkID(ctx, clerkID)
	if err != nil {
		return nil, errors.New("failed to get user by clerk ID")
	}

	return user, nil
}
