package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/okamyuji/matching-engine/internal/core/matching"
	"github.com/okamyuji/matching-engine/internal/modules/dating/domain"
	"github.com/okamyuji/matching-engine/internal/modules/dating/infrastructure/mapper"
)

// 注記: 完全なサービステストにはTestContainersを使用した結合テストが必要
// このテストはサービスの構築のみを検証する

func TestNewDatingMatchingService(t *testing.T) {
	// エンジン用の最小限の設定を作成
	config := &matching.MatchingConfig{
		Version: "1.0",
		Domain:  "test",
		Scoring: matching.ScoringConfig{
			Components: []matching.ComponentConfig{
				{
					Name:   "test",
					Type:   "euclidean",
					Fields: []string{"age"},
					Weight: 1.0,
				},
			},
		},
		Ranking: matching.RankingConfig{
			SortOrder: "desc",
		},
	}

	engine, err := matching.NewConfigurableEngine(config)
	if err != nil {
		t.Fatalf("NewConfigurableEngine() error = %v", err)
	}

	service := NewDatingMatchingService(
		engine,
		nil, // userRepo（構築テストでは未使用）
		nil, // profileRepo（構築テストでは未使用）
		nil, // preferenceRepo（構築テストでは未使用）
		nil, // matchRepo（構築テストでは未使用）
		mapper.NewDatingFeatureMapper(),
	)

	if service == nil {
		t.Fatal("NewDatingMatchingService() returned nil")
	}
	if service.engine == nil {
		t.Error("engine should not be nil")
	}
	if service.featureMapper == nil {
		t.Error("featureMapper should not be nil")
	}
}

func TestNewLikeService(t *testing.T) {
	service := NewLikeService(
		nil, // likeRepo（構築テストでは未使用）
		nil, // matchRepo（構築テストでは未使用）
	)

	if service == nil {
		t.Error("NewLikeService() returned nil")
	}
}

func TestGenerateLikeID(t *testing.T) {
	id1 := generateLikeID("user1", "user2")

	if id1 == "" {
		t.Error("generateLikeID() returned empty string")
	}

	// フォーマットに期待される要素が含まれているか確認
	if len(id1) < len("like_user1_user2_") {
		t.Error("generateLikeID() returned unexpected format")
	}

	// IDが期待されるプレフィックスで始まるか確認
	expectedPrefix := "like_user1_user2_"
	if len(id1) < len(expectedPrefix) || id1[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("generateLikeID() = %v, should start with %v", id1, expectedPrefix)
	}
}

func TestGenerateMatchID(t *testing.T) {
	id1 := generateMatchID("user1", "user2")

	if id1 == "" {
		t.Error("generateMatchID() returned empty string")
	}

	// フォーマットに期待される要素が含まれているか確認
	if len(id1) < len("match_user1_user2_") {
		t.Error("generateMatchID() returned unexpected format")
	}

	// IDが期待されるプレフィックスで始まるか確認
	expectedPrefix := "match_user1_user2_"
	if len(id1) < len(expectedPrefix) || id1[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("generateMatchID() = %v, should start with %v", id1, expectedPrefix)
	}
}

// 注記: 単純な構築テストにはモックリポジトリは不要
// TestContainersを使用した結合テストでは実際のデータベース実装を使用する

func TestLikeService_SendLike_NoMutual(t *testing.T) {
	likeRepo := newMockLikeRepository()
	matchRepo := newMockMatchRepository()
	service := NewLikeService(likeRepo, matchRepo)

	ctx := context.Background()
	resp, err := service.SendLike(ctx, "user1", "user2")

	if err != nil {
		t.Fatalf("SendLike() error = %v", err)
	}
	if resp.Matched {
		t.Error("SendLike() matched = true, want false")
	}
	if resp.MatchID != "" {
		t.Error("SendLike() MatchID should be empty when not matched")
	}
	if len(likeRepo.likes) != 1 {
		t.Errorf("SendLike() saved %v likes, want 1", len(likeRepo.likes))
	}
}

func TestLikeService_SendLike_Mutual(t *testing.T) {
	likeRepo := newMockLikeRepository()
	likeRepo.mutualCheck = true
	matchRepo := newMockMatchRepository()
	service := NewLikeService(likeRepo, matchRepo)

	ctx := context.Background()
	resp, err := service.SendLike(ctx, "user1", "user2")

	if err != nil {
		t.Fatalf("SendLike() error = %v", err)
	}
	if !resp.Matched {
		t.Error("SendLike() matched = false, want true")
	}
	if resp.MatchID == "" {
		t.Error("SendLike() MatchID should not be empty when matched")
	}
	if len(matchRepo.matches) != 1 {
		t.Errorf("SendLike() saved %v matches, want 1", len(matchRepo.matches))
	}
}

func TestLikeService_SendLike_SaveError(t *testing.T) {
	likeRepo := newMockLikeRepository()
	likeRepo.err = errors.New("save error")
	matchRepo := newMockMatchRepository()
	service := NewLikeService(likeRepo, matchRepo)

	ctx := context.Background()
	_, err := service.SendLike(ctx, "user1", "user2")

	if err == nil {
		t.Error("SendLike() error = nil, want error")
	}
}

func TestLikeService_SendLike_CheckMutualError(t *testing.T) {
	likeRepo := newMockLikeRepository()
	matchRepo := newMockMatchRepository()
	service := NewLikeService(likeRepo, matchRepo)

	ctx := context.Background()
	// First call succeeds to save the like
	if _, err := service.SendLike(ctx, "user1", "user2"); err != nil {
		t.Fatalf("First SendLike() unexpected error = %v", err)
	}

	// Now set error for CheckMutual
	likeRepo.err = errors.New("check mutual error")
	_, err := service.SendLike(ctx, "user3", "user4")

	if err == nil {
		t.Error("SendLike() error = nil, want error")
	}
}

func TestLikeService_SendLike_MatchSaveError(t *testing.T) {
	likeRepo := newMockLikeRepository()
	likeRepo.mutualCheck = true
	matchRepo := newMockMatchRepository()
	matchRepo.err = errors.New("match save error")
	service := NewLikeService(likeRepo, matchRepo)

	ctx := context.Background()
	_, err := service.SendLike(ctx, "user1", "user2")

	if err == nil {
		t.Error("SendLike() error = nil, want error")
	}
}

func TestLikeService_GetReceivedLikes_Success(t *testing.T) {
	likeRepo := newMockLikeRepository()
	matchRepo := newMockMatchRepository()
	service := NewLikeService(likeRepo, matchRepo)

	ctx := context.Background()
	likes, err := service.GetReceivedLikes(ctx, "user1")

	if err != nil {
		t.Fatalf("GetReceivedLikes() error = %v", err)
	}
	if likes == nil {
		t.Error("GetReceivedLikes() returned nil")
	}
}

func TestLikeService_GetReceivedLikes_Error(t *testing.T) {
	likeRepo := newMockLikeRepository()
	likeRepo.err = errors.New("find error")
	matchRepo := newMockMatchRepository()
	service := NewLikeService(likeRepo, matchRepo)

	ctx := context.Background()
	_, err := service.GetReceivedLikes(ctx, "user1")

	if err == nil {
		t.Error("GetReceivedLikes() error = nil, want error")
	}
}

func TestDatingMatchingService_FindMatches_UserNotFound(t *testing.T) {
	config := &matching.MatchingConfig{
		Version: "1.0",
		Domain:  "test",
		Scoring: matching.ScoringConfig{
			Components: []matching.ComponentConfig{
				{
					Name:   "test",
					Type:   "euclidean",
					Fields: []string{"age"},
					Weight: 1.0,
				},
			},
		},
		Ranking: matching.RankingConfig{
			SortOrder: "desc",
		},
	}

	engine, err := matching.NewConfigurableEngine(config)
	if err != nil {
		t.Fatalf("NewConfigurableEngine() error = %v", err)
	}

	userRepo := newMockUserRepository()
	profileRepo := newMockProfileRepository()
	featureMapper := mapper.NewDatingFeatureMapper()

	service := NewDatingMatchingService(engine, userRepo, profileRepo, nil, nil, featureMapper)

	ctx := context.Background()
	_, err = service.FindMatches(ctx, "nonexistent", 10)

	if err == nil {
		t.Error("FindMatches() error = nil, want error")
	}
}

func TestDatingMatchingService_FindMatches_Success(t *testing.T) {
	config := &matching.MatchingConfig{
		Version: "1.0",
		Domain:  "test",
		Scoring: matching.ScoringConfig{
			Components: []matching.ComponentConfig{
				{
					Name:   "test",
					Type:   "euclidean",
					Fields: []string{"age"},
					Weight: 1.0,
				},
			},
		},
		Ranking: matching.RankingConfig{
			SortOrder: "desc",
		},
	}

	engine, err := matching.NewConfigurableEngine(config)
	if err != nil {
		t.Fatalf("NewConfigurableEngine() error = %v", err)
	}

	userRepo := newMockUserRepository()
	birthDate := time.Now().AddDate(-25, 0, 0)
	userRepo.users["user1"] = &domain.User{
		ID:           "user1",
		Nickname:     "Test User",
		Gender:       "male",
		BirthDate:    birthDate,
		Prefecture:   "Tokyo",
		Verified:     true,
		EloRating:    1200,
		LastActiveAt: time.Now(),
		CreatedAt:    time.Now(),
	}

	profileRepo := newMockProfileRepository()
	profileRepo.profiles["user1"] = &domain.Profile{
		UserID:         "user1",
		Height:         175,
		BodyType:       "average",
		IncomeLevel:    700,
		Education:      "university",
		MarriageDesire: "yes_within_2years",
		Smoking:        "no",
		Drinking:       "sometimes",
		Tags: []domain.ProfileTag{
			{Tag: "sports"},
			{Tag: "travel"},
		},
	}

	featureMapper := mapper.NewDatingFeatureMapper()

	service := NewDatingMatchingService(engine, userRepo, profileRepo, nil, nil, featureMapper)

	ctx := context.Background()
	matches, err := service.FindMatches(ctx, "user1", 10)

	if err != nil {
		t.Fatalf("FindMatches() error = %v", err)
	}
	if matches == nil {
		t.Error("FindMatches() returned nil")
	}
}

func TestDatingMatchingService_FindMatches_ProfileNotFound(t *testing.T) {
	config := &matching.MatchingConfig{
		Version: "1.0",
		Domain:  "test",
		Scoring: matching.ScoringConfig{
			Components: []matching.ComponentConfig{
				{
					Name:   "test",
					Type:   "euclidean",
					Fields: []string{"age"},
					Weight: 1.0,
				},
			},
		},
		Ranking: matching.RankingConfig{
			SortOrder: "desc",
		},
	}

	engine, err := matching.NewConfigurableEngine(config)
	if err != nil {
		t.Fatalf("NewConfigurableEngine() error = %v", err)
	}

	userRepo := newMockUserRepository()
	birthDate := time.Now().AddDate(-25, 0, 0)
	userRepo.users["user1"] = &domain.User{
		ID:           "user1",
		Nickname:     "Test User",
		Gender:       "male",
		BirthDate:    birthDate,
		Prefecture:   "Tokyo",
		Verified:     true,
		EloRating:    1200,
		LastActiveAt: time.Now(),
		CreatedAt:    time.Now(),
	}

	profileRepo := newMockProfileRepository()
	// Profile intentionally not added

	featureMapper := mapper.NewDatingFeatureMapper()

	service := NewDatingMatchingService(engine, userRepo, profileRepo, nil, nil, featureMapper)

	ctx := context.Background()
	_, err = service.FindMatches(ctx, "user1", 10)

	if err == nil {
		t.Error("FindMatches() error = nil, want error")
	}
}

func TestDatingMatchingService_FindMatches_CandidatesError(t *testing.T) {
	config := &matching.MatchingConfig{
		Version: "1.0",
		Domain:  "test",
		Scoring: matching.ScoringConfig{
			Components: []matching.ComponentConfig{
				{
					Name:   "test",
					Type:   "euclidean",
					Fields: []string{"age"},
					Weight: 1.0,
				},
			},
		},
		Ranking: matching.RankingConfig{
			SortOrder: "desc",
		},
	}

	engine, err := matching.NewConfigurableEngine(config)
	if err != nil {
		t.Fatalf("NewConfigurableEngine() error = %v", err)
	}

	userRepo := newMockUserRepository()
	birthDate := time.Now().AddDate(-25, 0, 0)
	userRepo.users["user1"] = &domain.User{
		ID:           "user1",
		Nickname:     "Test User",
		Gender:       "male",
		BirthDate:    birthDate,
		Prefecture:   "Tokyo",
		Verified:     true,
		EloRating:    1200,
		LastActiveAt: time.Now(),
		CreatedAt:    time.Now(),
	}

	profileRepo := newMockProfileRepository()
	profileRepo.profiles["user1"] = &domain.Profile{
		UserID:         "user1",
		Height:         175,
		BodyType:       "average",
		IncomeLevel:    700,
		Education:      "university",
		MarriageDesire: "yes_within_2years",
		Smoking:        "no",
		Drinking:       "sometimes",
		Tags: []domain.ProfileTag{
			{Tag: "sports"},
			{Tag: "travel"},
		},
	}

	// Set error on FindCandidates
	userRepo.err = errors.New("find candidates error")

	featureMapper := mapper.NewDatingFeatureMapper()

	service := NewDatingMatchingService(engine, userRepo, profileRepo, nil, nil, featureMapper)

	ctx := context.Background()
	_, err = service.FindMatches(ctx, "user1", 10)

	if err == nil {
		t.Error("FindMatches() error = nil, want error")
	}
}

func TestDatingMatchingService_FindMatches_WithCandidates(t *testing.T) {
	config := &matching.MatchingConfig{
		Version: "1.0",
		Domain:  "test",
		Scoring: matching.ScoringConfig{
			Components: []matching.ComponentConfig{
				{
					Name:   "age",
					Type:   "euclidean",
					Fields: []string{"age"},
					Weight: 1.0,
				},
			},
		},
		Ranking: matching.RankingConfig{
			SortOrder: "desc",
		},
	}

	engine, err := matching.NewConfigurableEngine(config)
	if err != nil {
		t.Fatalf("NewConfigurableEngine() error = %v", err)
	}

	profileRepo := newMockProfileRepository()

	userRepo := newMockUserRepository()
	userRepo.profileRepo = profileRepo
	birthDate1 := time.Now().AddDate(-25, 0, 0)
	userRepo.users["user1"] = &domain.User{
		ID:           "user1",
		Nickname:     "User 1",
		Gender:       "male",
		BirthDate:    birthDate1,
		Prefecture:   "Tokyo",
		Verified:     true,
		EloRating:    1200,
		LastActiveAt: time.Now(),
		CreatedAt:    time.Now(),
	}

	birthDate2 := time.Now().AddDate(-26, 0, 0)
	userRepo.users["user2"] = &domain.User{
		ID:           "user2",
		Nickname:     "User 2",
		Gender:       "female",
		BirthDate:    birthDate2,
		Prefecture:   "Tokyo",
		Verified:     true,
		EloRating:    1300,
		LastActiveAt: time.Now(),
		CreatedAt:    time.Now(),
	}

	profileRepo.profiles["user1"] = &domain.Profile{
		UserID:         "user1",
		Height:         175,
		BodyType:       "average",
		IncomeLevel:    700,
		Education:      "university",
		MarriageDesire: "yes_within_2years",
		Smoking:        "no",
		Drinking:       "sometimes",
		Tags: []domain.ProfileTag{
			{Tag: "sports"},
		},
	}

	profileRepo.profiles["user2"] = &domain.Profile{
		UserID:         "user2",
		Height:         165,
		BodyType:       "slim",
		IncomeLevel:    700,
		Education:      "university",
		MarriageDesire: "yes_within_2years",
		Smoking:        "no",
		Drinking:       "sometimes",
		Tags: []domain.ProfileTag{
			{Tag: "travel"},
		},
	}

	featureMapper := mapper.NewDatingFeatureMapper()

	service := NewDatingMatchingService(engine, userRepo, profileRepo, nil, nil, featureMapper)

	ctx := context.Background()
	matches, err := service.FindMatches(ctx, "user1", 5)

	if err != nil {
		t.Fatalf("FindMatches() error = %v", err)
	}
	if matches == nil {
		t.Error("FindMatches() returned nil")
	}
	if len(matches) > 5 {
		t.Errorf("FindMatches() returned %v matches, want <= 5", len(matches))
	}
}

func TestDatingMatchingService_GetMutualMatches_Success(t *testing.T) {
	// マッチリポジトリをセットアップ
	matchRepo := newMockMatchRepository()
	matchRepo.matches = []*domain.Match{
		{
			ID:      "match-1",
			UserIDA: "user1",
			UserIDB: "user2",
			Score:   0.85,
			Breakdown: map[string]float64{
				"age":  0.9,
				"tags": 0.8,
			},
			MatchedAt: time.Now(),
		},
		{
			ID:      "match-2",
			UserIDA: "user1",
			UserIDB: "user3",
			Score:   0.75,
			Breakdown: map[string]float64{
				"age":  0.8,
				"tags": 0.7,
			},
			MatchedAt: time.Now().Add(-24 * time.Hour),
		},
	}

	service := NewDatingMatchingService(nil, nil, nil, nil, matchRepo, nil)

	ctx := context.Background()
	results, err := service.GetMutualMatches(ctx, "user1")

	if err != nil {
		t.Fatalf("GetMutualMatches() error = %v", err)
	}
	if results == nil {
		t.Fatal("GetMutualMatches() returned nil")
	}
	if len(results) != 2 {
		t.Errorf("GetMutualMatches() returned %d results, want 2", len(results))
	}

	// 最初の結果を確認
	if results[0].MatchID != "match-1" {
		t.Errorf("results[0].MatchID = %v, want match-1", results[0].MatchID)
	}
	if results[0].UserIDA != "user1" {
		t.Errorf("results[0].UserIDA = %v, want user1", results[0].UserIDA)
	}
	if results[0].UserIDB != "user2" {
		t.Errorf("results[0].UserIDB = %v, want user2", results[0].UserIDB)
	}
	if results[0].Score != 0.85 {
		t.Errorf("results[0].Score = %v, want 0.85", results[0].Score)
	}
	if results[0].Breakdown["age"] != 0.9 {
		t.Errorf("results[0].Breakdown[age] = %v, want 0.9", results[0].Breakdown["age"])
	}
}

func TestDatingMatchingService_GetMutualMatches_Error(t *testing.T) {
	// エラーを返すマッチリポジトリをセットアップ
	matchRepo := newMockMatchRepository()
	matchRepo.err = errors.New("database error")

	service := NewDatingMatchingService(nil, nil, nil, nil, matchRepo, nil)

	ctx := context.Background()
	results, err := service.GetMutualMatches(ctx, "user1")

	if err == nil {
		t.Fatal("GetMutualMatches() expected error, got nil")
	}
	if results != nil {
		t.Errorf("GetMutualMatches() returned results on error, want nil")
	}
}

func TestDatingMatchingService_GetMutualMatches_Empty(t *testing.T) {
	// 空のマッチリポジトリをセットアップ
	matchRepo := newMockMatchRepository()

	service := NewDatingMatchingService(nil, nil, nil, nil, matchRepo, nil)

	ctx := context.Background()
	results, err := service.GetMutualMatches(ctx, "user1")

	if err != nil {
		t.Fatalf("GetMutualMatches() error = %v", err)
	}
	if results == nil {
		t.Fatal("GetMutualMatches() returned nil, want empty slice")
	}
	if len(results) != 0 {
		t.Errorf("GetMutualMatches() returned %d results, want 0", len(results))
	}
}
