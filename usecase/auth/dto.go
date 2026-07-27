package auth

import (
	"go-api/domain/entity"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	Email  string `json:"email"`
	UserID string `json:"sub"`
	jwt.RegisteredClaims
}

type ValidateTokenInput struct {
	Token string
}

type ValidateTokenOutput struct {
	User   *entity.User
	Claims *JWTClaims
}
