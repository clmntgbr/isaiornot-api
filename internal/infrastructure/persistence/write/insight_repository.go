package write

import (
	"context"
	"errors"

	domaininsight "go-api/internal/domain/insight"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type insightWriteRepository struct {
	db *gorm.DB
}

func NewInsightWriteRepository(db *gorm.DB) domaininsight.InsightWriteRepository {
	return &insightWriteRepository{db: db}
}

func (r *insightWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *insightWriteRepository) Save(ctx context.Context, insight *domaininsight.Insight) error {
	return DBWithContext(ctx, r.db).Create(insightModelFromDomain(insight)).Error
}

func (r *insightWriteRepository) Update(ctx context.Context, insight *domaininsight.Insight) error {
	return DBWithContext(ctx, r.db).Save(insightModelFromDomain(insight)).Error
}

func (r *insightWriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domaininsight.Insight, error) {
	var model InsightModel
	err := DBWithContext(ctx, r.db).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return insightDomainFromModel(&model), nil
}

func insightModelFromDomain(i *domaininsight.Insight) *InsightModel {
	return &InsightModel{
		ID:          i.ID,
		Noise:       i.Noise,
		Compression: i.Compression,
		Frequency:   i.Frequency,
		Histogram:   i.Histogram,
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   i.UpdatedAt,
	}
}

func insightDomainFromModel(model *InsightModel) *domaininsight.Insight {
	return &domaininsight.Insight{
		ID:          model.ID,
		Noise:       model.Noise,
		Compression: model.Compression,
		Frequency:   model.Frequency,
		Histogram:   model.Histogram,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}
