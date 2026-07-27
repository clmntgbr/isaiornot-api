package auth

import (
	"context"
	"errors"

	"go-api/domain/port"
	"go-api/domain/repository"

	"github.com/golang-jwt/jwt/v5"
)

type ValidateTokenUseCase struct {
	tokenKeys port.TokenKeyProvider
	userRepo  repository.UserRepository
}

func NewValidateTokenUseCase(
	tokenKeys port.TokenKeyProvider,
	userRepo repository.UserRepository,
) *ValidateTokenUseCase {
	return &ValidateTokenUseCase{
		tokenKeys: tokenKeys,
		userRepo:  userRepo,
	}
}

func (uc *ValidateTokenUseCase) Execute(ctx context.Context, input ValidateTokenInput) (*ValidateTokenOutput, error) {
	token, err := jwt.ParseWithClaims(
		input.Token,
		&JWTClaims{},
		uc.tokenKeys.GetKeyfunc(),
		jwt.WithIssuer(uc.tokenKeys.GetIssuer()),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	user, err := uc.userRepo.GetByClerkID(ctx, claims.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return &ValidateTokenOutput{
		User:   user,
		Claims: claims,
	}, nil
}
