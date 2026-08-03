package user

import (
	"context"
	"errors"
	"log"

	"go-api/domain/entity"
	"go-api/domain/port"
	"go-api/domain/realtime"
	"go-api/domain/repository"
)

type UpdateUserUseCase struct {
	userRepo            repository.UserRepository
	centrifugoPublisher port.RealtimePublisher
}

func NewUpdateUserUseCase(
	userRepo repository.UserRepository,
	centrifugoPublisher port.RealtimePublisher,
) *UpdateUserUseCase {
	return &UpdateUserUseCase{
		userRepo:            userRepo,
		centrifugoPublisher: centrifugoPublisher,
	}
}

func (s *UpdateUserUseCase) Execute(ctx context.Context, user *entity.User) error {
	err := s.userRepo.Update(ctx, user)
	if err != nil {
		return errors.New("failed to update user")
	}

	event, err := realtime.NewUserUpdatedEvent(user)
	if err != nil {
		return errors.New("failed to build user updated event")
	}

	if err := s.centrifugoPublisher.PublishUserToUser(ctx, user.ID, event); err != nil {
		log.Printf("failed to publish user_updated to centrifugo for user %s: %v", user.ID, err)
	}

	return nil
}
