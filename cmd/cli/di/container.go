package di

import (
	"go-api/infrastructure/config"

	"gorm.io/gorm"
)

type Container struct {
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	return &Container{}
}
