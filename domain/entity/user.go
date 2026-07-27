package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClerkID   string    `gorm:"uniqueIndex;not null" json:"clerkId"`
	FirstName string    `gorm:"null" json:"firstName"`
	LastName  string    `gorm:"null" json:"lastName"`
	Banned    bool      `gorm:"default:false;index:idx_user_banned" json:"banned"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`

	SubscriptionID *uuid.UUID    `gorm:"type:uuid;index:idx_user_subscription_id" json:"subscriptionId"`
	Subscription   *Subscription `gorm:"foreignKey:SubscriptionID" json:"subscription"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (User) TableName() string {
	return "users"
}
