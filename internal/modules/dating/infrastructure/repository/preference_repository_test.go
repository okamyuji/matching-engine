package repository

import (
	"context"
	"testing"
	"time"

	"github.com/okamyuji/matching-engine/internal/modules/dating/domain"
	"github.com/okamyuji/matching-engine/internal/testutil"
)

func TestPreferenceRepository_FindByUserID_NormalCase(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewPreferenceRepository(td.Pool)
	ctx := context.Background()

	// テストユーザー作成
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
	userRepo := NewUserRepository(td.Pool)
	err := userRepo.Create(ctx, user)
	if err != nil {
		t.Fatalf("テストユーザー作成失敗: %v", err)
	}

	// テスト希望条件作成（関連データ含む）
	pref := &domain.Preference{
		UserID:    "user001",
		AgeMin:    25,
		AgeMax:    35,
		HeightMin: 160,
		HeightMax: 175,
		IncomeMin: 400,
		Prefectures: []domain.PreferencePrefecture{
			{UserID: "user001", Prefecture: domain.PrefectureTokyo},
			{UserID: "user001", Prefecture: domain.PrefectureOsaka},
		},
		Educations: []domain.PreferenceEducation{
			{UserID: "user001", Education: domain.EducationUniversity},
			{UserID: "user001", Education: domain.EducationGraduate},
		},
		MarriageDesires: []domain.PreferenceMarriageDesire{
			{UserID: "user001", MarriageDesire: domain.MarriageWantSoon},
		},
		SmokingStatuses: []domain.PreferenceSmokingStatus{
			{UserID: "user001", SmokingStatus: domain.SmokingNonSmoker},
		},
		DrinkingStatuses: []domain.PreferenceDrinkingStatus{
			{UserID: "user001", DrinkingStatus: domain.DrinkingSocial},
		},
		UpdatedAt: time.Now(),
	}

	err = repo.Upsert(ctx, pref)
	if err != nil {
		t.Fatalf("希望条件作成失敗: %v", err)
	}

	// FindByUserIDテスト
	found, err := repo.FindByUserID(ctx, "user001")
	if err != nil {
		t.Fatalf("FindByUserID失敗: %v", err)
	}

	if found.UserID != pref.UserID {
		t.Errorf("UserID = %v, want %v", found.UserID, pref.UserID)
	}
	if found.AgeMin != pref.AgeMin {
		t.Errorf("AgeMin = %v, want %v", found.AgeMin, pref.AgeMin)
	}
	if found.AgeMax != pref.AgeMax {
		t.Errorf("AgeMax = %v, want %v", found.AgeMax, pref.AgeMax)
	}
	if len(found.Prefectures) != 2 {
		t.Errorf("Prefectures count = %v, want 2", len(found.Prefectures))
	}
	if len(found.Educations) != 2 {
		t.Errorf("Educations count = %v, want 2", len(found.Educations))
	}
	if len(found.MarriageDesires) != 1 {
		t.Errorf("MarriageDesires count = %v, want 1", len(found.MarriageDesires))
	}
}

func TestPreferenceRepository_FindByUserID_EmptyRelations(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewPreferenceRepository(td.Pool)
	ctx := context.Background()

	// テストユーザー作成
	user := &domain.User{
		ID:           "user002",
		Nickname:     "user002",
		Gender:       domain.GenderFemale,
		BirthDate:    time.Date(1995, 3, 20, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureOsaka,
		Verified:     true,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}
	userRepo := NewUserRepository(td.Pool)
	err := userRepo.Create(ctx, user)
	if err != nil {
		t.Fatalf("テストユーザー作成失敗: %v", err)
	}

	// 関連データなしの希望条件作成
	pref := &domain.Preference{
		UserID:    "user002",
		AgeMin:    20,
		AgeMax:    30,
		HeightMin: 0,
		HeightMax: 0,
		IncomeMin: 0,
		UpdatedAt: time.Now(),
	}

	err = repo.Upsert(ctx, pref)
	if err != nil {
		t.Fatalf("希望条件作成失敗: %v", err)
	}

	// FindByUserIDテスト
	found, err := repo.FindByUserID(ctx, "user002")
	if err != nil {
		t.Fatalf("FindByUserID失敗: %v", err)
	}

	if found.UserID != pref.UserID {
		t.Errorf("UserID = %v, want %v", found.UserID, pref.UserID)
	}
	if len(found.Prefectures) != 0 {
		t.Errorf("Prefectures should be empty")
	}
	if len(found.Educations) != 0 {
		t.Errorf("Educations should be empty")
	}
}

func TestPreferenceRepository_FindByUserID_NotFound(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewPreferenceRepository(td.Pool)
	ctx := context.Background()

	// 存在しないユーザーIDで検索
	_, err := repo.FindByUserID(ctx, "nonexistent")
	if err == nil {
		t.Error("存在しない希望条件でエラーが発生すべき")
	}
}

func TestPreferenceRepository_FindByUserID_EmptyUserID(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewPreferenceRepository(td.Pool)
	ctx := context.Background()

	// 空のUserIDで検索
	_, err := repo.FindByUserID(ctx, "")
	if err == nil {
		t.Error("空のUserIDでエラーが発生すべき")
	}
}

func TestPreferenceRepository_Upsert_Insert(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewPreferenceRepository(td.Pool)
	ctx := context.Background()

	// テストユーザー作成
	user := &domain.User{
		ID:           "user003",
		Nickname:     "user003",
		Gender:       domain.GenderMale,
		BirthDate:    time.Date(1992, 7, 10, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureKyoto,
		Verified:     true,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}
	userRepo := NewUserRepository(td.Pool)
	err := userRepo.Create(ctx, user)
	if err != nil {
		t.Fatalf("テストユーザー作成失敗: %v", err)
	}

	// 新規希望条件作成
	pref := &domain.Preference{
		UserID:    "user003",
		AgeMin:    28,
		AgeMax:    38,
		HeightMin: 165,
		HeightMax: 180,
		IncomeMin: 500,
		Prefectures: []domain.PreferencePrefecture{
			{UserID: "user003", Prefecture: domain.PrefectureKyoto},
		},
		Educations: []domain.PreferenceEducation{
			{UserID: "user003", Education: domain.EducationUniversity},
		},
		MarriageDesires: []domain.PreferenceMarriageDesire{
			{UserID: "user003", MarriageDesire: domain.MarriageWantEventually},
		},
		SmokingStatuses: []domain.PreferenceSmokingStatus{
			{UserID: "user003", SmokingStatus: domain.SmokingNonSmoker},
			{UserID: "user003", SmokingStatus: domain.SmokingOccasional},
		},
		DrinkingStatuses: []domain.PreferenceDrinkingStatus{
			{UserID: "user003", DrinkingStatus: domain.DrinkingSocial},
			{UserID: "user003", DrinkingStatus: domain.DrinkingRegular},
		},
		UpdatedAt: time.Now(),
	}

	err = repo.Upsert(ctx, pref)
	if err != nil {
		t.Fatalf("Upsert(Insert)失敗: %v", err)
	}

	// 作成した希望条件を取得して確認
	found, err := repo.FindByUserID(ctx, "user003")
	if err != nil {
		t.Fatalf("FindByUserID失敗: %v", err)
	}

	if found.UserID != pref.UserID {
		t.Errorf("UserID = %v, want %v", found.UserID, pref.UserID)
	}
	if found.AgeMin != pref.AgeMin {
		t.Errorf("AgeMin = %v, want %v", found.AgeMin, pref.AgeMin)
	}
	if len(found.Prefectures) != 1 {
		t.Errorf("Prefectures count = %v, want 1", len(found.Prefectures))
	}
	if len(found.SmokingStatuses) != 2 {
		t.Errorf("SmokingStatuses count = %v, want 2", len(found.SmokingStatuses))
	}
	if len(found.DrinkingStatuses) != 2 {
		t.Errorf("DrinkingStatuses count = %v, want 2", len(found.DrinkingStatuses))
	}
}

func TestPreferenceRepository_Upsert_Update(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewPreferenceRepository(td.Pool)
	ctx := context.Background()

	// テストユーザー作成
	user := &domain.User{
		ID:           "user004",
		Nickname:     "user004",
		Gender:       domain.GenderFemale,
		BirthDate:    time.Date(1993, 8, 25, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureOsaka,
		Verified:     true,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}
	userRepo := NewUserRepository(td.Pool)
	err := userRepo.Create(ctx, user)
	if err != nil {
		t.Fatalf("テストユーザー作成失敗: %v", err)
	}

	// 初回の希望条件作成
	pref := &domain.Preference{
		UserID:    "user004",
		AgeMin:    25,
		AgeMax:    35,
		HeightMin: 160,
		HeightMax: 170,
		IncomeMin: 400,
		Prefectures: []domain.PreferencePrefecture{
			{UserID: "user004", Prefecture: domain.PrefectureTokyo},
		},
		Educations: []domain.PreferenceEducation{
			{UserID: "user004", Education: domain.EducationUniversity},
		},
		UpdatedAt: time.Now(),
	}

	err = repo.Upsert(ctx, pref)
	if err != nil {
		t.Fatalf("初回Upsert失敗: %v", err)
	}

	// 希望条件更新（年齢範囲変更、都道府県追加、学歴追加）
	pref.AgeMin = 27
	pref.AgeMax = 37
	pref.IncomeMin = 500
	pref.Prefectures = []domain.PreferencePrefecture{
		{UserID: "user004", Prefecture: domain.PrefectureTokyo},
		{UserID: "user004", Prefecture: domain.PrefectureOsaka},
		{UserID: "user004", Prefecture: domain.PrefectureKyoto},
	}
	pref.Educations = []domain.PreferenceEducation{
		{UserID: "user004", Education: domain.EducationUniversity},
		{UserID: "user004", Education: domain.EducationGraduate},
	}
	pref.UpdatedAt = time.Now()

	err = repo.Upsert(ctx, pref)
	if err != nil {
		t.Fatalf("更新Upsert失敗: %v", err)
	}

	// 更新を確認
	found, err := repo.FindByUserID(ctx, "user004")
	if err != nil {
		t.Fatalf("FindByUserID失敗: %v", err)
	}

	if found.AgeMin != 27 {
		t.Errorf("AgeMin = %v, want 27", found.AgeMin)
	}
	if found.AgeMax != 37 {
		t.Errorf("AgeMax = %v, want 37", found.AgeMax)
	}
	if found.IncomeMin != 500 {
		t.Errorf("IncomeMin = %v, want 500", found.IncomeMin)
	}
	if len(found.Prefectures) != 3 {
		t.Errorf("Prefectures count = %v, want 3", len(found.Prefectures))
	}
	if len(found.Educations) != 2 {
		t.Errorf("Educations count = %v, want 2", len(found.Educations))
	}
}

func TestPreferenceRepository_Upsert_MinimalFields(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewPreferenceRepository(td.Pool)
	ctx := context.Background()

	// テストユーザー作成
	user := &domain.User{
		ID:           "user005",
		Nickname:     "user005",
		Gender:       domain.GenderMale,
		BirthDate:    time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureTokyo,
		Verified:     true,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}
	userRepo := NewUserRepository(td.Pool)
	err := userRepo.Create(ctx, user)
	if err != nil {
		t.Fatalf("テストユーザー作成失敗: %v", err)
	}

	// 最小限のフィールドのみ
	pref := &domain.Preference{
		UserID:    "user005",
		AgeMin:    20,
		AgeMax:    40,
		HeightMin: 0,
		HeightMax: 0,
		IncomeMin: 0,
		UpdatedAt: time.Now(),
	}

	err = repo.Upsert(ctx, pref)
	if err != nil {
		t.Fatalf("最小限フィールドでのUpsert失敗: %v", err)
	}

	found, err := repo.FindByUserID(ctx, "user005")
	if err != nil {
		t.Fatalf("FindByUserID失敗: %v", err)
	}

	if found.UserID != pref.UserID {
		t.Errorf("UserID = %v, want %v", found.UserID, pref.UserID)
	}
	if found.AgeMin != pref.AgeMin {
		t.Errorf("AgeMin = %v, want %v", found.AgeMin, pref.AgeMin)
	}
}

func TestPreferenceRepository_Upsert_ClearRelations(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewPreferenceRepository(td.Pool)
	ctx := context.Background()

	// テストユーザー作成
	user := &domain.User{
		ID:           "user006",
		Nickname:     "user006",
		Gender:       domain.GenderFemale,
		BirthDate:    time.Date(1994, 5, 15, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureTokyo,
		Verified:     true,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}
	userRepo := NewUserRepository(td.Pool)
	err := userRepo.Create(ctx, user)
	if err != nil {
		t.Fatalf("テストユーザー作成失敗: %v", err)
	}

	// 関連データありの希望条件作成
	pref := &domain.Preference{
		UserID:    "user006",
		AgeMin:    25,
		AgeMax:    35,
		HeightMin: 160,
		HeightMax: 170,
		IncomeMin: 400,
		Prefectures: []domain.PreferencePrefecture{
			{UserID: "user006", Prefecture: domain.PrefectureTokyo},
			{UserID: "user006", Prefecture: domain.PrefectureOsaka},
		},
		Educations: []domain.PreferenceEducation{
			{UserID: "user006", Education: domain.EducationUniversity},
		},
		UpdatedAt: time.Now(),
	}

	err = repo.Upsert(ctx, pref)
	if err != nil {
		t.Fatalf("初回Upsert失敗: %v", err)
	}

	// 関連データを空配列で上書き（削除）
	pref.Prefectures = []domain.PreferencePrefecture{}
	pref.Educations = []domain.PreferenceEducation{}
	pref.UpdatedAt = time.Now()

	err = repo.Upsert(ctx, pref)
	if err != nil {
		t.Fatalf("関連データクリアUpsert失敗: %v", err)
	}

	// 削除を確認
	found, err := repo.FindByUserID(ctx, "user006")
	if err != nil {
		t.Fatalf("FindByUserID失敗: %v", err)
	}

	if len(found.Prefectures) != 0 {
		t.Errorf("Prefectures should be cleared but got %v items", len(found.Prefectures))
	}
	if len(found.Educations) != 0 {
		t.Errorf("Educations should be cleared but got %v items", len(found.Educations))
	}
}

func TestPreferenceRepository_Upsert_Idempotent(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewPreferenceRepository(td.Pool)
	ctx := context.Background()

	// テストユーザー作成
	user := &domain.User{
		ID:           "user007",
		Nickname:     "user007",
		Gender:       domain.GenderMale,
		BirthDate:    time.Date(1991, 11, 11, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureKyoto,
		Verified:     true,
		EloRating:    1500,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}
	userRepo := NewUserRepository(td.Pool)
	err := userRepo.Create(ctx, user)
	if err != nil {
		t.Fatalf("テストユーザー作成失敗: %v", err)
	}

	pref := &domain.Preference{
		UserID:    "user007",
		AgeMin:    30,
		AgeMax:    40,
		HeightMin: 165,
		HeightMax: 175,
		IncomeMin: 600,
		Prefectures: []domain.PreferencePrefecture{
			{UserID: "user007", Prefecture: domain.PrefectureKyoto},
		},
		Educations: []domain.PreferenceEducation{
			{UserID: "user007", Education: domain.EducationGraduate},
		},
		UpdatedAt: time.Now(),
	}

	// 同じ内容で3回Upsert
	for i := 0; i < 3; i++ {
		err = repo.Upsert(ctx, pref)
		if err != nil {
			t.Fatalf("Upsert %d回目失敗: %v", i+1, err)
		}
	}

	// 最終的に正しく保存されていることを確認
	found, err := repo.FindByUserID(ctx, "user007")
	if err != nil {
		t.Fatalf("FindByUserID失敗: %v", err)
	}

	if found.UserID != pref.UserID {
		t.Errorf("UserID = %v, want %v", found.UserID, pref.UserID)
	}
	if found.AgeMin != pref.AgeMin {
		t.Errorf("AgeMin = %v, want %v", found.AgeMin, pref.AgeMin)
	}
	if len(found.Prefectures) != 1 {
		t.Errorf("Prefectures count = %v, want 1", len(found.Prefectures))
	}
	if len(found.Educations) != 1 {
		t.Errorf("Educations count = %v, want 1", len(found.Educations))
	}
}
