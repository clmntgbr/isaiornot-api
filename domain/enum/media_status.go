package enum

type MediaStatus string

const (
	MediaStatusProcessing MediaStatus = "processing"
	MediaStatusUploaded   MediaStatus = "uploaded"
	MediaStatusAnalyzed   MediaStatus = "analyzed"
	MediaStatusFailed     MediaStatus = "failed"
)
