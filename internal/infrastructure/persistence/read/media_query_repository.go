package read

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	domainmedia "go-api/internal/domain/media"
	"go-api/internal/infrastructure/persistence/dbtype"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type mediaRow struct {
	ID          uuid.UUID
	ScanID      uuid.UUID
	Key         string
	Filename    string
	Thumbnail   string
	ContentType string
	Size        int64
	Status      string
	Statuses    dbtype.JSONB
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (mediaRow) TableName() string { return "medias" }

type mediaReadRepository struct {
	db *gorm.DB
}

func NewMediaReadRepository(db *gorm.DB) domainmedia.MediaReadRepository {
	return &mediaReadRepository{db: db}
}

func (r *mediaReadRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainmedia.MediaView, error) {
	var row mediaRow
	err := r.db.WithContext(ctx).
		Select(
			"id", "scan_id", "key", "filename", "thumbnail",
			"content_type", "size", "status", "statuses",
			"created_at", "updated_at",
		).
		First(&row, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toMediaView(row)
}

func (r *mediaReadRepository) FindByScanID(ctx context.Context, scanID uuid.UUID) ([]*domainmedia.MediaView, error) {
	byScan, err := r.FindByScanIDs(ctx, []uuid.UUID{scanID})
	if err != nil {
		return nil, err
	}
	return byScan[scanID], nil
}

func (r *mediaReadRepository) FindByScanIDs(ctx context.Context, scanIDs []uuid.UUID) (map[uuid.UUID][]*domainmedia.MediaView, error) {
	result := make(map[uuid.UUID][]*domainmedia.MediaView, len(scanIDs))
	if len(scanIDs) == 0 {
		return result, nil
	}

	var rows []mediaRow
	err := r.db.WithContext(ctx).
		Select(
			"id", "scan_id", "key", "filename", "thumbnail",
			"content_type", "size", "status", "statuses",
			"created_at", "updated_at",
		).
		Where("scan_id IN ?", scanIDs).
		Order("created_at asc").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		view, err := toMediaView(row)
		if err != nil {
			return nil, err
		}
		result[row.ScanID] = append(result[row.ScanID], view)
	}
	return result, nil
}

func toMediaView(row mediaRow) (*domainmedia.MediaView, error) {
	var raw []string
	if len(row.Statuses) > 0 {
		if err := json.Unmarshal(row.Statuses, &raw); err != nil {
			return nil, err
		}
	}

	statuses := make([]domainmedia.Status, 0, len(raw))
	for _, status := range raw {
		statuses = append(statuses, domainmedia.Status(status))
	}

	return &domainmedia.MediaView{
		ID:          row.ID,
		ScanID:      row.ScanID,
		Key:         row.Key,
		Filename:    row.Filename,
		Thumbnail:   row.Thumbnail,
		ContentType: row.ContentType,
		Size:        row.Size,
		Status:      domainmedia.Status(row.Status),
		Statuses:    statuses,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}
