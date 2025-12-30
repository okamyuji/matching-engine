package repository

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/yourorg/matching-engine/internal/modules/dating/domain"
)

// MatchRepository マッチデータアクセス用インターフェース
type MatchRepository interface {
	Save(ctx context.Context, match *domain.Match) error
	FindByUserID(ctx context.Context, userID string) ([]*domain.Match, error)
	FindMutual(ctx context.Context, userID string) ([]*domain.Match, error)
}

// matchRepository MatchRepositoryのBUN実装
type matchRepository struct {
	db *bun.DB
}

// NewMatchRepository 新しいMatchRepositoryを作成する
func NewMatchRepository(db *bun.DB) MatchRepository {
	return &matchRepository{db: db}
}

// Save 新しいマッチを挿入する
func (r *matchRepository) Save(ctx context.Context, match *domain.Match) error {
	_, err := r.db.NewInsert().
		Model(match).
		Exec(ctx)
	return err
}

// FindByUserID ユーザーの全マッチを取得する
func (r *matchRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Match, error) {
	var matches []*domain.Match
	err := r.db.NewSelect().
		Model(&matches).
		Where("user_id_a = ? OR user_id_b = ?", userID, userID).
		Order("matched_at DESC").
		Scan(ctx)
	return matches, err
}

// FindMutual ユーザーの相互マッチを取得する
// 相互マッチ = Matchが存在し、かつ両方向のLikeも存在するもの
func (r *matchRepository) FindMutual(ctx context.Context, userID string) ([]*domain.Match, error) {
	var matches []*domain.Match

	// Matchエントリーを取得し、両方向のLikeが存在することを確認
	err := r.db.NewSelect().
		Model(&matches).
		Where("user_id_a = ? OR user_id_b = ?", userID, userID).
		// 相互いいねが両方存在することを確認（A→BとB→A）
		Where("EXISTS (SELECT 1 FROM dating_likes WHERE from_user_id = user_id_a AND to_user_id = user_id_b)").
		Where("EXISTS (SELECT 1 FROM dating_likes WHERE from_user_id = user_id_b AND to_user_id = user_id_a)").
		Order("matched_at DESC").
		Scan(ctx)

	return matches, err
}
