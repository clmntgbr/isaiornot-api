package media

import "time"

const (
	EventTypeMediaCreated   = "media.created.v1"
	EventTypeMediaUpdated   = "media.updated.v1"
	EventTypeMediaUploaded  = "media.uploaded.v1"
	EventTypeMediaCompleted = "media.completed.v1"
	EventTypeMediaFailed    = "media.failed.v1"

	EventTypeMediaAnalyzeMetadata   = "media.analyze.metadata.v1"
	EventTypeMediaAnalyzeHeuristics = "media.analyze.heuristics.v1"
	EventTypeMediaAnalyzeAIModel    = "media.analyze.ai_model.v1"

	EventTypeMediaAnalyzeMetadataDone     = "media.analyze.metadata.done.v1"
	EventTypeMediaAnalyzeMetadataFailed   = "media.analyze.metadata.failed.v1"
	EventTypeMediaAnalyzeHeuristicsDone   = "media.analyze.heuristics.done.v1"
	EventTypeMediaAnalyzeHeuristicsFailed = "media.analyze.heuristics.failed.v1"
	EventTypeMediaAnalyzeAIModelDone      = "media.analyze.ai_model.done.v1"
	EventTypeMediaAnalyzeAIModelFailed    = "media.analyze.ai_model.failed.v1"

	StageMetadata   = "metadata"
	StageHeuristics = "heuristics"
	StageAIModel    = "ai_model"
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

type MediaUploaded struct {
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

func (e MediaUploaded) EventID() string       { return e.ID }
func (e MediaUploaded) EventType() string     { return EventTypeMediaUploaded }
func (e MediaUploaded) AggregateID() string   { return e.MediaID }
func (e MediaUploaded) OccurredAt() time.Time { return e.Timestamp }

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

type MediaAnalyzeRequested struct {
	ID        string    `json:"eventId"`
	MediaID   string    `json:"mediaId"`
	ScanID    string    `json:"scanId"`
	Stage     string    `json:"stage"`
	Timestamp time.Time `json:"timestamp"`
}

type MediaAnalyzeMetadataDone struct {
	ID        string    `json:"eventId"`
	MediaID   string    `json:"mediaId"`
	ScanID    string    `json:"scanId"`
	Stage     string    `json:"stage"`
	Timestamp time.Time `json:"timestamp"`
}

func (e MediaAnalyzeMetadataDone) EventID() string       { return e.ID }
func (e MediaAnalyzeMetadataDone) EventType() string     { return EventTypeMediaAnalyzeMetadataDone }
func (e MediaAnalyzeMetadataDone) AggregateID() string   { return e.MediaID }
func (e MediaAnalyzeMetadataDone) OccurredAt() time.Time { return e.Timestamp }

type MediaAnalyzeMetadataFailed struct {
	ID        string    `json:"eventId"`
	MediaID   string    `json:"mediaId"`
	ScanID    string    `json:"scanId"`
	Stage     string    `json:"stage"`
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

func (e MediaAnalyzeMetadataFailed) EventID() string       { return e.ID }
func (e MediaAnalyzeMetadataFailed) EventType() string     { return EventTypeMediaAnalyzeMetadataFailed }
func (e MediaAnalyzeMetadataFailed) AggregateID() string   { return e.MediaID }
func (e MediaAnalyzeMetadataFailed) OccurredAt() time.Time { return e.Timestamp }

type MediaAnalyzeHeuristicsDone struct {
	ID        string    `json:"eventId"`
	MediaID   string    `json:"mediaId"`
	ScanID    string    `json:"scanId"`
	Stage     string    `json:"stage"`
	Timestamp time.Time `json:"timestamp"`
}

func (e MediaAnalyzeHeuristicsDone) EventID() string       { return e.ID }
func (e MediaAnalyzeHeuristicsDone) EventType() string     { return EventTypeMediaAnalyzeHeuristicsDone }
func (e MediaAnalyzeHeuristicsDone) AggregateID() string   { return e.MediaID }
func (e MediaAnalyzeHeuristicsDone) OccurredAt() time.Time { return e.Timestamp }

type MediaAnalyzeHeuristicsFailed struct {
	ID        string    `json:"eventId"`
	MediaID   string    `json:"mediaId"`
	ScanID    string    `json:"scanId"`
	Stage     string    `json:"stage"`
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

func (e MediaAnalyzeHeuristicsFailed) EventID() string { return e.ID }
func (e MediaAnalyzeHeuristicsFailed) EventType() string {
	return EventTypeMediaAnalyzeHeuristicsFailed
}
func (e MediaAnalyzeHeuristicsFailed) AggregateID() string   { return e.MediaID }
func (e MediaAnalyzeHeuristicsFailed) OccurredAt() time.Time { return e.Timestamp }

type MediaAnalyzeAIModelDone struct {
	ID        string    `json:"eventId"`
	MediaID   string    `json:"mediaId"`
	ScanID    string    `json:"scanId"`
	Stage     string    `json:"stage"`
	Timestamp time.Time `json:"timestamp"`
}

func (e MediaAnalyzeAIModelDone) EventID() string       { return e.ID }
func (e MediaAnalyzeAIModelDone) EventType() string     { return EventTypeMediaAnalyzeAIModelDone }
func (e MediaAnalyzeAIModelDone) AggregateID() string   { return e.MediaID }
func (e MediaAnalyzeAIModelDone) OccurredAt() time.Time { return e.Timestamp }

type MediaAnalyzeAIModelFailed struct {
	ID        string    `json:"eventId"`
	MediaID   string    `json:"mediaId"`
	ScanID    string    `json:"scanId"`
	Stage     string    `json:"stage"`
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

func (e MediaAnalyzeAIModelFailed) EventID() string       { return e.ID }
func (e MediaAnalyzeAIModelFailed) EventType() string     { return EventTypeMediaAnalyzeAIModelFailed }
func (e MediaAnalyzeAIModelFailed) AggregateID() string   { return e.MediaID }
func (e MediaAnalyzeAIModelFailed) OccurredAt() time.Time { return e.Timestamp }
