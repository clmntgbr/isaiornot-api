package user

import (
	"context"
	"errors"
	"log"
	"time"

	"go-api/domain/entity"
	"go-api/domain/port"
	"go-api/domain/realtime"
	"go-api/domain/repository"
	"go-api/usecase/subscription"

	"github.com/google/uuid"
)

const userCreatedEventDelay = 5 * time.Second

type CreateUserUseCase struct {
	userRepo                      repository.UserRepository
	createFreeSubscriptionUseCase *subscription.CreateFreeSubscriptionUseCase
	centrifugoPublisher           port.RealtimePublisher
}

func NewCreateUserUseCase(
	userRepo repository.UserRepository,
	createFreeSubscriptionUseCase *subscription.CreateFreeSubscriptionUseCase,
	centrifugoPublisher port.RealtimePublisher,
) *CreateUserUseCase {
	return &CreateUserUseCase{
		userRepo:                      userRepo,
		createFreeSubscriptionUseCase: createFreeSubscriptionUseCase,
		centrifugoPublisher:           centrifugoPublisher,
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

	err := u.userRepo.Create(ctx, &user)
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	if _, err := u.createFreeSubscriptionUseCase.Execute(ctx, &user); err != nil {
		return nil, err
	}

	event, err := realtime.NewUserCreatedEvent(&user)
	if err != nil {
		return nil, errors.New("failed to build user created event")
	}

	go u.publishUserCreated(user.ID, event)

	return &user, nil
}

func (u *CreateUserUseCase) publishUserCreated(userID uuid.UUID, event realtime.UserEvent) {
	time.Sleep(userCreatedEventDelay)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := u.centrifugoPublisher.PublishUserToUser(ctx, userID, event); err != nil {
		log.Printf("failed to publish user_created to centrifugo for user %s: %v", userID, err)
	}
}
