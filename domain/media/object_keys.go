package media

import (
	"fmt"
	"mime"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type PresignUploadInput struct {
	Filename    string
	ContentType string
}

var allowedExtensions = map[string]struct{}{
	"jpg":  {},
	"jpeg": {},
	"png":  {},
	"webp": {},
	"mp4":  {},
	"mov":  {},
	"avi":  {},
	"mkv":  {},
	"m4v":  {},
	"mpeg": {},
	"mpg":  {},
	"wmv":  {},
	"asf":  {},
	"flv":  {},
	"webm": {},
	"ogg":  {},
	"ogv":  {},
	"mka":  {},
}

func ValidatePresignUploadInput(input PresignUploadInput) error {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(input.Filename), "."))
	if ext == "" {
		return fmt.Errorf("filename must have a supported extension")
	}

	if _, ok := allowedExtensions[ext]; !ok {
		return fmt.Errorf("unsupported file type: .%s", ext)
	}

	return nil
}

func ContentTypeFromKey(key string, fallback string) string {
	if fallback != "" {
		return fallback
	}

	if contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(key))); contentType != "" {
		return contentType
	}

	return "application/octet-stream"
}

func IsImageContentType(contentType string) bool {
	return strings.HasPrefix(contentType, "image/")
}

func IsVideoContentType(contentType string) bool {
	return strings.HasPrefix(contentType, "video/")
}

const maxVideoFrames = 10

func MaxVideoFrames() int {
	return maxVideoFrames
}

func NewFrameFileKey() string {
	return "frames/" + uuid.NewString() + ".jpg"
}

func IsFrameObjectKey(decodedKey string) bool {
	return strings.Contains(decodedKey, "/frames/")
}

func NewFileKey(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	return uuid.NewString() + ext
}

func NewObjectKey(userID uuid.UUID, fileKey string) string {
	return userID.String() + "/" + fileKey
}

func NewObjectKeyFromFilename(userID uuid.UUID, filename string) string {
	return NewObjectKey(userID, NewFileKey(filename))
}

func NewThumbnailFileKey(mediaID uuid.UUID) string {
	return mediaID.String() + ".jpg"
}

func NewThumbnailObjectKey(userID uuid.UUID, mediaID uuid.UUID) string {
	return NewObjectKey(userID, NewThumbnailFileKey(mediaID))
}

func DecodeObjectKey(key string) (string, error) {
	decoded, err := url.QueryUnescape(key)
	if err != nil {
		return "", fmt.Errorf("invalid media key: %w", err)
	}

	return decoded, nil
}

func UserIDFromKey(encodedKey string) (uuid.UUID, error) {
	key, err := DecodeObjectKey(encodedKey)
	if err != nil {
		return uuid.Nil, err
	}

	userID, _, ok := strings.Cut(key, "/")
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid media key: %q", key)
	}

	return uuid.Parse(userID)
}

func FileKeyFromObjectKey(encodedKey string) (string, error) {
	key, err := DecodeObjectKey(encodedKey)
	if err != nil {
		return "", err
	}

	_, fileKey, ok := strings.Cut(key, "/")
	if !ok {
		return "", fmt.Errorf("invalid media key: %q", key)
	}

	return fileKey, nil
}
