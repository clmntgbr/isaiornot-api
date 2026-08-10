package write

import (
	"context"
	"encoding/json"
	"errors"

	domainsignal "go-api/internal/domain/signal"
	"go-api/internal/infrastructure/persistence/dbtype"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type signalWriteRepository struct {
	db *gorm.DB
}

func NewSignalWriteRepository(db *gorm.DB) domainsignal.SignalWriteRepository {
	return &signalWriteRepository{db: db}
}

func (r *signalWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *signalWriteRepository) Save(ctx context.Context, signal *domainsignal.Signal) error {
	model, err := signalModelFromDomain(signal)
	if err != nil {
		return err
	}
	return DBWithContext(ctx, r.db).Create(model).Error
}

func (r *signalWriteRepository) Update(ctx context.Context, signal *domainsignal.Signal) error {
	model, err := signalModelFromDomain(signal)
	if err != nil {
		return err
	}
	return DBWithContext(ctx, r.db).Save(model).Error
}

func (r *signalWriteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return DBWithContext(ctx, r.db).Delete(&SignalModel{}, "id = ?", id).Error
}

func (r *signalWriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainsignal.Signal, error) {
	var model SignalModel
	err := DBWithContext(ctx, r.db).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return signalDomainFromModel(&model)
}

func (r *signalWriteRepository) GetByMediaIDAndName(
	ctx context.Context,
	mediaID uuid.UUID,
	name string,
) (*domainsignal.Signal, error) {
	var model SignalModel
	err := DBWithContext(ctx, r.db).Where("media_id = ? AND name = ?", mediaID, name).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return signalDomainFromModel(&model)
}

func (r *signalWriteRepository) GetByMediaID(ctx context.Context, mediaID uuid.UUID) ([]*domainsignal.Signal, error) {
	var models []SignalModel
	err := DBWithContext(ctx, r.db).Where("media_id = ?", mediaID).Order("created_at asc").Find(&models).Error
	if err != nil {
		return nil, err
	}
	result := make([]*domainsignal.Signal, 0, len(models))
	for i := range models {
		signalEntity, err := signalDomainFromModel(&models[i])
		if err != nil {
			return nil, err
		}
		result = append(result, signalEntity)
	}
	return result, nil
}

func signalModelFromDomain(s *domainsignal.Signal) (*SignalModel, error) {
	details := s.Details
	if details == nil {
		details = []string{}
	}
	payload, err := json.Marshal(details)
	if err != nil {
		return nil, err
	}

	return &SignalModel{
		ID:         s.ID,
		MediaID:    s.MediaID,
		Name:       s.Name,
		Score:      s.Score,
		Confidence: string(s.Confidence),
		Details:    dbtype.JSONB(payload),
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}, nil
}

func signalDomainFromModel(model *SignalModel) (*domainsignal.Signal, error) {
	var details []string
	if len(model.Details) > 0 {
		if err := json.Unmarshal(model.Details, &details); err != nil {
			return nil, err
		}
	}
	if details == nil {
		details = []string{}
	}

	return &domainsignal.Signal{
		ID:         model.ID,
		MediaID:    model.MediaID,
		Name:       model.Name,
		Score:      model.Score,
		Confidence: domainsignal.ConfidenceLevel(model.Confidence),
		Details:    details,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
	}, nil
}
