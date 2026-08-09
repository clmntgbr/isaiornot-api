package write

import (
	"context"
	"encoding/json"
	"errors"

	domainmedia "go-api/internal/domain/media"
	"go-api/internal/infrastructure/persistence/dbtype"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type mediaWriteRepository struct {
	db *gorm.DB
}

func NewMediaWriteRepository(db *gorm.DB) domainmedia.MediaWriteRepository {
	return &mediaWriteRepository{db: db}
}

func (r *mediaWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *mediaWriteRepository) Save(ctx context.Context, media *domainmedia.Media) error {
	model, err := mediaModelFromDomain(media)
	if err != nil {
		return err
	}
	return DBWithContext(ctx, r.db).Create(model).Error
}

func (r *mediaWriteRepository) Update(ctx context.Context, media *domainmedia.Media) error {
	model, err := mediaModelFromDomain(media)
	if err != nil {
		return err
	}
	return DBWithContext(ctx, r.db).Save(model).Error
}

func (r *mediaWriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainmedia.Media, error) {
	var model MediaModel
	err := DBWithContext(ctx, r.db).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mediaDomainFromModel(&model)
}

func (r *mediaWriteRepository) GetByKey(ctx context.Context, key string) (*domainmedia.Media, error) {
	var model MediaModel
	err := DBWithContext(ctx, r.db).Where("key = ?", key).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mediaDomainFromModel(&model)
}

func mediaModelFromDomain(m *domainmedia.Media) (*MediaModel, error) {
	statuses := make([]string, 0, len(m.Statuses))
	for _, status := range m.Statuses {
		statuses = append(statuses, string(status))
	}
	payload, err := json.Marshal(statuses)
	if err != nil {
		return nil, err
	}

	return &MediaModel{
		ID:          m.ID,
		ScanID:      m.ScanID,
		Key:         m.Key,
		Filename:    m.Filename,
		Thumbnail:   m.Thumbnail,
		ContentType: m.ContentType,
		Size:        m.Size,
		Status:      string(m.Status),
		Statuses:    dbtype.JSONB(payload),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}, nil
}

func mediaDomainFromModel(model *MediaModel) (*domainmedia.Media, error) {
	var raw []string
	if len(model.Statuses) > 0 {
		if err := json.Unmarshal(model.Statuses, &raw); err != nil {
			return nil, err
		}
	}

	statuses := make([]domainmedia.Status, 0, len(raw))
	for _, status := range raw {
		statuses = append(statuses, domainmedia.Status(status))
	}

	return &domainmedia.Media{
		ID:          model.ID,
		ScanID:      model.ScanID,
		Key:         model.Key,
		Filename:    model.Filename,
		Thumbnail:   model.Thumbnail,
		ContentType: model.ContentType,
		Size:        model.Size,
		Status:      domainmedia.Status(model.Status),
		Statuses:    statuses,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}, nil
}
