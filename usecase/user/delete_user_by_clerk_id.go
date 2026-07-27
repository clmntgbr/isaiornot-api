package user

import (
	"context"
	"errors"
	"go-api/domain/repository"
)

type DeleteUserByClerkIDUseCase struct {
	userRepo repository.UserRepository
}

func NewDeleteUserByClerkIDUseCase(userRepo repository.UserRepository) *DeleteUserByClerkIDUseCase {
	return &DeleteUserByClerkIDUseCase{userRepo: userRepo}
}

func (u *DeleteUserByClerkIDUseCase) Execute(ctx context.Context, clerkID string) error {
	err := u.userRepo.DeleteByClerkID(ctx, clerkID)
	if err != nil {
		return errors.New("failed to delete user by clerk ID")
	}

	return nil
}
