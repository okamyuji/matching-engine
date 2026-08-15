package application

import (
	"context"
	"fmt"

	"github.com/okamyuji/matching-engine/internal/core/matching"
	"github.com/okamyuji/matching-engine/internal/modules/dating/domain"
	"github.com/okamyuji/matching-engine/internal/modules/dating/infrastructure/mapper"
	"github.com/okamyuji/matching-engine/internal/modules/dating/infrastructure/repository"
)

// DatingMatchingService デートマッチング処理を統括する
type DatingMatchingService struct {
	engine         *matching.ConfigurableEngine
	userRepo       repository.UserRepository
	profileRepo    repository.ProfileRepository
	preferenceRepo repository.PreferenceRepository
	matchRepo      repository.MatchRepository
	featureMapper  *mapper.DatingFeatureMapper
}

// NewDatingMatchingService 新しいDatingMatchingServiceを作成する
func NewDatingMatchingService(
	engine *matching.ConfigurableEngine,
	userRepo repository.UserRepository,
	profileRepo repository.ProfileRepository,
	preferenceRepo repository.PreferenceRepository,
	matchRepo repository.MatchRepository,
	featureMapper *mapper.DatingFeatureMapper,
) *DatingMatchingService {
	return &DatingMatchingService{
		engine:         engine,
		userRepo:       userRepo,
		profileRepo:    profileRepo,
		preferenceRepo: preferenceRepo,
		matchRepo:      matchRepo,
		featureMapper:  featureMapper,
	}
}

// FindMatches ユーザーのマッチング候補を取得する
// 処理:
// 1. ユーザー情報を取得
// 2. 候補を取得（設定によりフィルタリング）
// 3. 特徴ベクトルに変換
// 4. マッチングエンジンを実行
// 5. DTOに変換
func (s *DatingMatchingService) FindMatches(
	ctx context.Context,
	userID string,
	limit int,
) ([]*MatchResult, error) {
	// 1. ユーザー情報を取得
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	profile, err := s.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find profile: %w", err)
	}

	// ユーザーの希望条件を取得
	var pref *domain.Preference
	if s.preferenceRepo != nil {
		pref, err = s.preferenceRepo.FindByUserID(ctx, userID)
		if err != nil {
			// 設定が見つからない場合はデフォルト設定を使用
			pref = &domain.Preference{
				UserID: userID,
				AgeMin: 18,
				AgeMax: 80,
			}
		}
	} else {
		// preferenceRepoがnilの場合はデフォルト設定を使用（テスト時）
		pref = &domain.Preference{
			UserID: userID,
			AgeMin: 18,
			AgeMax: 80,
		}
	}

	// 2. 候補を取得（設定によりフィルタリング）
	candidates, err := s.userRepo.FindCandidates(ctx, userID, pref)
	if err != nil {
		return nil, fmt.Errorf("failed to find candidates: %w", err)
	}

	if len(candidates) == 0 {
		return []*MatchResult{}, nil
	}

	// 3. 特徴ベクトルに変換
	sourceVector := s.featureMapper.ToFeatureVector(user, profile)

	candidateVectors := make([]*matching.FeatureVector, len(candidates))
	for i, c := range candidates {
		candidateVectors[i] = s.featureMapper.ToFeatureVector(c.User, c.Profile)
	}

	// 4. Execute matching engine
	matches, err := s.engine.FindMatches(ctx, sourceVector, candidateVectors)
	if err != nil {
		return nil, fmt.Errorf("failed to compute matches: %w", err)
	}

	// 5. Convert to DTOs
	results := make([]*MatchResult, 0, len(matches))
	for _, m := range matches {
		if len(results) >= limit {
			break
		}
		results = append(results, &MatchResult{
			UserID:    m.Candidate.ID,
			Score:     m.Score,
			Rank:      m.Rank,
			Breakdown: m.Breakdown,
		})
	}

	return results, nil
}

// GetMutualMatches ユーザーの相互マッチを取得する
// 処理:
// 1. MatchRepositoryから相互マッチを取得
// 2. DTOに変換して返す
func (s *DatingMatchingService) GetMutualMatches(
	ctx context.Context,
	userID string,
) ([]*MutualMatchResult, error) {
	// 1. 相互マッチを取得
	matches, err := s.matchRepo.FindMutual(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find mutual matches: %w", err)
	}

	// 2. DTOに変換
	results := make([]*MutualMatchResult, 0, len(matches))
	for _, m := range matches {
		results = append(results, &MutualMatchResult{
			MatchID:   m.ID,
			UserIDA:   m.UserIDA,
			UserIDB:   m.UserIDB,
			Score:     m.Score,
			Breakdown: m.Breakdown,
			MatchedAt: m.MatchedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return results, nil
}
