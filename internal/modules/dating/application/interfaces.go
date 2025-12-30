package application

import (
	"context"

	"github.com/yourorg/matching-engine/internal/modules/dating/infrastructure/repository"
)

// MatchingServiceInterface マッチングサービスインターフェース
type MatchingServiceInterface interface {
	FindMatches(ctx context.Context, userID string, limit int) ([]*MatchResult, error)
	GetMutualMatches(ctx context.Context, userID string) ([]*MutualMatchResult, error)
}

// LikeServiceInterface いいねサービスインターフェース
type LikeServiceInterface interface {
	SendLike(ctx context.Context, fromUserID, toUserID string) (*LikeResponse, error)
	GetReceivedLikes(ctx context.Context, userID string) ([]*repository.Like, error)
}
