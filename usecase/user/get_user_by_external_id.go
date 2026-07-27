package user

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"
)

type GetUserByExternalIDUseCase struct {
	userRepo repository.UserRepository
}

func NewGetUserByExternalIDUseCase(userRepo repository.UserRepository) *GetUserByExternalIDUseCase {
	return &GetUserByExternalIDUseCase{userRepo: userRepo}
}

func (u *GetUserByExternalIDUseCase) Execute(ctx context.Context, externalID string) (*entity.User, error) {
	user, err := u.userRepo.GetByClerkID(ctx, externalID)
	if err != nil {
		return nil, errors.New("failed to get user by external ID")
	}

	return user, nil
}
