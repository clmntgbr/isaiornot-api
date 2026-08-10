package read

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"go-api/internal/domain/aggregate"
	"go-api/internal/domain/paginate"
	domainscan "go-api/internal/domain/scan"
	"go-api/internal/infrastructure/persistence/dbtype"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type scanRow struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Status     string
	Statuses   dbtype.JSONB
	Message    string
	FinalScore float64
	Confidence string
	Verdict    string
	Duration   int
	RetryCount int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (scanRow) TableName() string { return "scans" }

type scanReadRepository struct {
	db *gorm.DB
}

func NewScanReadRepository(db *gorm.DB) domainscan.ScanReadRepository {
	return &scanReadRepository{db: db}
}

func (r *scanReadRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainscan.ScanView, error) {
	var row scanRow
	err := r.db.WithContext(ctx).
		Select(
			"id", "user_id", "status", "statuses", "message",
			"final_score", "confidence", "verdict", "duration", "retry_count",
			"created_at", "updated_at",
		).
		First(&row, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toScanView(row)
}

func (r *scanReadRepository) FindByUserID(
	ctx context.Context,
	userID uuid.UUID,
	query paginate.PaginateQuery,
) ([]*domainscan.ScanView, int64, error) {
	query.Normalize()
	if query.SortBy == "" {
		query.SortBy = "created_at"
	}
	if query.OrderBy != paginate.OrderByAsc {
		query.OrderBy = paginate.OrderByDesc
	}

	db := r.db.WithContext(ctx).Model(&scanRow{}).Where("user_id = ?", userID)

	db, total, err := Paginate(db, query)
	if err != nil {
		return nil, 0, err
	}

	var rows []scanRow
	if err := db.
		Select(
			"id", "user_id", "status", "statuses", "message",
			"final_score", "confidence", "verdict", "duration", "retry_count",
			"created_at", "updated_at",
		).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	views := make([]*domainscan.ScanView, 0, len(rows))
	for _, row := range rows {
		view, err := toScanView(row)
		if err != nil {
			return nil, 0, err
		}
		views = append(views, view)
	}

	return views, total, nil
}

func (r *scanReadRepository) GetStatisticsByUserID(ctx context.Context, userID uuid.UUID) (*domainscan.StatisticsView, error) {
	var stats domainscan.StatisticsView
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*) FILTER (WHERE verdict <> '') AS scans_count,
			COUNT(*) FILTER (WHERE verdict = ?) AS real_image_count,
			COUNT(*) FILTER (WHERE verdict = ?) AS ai_image_count,
			COALESCE(AVG(final_score) FILTER (WHERE verdict <> '' AND final_score >= 0), 0) AS average_score
		FROM scans
		WHERE user_id = ?
	`, aggregate.VerdictLikelyReal, aggregate.VerdictLikelyAI, userID).Scan(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func toScanView(row scanRow) (*domainscan.ScanView, error) {
	var raw []string
	if len(row.Statuses) > 0 {
		if err := json.Unmarshal(row.Statuses, &raw); err != nil {
			return nil, err
		}
	}

	statuses := make([]domainscan.Status, 0, len(raw))
	for _, status := range raw {
		statuses = append(statuses, domainscan.Status(status))
	}

	return &domainscan.ScanView{
		ID:         row.ID,
		UserID:     row.UserID,
		Status:     domainscan.Status(row.Status),
		Statuses:   statuses,
		Message:    row.Message,
		FinalScore: row.FinalScore,
		Confidence: domainscan.ConfidenceLevel(row.Confidence),
		Verdict:    row.Verdict,
		Duration:   row.Duration,
		RetryCount: row.RetryCount,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}, nil
}
