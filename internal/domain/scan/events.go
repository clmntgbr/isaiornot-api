package scan

import "time"

const (
	EventTypeScanCreated   = "scan.created.v1"
	EventTypeScanUpdated   = "scan.updated.v1"
	EventTypeScanCompleted = "scan.completed.v1"
	EventTypeScanFailed    = "scan.failed.v1"
	EventTypeScanFinalize  = "scan.finalize.v1"
)

type ScanCreated struct {
	ID        string    `json:"eventId"`
	ScanID    string    `json:"scanId"`
	UserID    string    `json:"userId"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

func (e ScanCreated) EventID() string       { return e.ID }
func (e ScanCreated) EventType() string     { return EventTypeScanCreated }
func (e ScanCreated) AggregateID() string   { return e.ScanID }
func (e ScanCreated) OccurredAt() time.Time { return e.Timestamp }

type ScanUpdated struct {
	ID         string    `json:"eventId"`
	ScanID     string    `json:"scanId"`
	UserID     string    `json:"userId"`
	Status     string    `json:"status"`
	Message    string    `json:"message,omitempty"`
	RetryCount int       `json:"retryCount,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

func (e ScanUpdated) EventID() string       { return e.ID }
func (e ScanUpdated) EventType() string     { return EventTypeScanUpdated }
func (e ScanUpdated) AggregateID() string   { return e.ScanID }
func (e ScanUpdated) OccurredAt() time.Time { return e.Timestamp }

type ScanCompleted struct {
	ID         string    `json:"eventId"`
	ScanID     string    `json:"scanId"`
	UserID     string    `json:"userId"`
	Status     string    `json:"status"`
	FinalScore float64   `json:"finalScore"`
	Confidence string    `json:"confidence"`
	Verdict    string    `json:"verdict"`
	Duration   int       `json:"duration"`
	Timestamp  time.Time `json:"timestamp"`
}

func (e ScanCompleted) EventID() string       { return e.ID }
func (e ScanCompleted) EventType() string     { return EventTypeScanCompleted }
func (e ScanCompleted) AggregateID() string   { return e.ScanID }
func (e ScanCompleted) OccurredAt() time.Time { return e.Timestamp }

type ScanFailed struct {
	ID        string    `json:"eventId"`
	ScanID    string    `json:"scanId"`
	UserID    string    `json:"userId"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Duration  int       `json:"duration"`
	Timestamp time.Time `json:"timestamp"`
}

func (e ScanFailed) EventID() string       { return e.ID }
func (e ScanFailed) EventType() string     { return EventTypeScanFailed }
func (e ScanFailed) AggregateID() string   { return e.ScanID }
func (e ScanFailed) OccurredAt() time.Time { return e.Timestamp }

type ScanFinalizeRequested struct {
	ID        string    `json:"eventId"`
	ScanID    string    `json:"scanId"`
	Timestamp time.Time `json:"timestamp"`
}

func (e ScanFinalizeRequested) EventID() string       { return e.ID }
func (e ScanFinalizeRequested) EventType() string     { return EventTypeScanFinalize }
func (e ScanFinalizeRequested) AggregateID() string   { return e.ScanID }
func (e ScanFinalizeRequested) OccurredAt() time.Time { return e.Timestamp }
