package read

import (
	"context"
	"errors"
	"time"

	domainquota "go-api/internal/domain/quota"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type quotaRow struct {
	ID                uuid.UUID
	MaxImagesPerMonth int
	MaxVideosPerMonth int
	MaxFileSizeImage  int64
	MaxFileSizeVideo  int64
	FullPipeline      bool
	HistoryRetention  int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (quotaRow) TableName() string { return "quotas" }

type quotaReadRepository struct {
	db *gorm.DB
}

func NewQuotaReadRepository(db *gorm.DB) domainquota.QuotaReadRepository {
	return &quotaReadRepository{db: db}
}

func (r *quotaReadRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainquota.QuotaView, error) {
	var row quotaRow
	err := r.db.WithContext(ctx).
		Select(
			"id",
			"max_images_per_month",
			"max_videos_per_month",
			"max_file_size_image",
			"max_file_size_video",
			"full_pipeline",
			"history_retention",
			"created_at",
			"updated_at",
		).
		First(&row, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toQuotaView(row), nil
}

func (r *quotaReadRepository) FindAll(ctx context.Context) ([]*domainquota.QuotaView, error) {
	var rows []quotaRow
	err := r.db.WithContext(ctx).
		Select(
			"id",
			"max_images_per_month",
			"max_videos_per_month",
			"max_file_size_image",
			"max_file_size_video",
			"full_pipeline",
			"history_retention",
			"created_at",
			"updated_at",
		).
		Order("created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	views := make([]*domainquota.QuotaView, 0, len(rows))
	for _, row := range rows {
		views = append(views, toQuotaView(row))
	}
	return views, nil
}

func toQuotaView(row quotaRow) *domainquota.QuotaView {
	return &domainquota.QuotaView{
		ID:                row.ID,
		MaxImagesPerMonth: row.MaxImagesPerMonth,
		MaxVideosPerMonth: row.MaxVideosPerMonth,
		MaxFileSizeImage:  row.MaxFileSizeImage,
		MaxFileSizeVideo:  row.MaxFileSizeVideo,
		FullPipeline:      row.FullPipeline,
		HistoryRetention:  time.Duration(row.HistoryRetention),
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}
}
