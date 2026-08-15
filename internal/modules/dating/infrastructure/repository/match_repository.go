package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/okamyuji/matching-engine/internal/modules/dating/domain"
	"github.com/okamyuji/matching-engine/internal/modules/dating/infrastructure/repository/sqlcgen"
)

// MatchRepository マッチデータアクセス用インターフェース
type MatchRepository interface {
	Save(ctx context.Context, match *domain.Match) error
	FindByUserID(ctx context.Context, userID string) ([]*domain.Match, error)
	FindMutual(ctx context.Context, userID string) ([]*domain.Match, error)
}

// matchRepository MatchRepository の sqlc 実装
type matchRepository struct {
	q *sqlcgen.Queries
}

// NewMatchRepository 新しい MatchRepository を作成する
func NewMatchRepository(db DB) MatchRepository {
	return &matchRepository{q: sqlcgen.New(db)}
}

// Save 新しいマッチを挿入する
func (r *matchRepository) Save(ctx context.Context, match *domain.Match) error {
	breakdown, err := marshalBreakdown(match.Breakdown)
	if err != nil {
		return err
	}
	matchedAt := match.MatchedAt
	if matchedAt.IsZero() {
		matchedAt = time.Now()
	}
	return r.q.InsertMatch(ctx, sqlcgen.InsertMatchParams{
		ID:        match.ID,
		UserIDA:   match.UserIDA,
		UserIDB:   match.UserIDB,
		Score:     match.Score,
		Breakdown: breakdown,
		MatchedAt: matchedAt,
	})
}

// FindByUserID ユーザーが関わる全てのマッチを新しい順に取得する
func (r *matchRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Match, error) {
	rows, err := r.q.ListMatchesByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return matchesFromRows(rows)
}

// FindMutual 双方向にいいねが存在する（相互）マッチだけを取得する
func (r *matchRepository) FindMutual(ctx context.Context, userID string) ([]*domain.Match, error) {
	rows, err := r.q.ListMutualMatchesByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return matchesFromRows(rows)
}

func matchesFromRows(rows []sqlcgen.DatingMatch) ([]*domain.Match, error) {
	matches := make([]*domain.Match, 0, len(rows))
	for _, row := range rows {
		breakdown, err := unmarshalBreakdown(row.Breakdown)
		if err != nil {
			return nil, fmt.Errorf("match %s breakdown: %w", row.ID, err)
		}
		matches = append(matches, &domain.Match{
			ID:        row.ID,
			UserIDA:   row.UserIDA,
			UserIDB:   row.UserIDB,
			Score:     row.Score,
			Breakdown: breakdown,
			MatchedAt: row.MatchedAt,
		})
	}
	return matches, nil
}

func marshalBreakdown(b map[string]float64) ([]byte, error) {
	if b == nil {
		return nil, nil //nolint:nilnil // 内訳なしは NULL として保存する
	}
	data, err := json.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("marshal breakdown: %w", err)
	}
	return data, nil
}

func unmarshalBreakdown(data []byte) (map[string]float64, error) {
	if len(data) == 0 {
		return nil, nil //nolint:nilnil // 内訳なし（NULL）は正常系
	}
	var b map[string]float64
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, err
	}
	return b, nil
}
