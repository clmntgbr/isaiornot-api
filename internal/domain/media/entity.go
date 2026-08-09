package media

import (
	"time"

	"go-api/internal/domain/event"
	"go-api/internal/domain/signal"

	"github.com/google/uuid"
)

type Media struct {
	ID     uuid.UUID
	ScanID uuid.UUID

	Key         string
	Filename    string
	Thumbnail   string
	ContentType string
	Size        int64

	Status   Status
	Statuses []Status

	CreatedAt time.Time
	UpdatedAt time.Time

	Signals []signal.Signal
	events  []event.DomainEvent
}

func NewMedia(
	scanID uuid.UUID,
	key string,
	filename string,
	contentType string,
	size int64,
) *Media {
	now := time.Now().UTC()
	m := &Media{
		ID:          uuid.New(),
		ScanID:      scanID,
		Key:         key,
		Filename:    filename,
		ContentType: contentType,
		Size:        size,
		Status:      StatusProcessing,
		Statuses:    []Status{StatusProcessing},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	m.recordEvent(MediaCreated{
		ID:          uuid.New().String(),
		MediaID:     m.ID.String(),
		ScanID:      scanID.String(),
		Key:         key,
		Filename:    filename,
		ContentType: contentType,
		Size:        size,
		Status:      string(m.Status),
		Timestamp:   now,
	})
	return m
}

func (m *Media) PullEvents() []event.DomainEvent {
	events := m.events
	m.events = nil
	return events
}

func (m *Media) recordEvent(e event.DomainEvent) {
	m.events = append(m.events, e)
}

func (m *Media) SetThumbnail(thumbnail string) {
	m.Thumbnail = thumbnail
	m.UpdatedAt = time.Now().UTC()
	m.recordUpdated()
}

func (m *Media) MarkProcessing() {
	if m.Status == StatusProcessing {
		return
	}
	m.setStatus(StatusProcessing)
	m.recordUpdated()
}

func (m *Media) MarkUploaded() {
	if m.Status == StatusUploaded {
		return
	}
	now := time.Now().UTC()
	m.setStatus(StatusUploaded)
	m.UpdatedAt = now
	m.recordEvent(MediaUploaded{
		ID:          uuid.New().String(),
		MediaID:     m.ID.String(),
		ScanID:      m.ScanID.String(),
		Key:         m.Key,
		Filename:    m.Filename,
		ContentType: m.ContentType,
		Size:        m.Size,
		Status:      string(m.Status),
		Timestamp:   now,
	})
}

// RecordUpload updates metadata from the storage webhook and marks the media as uploaded.
func (m *Media) RecordUpload(contentType string, size int64) {
	m.ContentType = contentType
	m.Size = size
	if m.Status == StatusUploaded {
		m.UpdatedAt = time.Now().UTC()
		m.recordUpdated()
		return
	}
	m.MarkUploaded()
}

// ReplaceWithFrame rewrites the media to point at an extracted video frame image.
func (m *Media) ReplaceWithFrame(key, filename string, size int64) {
	m.Key = key
	m.Filename = filename
	m.ContentType = "image/jpeg"
	m.Size = size
	m.UpdatedAt = time.Now().UTC()
	m.recordUpdated()
}

func (m *Media) MarkCompleted() {
	if m.Status == StatusCompleted {
		return
	}
	now := time.Now().UTC()
	m.setStatus(StatusCompleted)
	m.UpdatedAt = now
	m.recordEvent(MediaCompleted{
		ID:        uuid.New().String(),
		MediaID:   m.ID.String(),
		ScanID:    m.ScanID.String(),
		Status:    string(m.Status),
		Timestamp: now,
	})
}

func (m *Media) MarkFailed() {
	if m.Status == StatusFailed {
		return
	}
	now := time.Now().UTC()
	m.setStatus(StatusFailed)
	m.UpdatedAt = now
	m.recordEvent(MediaFailed{
		ID:        uuid.New().String(),
		MediaID:   m.ID.String(),
		ScanID:    m.ScanID.String(),
		Status:    string(m.Status),
		Timestamp: now,
	})
}

func (m *Media) setStatus(status Status) {
	m.Status = status
	m.Statuses = append(m.Statuses, status)
	m.UpdatedAt = time.Now().UTC()
}

func (m *Media) recordUpdated() {
	m.recordEvent(MediaUpdated{
		ID:          uuid.New().String(),
		MediaID:     m.ID.String(),
		ScanID:      m.ScanID.String(),
		Key:         m.Key,
		Filename:    m.Filename,
		Thumbnail:   m.Thumbnail,
		ContentType: m.ContentType,
		Size:        m.Size,
		Status:      string(m.Status),
		Timestamp:   m.UpdatedAt,
	})
}
