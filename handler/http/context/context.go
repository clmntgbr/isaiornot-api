package context

import (
	"errors"
	"go-api/domain/entity"

	"github.com/gofiber/fiber/v3"
)

const (
	UserKey = "user"
)

func GetUser(c fiber.Ctx) (*entity.User, error) {
	user, ok := c.Locals(UserKey).(*entity.User)
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func SetUser(c fiber.Ctx, user entity.User) {
	c.Locals(UserKey, &user)
}
