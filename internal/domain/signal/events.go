package signal

import "time"

const (
	EventTypeSignalCreated = "signal.created.v1"
	EventTypeSignalUpdated = "signal.updated.v1"
	EventTypeSignalDeleted = "signal.deleted.v1"
)

type SignalCreated struct {
	ID         string    `json:"eventId"`
	SignalID   string    `json:"signalId"`
	MediaID    string    `json:"mediaId"`
	Name       string    `json:"name"`
	Score      int       `json:"score"`
	Confidence string    `json:"confidence"`
	Details    []string  `json:"details"`
	Timestamp  time.Time `json:"timestamp"`
}

func (e SignalCreated) EventID() string       { return e.ID }
func (e SignalCreated) EventType() string     { return EventTypeSignalCreated }
func (e SignalCreated) AggregateID() string   { return e.SignalID }
func (e SignalCreated) OccurredAt() time.Time { return e.Timestamp }

type SignalUpdated struct {
	ID         string    `json:"eventId"`
	SignalID   string    `json:"signalId"`
	MediaID    string    `json:"mediaId"`
	Name       string    `json:"name"`
	Score      int       `json:"score"`
	Confidence string    `json:"confidence"`
	Details    []string  `json:"details"`
	Timestamp  time.Time `json:"timestamp"`
}

func (e SignalUpdated) EventID() string       { return e.ID }
func (e SignalUpdated) EventType() string     { return EventTypeSignalUpdated }
func (e SignalUpdated) AggregateID() string   { return e.SignalID }
func (e SignalUpdated) OccurredAt() time.Time { return e.Timestamp }

type SignalDeleted struct {
	ID        string    `json:"eventId"`
	SignalID  string    `json:"signalId"`
	MediaID   string    `json:"mediaId"`
	Timestamp time.Time `json:"timestamp"`
}

func (e SignalDeleted) EventID() string       { return e.ID }
func (e SignalDeleted) EventType() string     { return EventTypeSignalDeleted }
func (e SignalDeleted) AggregateID() string   { return e.SignalID }
func (e SignalDeleted) OccurredAt() time.Time { return e.Timestamp }
