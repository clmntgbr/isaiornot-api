package clerk

import (
	"context"

	"go-api/domain/port"
)

type FetchUserUseCase struct {
	userFetcher port.ClerkUserFetcher
}

func NewFetchUserUseCase(userFetcher port.ClerkUserFetcher) *FetchUserUseCase {
	return &FetchUserUseCase{userFetcher: userFetcher}
}

func (s *FetchUserUseCase) Execute(ctx context.Context, clerkID string) (port.ClerkUser, error) {
	return s.userFetcher.Get(ctx, clerkID)
}
