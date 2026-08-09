package media

import "time"

const (
	EventTypeMediaCreated   = "media.created.v1"
	EventTypeMediaUpdated   = "media.updated.v1"
	EventTypeMediaCompleted = "media.completed.v1"
	EventTypeMediaFailed    = "media.failed.v1"
)

type MediaCreated struct {
	ID          string    `json:"eventId"`
	MediaID     string    `json:"mediaId"`
	ScanID      string    `json:"scanId"`
	Key         string    `json:"key"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"contentType"`
	Size        int64     `json:"size"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e MediaCreated) EventID() string       { return e.ID }
func (e MediaCreated) EventType() string     { return EventTypeMediaCreated }
func (e MediaCreated) AggregateID() string   { return e.MediaID }
func (e MediaCreated) OccurredAt() time.Time { return e.Timestamp }

type MediaUpdated struct {
	ID          string    `json:"eventId"`
	MediaID     string    `json:"mediaId"`
	ScanID      string    `json:"scanId"`
	Key         string    `json:"key"`
	Filename    string    `json:"filename"`
	Thumbnail   string    `json:"thumbnail"`
	ContentType string    `json:"contentType"`
	Size        int64     `json:"size"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e MediaUpdated) EventID() string       { return e.ID }
func (e MediaUpdated) EventType() string     { return EventTypeMediaUpdated }
func (e MediaUpdated) AggregateID() string   { return e.MediaID }
func (e MediaUpdated) OccurredAt() time.Time { return e.Timestamp }

type MediaCompleted struct {
	ID        string    `json:"eventId"`
	MediaID   string    `json:"mediaId"`
	ScanID    string    `json:"scanId"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

func (e MediaCompleted) EventID() string       { return e.ID }
func (e MediaCompleted) EventType() string     { return EventTypeMediaCompleted }
func (e MediaCompleted) AggregateID() string   { return e.MediaID }
func (e MediaCompleted) OccurredAt() time.Time { return e.Timestamp }

type MediaFailed struct {
	ID        string    `json:"eventId"`
	MediaID   string    `json:"mediaId"`
	ScanID    string    `json:"scanId"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

func (e MediaFailed) EventID() string       { return e.ID }
func (e MediaFailed) EventType() string     { return EventTypeMediaFailed }
func (e MediaFailed) AggregateID() string   { return e.MediaID }
func (e MediaFailed) OccurredAt() time.Time { return e.Timestamp }
