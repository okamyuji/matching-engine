package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/okamyuji/matching-engine/internal/modules/dating/domain"
	"github.com/okamyuji/matching-engine/internal/testutil"
)

func TestUserRepository_FindByID_NormalCase(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewUserRepository(td.Pool)
	ctx := context.Background()

	// テストデータ作成
	user := &domain.User{
		ID:           "user001",
		Nickname:     "user001",
		Gender:       domain.GenderMale,
		BirthDate:    time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureTokyo,
		Verified:     true,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}

	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("テストデータ作成失敗: %v", err)
	}

	// FindByIDテスト
	found, err := repo.FindByID(ctx, "user001")
	if err != nil {
		t.Fatalf("FindByID失敗: %v", err)
	}

	if found.ID != user.ID {
		t.Errorf("ID = %v, want %v", found.ID, user.ID)
	}
	if found.Nickname != user.Nickname {
		t.Errorf("Nickname = %v, want %v", found.Nickname, user.Nickname)
	}
	if found.Gender != user.Gender {
		t.Errorf("Gender = %v, want %v", found.Gender, user.Gender)
	}
}

func TestUserRepository_FindByID_UserNotFound(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewUserRepository(td.Pool)
	ctx := context.Background()

	// 存在しないユーザーIDで検索
	_, err := repo.FindByID(ctx, "nonexistent")
	if err == nil {
		t.Error("存在しないユーザーでエラーが発生すべき")
	}
}

func TestUserRepository_FindByID_EmptyID(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewUserRepository(td.Pool)
	ctx := context.Background()

	// 空のIDで検索
	_, err := repo.FindByID(ctx, "")
	if err == nil {
		t.Error("空のIDでエラーが発生すべき")
	}
}

func TestUserRepository_Create_NormalCase(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewUserRepository(td.Pool)
	ctx := context.Background()

	user := &domain.User{
		ID:           "user002",
		Nickname:     "user002",
		Gender:       domain.GenderFemale,
		BirthDate:    time.Date(1995, 3, 20, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureOsaka,
		Verified:     false,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}

	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Create失敗: %v", err)
	}

	// 作成したユーザーを取得して確認
	found, err := repo.FindByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("FindByID失敗: %v", err)
	}

	if found.ID != user.ID {
		t.Errorf("ID = %v, want %v", found.ID, user.ID)
	}
	if found.Nickname != user.Nickname {
		t.Errorf("Nickname = %v, want %v", found.Nickname, user.Nickname)
	}
}

func TestUserRepository_Create_DuplicateID(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewUserRepository(td.Pool)
	ctx := context.Background()

	user := &domain.User{
		ID:           "user003",
		Nickname:     "user003",
		Gender:       domain.GenderMale,
		BirthDate:    time.Date(1992, 7, 10, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureTokyo,
		Verified:     true,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}

	// 1回目の作成
	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("1回目のCreate失敗: %v", err)
	}

	// 2回目の作成（重複）
	duplicate := &domain.User{
		ID:           "user003", // 同じID
		Nickname:     "different",
		Gender:       domain.GenderFemale,
		BirthDate:    time.Date(1993, 8, 11, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureOsaka,
		Verified:     false,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}

	err = repo.Create(ctx, duplicate)
	if err == nil {
		t.Error("重複IDでエラーが発生すべき")
	}
}

func TestUserRepository_Create_MinimalFields(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewUserRepository(td.Pool)
	ctx := context.Background()

	// 最小限の必須フィールドのみ
	user := &domain.User{
		ID:           "user004",
		Nickname:     "user004",
		Gender:       domain.GenderMale,
		BirthDate:    time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureTokyo,
		Verified:     false,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}

	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("最小限フィールドでのCreate失敗: %v", err)
	}

	found, err := repo.FindByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("FindByID失敗: %v", err)
	}

	if found.ID != user.ID {
		t.Errorf("ID = %v, want %v", found.ID, user.ID)
	}
}

func TestUserRepository_Update_NormalCase(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewUserRepository(td.Pool)
	ctx := context.Background()

	// テストデータ作成
	user := &domain.User{
		ID:           "user005",
		Nickname:     "user005",
		Gender:       domain.GenderMale,
		BirthDate:    time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureTokyo,
		Verified:     false,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}

	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("テストデータ作成失敗: %v", err)
	}

	// 更新
	user.Verified = true
	user.EloRating = 1600
	user.LastActiveAt = time.Now()

	err = repo.Update(ctx, user)
	if err != nil {
		t.Fatalf("Update失敗: %v", err)
	}

	// 更新を確認
	found, err := repo.FindByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("FindByID失敗: %v", err)
	}

	if !found.Verified {
		t.Error("Verifiedが更新されていない")
	}
	if found.EloRating != 1600 {
		t.Errorf("EloRating = %v, want 1600", found.EloRating)
	}
}

func TestUserRepository_Update_NonExistentUser(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewUserRepository(td.Pool)
	ctx := context.Background()

	// 存在しないユーザーを更新
	user := &domain.User{
		ID:           "nonexistent",
		Nickname:     "nonexistent",
		Gender:       domain.GenderMale,
		BirthDate:    time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureTokyo,
		Verified:     true,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}

	err := repo.Update(ctx, user)
	// 対象行が無い UPDATE はエラーにならない（更新件数0で正常終了する）
	if err != nil {
		t.Logf("Update実行: %v", err)
	}

	// 存在しないことを確認
	_, err = repo.FindByID(ctx, user.ID)
	if err == nil {
		t.Error("存在しないユーザーが見つかってしまった")
	}
}

func TestUserRepository_Update_PartialUpdate(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewUserRepository(td.Pool)
	ctx := context.Background()

	// テストデータ作成
	originalNickname := "user006"
	user := &domain.User{
		ID:           "user006",
		Nickname:     originalNickname,
		Gender:       domain.GenderFemale,
		BirthDate:    time.Date(1995, 3, 20, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureOsaka,
		Verified:     false,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}

	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("テストデータ作成失敗: %v", err)
	}

	// EloRatingのみ更新
	user.EloRating = 1700
	user.LastActiveAt = time.Now()

	err = repo.Update(ctx, user)
	if err != nil {
		t.Fatalf("Update失敗: %v", err)
	}

	// 更新を確認
	found, err := repo.FindByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("FindByID失敗: %v", err)
	}

	if found.EloRating != 1700 {
		t.Errorf("EloRating = %v, want 1700", found.EloRating)
	}
	if found.Nickname != originalNickname {
		t.Errorf("Nickname = %v, want %v (変更されるべきでない)", found.Nickname, originalNickname)
	}
}

func TestUserRepository_FindCandidates_NormalCase(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewUserRepository(td.Pool)
	ctx := context.Background()

	// 検索ユーザー作成
	searchUser := &domain.User{
		ID:           "searcher",
		Nickname:     "searcher",
		Gender:       domain.GenderMale,
		BirthDate:    time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureTokyo,
		Verified:     true,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}
	err := repo.Create(ctx, searchUser)
	if err != nil {
		t.Fatalf("検索ユーザー作成失敗: %v", err)
	}

	// 候補ユーザー作成
	candidate := &domain.User{
		ID:           "candidate001",
		Nickname:     "candidate001",
		Gender:       domain.GenderFemale,
		BirthDate:    time.Date(1992, 6, 15, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureTokyo,
		Verified:     true,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}
	err = repo.Create(ctx, candidate)
	if err != nil {
		t.Fatalf("候補ユーザー作成失敗: %v", err)
	}

	// プロフィール作成
	profile := &domain.Profile{
		UserID:           "candidate001",
		Height:           165,
		BodyType:         domain.BodyTypeAverage,
		Education:        domain.EducationUniversity,
		Occupation:       "エンジニア",
		IncomeLevel:      500,
		MarriageDesire:   domain.MarriageWantSoon,
		ChildrenDesire:   domain.ChildrenWant,
		Smoking:          domain.SmokingNonSmoker,
		Drinking:         domain.DrinkingSocial,
		SelfIntroduction: "テストプロフィール",
		UpdatedAt:        time.Now(),
	}
	err = NewProfileRepository(td.Pool).Upsert(ctx, profile)
	if err != nil {
		t.Fatalf("プロフィール作成失敗: %v", err)
	}

	// 検索条件作成
	pref := &domain.Preference{
		UserID:    "searcher",
		AgeMin:    25,
		AgeMax:    40,
		HeightMin: 160,
		HeightMax: 170,
		IncomeMin: 400,
		Prefectures: []domain.PreferencePrefecture{
			{UserID: "searcher", Prefecture: domain.PrefectureTokyo},
		},
	}

	// FindCandidatesテスト
	results, err := repo.FindCandidates(ctx, "searcher", pref)
	if err != nil {
		t.Fatalf("FindCandidates失敗: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("候補数 = %v, want 1", len(results))
	}

	if results[0].User.ID != "candidate001" {
		t.Errorf("候補ユーザーID = %v, want candidate001", results[0].User.ID)
	}

	if results[0].Profile == nil {
		t.Error("プロフィールがnilです")
	} else if results[0].Profile.Height != 165 {
		t.Errorf("プロフィール身長 = %v, want 165", results[0].Profile.Height)
	}
}

func TestUserRepository_FindCandidates_EmptyPreferences(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewUserRepository(td.Pool)
	ctx := context.Background()

	// 検索ユーザー作成
	searchUser := &domain.User{
		ID:           "searcher2",
		Nickname:     "searcher2",
		Gender:       domain.GenderMale,
		BirthDate:    time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureTokyo,
		Verified:     true,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}
	err := repo.Create(ctx, searchUser)
	if err != nil {
		t.Fatalf("検索ユーザー作成失敗: %v", err)
	}

	// 候補ユーザー作成
	for i := 0; i < 3; i++ {
		candidateID := fmt.Sprintf("candidate_empty_%d", i)
		candidate := &domain.User{
			ID:           candidateID,
			Nickname:     candidateID,
			Gender:       domain.GenderFemale,
			BirthDate:    time.Date(1992, 1, 1, 0, 0, 0, 0, time.UTC),
			Prefecture:   domain.PrefectureTokyo,
			Verified:     true,
			EloRating:    1500,
			CreatedAt:    time.Now(),
			LastActiveAt: time.Now(),
		}
		err := repo.Create(ctx, candidate)
		if err != nil {
			t.Fatalf("候補ユーザー%d作成失敗: %v", i, err)
		}

		// プロフィール作成
		profile := &domain.Profile{
			UserID:           candidate.ID,
			Height:           160,
			BodyType:         domain.BodyTypeAverage,
			Education:        domain.EducationUniversity,
			Occupation:       "営業",
			IncomeLevel:      400,
			MarriageDesire:   domain.MarriageWantSoon,
			ChildrenDesire:   domain.ChildrenWant,
			Smoking:          domain.SmokingNonSmoker,
			Drinking:         domain.DrinkingSocial,
			SelfIntroduction: "テスト",
			UpdatedAt:        time.Now(),
		}
		err = NewProfileRepository(td.Pool).Upsert(ctx, profile)
		if err != nil {
			t.Fatalf("プロフィール%d作成失敗: %v", i, err)
		}
	}

	// 空の検索条件
	pref := &domain.Preference{
		UserID: "searcher2",
	}

	// FindCandidatesテスト
	results, err := repo.FindCandidates(ctx, "searcher2", pref)
	if err != nil {
		t.Fatalf("FindCandidates失敗: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("候補数 = %v, want 3", len(results))
	}
}

func TestUserRepository_FindCandidates_AgeFilter(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewUserRepository(td.Pool)
	ctx := context.Background()

	// 検索ユーザー作成
	searchUser := &domain.User{
		ID:           "searcher3",
		Nickname:     "searcher3",
		Gender:       domain.GenderMale,
		BirthDate:    time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureTokyo,
		Verified:     true,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}
	err := repo.Create(ctx, searchUser)
	if err != nil {
		t.Fatalf("検索ユーザー作成失敗: %v", err)
	}

	// 年齢が異なる候補ユーザー作成（time.Now()基準で年齢を算出）
	now := time.Now()
	testCases := []struct {
		id          string
		birthDate   time.Time
		shouldMatch bool
	}{
		{"young", now.AddDate(-20, 0, 0), false}, // 20歳 (範囲外)
		{"match1", now.AddDate(-25, 0, 0), true}, // 25歳
		{"match2", now.AddDate(-30, 0, 0), true}, // 30歳
		{"match3", now.AddDate(-35, 0, 0), true}, // 35歳
		{"old", now.AddDate(-45, 0, 0), false},   // 45歳 (範囲外)
	}

	for _, tc := range testCases {
		candidate := &domain.User{
			ID:           tc.id,
			Nickname:     tc.id,
			Gender:       domain.GenderFemale,
			BirthDate:    tc.birthDate,
			Prefecture:   domain.PrefectureTokyo,
			Verified:     true,
			EloRating:    1500,
			CreatedAt:    time.Now(),
			LastActiveAt: time.Now(),
		}
		err := repo.Create(ctx, candidate)
		if err != nil {
			t.Fatalf("候補ユーザー%s作成失敗: %v", tc.id, err)
		}

		// プロフィール作成
		profile := &domain.Profile{
			UserID:           tc.id,
			Height:           160,
			BodyType:         domain.BodyTypeAverage,
			Education:        domain.EducationUniversity,
			Occupation:       "営業",
			IncomeLevel:      400,
			MarriageDesire:   domain.MarriageWantSoon,
			ChildrenDesire:   domain.ChildrenWant,
			Smoking:          domain.SmokingNonSmoker,
			Drinking:         domain.DrinkingSocial,
			SelfIntroduction: "テスト",
			UpdatedAt:        time.Now(),
		}
		err = NewProfileRepository(td.Pool).Upsert(ctx, profile)
		if err != nil {
			t.Fatalf("プロフィール%s作成失敗: %v", tc.id, err)
		}
	}

	// 年齢フィルタ: 25-35歳
	pref := &domain.Preference{
		UserID: "searcher3",
		AgeMin: 25,
		AgeMax: 35,
	}

	results, err := repo.FindCandidates(ctx, "searcher3", pref)
	if err != nil {
		t.Fatalf("FindCandidates失敗: %v", err)
	}

	// マッチすべきユーザーのみが返されることを確認
	expectedCount := 0
	for _, tc := range testCases {
		if tc.shouldMatch {
			expectedCount++
		}
	}

	if len(results) != expectedCount {
		t.Errorf("候補数 = %v, want %v", len(results), expectedCount)
	}

	// マッチしたユーザーIDを確認
	matchedIDs := make(map[string]bool)
	for _, result := range results {
		matchedIDs[result.User.ID] = true
	}

	for _, tc := range testCases {
		if tc.shouldMatch && !matchedIDs[tc.id] {
			t.Errorf("ユーザー%sがマッチすべきだがマッチしなかった", tc.id)
		}
		if !tc.shouldMatch && matchedIDs[tc.id] {
			t.Errorf("ユーザー%sがマッチすべきでないのにマッチした", tc.id)
		}
	}
}

func TestUserRepository_FindCandidates_ExcludeSelf(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewUserRepository(td.Pool)
	ctx := context.Background()

	// 検索ユーザー作成
	searchUser := &domain.User{
		ID:           "searcher4",
		Nickname:     "searcher4",
		Gender:       domain.GenderMale,
		BirthDate:    time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureTokyo,
		Verified:     true,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}
	err := repo.Create(ctx, searchUser)
	if err != nil {
		t.Fatalf("検索ユーザー作成失敗: %v", err)
	}

	// プロフィール作成
	profile := &domain.Profile{
		UserID:           "searcher4",
		Height:           170,
		BodyType:         domain.BodyTypeAverage,
		Education:        domain.EducationUniversity,
		Occupation:       "営業",
		IncomeLevel:      500,
		MarriageDesire:   domain.MarriageWantSoon,
		ChildrenDesire:   domain.ChildrenWant,
		Smoking:          domain.SmokingNonSmoker,
		Drinking:         domain.DrinkingSocial,
		SelfIntroduction: "テスト",
		UpdatedAt:        time.Now(),
	}
	err = NewProfileRepository(td.Pool).Upsert(ctx, profile)
	if err != nil {
		t.Fatalf("プロフィール作成失敗: %v", err)
	}

	// 空の検索条件
	pref := &domain.Preference{
		UserID: "searcher4",
	}

	// FindCandidatesテスト
	results, err := repo.FindCandidates(ctx, "searcher4", pref)
	if err != nil {
		t.Fatalf("FindCandidates失敗: %v", err)
	}

	// 自分自身は除外される
	for _, result := range results {
		if result.User.ID == "searcher4" {
			t.Error("自分自身が候補に含まれている")
		}
	}
}

func TestUserRepository_FindCandidates_VerifiedOnly(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewUserRepository(td.Pool)
	ctx := context.Background()

	// 検索ユーザー作成
	searchUser := &domain.User{
		ID:           "searcher5",
		Nickname:     "searcher5",
		Gender:       domain.GenderMale,
		BirthDate:    time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureTokyo,
		Verified:     true,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}
	err := repo.Create(ctx, searchUser)
	if err != nil {
		t.Fatalf("検索ユーザー作成失敗: %v", err)
	}

	// 認証済みユーザー作成
	verifiedUser := &domain.User{
		ID:           "verified",
		Nickname:     "verified",
		Gender:       domain.GenderFemale,
		BirthDate:    time.Date(1992, 1, 1, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureTokyo,
		Verified:     true,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}
	err = repo.Create(ctx, verifiedUser)
	if err != nil {
		t.Fatalf("認証済みユーザー作成失敗: %v", err)
	}

	verifiedProfile := &domain.Profile{
		UserID:           "verified",
		Height:           165,
		BodyType:         domain.BodyTypeAverage,
		Education:        domain.EducationUniversity,
		Occupation:       "営業",
		IncomeLevel:      500,
		MarriageDesire:   domain.MarriageWantSoon,
		ChildrenDesire:   domain.ChildrenWant,
		Smoking:          domain.SmokingNonSmoker,
		Drinking:         domain.DrinkingSocial,
		SelfIntroduction: "テスト",
		UpdatedAt:        time.Now(),
	}
	err = NewProfileRepository(td.Pool).Upsert(ctx, verifiedProfile)
	if err != nil {
		t.Fatalf("認証済みユーザープロフィール作成失敗: %v", err)
	}

	// 未認証ユーザー作成
	unverifiedUser := &domain.User{
		ID:           "unverified",
		Nickname:     "unverified",
		Gender:       domain.GenderFemale,
		BirthDate:    time.Date(1993, 1, 1, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureTokyo,
		Verified:     false,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}
	err = repo.Create(ctx, unverifiedUser)
	if err != nil {
		t.Fatalf("未認証ユーザー作成失敗: %v", err)
	}

	unverifiedProfile := &domain.Profile{
		UserID:           "unverified",
		Height:           165,
		BodyType:         domain.BodyTypeAverage,
		Education:        domain.EducationUniversity,
		Occupation:       "営業",
		IncomeLevel:      500,
		MarriageDesire:   domain.MarriageWantSoon,
		ChildrenDesire:   domain.ChildrenWant,
		Smoking:          domain.SmokingNonSmoker,
		Drinking:         domain.DrinkingSocial,
		SelfIntroduction: "テスト",
		UpdatedAt:        time.Now(),
	}
	err = NewProfileRepository(td.Pool).Upsert(ctx, unverifiedProfile)
	if err != nil {
		t.Fatalf("未認証ユーザープロフィール作成失敗: %v", err)
	}

	// 空の検索条件
	pref := &domain.Preference{
		UserID: "searcher5",
	}

	results, err := repo.FindCandidates(ctx, "searcher5", pref)
	if err != nil {
		t.Fatalf("FindCandidates失敗: %v", err)
	}

	// 認証済みユーザーのみが返される
	for _, result := range results {
		if !result.User.Verified {
			t.Errorf("未認証ユーザー%sが候補に含まれている", result.User.ID)
		}
	}

	// 認証済みユーザーが含まれていることを確認
	found := false
	for _, result := range results {
		if result.User.ID == "verified" {
			found = true
		}
	}
	if !found {
		t.Error("認証済みユーザーが候補に含まれていない")
	}
}

func TestUserRepository_FindCandidates_MultipleFilters(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewUserRepository(td.Pool)
	ctx := context.Background()

	// 検索ユーザー作成
	searchUser := &domain.User{
		ID:           "searcher6",
		Nickname:     "searcher6",
		Gender:       domain.GenderMale,
		BirthDate:    time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureTokyo,
		Verified:     true,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}
	err := repo.Create(ctx, searchUser)
	if err != nil {
		t.Fatalf("検索ユーザー作成失敗: %v", err)
	}

	// マッチする候補ユーザー作成
	matchCandidate := &domain.User{
		ID:           "match_all",
		Nickname:     "match_all",
		Gender:       domain.GenderFemale,
		BirthDate:    time.Date(1995, 6, 15, 0, 0, 0, 0, time.UTC), // 30歳
		Prefecture:   domain.PrefectureTokyo,
		Verified:     true,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}
	err = repo.Create(ctx, matchCandidate)
	if err != nil {
		t.Fatalf("マッチ候補ユーザー作成失敗: %v", err)
	}

	matchProfile := &domain.Profile{
		UserID:           "match_all",
		Height:           165,
		BodyType:         domain.BodyTypeAverage,
		Education:        domain.EducationUniversity,
		Occupation:       "エンジニア",
		IncomeLevel:      600,
		MarriageDesire:   domain.MarriageWantSoon,
		ChildrenDesire:   domain.ChildrenWant,
		Smoking:          domain.SmokingNonSmoker,
		Drinking:         domain.DrinkingSocial,
		SelfIntroduction: "テスト",
		UpdatedAt:        time.Now(),
	}
	err = NewProfileRepository(td.Pool).Upsert(ctx, matchProfile)
	if err != nil {
		t.Fatalf("マッチプロフィール作成失敗: %v", err)
	}

	// マッチしない候補ユーザー（身長が低い）
	noMatchCandidate := &domain.User{
		ID:           "no_match",
		Nickname:     "no_match",
		Gender:       domain.GenderFemale,
		BirthDate:    time.Date(1995, 6, 15, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureTokyo,
		Verified:     true,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}
	err = repo.Create(ctx, noMatchCandidate)
	if err != nil {
		t.Fatalf("非マッチ候補ユーザー作成失敗: %v", err)
	}

	noMatchProfile := &domain.Profile{
		UserID:           "no_match",
		Height:           150, // 身長が基準以下
		BodyType:         domain.BodyTypeAverage,
		Education:        domain.EducationUniversity,
		Occupation:       "営業",
		IncomeLevel:      600,
		MarriageDesire:   domain.MarriageWantSoon,
		ChildrenDesire:   domain.ChildrenWant,
		Smoking:          domain.SmokingNonSmoker,
		Drinking:         domain.DrinkingSocial,
		SelfIntroduction: "テスト",
		UpdatedAt:        time.Now(),
	}
	err = NewProfileRepository(td.Pool).Upsert(ctx, noMatchProfile)
	if err != nil {
		t.Fatalf("非マッチプロフィール作成失敗: %v", err)
	}

	// 複数フィルタ条件
	pref := &domain.Preference{
		UserID:    "searcher6",
		AgeMin:    25,
		AgeMax:    35,
		HeightMin: 160,
		HeightMax: 170,
		IncomeMin: 500,
		Prefectures: []domain.PreferencePrefecture{
			{UserID: "searcher6", Prefecture: domain.PrefectureTokyo},
		},
		Educations: []domain.PreferenceEducation{
			{UserID: "searcher6", Education: domain.EducationUniversity},
		},
		MarriageDesires: []domain.PreferenceMarriageDesire{
			{UserID: "searcher6", MarriageDesire: domain.MarriageWantSoon},
		},
		SmokingStatuses: []domain.PreferenceSmokingStatus{
			{UserID: "searcher6", SmokingStatus: domain.SmokingNonSmoker},
		},
		DrinkingStatuses: []domain.PreferenceDrinkingStatus{
			{UserID: "searcher6", DrinkingStatus: domain.DrinkingSocial},
		},
	}

	results, err := repo.FindCandidates(ctx, "searcher6", pref)
	if err != nil {
		t.Fatalf("FindCandidates失敗: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("候補数 = %v, want 1", len(results))
	}

	if results[0].User.ID != "match_all" {
		t.Errorf("候補ユーザーID = %v, want match_all", results[0].User.ID)
	}
}
