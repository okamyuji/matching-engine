package repository

import (
	"context"
	"time"

	"github.com/okamyuji/matching-engine/internal/modules/dating/infrastructure/repository/sqlcgen"
)

// Like あるユーザーから別のユーザーへのいいね
type Like struct {
	ID         string
	FromUserID string
	ToUserID   string
	CreatedAt  time.Time
}

// LikeRepository いいねデータアクセス用インターフェース
type LikeRepository interface {
	Save(ctx context.Context, like *Like) error
	FindByTargetUserID(ctx context.Context, targetUserID string) ([]*Like, error)
	CheckMutual(ctx context.Context, userA, userB string) (bool, error)
	FindByFromUserID(ctx context.Context, fromUserID string) ([]*Like, error)
}

// likeRepository LikeRepository の sqlc 実装
type likeRepository struct {
	q *sqlcgen.Queries
}

// NewLikeRepository 新しい LikeRepository を作成する
func NewLikeRepository(db DB) LikeRepository {
	return &likeRepository{q: sqlcgen.New(db)}
}

// Save 新しいいいねを挿入する
func (r *likeRepository) Save(ctx context.Context, like *Like) error {
	createdAt := like.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	return r.q.InsertLike(ctx, sqlcgen.InsertLikeParams{
		ID:         like.ID,
		FromUserID: like.FromUserID,
		ToUserID:   like.ToUserID,
		CreatedAt:  createdAt,
	})
}

// FindByTargetUserID ユーザーが受け取った全てのいいねを新しい順に取得する
func (r *likeRepository) FindByTargetUserID(ctx context.Context, targetUserID string) ([]*Like, error) {
	rows, err := r.q.ListLikesByToUser(ctx, targetUserID)
	if err != nil {
		return nil, err
	}
	return likesFromRows(rows), nil
}

// CheckMutual 2人のユーザー間に相互いいねが存在するかチェックする
func (r *likeRepository) CheckMutual(ctx context.Context, userA, userB string) (bool, error) {
	ab, err := r.q.LikeExists(ctx, sqlcgen.LikeExistsParams{FromUserID: userA, ToUserID: userB})
	if err != nil {
		return false, err
	}
	if !ab {
		return false, nil
	}
	ba, err := r.q.LikeExists(ctx, sqlcgen.LikeExistsParams{FromUserID: userB, ToUserID: userA})
	if err != nil {
		return false, err
	}
	return ba, nil
}

// FindByFromUserID ユーザーが送った全てのいいねを新しい順に取得する
func (r *likeRepository) FindByFromUserID(ctx context.Context, fromUserID string) ([]*Like, error) {
	rows, err := r.q.ListLikesByFromUser(ctx, fromUserID)
	if err != nil {
		return nil, err
	}
	return likesFromRows(rows), nil
}

func likesFromRows(rows []sqlcgen.DatingLike) []*Like {
	likes := make([]*Like, 0, len(rows))
	for _, row := range rows {
		likes = append(likes, &Like{ID: row.ID, FromUserID: row.FromUserID, ToUserID: row.ToUserID, CreatedAt: row.CreatedAt})
	}
	return likes
}
