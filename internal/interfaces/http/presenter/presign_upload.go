package presenter

import (
	cmdscan "go-api/internal/application/command/scan"

	"github.com/google/uuid"
)

type PresignUploadResponse struct {
	URL    string    `json:"url"`
	ScanID uuid.UUID `json:"scanId"`
}

func NewPresignUploadResponse(result *cmdscan.PresignUploadResult) PresignUploadResponse {
	return PresignUploadResponse{
		URL:    result.URL,
		ScanID: result.ScanID,
	}
}
