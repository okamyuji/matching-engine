package repository

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// Like あるユーザーから別のユーザーへのいいね
type Like struct {
	bun.BaseModel `bun:"table:dating_likes"`

	ID         string    `bun:"id,pk"`
	FromUserID string    `bun:"from_user_id,notnull"`
	ToUserID   string    `bun:"to_user_id,notnull"`
	CreatedAt  time.Time `bun:"created_at,notnull"`
}

// LikeRepository いいねデータアクセス用インターフェース
type LikeRepository interface {
	Save(ctx context.Context, like *Like) error
	FindByTargetUserID(ctx context.Context, targetUserID string) ([]*Like, error)
	CheckMutual(ctx context.Context, userA, userB string) (bool, error)
	FindByFromUserID(ctx context.Context, fromUserID string) ([]*Like, error)
}

// likeRepository LikeRepositoryのBUN実装
type likeRepository struct {
	db *bun.DB
}

// NewLikeRepository 新しいLikeRepositoryを作成する
func NewLikeRepository(db *bun.DB) LikeRepository {
	return &likeRepository{db: db}
}

// Save 新しいいいねを挿入する
func (r *likeRepository) Save(ctx context.Context, like *Like) error {
	_, err := r.db.NewInsert().
		Model(like).
		Exec(ctx)
	return err
}

// FindByTargetUserID ユーザーが受け取った全てのいいねを取得する
func (r *likeRepository) FindByTargetUserID(ctx context.Context, targetUserID string) ([]*Like, error) {
	var likes []*Like
	err := r.db.NewSelect().
		Model(&likes).
		Where("to_user_id = ?", targetUserID).
		Order("created_at DESC").
		Scan(ctx)
	return likes, err
}

// CheckMutual 2人のユーザー間に相互いいねが存在するかチェックする
func (r *likeRepository) CheckMutual(ctx context.Context, userA, userB string) (bool, error) {
	// ユーザーAがユーザーBにいいねしているかチェック
	countAB, err := r.db.NewSelect().
		Model((*Like)(nil)).
		Where("from_user_id = ? AND to_user_id = ?", userA, userB).
		Count(ctx)
	if err != nil {
		return false, err
	}

	// ユーザーBがユーザーAにいいねしているかチェック
	countBA, err := r.db.NewSelect().
		Model((*Like)(nil)).
		Where("from_user_id = ? AND to_user_id = ?", userB, userA).
		Count(ctx)
	if err != nil {
		return false, err
	}

	return countAB > 0 && countBA > 0, nil
}

// FindByFromUserID ユーザーが送った全てのいいねを取得する
func (r *likeRepository) FindByFromUserID(ctx context.Context, fromUserID string) ([]*Like, error) {
	var likes []*Like
	err := r.db.NewSelect().
		Model(&likes).
		Where("from_user_id = ?", fromUserID).
		Order("created_at DESC").
		Scan(ctx)
	return likes, err
}
