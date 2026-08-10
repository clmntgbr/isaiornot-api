package quota

import (
	"time"

	"go-api/internal/domain/event"

	"github.com/google/uuid"
)

type Quota struct {
	ID uuid.UUID

	MaxImagesPerMonth int
	MaxVideosPerMonth int
	MaxFileSizeImage  int64
	MaxFileSizeVideo  int64
	FullPipeline      bool
	HistoryRetention  time.Duration

	CreatedAt time.Time
	UpdatedAt time.Time

	events []event.DomainEvent
}

func NewQuota(
	maxImagesPerMonth int,
	maxVideosPerMonth int,
	maxFileSizeImage int64,
	maxFileSizeVideo int64,
	fullPipeline bool,
	historyRetention time.Duration,
) *Quota {
	now := time.Now().UTC()
	q := &Quota{
		ID:                uuid.New(),
		MaxImagesPerMonth: maxImagesPerMonth,
		MaxVideosPerMonth: maxVideosPerMonth,
		MaxFileSizeImage:  maxFileSizeImage,
		MaxFileSizeVideo:  maxFileSizeVideo,
		FullPipeline:      fullPipeline,
		HistoryRetention:  historyRetention,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	q.recordEvent(QuotaCreated{
		ID:                uuid.New().String(),
		QuotaID:           q.ID.String(),
		MaxImagesPerMonth: q.MaxImagesPerMonth,
		MaxVideosPerMonth: q.MaxVideosPerMonth,
		MaxFileSizeImage:  q.MaxFileSizeImage,
		MaxFileSizeVideo:  q.MaxFileSizeVideo,
		FullPipeline:      q.FullPipeline,
		HistoryRetention:  q.HistoryRetention.Nanoseconds(),
		Timestamp:         now,
	})
	return q
}

func (q *Quota) PullEvents() []event.DomainEvent {
	events := q.events
	q.events = nil
	return events
}

func (q *Quota) recordEvent(e event.DomainEvent) {
	q.events = append(q.events, e)
}

func (q *Quota) ApplyUpdate(
	maxImagesPerMonth int,
	maxVideosPerMonth int,
	maxFileSizeImage int64,
	maxFileSizeVideo int64,
	fullPipeline bool,
	historyRetention time.Duration,
) {
	q.MaxImagesPerMonth = maxImagesPerMonth
	q.MaxVideosPerMonth = maxVideosPerMonth
	q.MaxFileSizeImage = maxFileSizeImage
	q.MaxFileSizeVideo = maxFileSizeVideo
	q.FullPipeline = fullPipeline
	q.HistoryRetention = historyRetention
	q.UpdatedAt = time.Now().UTC()
	q.recordEvent(QuotaUpdated{
		ID:                uuid.New().String(),
		QuotaID:           q.ID.String(),
		MaxImagesPerMonth: q.MaxImagesPerMonth,
		MaxVideosPerMonth: q.MaxVideosPerMonth,
		MaxFileSizeImage:  q.MaxFileSizeImage,
		MaxFileSizeVideo:  q.MaxFileSizeVideo,
		FullPipeline:      q.FullPipeline,
		HistoryRetention:  q.HistoryRetention.Nanoseconds(),
		Timestamp:         q.UpdatedAt,
	})
}

func (q *Quota) MarkDeleted() {
	q.UpdatedAt = time.Now().UTC()
	q.recordEvent(QuotaDeleted{
		ID:        uuid.New().String(),
		QuotaID:   q.ID.String(),
		Timestamp: q.UpdatedAt,
	})
}
