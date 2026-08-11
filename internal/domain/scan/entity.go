package scan

import (
	"time"

	"go-api/internal/domain/event"

	"github.com/google/uuid"
)

type Scan struct {
	ID     uuid.UUID
	UserID uuid.UUID

	Type     ScanType
	Status   Status
	Statuses []Status
	Message  string

	FinalScore float64
	Confidence ConfidenceLevel
	Verdict    string

	Duration   int
	RetryCount int

	CreatedAt time.Time
	UpdatedAt time.Time

	events []event.DomainEvent
}

func NewScan(userID uuid.UUID, scanType ScanType) *Scan {
	now := time.Now().UTC()
	s := &Scan{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      scanType,
		Status:    StatusUploaded,
		Statuses:  []Status{StatusUploaded},
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.recordEvent(ScanCreated{
		ID:        uuid.New().String(),
		ScanID:    s.ID.String(),
		UserID:    userID.String(),
		Type:      string(s.Type),
		Status:    string(s.Status),
		Timestamp: now,
	})
	return s
}

func (s *Scan) PullEvents() []event.DomainEvent {
	events := s.events
	s.events = nil
	return events
}

func (s *Scan) recordEvent(e event.DomainEvent) {
	s.events = append(s.events, e)
}

func (s *Scan) MarkProcessing() {
	if s.Status == StatusProcessing {
		return
	}
	s.Status = StatusProcessing
	s.Statuses = append(s.Statuses, StatusProcessing)
	s.UpdatedAt = time.Now().UTC()
	s.recordEvent(ScanUpdated{
		ID:        uuid.New().String(),
		ScanID:    s.ID.String(),
		UserID:    s.UserID.String(),
		Status:    string(s.Status),
		Message:   s.Message,
		Timestamp: s.UpdatedAt,
	})
}

func (s *Scan) MarkCompleted(finalScore float64, confidence ConfidenceLevel, verdict string) {
	now := time.Now().UTC()
	s.Status = StatusCompleted
	s.Statuses = append(s.Statuses, StatusCompleted)
	s.FinalScore = finalScore
	s.Confidence = confidence
	s.Verdict = verdict
	s.Message = ""
	s.freezeDuration(now)
	s.UpdatedAt = now
	s.recordEvent(ScanCompleted{
		ID:         uuid.New().String(),
		ScanID:     s.ID.String(),
		UserID:     s.UserID.String(),
		Status:     string(s.Status),
		FinalScore: s.FinalScore,
		Confidence: string(s.Confidence),
		Verdict:    s.Verdict,
		Duration:   s.Duration,
		Timestamp:  now,
	})
}

func (s *Scan) MarkFailed(message string) {
	now := time.Now().UTC()
	s.Status = StatusFailed
	s.Statuses = append(s.Statuses, StatusFailed)
	s.Message = message
	s.freezeDuration(now)
	s.UpdatedAt = now
	s.recordEvent(ScanFailed{
		ID:        uuid.New().String(),
		ScanID:    s.ID.String(),
		UserID:    s.UserID.String(),
		Status:    string(s.Status),
		Message:   s.Message,
		Duration:  s.Duration,
		Timestamp: now,
	})
}

func (s *Scan) IncrementRetry() {
	s.RetryCount++
	s.UpdatedAt = time.Now().UTC()
	s.recordEvent(ScanUpdated{
		ID:         uuid.New().String(),
		ScanID:     s.ID.String(),
		UserID:     s.UserID.String(),
		Status:     string(s.Status),
		Message:    s.Message,
		RetryCount: s.RetryCount,
		Timestamp:  s.UpdatedAt,
	})
}

func (s *Scan) freezeDuration(at time.Time) {
	if s.Duration > 0 {
		return
	}
	s.Duration = int(at.Sub(s.CreatedAt).Milliseconds())
	if s.Duration < 0 {
		s.Duration = 0
	}
}
