package repository

import (
	"context"
	"testing"
	"time"

	"github.com/okamyuji/matching-engine/internal/modules/dating/domain"
	"github.com/okamyuji/matching-engine/internal/testutil"
)

func TestProfileRepository_FindByUserID_NormalCase(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewProfileRepository(td.Pool)
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

	// テストプロフィール作成
	profile := &domain.Profile{
		UserID:           "user001",
		Height:           175,
		BodyType:         domain.BodyTypeAthletic,
		Education:        domain.EducationUniversity,
		Occupation:       "エンジニア",
		IncomeLevel:      600,
		MarriageDesire:   domain.MarriageWantSoon,
		ChildrenDesire:   domain.ChildrenWant,
		Smoking:          domain.SmokingNonSmoker,
		Drinking:         domain.DrinkingSocial,
		SelfIntroduction: "テストプロフィール",
		UpdatedAt:        time.Now(),
	}

	err = repo.Upsert(ctx, profile)
	if err != nil {
		t.Fatalf("プロフィール作成失敗: %v", err)
	}

	// FindByUserIDテスト
	found, err := repo.FindByUserID(ctx, "user001")
	if err != nil {
		t.Fatalf("FindByUserID失敗: %v", err)
	}

	if found.UserID != profile.UserID {
		t.Errorf("UserID = %v, want %v", found.UserID, profile.UserID)
	}
	if found.Height != profile.Height {
		t.Errorf("Height = %v, want %v", found.Height, profile.Height)
	}
	if found.BodyType != profile.BodyType {
		t.Errorf("BodyType = %v, want %v", found.BodyType, profile.BodyType)
	}
	if found.Education != profile.Education {
		t.Errorf("Education = %v, want %v", found.Education, profile.Education)
	}
}

func TestProfileRepository_FindByUserID_NotFound(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewProfileRepository(td.Pool)
	ctx := context.Background()

	// 存在しないユーザーIDで検索
	_, err := repo.FindByUserID(ctx, "nonexistent")
	if err == nil {
		t.Error("存在しないプロフィールでエラーが発生すべき")
	}
}

func TestProfileRepository_FindByUserID_EmptyUserID(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewProfileRepository(td.Pool)
	ctx := context.Background()

	// 空のUserIDで検索
	_, err := repo.FindByUserID(ctx, "")
	if err == nil {
		t.Error("空のUserIDでエラーが発生すべき")
	}
}

func TestProfileRepository_Upsert_Insert(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewProfileRepository(td.Pool)
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

	// 新規プロフィール作成
	profile := &domain.Profile{
		UserID:           "user002",
		Height:           165,
		BodyType:         domain.BodyTypeSlim,
		Education:        domain.EducationGraduate,
		Occupation:       "デザイナー",
		IncomeLevel:      500,
		MarriageDesire:   domain.MarriageWantEventually,
		ChildrenDesire:   domain.ChildrenUndecided,
		Smoking:          domain.SmokingNonSmoker,
		Drinking:         domain.DrinkingSocial,
		SelfIntroduction: "よろしくお願いします",
		UpdatedAt:        time.Now(),
	}

	err = repo.Upsert(ctx, profile)
	if err != nil {
		t.Fatalf("Upsert(Insert)失敗: %v", err)
	}

	// 作成したプロフィールを取得して確認
	found, err := repo.FindByUserID(ctx, "user002")
	if err != nil {
		t.Fatalf("FindByUserID失敗: %v", err)
	}

	if found.UserID != profile.UserID {
		t.Errorf("UserID = %v, want %v", found.UserID, profile.UserID)
	}
	if found.Height != profile.Height {
		t.Errorf("Height = %v, want %v", found.Height, profile.Height)
	}
}

func TestProfileRepository_Upsert_Update(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewProfileRepository(td.Pool)
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

	// 初回のプロフィール作成
	profile := &domain.Profile{
		UserID:           "user003",
		Height:           170,
		BodyType:         domain.BodyTypeAverage,
		Education:        domain.EducationUniversity,
		Occupation:       "営業",
		IncomeLevel:      400,
		MarriageDesire:   domain.MarriageWantSoon,
		ChildrenDesire:   domain.ChildrenWant,
		Smoking:          domain.SmokingNonSmoker,
		Drinking:         domain.DrinkingSocial,
		SelfIntroduction: "初回",
		UpdatedAt:        time.Now(),
	}

	err = repo.Upsert(ctx, profile)
	if err != nil {
		t.Fatalf("初回Upsert失敗: %v", err)
	}

	// プロフィール更新
	profile.Height = 172
	profile.IncomeLevel = 500
	profile.SelfIntroduction = "更新後"
	profile.UpdatedAt = time.Now()

	err = repo.Upsert(ctx, profile)
	if err != nil {
		t.Fatalf("更新Upsert失敗: %v", err)
	}

	// 更新を確認
	found, err := repo.FindByUserID(ctx, "user003")
	if err != nil {
		t.Fatalf("FindByUserID失敗: %v", err)
	}

	if found.Height != 172 {
		t.Errorf("Height = %v, want 172", found.Height)
	}
	if found.IncomeLevel != 500 {
		t.Errorf("IncomeLevel = %v, want 500", found.IncomeLevel)
	}
	if found.SelfIntroduction != "更新後" {
		t.Errorf("SelfIntroduction = %v, want 更新後", found.SelfIntroduction)
	}
}

func TestProfileRepository_Upsert_MinimalFields(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewProfileRepository(td.Pool)
	ctx := context.Background()

	// テストユーザー作成
	user := &domain.User{
		ID:           "user004",
		Nickname:     "user004",
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
	profile := &domain.Profile{
		UserID:           "user004",
		Height:           0,
		BodyType:         domain.BodyTypeAverage,
		Education:        domain.EducationUniversity,
		Occupation:       "",
		IncomeLevel:      0,
		MarriageDesire:   domain.MarriageWantSoon,
		ChildrenDesire:   domain.ChildrenWant,
		Smoking:          domain.SmokingNonSmoker,
		Drinking:         domain.DrinkingSocial,
		SelfIntroduction: "",
		UpdatedAt:        time.Now(),
	}

	err = repo.Upsert(ctx, profile)
	if err != nil {
		t.Fatalf("最小限フィールドでのUpsert失敗: %v", err)
	}

	found, err := repo.FindByUserID(ctx, "user004")
	if err != nil {
		t.Fatalf("FindByUserID失敗: %v", err)
	}

	if found.UserID != profile.UserID {
		t.Errorf("UserID = %v, want %v", found.UserID, profile.UserID)
	}
}

func TestProfileRepository_Upsert_Idempotent(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewProfileRepository(td.Pool)
	ctx := context.Background()

	// テストユーザー作成
	user := &domain.User{
		ID:           "user005",
		Nickname:     "user005",
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

	profile := &domain.Profile{
		UserID:           "user005",
		Height:           160,
		BodyType:         domain.BodyTypeSlim,
		Education:        domain.EducationUniversity,
		Occupation:       "看護師",
		IncomeLevel:      450,
		MarriageDesire:   domain.MarriageWantSoon,
		ChildrenDesire:   domain.ChildrenWant,
		Smoking:          domain.SmokingNonSmoker,
		Drinking:         domain.DrinkingNonDrinker,
		SelfIntroduction: "同じ内容で複数回Upsert",
		UpdatedAt:        time.Now(),
	}

	// 同じ内容で3回Upsert
	for i := 0; i < 3; i++ {
		err = repo.Upsert(ctx, profile)
		if err != nil {
			t.Fatalf("Upsert %d回目失敗: %v", i+1, err)
		}
	}

	// 最終的に1件のみ存在することを確認
	found, err := repo.FindByUserID(ctx, "user005")
	if err != nil {
		t.Fatalf("FindByUserID失敗: %v", err)
	}

	if found.UserID != profile.UserID {
		t.Errorf("UserID = %v, want %v", found.UserID, profile.UserID)
	}
	if found.SelfIntroduction != profile.SelfIntroduction {
		t.Errorf("SelfIntroduction = %v, want %v", found.SelfIntroduction, profile.SelfIntroduction)
	}
}

func TestProfileRepository_Upsert_AllFields(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewProfileRepository(td.Pool)
	ctx := context.Background()

	// テストユーザー作成
	user := &domain.User{
		ID:           "user006",
		Nickname:     "user006",
		Gender:       domain.GenderMale,
		BirthDate:    time.Date(1988, 12, 1, 0, 0, 0, 0, time.UTC),
		Prefecture:   domain.PrefectureAichi,
		Verified:     true,
		EloRating:    1600,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}
	userRepo := NewUserRepository(td.Pool)
	err := userRepo.Create(ctx, user)
	if err != nil {
		t.Fatalf("テストユーザー作成失敗: %v", err)
	}

	// 全フィールドを設定
	profile := &domain.Profile{
		UserID:           "user006",
		Height:           180,
		BodyType:         domain.BodyTypeAthletic,
		Education:        domain.EducationGraduate,
		Occupation:       "医師",
		IncomeLevel:      1000,
		MarriageDesire:   domain.MarriageWantEventually,
		ChildrenDesire:   domain.ChildrenWant,
		Smoking:          domain.SmokingNonSmoker,
		Drinking:         domain.DrinkingSocial,
		SelfIntroduction: "全フィールド設定のテスト。趣味はスポーツと読書です。",
		UpdatedAt:        time.Now(),
	}

	err = repo.Upsert(ctx, profile)
	if err != nil {
		t.Fatalf("Upsert失敗: %v", err)
	}

	found, err := repo.FindByUserID(ctx, "user006")
	if err != nil {
		t.Fatalf("FindByUserID失敗: %v", err)
	}

	// 全フィールドの確認
	if found.UserID != profile.UserID {
		t.Errorf("UserID = %v, want %v", found.UserID, profile.UserID)
	}
	if found.Height != profile.Height {
		t.Errorf("Height = %v, want %v", found.Height, profile.Height)
	}
	if found.BodyType != profile.BodyType {
		t.Errorf("BodyType = %v, want %v", found.BodyType, profile.BodyType)
	}
	if found.Education != profile.Education {
		t.Errorf("Education = %v, want %v", found.Education, profile.Education)
	}
	if found.Occupation != profile.Occupation {
		t.Errorf("Occupation = %v, want %v", found.Occupation, profile.Occupation)
	}
	if found.IncomeLevel != profile.IncomeLevel {
		t.Errorf("IncomeLevel = %v, want %v", found.IncomeLevel, profile.IncomeLevel)
	}
	if found.MarriageDesire != profile.MarriageDesire {
		t.Errorf("MarriageDesire = %v, want %v", found.MarriageDesire, profile.MarriageDesire)
	}
	if found.ChildrenDesire != profile.ChildrenDesire {
		t.Errorf("ChildrenDesire = %v, want %v", found.ChildrenDesire, profile.ChildrenDesire)
	}
	if found.Smoking != profile.Smoking {
		t.Errorf("Smoking = %v, want %v", found.Smoking, profile.Smoking)
	}
	if found.Drinking != profile.Drinking {
		t.Errorf("Drinking = %v, want %v", found.Drinking, profile.Drinking)
	}
	if found.SelfIntroduction != profile.SelfIntroduction {
		t.Errorf("SelfIntroduction = %v, want %v", found.SelfIntroduction, profile.SelfIntroduction)
	}
}
