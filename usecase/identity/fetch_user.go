package identity

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

func (s *FetchUserUseCase) Execute(ctx context.Context, externalID string) (port.ClerkUser, error) {
	return s.userFetcher.Get(ctx, externalID)
}
