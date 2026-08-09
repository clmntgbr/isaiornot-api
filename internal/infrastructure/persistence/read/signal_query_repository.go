package read

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	domainsignal "go-api/internal/domain/signal"
	"go-api/internal/infrastructure/persistence/dbtype"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type signalRow struct {
	ID         uuid.UUID
	MediaID    uuid.UUID
	Name       string
	Score      int
	Confidence string
	Details    dbtype.JSONB
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (signalRow) TableName() string { return "signals" }

type signalReadRepository struct {
	db *gorm.DB
}

func NewSignalReadRepository(db *gorm.DB) domainsignal.SignalReadRepository {
	return &signalReadRepository{db: db}
}

func (r *signalReadRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainsignal.SignalView, error) {
	var row signalRow
	err := r.db.WithContext(ctx).
		Select("id", "media_id", "name", "score", "confidence", "details", "created_at", "updated_at").
		First(&row, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toSignalView(row)
}

func (r *signalReadRepository) FindByMediaID(ctx context.Context, mediaID uuid.UUID) ([]*domainsignal.SignalView, error) {
	byMedia, err := r.FindByMediaIDs(ctx, []uuid.UUID{mediaID})
	if err != nil {
		return nil, err
	}
	return byMedia[mediaID], nil
}

func (r *signalReadRepository) FindByMediaIDs(ctx context.Context, mediaIDs []uuid.UUID) (map[uuid.UUID][]*domainsignal.SignalView, error) {
	result := make(map[uuid.UUID][]*domainsignal.SignalView, len(mediaIDs))
	if len(mediaIDs) == 0 {
		return result, nil
	}

	var rows []signalRow
	err := r.db.WithContext(ctx).
		Select("id", "media_id", "name", "score", "confidence", "details", "created_at", "updated_at").
		Where("media_id IN ?", mediaIDs).
		Order("created_at asc").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		view, err := toSignalView(row)
		if err != nil {
			return nil, err
		}
		result[row.MediaID] = append(result[row.MediaID], view)
	}
	return result, nil
}

func toSignalView(row signalRow) (*domainsignal.SignalView, error) {
	var details []string
	if len(row.Details) > 0 {
		if err := json.Unmarshal(row.Details, &details); err != nil {
			return nil, err
		}
	}
	if details == nil {
		details = []string{}
	}

	return &domainsignal.SignalView{
		ID:         row.ID,
		MediaID:    row.MediaID,
		Name:       row.Name,
		Score:      row.Score,
		Confidence: domainsignal.ConfidenceLevel(row.Confidence),
		Details:    details,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}, nil
}
