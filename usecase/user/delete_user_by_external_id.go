package user

import (
	"context"
	"errors"
	"go-api/domain/repository"
)

type DeleteUserByExternalIDUseCase struct {
	userRepo repository.UserRepository
}

func NewDeleteUserByExternalIDUseCase(userRepo repository.UserRepository) *DeleteUserByExternalIDUseCase {
	return &DeleteUserByExternalIDUseCase{userRepo: userRepo}
}

func (u *DeleteUserByExternalIDUseCase) Execute(ctx context.Context, externalID string) error {
	err := u.userRepo.DeleteByClerkID(ctx, externalID)
	if err != nil {
		return errors.New("failed to delete user by external ID")
	}

	return nil
}
