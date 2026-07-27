package enum

type MediaStatus string

const (
	MediaStatusProcessing MediaStatus = "processing"
	MediaStatusUploaded   MediaStatus = "uploaded"
	MediaStatusCompleted  MediaStatus = "completed"
	MediaStatusFailed     MediaStatus = "failed"
)
