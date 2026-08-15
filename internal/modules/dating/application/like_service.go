package application

import (
	"context"
	"fmt"
	"time"

	"github.com/okamyuji/matching-engine/internal/modules/dating/domain"
	"github.com/okamyuji/matching-engine/internal/modules/dating/infrastructure/repository"
)

// LikeService いいね/いいえアクションと相互マッチング検出を処理する
type LikeService struct {
	likeRepo  repository.LikeRepository
	matchRepo repository.MatchRepository
}

// NewLikeService 新しいLikeServiceを作成する
func NewLikeService(
	likeRepo repository.LikeRepository,
	matchRepo repository.MatchRepository,
) *LikeService {
	return &LikeService{
		likeRepo:  likeRepo,
		matchRepo: matchRepo,
	}
}

// SendLike あるユーザーから別のユーザーへいいねを送る
// 処理:
// 1. いいねを保存
// 2. 対象ユーザーもこのユーザーにいいねをしているかチェック
// 3. 相互いいねならマッチを作成
func (s *LikeService) SendLike(
	ctx context.Context,
	fromUserID string,
	targetUserID string,
) (*LikeResponse, error) {
	// 1. いいねを保存
	like := &repository.Like{
		ID:         generateLikeID(fromUserID, targetUserID),
		FromUserID: fromUserID,
		ToUserID:   targetUserID,
		CreatedAt:  time.Now(),
	}

	err := s.likeRepo.Save(ctx, like)
	if err != nil {
		return nil, fmt.Errorf("failed to save like: %w", err)
	}

	// 2. 相互いいねが存在するかチェック
	isMutual, err := s.likeRepo.CheckMutual(ctx, fromUserID, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check mutual like: %w", err)
	}

	// 3. 相互いいねならマッチを作成
	if isMutual {
		match := &domain.Match{
			ID:        generateMatchID(fromUserID, targetUserID),
			UserIDA:   fromUserID,
			UserIDB:   targetUserID,
			Score:     0.0, // Score can be calculated later if needed
			Breakdown: make(map[string]float64),
			MatchedAt: time.Now(),
		}

		err = s.matchRepo.Save(ctx, match)
		if err != nil {
			return nil, fmt.Errorf("failed to save match: %w", err)
		}

		return &LikeResponse{
			Matched: true,
			MatchID: match.ID,
		}, nil
	}

	return &LikeResponse{
		Matched: false,
	}, nil
}

// GetReceivedLikes ユーザーが受け取った全てのいいねを取得する
func (s *LikeService) GetReceivedLikes(
	ctx context.Context,
	userID string,
) ([]*repository.Like, error) {
	likes, err := s.likeRepo.FindByTargetUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find received likes: %w", err)
	}
	return likes, nil
}

// generateLikeID いいね用のユニークIDを生成する
func generateLikeID(fromUserID, toUserID string) string {
	return fmt.Sprintf("like_%s_%s_%d", fromUserID, toUserID, time.Now().UnixNano())
}

// generateMatchID マッチ用のユニークIDを生成する
func generateMatchID(userA, userB string) string {
	return fmt.Sprintf("match_%s_%s_%d", userA, userB, time.Now().UnixNano())
}
