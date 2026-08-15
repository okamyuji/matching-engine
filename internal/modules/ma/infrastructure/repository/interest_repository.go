package repository

import (
	"context"
	"time"

	"github.com/okamyuji/matching-engine/internal/modules/ma/domain"
	"github.com/okamyuji/matching-engine/internal/modules/ma/infrastructure/repository/sqlcgen"
)

// InterestRepository 関心表明データアクセス用インターフェース
type InterestRepository interface {
	Save(ctx context.Context, interest *domain.Interest) error
	FindByToCompany(ctx context.Context, toCompanyID string) ([]*domain.Interest, error)
	Exists(ctx context.Context, fromCompanyID, toCompanyID string) (bool, error)
}

// interestRepository InterestRepository の sqlc 実装
type interestRepository struct {
	q *sqlcgen.Queries
}

// NewInterestRepository 新しい InterestRepository を作成する
func NewInterestRepository(db DB) InterestRepository {
	return &interestRepository{q: sqlcgen.New(db)}
}

// Save 新しい関心表明を挿入する
func (r *interestRepository) Save(ctx context.Context, interest *domain.Interest) error {
	createdAt := interest.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	return r.q.InsertInterest(ctx, sqlcgen.InsertInterestParams{
		ID:            interest.ID,
		FromCompanyID: interest.FromCompanyID,
		ToCompanyID:   interest.ToCompanyID,
		CreatedAt:     createdAt,
	})
}

// FindByToCompany 企業が受け取った関心表明を新しい順に取得する
func (r *interestRepository) FindByToCompany(ctx context.Context, toCompanyID string) ([]*domain.Interest, error) {
	rows, err := r.q.ListInterestsByToCompany(ctx, toCompanyID)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Interest, 0, len(rows))
	for _, row := range rows {
		out = append(out, &domain.Interest{ID: row.ID, FromCompanyID: row.FromCompanyID, ToCompanyID: row.ToCompanyID, CreatedAt: row.CreatedAt})
	}
	return out, nil
}

// Exists 指定の向きの関心表明が存在するかを返す
func (r *interestRepository) Exists(ctx context.Context, fromCompanyID, toCompanyID string) (bool, error) {
	return r.q.InterestExists(ctx, sqlcgen.InterestExistsParams{FromCompanyID: fromCompanyID, ToCompanyID: toCompanyID})
}
