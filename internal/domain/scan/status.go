package scan

import "strings"

type Status string

const (
	StatusUploaded   Status = "uploaded"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

type ScanType string

const (
	ScanTypeImage ScanType = "image"
	ScanTypeVideo ScanType = "video"
)

func TypeFromContentType(contentType string) (ScanType, bool) {
	normalized := strings.ToLower(contentType)
	switch {
	case strings.HasPrefix(normalized, "image/"):
		return ScanTypeImage, true
	case strings.HasPrefix(normalized, "video/"):
		return ScanTypeVideo, true
	default:
		return "", false
	}
}

type ConfidenceLevel string

const (
	ConfidenceUncertain ConfidenceLevel = "uncertain"
	ConfidenceLow       ConfidenceLevel = "low"
	ConfidenceMedium    ConfidenceLevel = "medium"
	ConfidenceHigh      ConfidenceLevel = "high"
)
