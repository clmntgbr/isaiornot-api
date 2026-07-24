package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClerkID   string    `gorm:"uniqueIndex;not null" json:"clerk_id"`
	FirstName string    `gorm:"null" json:"first_name"`
	LastName  string    `gorm:"null" json:"last_name"`
	Banned    bool      `gorm:"default:false;index:idx_user_banned" json:"banned"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`

	SubscriptionID *uuid.UUID    `gorm:"type:uuid;index:idx_user_subscription_id" json:"subscription_id"`
	Subscription   *Subscription `gorm:"foreignKey:SubscriptionID" json:"subscription"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
