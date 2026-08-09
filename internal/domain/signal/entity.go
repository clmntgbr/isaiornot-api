package signal

import (
	"time"

	"go-api/internal/domain/event"

	"github.com/google/uuid"
)

type Signal struct {
	ID      uuid.UUID
	MediaID uuid.UUID

	Name       string
	Score      int
	Confidence ConfidenceLevel
	Details    []string

	CreatedAt time.Time
	UpdatedAt time.Time

	events []event.DomainEvent
}

func NewSignal(
	mediaID uuid.UUID,
	name string,
	score int,
	confidence ConfidenceLevel,
	details []string,
) *Signal {
	now := time.Now().UTC()
	if details == nil {
		details = []string{}
	}

	s := &Signal{
		ID:         uuid.New(),
		MediaID:    mediaID,
		Name:       name,
		Score:      score,
		Confidence: confidence,
		Details:    details,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	s.recordEvent(SignalCreated{
		ID:         uuid.New().String(),
		SignalID:   s.ID.String(),
		MediaID:    mediaID.String(),
		Name:       name,
		Score:      score,
		Confidence: string(confidence),
		Details:    details,
		Timestamp:  now,
	})
	return s
}

func (s *Signal) PullEvents() []event.DomainEvent {
	events := s.events
	s.events = nil
	return events
}

func (s *Signal) recordEvent(e event.DomainEvent) {
	s.events = append(s.events, e)
}

func (s *Signal) Update(score int, confidence ConfidenceLevel, details []string) {
	if details == nil {
		details = []string{}
	}
	s.Score = score
	s.Confidence = confidence
	s.Details = details
	s.UpdatedAt = time.Now().UTC()
	s.recordEvent(SignalUpdated{
		ID:         uuid.New().String(),
		SignalID:   s.ID.String(),
		MediaID:    s.MediaID.String(),
		Name:       s.Name,
		Score:      s.Score,
		Confidence: string(s.Confidence),
		Details:    s.Details,
		Timestamp:  s.UpdatedAt,
	})
}

func (s *Signal) MarkDeleted() {
	s.UpdatedAt = time.Now().UTC()
	s.recordEvent(SignalDeleted{
		ID:        uuid.New().String(),
		SignalID:  s.ID.String(),
		MediaID:   s.MediaID.String(),
		Timestamp: s.UpdatedAt,
	})
}
