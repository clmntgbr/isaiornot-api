package gorm

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type mediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) repository.MediaRepository {
	return &mediaRepository{db: db}
}

func (r *mediaRepository) Create(ctx context.Context, media *entity.Media) error {
	return dbWithContext(ctx, r.db).Create(media).Error
}

func (r *mediaRepository) Update(ctx context.Context, media *entity.Media) error {
	return dbWithContext(ctx, r.db).Save(media).Error
}

func (r *mediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return dbWithContext(ctx, r.db).Delete(&entity.Media{}, id).Error
}

func (r *mediaRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Media, error) {
	var media entity.Media
	err := dbWithContext(ctx, r.db).
		Where("id = ?", id).
		Preload("Signals").
		Preload("Insight").
		Preload("Scan").
		First(&media).Error
	if err != nil {
		return nil, err
	}
	if media.ID == uuid.Nil {
		return nil, errors.New("media not found")
	}
	return &media, nil
}

func (r *mediaRepository) GetByKey(ctx context.Context, key string) (*entity.Media, error) {
	var media entity.Media
	err := dbWithContext(ctx, r.db).
		Where("key = ?", key).
		Preload("Scan").
		First(&media).Error
	if err != nil {
		return nil, err
	}
	if media.ID == uuid.Nil {
		return nil, errors.New("media not found")
	}
	return &media, nil
}

func (r *mediaRepository) CountUsageInPeriod(ctx context.Context, userID uuid.UUID, from, to time.Time) (*repository.MediaUsageCounts, error) {
	base := dbWithContext(ctx, r.db).Model(&entity.Media{}).
		Joins("JOIN scans ON scans.id = medias.scan_id").
		Where("medias.user_id = ? AND medias.created_at >= ? AND medias.created_at < ?", userID, from, to).
		Where("scans.status <> ?", "failed")

	var images int64
	if err := base.Session(&gorm.Session{}).
		Where("medias.content_type LIKE ? AND medias.key NOT LIKE ?", "image/%", "frames/%").
		Count(&images).Error; err != nil {
		return nil, err
	}

	var videos int64
	if err := base.Session(&gorm.Session{}).
		Where("medias.key LIKE ? OR medias.content_type LIKE ?", "frames/%", "video/%").
		Distinct("medias.scan_id").
		Count(&videos).Error; err != nil {
		return nil, err
	}

	return &repository.MediaUsageCounts{
		Images: images,
		Videos: videos,
	}, nil
}
