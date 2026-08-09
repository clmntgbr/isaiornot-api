package presenter

import (
	"time"

	domainmedia "go-api/internal/domain/media"
)

type MediaResponse struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Filename    string    `json:"filename"`
	Thumbnail   *string   `json:"thumbnail"`
	ContentType string    `json:"contentType"`
	Status      string    `json:"status"`
	Statuses    []string  `json:"statuses"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func NewMediaResponse(view *domainmedia.MediaView) MediaResponse {
	statuses := make([]string, 0, len(view.Statuses))
	for _, status := range view.Statuses {
		statuses = append(statuses, string(status))
	}

	return MediaResponse{
		ID:          view.ID.String(),
		Key:         view.Key,
		Filename:    view.Filename,
		Thumbnail:   optionalString(thumbnailURL(view.ID.String(), view.Thumbnail)),
		ContentType: view.ContentType,
		Status:      string(view.Status),
		Statuses:    statuses,
		Size:        view.Size,
		CreatedAt:   view.CreatedAt,
		UpdatedAt:   view.UpdatedAt,
	}
}

func thumbnailURL(mediaID, thumbnail string) string {
	if thumbnail == "" {
		return ""
	}
	return "/api/medias/" + mediaID + "/thumbnail"
}

func NewMediaResponses(views []*domainmedia.MediaView) []MediaResponse {
	responses := make([]MediaResponse, 0, len(views))
	for _, view := range views {
		if view == nil {
			continue
		}
		responses = append(responses, NewMediaResponse(view))
	}
	return responses
}
