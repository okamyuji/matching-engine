//go:build integration
// +build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/matching-engine/internal/modules/dating/domain"
	"github.com/yourorg/matching-engine/internal/testutil"
)

func TestUserRepository_Integration(t *testing.T) {
	testDB := testutil.GetSharedTestDB(t)
	testDB.CleanTables(t)

	repo := NewUserRepository(testDB.DB)
	ctx := context.Background()

	t.Run("Create と FindByID - 正常系", func(t *testing.T) {
		user := &domain.User{
			ID:           "test_user_1",
			Nickname:     "TestUser",
			Gender:       "male",
			BirthDate:    time.Now().AddDate(-25, 0, 0),
			Prefecture:   "Tokyo",
			Verified:     true,
			EloRating:    1200,
			LastActiveAt: time.Now(),
			CreatedAt:    time.Now(),
		}

		// 作成
		err := repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// 取得
		found, err := repo.FindByID(ctx, "test_user_1")
		if err != nil {
			t.Fatalf("FindByID() error = %v", err)
		}

		if found.ID != user.ID {
			t.Errorf("ID = %v, want %v", found.ID, user.ID)
		}
		if found.Nickname != user.Nickname {
			t.Errorf("Nickname = %v, want %v", found.Nickname, user.Nickname)
		}
	})

	t.Run("FindByID - 存在しないユーザー（異常系）", func(t *testing.T) {
		_, err := repo.FindByID(ctx, "nonexistent")
		if err == nil {
			t.Error("FindByID() error = nil, want error")
		}
	})

	t.Run("Update - 正常系", func(t *testing.T) {
		user := &domain.User{
			ID:           "test_user_2",
			Nickname:     "OldName",
			Gender:       "female",
			BirthDate:    time.Now().AddDate(-30, 0, 0),
			Prefecture:   "Osaka",
			Verified:     false,
			EloRating:    1000,
			LastActiveAt: time.Now(),
			CreatedAt:    time.Now(),
		}

		err := repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// 更新
		user.Nickname = "NewName"
		user.Verified = true
		user.EloRating = 1500

		err = repo.Update(ctx, user)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		// 確認
		updated, err := repo.FindByID(ctx, "test_user_2")
		if err != nil {
			t.Fatalf("FindByID() error = %v", err)
		}

		if updated.Nickname != "NewName" {
			t.Errorf("Nickname = %v, want NewName", updated.Nickname)
		}
		if !updated.Verified {
			t.Error("Verified should be true")
		}
		if updated.EloRating != 1500 {
			t.Errorf("EloRating = %v, want 1500", updated.EloRating)
		}
	})

	t.Run("FindCandidates - 正常系", func(t *testing.T) {
		testDB.SeedTestData(t)

		pref := &domain.Preference{
			UserID: "user1",
			AgeMin: 20,
			AgeMax: 35,
			Prefectures: []domain.PreferencePrefecture{
				{Prefecture: domain.Prefecture("Tokyo")},
			},
		}

		candidates, err := repo.FindCandidates(ctx, "user1", pref)
		if err != nil {
			t.Fatalf("FindCandidates() error = %v", err)
		}

		// user1以外のTokyoのユーザーが含まれるべき
		if len(candidates) == 0 {
			t.Error("FindCandidates() returned no candidates")
		}

		// 自分自身は含まれないこと確認
		for _, c := range candidates {
			if c.User.ID == "user1" {
				t.Error("FindCandidates() included self")
			}
		}
	})

	t.Run("FindCandidates - 空の結果（境界値）", func(t *testing.T) {
		pref := &domain.Preference{
			UserID: "user1",
			AgeMin: 99,
			AgeMax: 100,
			Prefectures: []domain.PreferencePrefecture{
				{Prefecture: domain.Prefecture("Tokyo")},
			},
		}

		candidates, err := repo.FindCandidates(ctx, "user1", pref)
		if err != nil {
			t.Fatalf("FindCandidates() error = %v", err)
		}

		if len(candidates) != 0 {
			t.Errorf("FindCandidates() returned %v candidates, want 0", len(candidates))
		}
	})
}

func TestProfileRepository_Integration(t *testing.T) {
	testDB := testutil.GetSharedTestDB(t)
	testDB.CleanTables(t)
	testDB.SeedTestData(t)

	repo := NewProfileRepository(testDB.DB)
	ctx := context.Background()

	t.Run("FindByUserID - 正常系", func(t *testing.T) {
		profile, err := repo.FindByUserID(ctx, "user1")
		if err != nil {
			t.Fatalf("FindByUserID() error = %v", err)
		}

		if profile.UserID != "user1" {
			t.Errorf("UserID = %v, want user1", profile.UserID)
		}
		if profile.Height <= 0 {
			t.Error("Height should be positive")
		}
	})

	t.Run("FindByUserID - 存在しないプロフィール（異常系）", func(t *testing.T) {
		_, err := repo.FindByUserID(ctx, "nonexistent")
		if err == nil {
			t.Error("FindByUserID() error = nil, want error")
		}
	})

	t.Run("Upsert - 新規作成（正常系）", func(t *testing.T) {
		profile := &domain.Profile{
			UserID:         "user3",
			Height:         180,
			BodyType:       domain.BodyTypeAthletic,
			IncomeLevel:    900,
			Education:      domain.EducationGraduate,
			MarriageDesire: domain.MarriageUndecided,
			ChildrenDesire: domain.ChildrenWant,
			Smoking:        domain.SmokingNonSmoker,
			Drinking:       domain.DrinkingRegular,
			Tags: []domain.ProfileTag{
				{UserID: "user3", Tag: "sports"},
				{UserID: "user3", Tag: "cooking"},
			},
		}

		err := repo.Upsert(ctx, profile)
		if err != nil {
			t.Fatalf("Upsert() error = %v", err)
		}

		// 確認
		found, err := repo.FindByUserID(ctx, "user3")
		if err != nil {
			t.Fatalf("FindByUserID() error = %v", err)
		}

		if found.Height != 180 {
			t.Errorf("Height = %v, want 180", found.Height)
		}
	})

	t.Run("Upsert - 更新（正常系）", func(t *testing.T) {
		// user1のプロフィールを更新
		profile := &domain.Profile{
			UserID:         "user1",
			Height:         185,
			BodyType:       domain.BodyTypeAthletic,
			IncomeLevel:    1250,
			Education:      domain.EducationGraduate,
			MarriageDesire: domain.MarriageWantSoon,
			ChildrenDesire: domain.ChildrenWant,
			Smoking:        domain.SmokingNonSmoker,
			Drinking:       domain.DrinkingNonDrinker,
			Tags: []domain.ProfileTag{
				{UserID: "user1", Tag: "technology"},
			},
		}

		err := repo.Upsert(ctx, profile)
		if err != nil {
			t.Fatalf("Upsert() error = %v", err)
		}

		// 確認
		updated, err := repo.FindByUserID(ctx, "user1")
		if err != nil {
			t.Fatalf("FindByUserID() error = %v", err)
		}

		if updated.Height != 185 {
			t.Errorf("Height = %v, want 185", updated.Height)
		}
	})
}

func TestLikeRepository_Integration(t *testing.T) {
	testDB := testutil.GetSharedTestDB(t)
	testDB.CleanTables(t)
	testDB.SeedTestData(t)

	repo := NewLikeRepository(testDB.DB)
	ctx := context.Background()

	t.Run("Save と FindByTargetUserID - 正常系", func(t *testing.T) {
		like := &Like{
			ID:         "like_1",
			FromUserID: "user1",
			ToUserID:   "user2",
			CreatedAt:  time.Now(),
		}

		err := repo.Save(ctx, like)
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// 取得
		likes, err := repo.FindByTargetUserID(ctx, "user2")
		if err != nil {
			t.Fatalf("FindByTargetUserID() error = %v", err)
		}

		if len(likes) == 0 {
			t.Error("FindByTargetUserID() returned no likes")
		}

		found := false
		for _, l := range likes {
			if l.ID == "like_1" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected like not found")
		}
	})

	t.Run("Save - 重複いいね（異常系）", func(t *testing.T) {
		like1 := &Like{
			ID:         "like_2",
			FromUserID: "user1",
			ToUserID:   "user3",
			CreatedAt:  time.Now(),
		}

		err := repo.Save(ctx, like1)
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// 同じ組み合わせで再度保存（エラーになるべき）
		like2 := &Like{
			ID:         "like_3",
			FromUserID: "user1",
			ToUserID:   "user3",
			CreatedAt:  time.Now(),
		}

		err = repo.Save(ctx, like2)
		if err == nil {
			t.Error("Save() should fail for duplicate like")
		}
	})

	t.Run("CheckMutual - 相互いいねあり（正常系）", func(t *testing.T) {
		// user1 -> user4
		like1 := &Like{
			ID:         "like_4",
			FromUserID: "user1",
			ToUserID:   "user4",
			CreatedAt:  time.Now(),
		}
		repo.Save(ctx, like1)

		// user4 -> user1
		like2 := &Like{
			ID:         "like_5",
			FromUserID: "user4",
			ToUserID:   "user1",
			CreatedAt:  time.Now(),
		}
		repo.Save(ctx, like2)

		// 相互いいねチェック
		isMutual, err := repo.CheckMutual(ctx, "user1", "user4")
		if err != nil {
			t.Fatalf("CheckMutual() error = %v", err)
		}

		if !isMutual {
			t.Error("CheckMutual() = false, want true")
		}
	})

	t.Run("CheckMutual - 相互いいねなし（正常系）", func(t *testing.T) {
		// user1 -> user5 のみ
		like := &Like{
			ID:         "like_6",
			FromUserID: "user1",
			ToUserID:   "user5",
			CreatedAt:  time.Now(),
		}
		repo.Save(ctx, like)

		isMutual, err := repo.CheckMutual(ctx, "user1", "user5")
		if err != nil {
			t.Fatalf("CheckMutual() error = %v", err)
		}

		if isMutual {
			t.Error("CheckMutual() = true, want false")
		}
	})

	t.Run("FindByFromUserID - 正常系", func(t *testing.T) {
		likes, err := repo.FindByFromUserID(ctx, "user1")
		if err != nil {
			t.Fatalf("FindByFromUserID() error = %v", err)
		}

		if len(likes) == 0 {
			t.Error("FindByFromUserID() returned no likes")
		}

		// すべてのいいねがuser1から送られていることを確認
		for _, like := range likes {
			if like.FromUserID != "user1" {
				t.Errorf("FromUserID = %v, want user1", like.FromUserID)
			}
		}
	})

	t.Run("FindByTargetUserID - 空の結果（境界値）", func(t *testing.T) {
		likes, err := repo.FindByTargetUserID(ctx, "nonexistent_user")
		if err != nil {
			t.Fatalf("FindByTargetUserID() error = %v", err)
		}

		if len(likes) != 0 {
			t.Errorf("FindByTargetUserID() returned %v likes, want 0", len(likes))
		}
	})
}

func TestMatchRepository_Integration(t *testing.T) {
	testDB := testutil.GetSharedTestDB(t)
	testDB.CleanTables(t)
	testDB.SeedTestData(t)

	repo := NewMatchRepository(testDB.DB)
	ctx := context.Background()

	t.Run("Save と FindByUserID - 正常系", func(t *testing.T) {
		match := &domain.Match{
			ID:        "match_1",
			UserIDA:   "user1",
			UserIDB:   "user2",
			Score:     0.85,
			Breakdown: map[string]float64{"age": 0.9, "location": 0.8},
			MatchedAt: time.Now(),
		}

		err := repo.Save(ctx, match)
		if err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// user1のマッチを取得
		matches, err := repo.FindByUserID(ctx, "user1")
		if err != nil {
			t.Fatalf("FindByUserID() error = %v", err)
		}

		if len(matches) == 0 {
			t.Error("FindByUserID() returned no matches")
		}

		found := false
		for _, m := range matches {
			if m.ID == "match_1" {
				found = true
				if m.Score != 0.85 {
					t.Errorf("Score = %v, want 0.85", m.Score)
				}
				break
			}
		}
		if !found {
			t.Error("Expected match not found")
		}
	})

	t.Run("FindByUserID - 両方向検索（正常系）", func(t *testing.T) {
		// user2のマッチも取得できるべき
		matches, err := repo.FindByUserID(ctx, "user2")
		if err != nil {
			t.Fatalf("FindByUserID() error = %v", err)
		}

		found := false
		for _, m := range matches {
			if m.ID == "match_1" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Match should be found from both users")
		}
	})

	t.Run("FindMutual - 正常系", func(t *testing.T) {
		// FindMutualは相互いいねが存在するマッチのみを返す
		// テストのために両方向のいいねを作成
		likeRepo := NewLikeRepository(testDB.DB)

		like1 := &Like{
			ID:         "like_user1_user2",
			FromUserID: "user1",
			ToUserID:   "user2",
			CreatedAt:  time.Now(),
		}
		like2 := &Like{
			ID:         "like_user2_user1",
			FromUserID: "user2",
			ToUserID:   "user1",
			CreatedAt:  time.Now(),
		}

		if err := likeRepo.Save(ctx, like1); err != nil {
			t.Fatalf("Save like1 error = %v", err)
		}
		if err := likeRepo.Save(ctx, like2); err != nil {
			t.Fatalf("Save like2 error = %v", err)
		}

		matches, err := repo.FindMutual(ctx, "user1")
		if err != nil {
			t.Fatalf("FindMutual() error = %v", err)
		}

		if len(matches) == 0 {
			t.Error("FindMutual() returned no matches")
		}
	})

	t.Run("Save - 重複マッチ（異常系）", func(t *testing.T) {
		match := &domain.Match{
			ID:        "match_2",
			UserIDA:   "user1",
			UserIDB:   "user2",
			Score:     0.90,
			Breakdown: map[string]float64{},
			MatchedAt: time.Now(),
		}

		err := repo.Save(ctx, match)
		if err == nil {
			t.Error("Save() should fail for duplicate match")
		}
	})

	t.Run("FindByUserID - 空の結果（境界値）", func(t *testing.T) {
		matches, err := repo.FindByUserID(ctx, "nonexistent_user")
		if err != nil {
			t.Fatalf("FindByUserID() error = %v", err)
		}

		if len(matches) != 0 {
			t.Errorf("FindByUserID() returned %v matches, want 0", len(matches))
		}
	})
}
