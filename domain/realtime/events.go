package realtime

import (
	"encoding/json"
	"time"

	"go-api/domain/entity"
)

const (
	EventScanStarted   = "scan_started"
	EventScanCompleted = "scan_completed"
	EventScanFailed    = "scan_failed"
)

type MediaEvent struct {
	Type       string          `json:"type"`
	ScanID     string          `json:"scanId"`
	MediaID    string          `json:"mediaId,omitempty"`
	UserID     string          `json:"userId"`
	Status     string          `json:"status"`
	Message    string          `json:"message,omitempty"`
	FinalScore float64         `json:"finalScore,omitempty"`
	Confidence string          `json:"confidence,omitempty"`
	Verdict    string          `json:"verdict,omitempty"`
	Signals    []SignalPayload `json:"signals,omitempty"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}

type SignalPayload struct {
	Name       string   `json:"name"`
	Score      int      `json:"score"`
	Confidence string   `json:"confidence"`
	Details    []string `json:"details"`
}

func NewScanStartedEvent(media *entity.Media) (MediaEvent, error) {
	if media == nil {
		return MediaEvent{}, ErrInvalidMedia
	}

	return MediaEvent{
		Type:      EventScanStarted,
		ScanID:    media.ScanID.String(),
		MediaID:   media.ID.String(),
		UserID:    media.UserID.String(),
		Status:    string(media.Status),
		UpdatedAt: media.UpdatedAt,
	}, nil
}

func NewScanCompletedEvent(scan *entity.Scan, media *entity.Media, signals []*entity.Signal) (MediaEvent, error) {
	if scan == nil || media == nil {
		return MediaEvent{}, ErrInvalidMedia
	}

	event := MediaEvent{
		Type:       EventScanCompleted,
		ScanID:     scan.ID.String(),
		MediaID:    media.ID.String(),
		UserID:     scan.UserID.String(),
		Status:     string(media.Status),
		FinalScore: scan.FinalScore,
		Confidence: string(scan.Confidence),
		Verdict:    scan.Verdict,
		UpdatedAt:  scan.UpdatedAt,
		Signals:    make([]SignalPayload, 0, len(signals)),
	}

	for _, signal := range signals {
		if signal == nil {
			continue
		}

		event.Signals = append(event.Signals, SignalPayload{
			Name:       signal.Name,
			Score:      signal.Score,
			Confidence: string(signal.Confidence),
			Details:    signal.Details,
		})
	}

	return event, nil
}

func NewScanFailedEvent(scan *entity.Scan) (MediaEvent, error) {
	if scan == nil {
		return MediaEvent{}, ErrInvalidMedia
	}

	mediaID := ""
	if len(scan.Medias) > 0 {
		mediaID = scan.Medias[0].ID.String()
	}

	return MediaEvent{
		Type:      EventScanFailed,
		ScanID:    scan.ID.String(),
		MediaID:   mediaID,
		UserID:    scan.UserID.String(),
		Status:    string(scan.Status),
		Message:   scan.Message,
		UpdatedAt: scan.UpdatedAt,
	}, nil
}

func (e MediaEvent) Marshal() ([]byte, error) {
	return json.Marshal(e)
}
