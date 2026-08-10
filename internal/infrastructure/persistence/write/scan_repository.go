package write

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	domainscan "go-api/internal/domain/scan"
	"go-api/internal/infrastructure/persistence/dbtype"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type scanWriteRepository struct {
	db *gorm.DB
}

func NewScanWriteRepository(db *gorm.DB) domainscan.ScanWriteRepository {
	return &scanWriteRepository{db: db}
}

func (r *scanWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *scanWriteRepository) Save(ctx context.Context, scan *domainscan.Scan) error {
	model, err := scanModelFromDomain(scan)
	if err != nil {
		return err
	}
	return DBWithContext(ctx, r.db).Create(model).Error
}

func (r *scanWriteRepository) Update(ctx context.Context, scan *domainscan.Scan) error {
	model, err := scanModelFromDomain(scan)
	if err != nil {
		return err
	}
	return DBWithContext(ctx, r.db).Save(model).Error
}

func (r *scanWriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainscan.Scan, error) {
	var model ScanModel
	err := DBWithContext(ctx, r.db).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return scanDomainFromModel(&model)
}

func (r *scanWriteRepository) ListInProgressCreatedBefore(
	ctx context.Context,
	before time.Time,
) ([]*domainscan.Scan, error) {
	var models []ScanModel
	err := DBWithContext(ctx, r.db).
		Where("status IN ?", []string{
			string(domainscan.StatusUploaded),
			string(domainscan.StatusProcessing),
		}).
		Where("created_at < ?", before).
		Order("created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	scans := make([]*domainscan.Scan, 0, len(models))
	for i := range models {
		scanEntity, err := scanDomainFromModel(&models[i])
		if err != nil {
			return nil, err
		}
		scans = append(scans, scanEntity)
	}
	return scans, nil
}

func scanModelFromDomain(s *domainscan.Scan) (*ScanModel, error) {
	statuses := make([]string, 0, len(s.Statuses))
	for _, status := range s.Statuses {
		statuses = append(statuses, string(status))
	}
	payload, err := json.Marshal(statuses)
	if err != nil {
		return nil, err
	}

	return &ScanModel{
		ID:         s.ID,
		UserID:     s.UserID,
		Status:     string(s.Status),
		Statuses:   dbtype.JSONB(payload),
		Message:    s.Message,
		FinalScore: s.FinalScore,
		Confidence: string(s.Confidence),
		Verdict:    s.Verdict,
		Duration:   s.Duration,
		RetryCount: s.RetryCount,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}, nil
}

func scanDomainFromModel(m *ScanModel) (*domainscan.Scan, error) {
	var raw []string
	if len(m.Statuses) > 0 {
		if err := json.Unmarshal(m.Statuses, &raw); err != nil {
			return nil, err
		}
	}

	statuses := make([]domainscan.Status, 0, len(raw))
	for _, status := range raw {
		statuses = append(statuses, domainscan.Status(status))
	}

	return &domainscan.Scan{
		ID:         m.ID,
		UserID:     m.UserID,
		Status:     domainscan.Status(m.Status),
		Statuses:   statuses,
		Message:    m.Message,
		FinalScore: m.FinalScore,
		Confidence: domainscan.ConfidenceLevel(m.Confidence),
		Verdict:    m.Verdict,
		Duration:   m.Duration,
		RetryCount: m.RetryCount,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}, nil
}
