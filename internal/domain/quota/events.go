package quota

import "time"

const (
	EventTypeQuotaCreated = "quota.created.v1"
	EventTypeQuotaUpdated = "quota.updated.v1"
	EventTypeQuotaDeleted = "quota.deleted.v1"
)

type QuotaCreated struct {
	ID                string    `json:"eventId"`
	QuotaID           string    `json:"quotaId"`
	MaxImagesPerMonth int       `json:"maxImagesPerMonth"`
	MaxVideosPerMonth int       `json:"maxVideosPerMonth"`
	MaxFileSizeImage  int64     `json:"maxFileSizeImage"`
	MaxFileSizeVideo  int64     `json:"maxFileSizeVideo"`
	FullPipeline      bool      `json:"fullPipeline"`
	HistoryRetention  int64     `json:"historyRetention"`
	Timestamp         time.Time `json:"timestamp"`
}

func (e QuotaCreated) EventID() string       { return e.ID }
func (e QuotaCreated) EventType() string     { return EventTypeQuotaCreated }
func (e QuotaCreated) AggregateID() string   { return e.QuotaID }
func (e QuotaCreated) OccurredAt() time.Time { return e.Timestamp }

type QuotaUpdated struct {
	ID                string    `json:"eventId"`
	QuotaID           string    `json:"quotaId"`
	MaxImagesPerMonth int       `json:"maxImagesPerMonth"`
	MaxVideosPerMonth int       `json:"maxVideosPerMonth"`
	MaxFileSizeImage  int64     `json:"maxFileSizeImage"`
	MaxFileSizeVideo  int64     `json:"maxFileSizeVideo"`
	FullPipeline      bool      `json:"fullPipeline"`
	HistoryRetention  int64     `json:"historyRetention"`
	Timestamp         time.Time `json:"timestamp"`
}

func (e QuotaUpdated) EventID() string       { return e.ID }
func (e QuotaUpdated) EventType() string     { return EventTypeQuotaUpdated }
func (e QuotaUpdated) AggregateID() string   { return e.QuotaID }
func (e QuotaUpdated) OccurredAt() time.Time { return e.Timestamp }

type QuotaDeleted struct {
	ID        string    `json:"eventId"`
	QuotaID   string    `json:"quotaId"`
	Timestamp time.Time `json:"timestamp"`
}

func (e QuotaDeleted) EventID() string       { return e.ID }
func (e QuotaDeleted) EventType() string     { return EventTypeQuotaDeleted }
func (e QuotaDeleted) AggregateID() string   { return e.QuotaID }
func (e QuotaDeleted) OccurredAt() time.Time { return e.Timestamp }
