package gorm

import (
	"context"
	"errors"
	"go-api/domain/aggregate"
	"go-api/domain/entity"
	"go-api/domain/repository"
	"go-api/infrastructure/paginate"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type scanRepository struct {
	db *gorm.DB
}

func NewScanRepository(db *gorm.DB) repository.ScanRepository {
	return &scanRepository{db: db}
}

func (r *scanRepository) Create(ctx context.Context, scan *entity.Scan) error {
	return dbWithContext(ctx, r.db).Create(scan).Error
}

func (r *scanRepository) Update(ctx context.Context, scan *entity.Scan) error {
	return dbWithContext(ctx, r.db).Save(scan).Error
}

func (r *scanRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return dbWithContext(ctx, r.db).Delete(&entity.Scan{}, id).Error
}

func (r *scanRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Scan, error) {
	var scan entity.Scan
	err := dbWithContext(ctx, r.db).
		Where("id = ?", id).
		Preload("Medias", func(db *gorm.DB) *gorm.DB {
			return db.Order("medias.created_at ASC")
		}).
		Preload("Medias.Signals").
		Preload("Medias.Insight").
		First(&scan).Error
	if err != nil {
		return nil, err
	}
	if scan.ID == uuid.Nil {
		return nil, errors.New("scan not found")
	}
	return &scan, nil
}

func (r *scanRepository) GetByUserID(ctx context.Context, userID uuid.UUID, query paginate.PaginateQuery) ([]*entity.Scan, int64, error) {
	var scans []*entity.Scan

	db := dbWithContext(ctx, r.db).Model(&entity.Scan{}).
		Where("scans.user_id = ?", userID)

	if query.Search != "" {
		db = db.Joins("JOIN medias ON medias.scan_id = scans.id").
			Where("medias.filename ILIKE ? OR medias.key ILIKE ?", "%"+query.Search+"%", "%"+query.Search+"%").
			Distinct()
	}

	db, total, err := Paginate(db, query)
	if err != nil {
		return nil, 0, err
	}

	err = db.
		Preload("Medias", func(db *gorm.DB) *gorm.DB {
			return db.Order("medias.created_at ASC")
		}).
		Find(&scans).Error
	if err != nil {
		return nil, 0, err
	}

	return scans, total, nil
}

func (r *scanRepository) GetStatisticsByUserID(ctx context.Context, userID uuid.UUID) (*entity.MediaStatistics, error) {
	var stats entity.MediaStatistics

	err := dbWithContext(ctx, r.db).Raw(`
		SELECT
			COUNT(*) FILTER (WHERE verdict <> '') AS scans_count,
			COUNT(*) FILTER (WHERE verdict = ?) AS real_image_count,
			COUNT(*) FILTER (WHERE verdict = ?) AS ai_image_count,
			COALESCE(AVG(final_score) FILTER (WHERE verdict <> '' AND final_score >= 0), 0) AS average_score
		FROM scans
		WHERE user_id = ?
	`,
		aggregate.VerdictLikelyReal,
		aggregate.VerdictLikelyAI,
		userID,
	).Scan(&stats).Error
	if err != nil {
		return nil, err
	}

	return &stats, nil
}
