package gorm

import (
	"context"
	"errors"
	"go-api/domain/entity"
	"go-api/domain/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type signalRepository struct {
	db *gorm.DB
}

func NewSignalRepository(db *gorm.DB) repository.SignalRepository {
	return &signalRepository{db: db}
}

func (r *signalRepository) Create(ctx context.Context, signal *entity.Signal) error {
	return dbWithContext(ctx, r.db).Create(signal).Error
}

func (r *signalRepository) Update(ctx context.Context, signal *entity.Signal) error {
	return dbWithContext(ctx, r.db).Save(signal).Error
}

func (r *signalRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return dbWithContext(ctx, r.db).Delete(&entity.Signal{}, id).Error
}

func (r *signalRepository) DeleteByMediaID(ctx context.Context, mediaID uuid.UUID) error {
	return dbWithContext(ctx, r.db).Where("media_id = ?", mediaID).Delete(&entity.Signal{}).Error
}

func (r *signalRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Signal, error) {
	var signal entity.Signal
	err := dbWithContext(ctx, r.db).Where("id = ?", id).First(&signal).Error
	if err != nil {
		return nil, err
	}
	if signal.ID == uuid.Nil {
		return nil, errors.New("signal not found")
	}

	return &signal, nil
}

func (r *signalRepository) GetByMediaID(ctx context.Context, mediaID uuid.UUID) ([]*entity.Signal, error) {
	var signals []*entity.Signal
	err := dbWithContext(ctx, r.db).
		Where("media_id = ?", mediaID).
		Order("created_at ASC").
		Find(&signals).Error
	if err != nil {
		return nil, err
	}

	return signals, nil
}
