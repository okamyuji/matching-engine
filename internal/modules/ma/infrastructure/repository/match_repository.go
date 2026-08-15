package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/okamyuji/matching-engine/internal/modules/ma/domain"
	"github.com/okamyuji/matching-engine/internal/modules/ma/infrastructure/repository/sqlcgen"
)

// MAMatchRepository M&Aマッチデータアクセス用インターフェース
type MAMatchRepository interface {
	Save(ctx context.Context, match *domain.MAMatch) error
	FindMutual(ctx context.Context, companyID string) ([]*domain.MAMatch, error)
	FindByID(ctx context.Context, matchID string) (*domain.MAMatch, error)
}

// maMatchRepository MAMatchRepository の sqlc 実装
type maMatchRepository struct {
	q *sqlcgen.Queries
}

// NewMAMatchRepository 新しい MAMatchRepository を作成する
func NewMAMatchRepository(db DB) MAMatchRepository {
	return &maMatchRepository{q: sqlcgen.New(db)}
}

// Save 新しいマッチを挿入する
func (r *maMatchRepository) Save(ctx context.Context, match *domain.MAMatch) error {
	breakdown, err := marshalJSON(match.Breakdown)
	if err != nil {
		return fmt.Errorf("marshal breakdown: %w", err)
	}
	var synergy []byte
	if match.SynergySummary != nil {
		synergy, err = json.Marshal(match.SynergySummary)
		if err != nil {
			return fmt.Errorf("marshal synergy summary: %w", err)
		}
	}
	matchedAt := match.MatchedAt
	if matchedAt.IsZero() {
		matchedAt = time.Now()
	}
	return r.q.InsertMAMatch(ctx, sqlcgen.InsertMAMatchParams{
		ID:             match.ID,
		CompanyIDA:     match.CompanyIDA,
		CompanyIDB:     match.CompanyIDB,
		Score:          match.Score,
		Breakdown:      breakdown,
		SynergySummary: synergy,
		MatchedAt:      matchedAt,
	})
}

// FindMutual 双方向に関心表明がある（相互）マッチだけを取得する
func (r *maMatchRepository) FindMutual(ctx context.Context, companyID string) ([]*domain.MAMatch, error) {
	rows, err := r.q.ListMutualMAMatchesByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.MAMatch, 0, len(rows))
	for _, row := range rows {
		m, err := maMatchFromRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

// FindByID IDによりマッチを取得する
func (r *maMatchRepository) FindByID(ctx context.Context, matchID string) (*domain.MAMatch, error) {
	row, err := r.q.GetMAMatch(ctx, matchID)
	if err != nil {
		return nil, err
	}
	return maMatchFromRow(row)
}

func maMatchFromRow(row sqlcgen.MaMatch) (*domain.MAMatch, error) {
	var breakdown map[string]float64
	if len(row.Breakdown) > 0 {
		if err := json.Unmarshal(row.Breakdown, &breakdown); err != nil {
			return nil, fmt.Errorf("match %s breakdown: %w", row.ID, err)
		}
	}
	var synergy *domain.SynergySummary
	if len(row.SynergySummary) > 0 {
		synergy = &domain.SynergySummary{}
		if err := json.Unmarshal(row.SynergySummary, synergy); err != nil {
			return nil, fmt.Errorf("match %s synergy summary: %w", row.ID, err)
		}
	}
	return &domain.MAMatch{
		ID:             row.ID,
		CompanyIDA:     row.CompanyIDA,
		CompanyIDB:     row.CompanyIDB,
		Score:          row.Score,
		Breakdown:      breakdown,
		SynergySummary: synergy,
		MatchedAt:      row.MatchedAt,
	}, nil
}

func marshalJSON(v map[string]float64) ([]byte, error) {
	if v == nil {
		return nil, nil //nolint:nilnil // 内訳なしは NULL として保存する
	}
	return json.Marshal(v)
}
