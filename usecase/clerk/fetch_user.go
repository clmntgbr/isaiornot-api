package clerk

import (
	"context"
	"errors"
	clerkdto "go-api/infrastructure/clerk"
	"go-api/infrastructure/config"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkuser "github.com/clerk/clerk-sdk-go/v2/user"
)

type FetchUserUseCase struct {
	config *config.Config
}

func NewFetchUserUseCase(cfg *config.Config) *FetchUserUseCase {
	clerk.SetKey(cfg.ClerkSecretKey)
	return &FetchUserUseCase{config: cfg}
}

func (s *FetchUserUseCase) Execute(ctx context.Context, clerkID string) (clerkdto.ClerkUser, error) {
	clerkUser, err := clerkuser.Get(context.Background(), clerkID)
	if err != nil {
		return clerkdto.ClerkUser{}, errors.New("failed to get user")
	}

	firstName := ""
	if clerkUser.FirstName != nil {
		firstName = *clerkUser.FirstName
	}

	lastName := ""
	if clerkUser.LastName != nil {
		lastName = *clerkUser.LastName
	}

	banned := clerkUser.Banned

	email := ""
	for _, address := range clerkUser.EmailAddresses {
		if clerkUser.PrimaryEmailAddressID != nil && address.ID == *clerkUser.PrimaryEmailAddressID {
			email = address.EmailAddress
			break
		}
	}
	if email == "" && len(clerkUser.EmailAddresses) > 0 {
		email = clerkUser.EmailAddresses[0].EmailAddress
	}

	return clerkdto.ClerkUser{
		ID:        clerkUser.ID,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Banned:    banned,
	}, nil
}
