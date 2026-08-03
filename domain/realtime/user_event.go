package realtime

import (
	"encoding/json"
	"time"

	"go-api/domain/entity"
)

const (
	EventUserCreated = "user_created"
	EventUserUpdated = "user_updated"
)

type UserEvent struct {
	Type      string    `json:"type"`
	UserID    string    `json:"userId"`
	ClerkID   string    `json:"clerkId"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewUserCreatedEvent(user *entity.User) (UserEvent, error) {
	return newUserEvent(user, EventUserCreated)
}

func NewUserUpdatedEvent(user *entity.User) (UserEvent, error) {
	return newUserEvent(user, EventUserUpdated)
}

func newUserEvent(user *entity.User, eventType string) (UserEvent, error) {
	if user == nil {
		return UserEvent{}, ErrInvalidUser
	}

	return UserEvent{
		Type:      eventType,
		UserID:    user.ID.String(),
		ClerkID:   user.ClerkID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (e UserEvent) Marshal() ([]byte, error) {
	return json.Marshal(e)
}
