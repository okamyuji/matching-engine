package repository

import (
	"context"
	"testing"
	"time"

	"github.com/okamyuji/matching-engine/internal/modules/dating/domain"
	"github.com/okamyuji/matching-engine/internal/testutil"
)

func TestLikeRepository_Save_NormalCase(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewLikeRepository(td.Pool)
	ctx := context.Background()

	// テストユーザー2人作成
	userRepo := NewUserRepository(td.Pool)
	user1 := &domain.User{
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
	user2 := &domain.User{
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
	if err := userRepo.Create(ctx, user1); err != nil {
		t.Fatalf("ユーザー1作成失敗: %v", err)
	}
	if err := userRepo.Create(ctx, user2); err != nil {
		t.Fatalf("ユーザー2作成失敗: %v", err)
	}

	// いいね作成
	like := &Like{
		ID:         "like001",
		FromUserID: "user001",
		ToUserID:   "user002",
		CreatedAt:  time.Now(),
	}

	err := repo.Save(ctx, like)
	if err != nil {
		t.Fatalf("いいね保存失敗: %v", err)
	}

	// 保存されたことを確認
	likes, err := repo.FindByTargetUserID(ctx, "user002")
	if err != nil {
		t.Fatalf("いいね取得失敗: %v", err)
	}
	if len(likes) != 1 {
		t.Errorf("いいね数 = %d, want 1", len(likes))
	}
	if likes[0].FromUserID != "user001" {
		t.Errorf("FromUserID = %s, want user001", likes[0].FromUserID)
	}
}

func TestLikeRepository_Save_DuplicateLike(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewLikeRepository(td.Pool)
	ctx := context.Background()

	// テストユーザー2人作成
	userRepo := NewUserRepository(td.Pool)
	user1 := &domain.User{
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
	user2 := &domain.User{
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
	if err := userRepo.Create(ctx, user1); err != nil {
		t.Fatalf("ユーザー1作成失敗: %v", err)
	}
	if err := userRepo.Create(ctx, user2); err != nil {
		t.Fatalf("ユーザー2作成失敗: %v", err)
	}

	// 1回目のいいね
	like1 := &Like{
		ID:         "like001",
		FromUserID: "user001",
		ToUserID:   "user002",
		CreatedAt:  time.Now(),
	}
	err := repo.Save(ctx, like1)
	if err != nil {
		t.Fatalf("1回目のいいね保存失敗: %v", err)
	}

	// 2回目のいいね（同じID）
	like2 := &Like{
		ID:         "like001",
		FromUserID: "user001",
		ToUserID:   "user002",
		CreatedAt:  time.Now(),
	}
	err = repo.Save(ctx, like2)
	if err == nil {
		t.Error("重複IDでエラーが発生すべき")
	}
}

func TestLikeRepository_FindByTargetUserID_NormalCase(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewLikeRepository(td.Pool)
	ctx := context.Background()

	// テストユーザー3人作成（user002が2人からいいねを受け取る）
	userRepo := NewUserRepository(td.Pool)
	users := []*domain.User{
		{
			ID:           "user001",
			Nickname:     "user001",
			Gender:       domain.GenderMale,
			BirthDate:    time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
			Prefecture:   domain.PrefectureTokyo,
			Verified:     true,
			EloRating:    1500,
			CreatedAt:    time.Now(),
			LastActiveAt: time.Now(),
		},
		{
			ID:           "user002",
			Nickname:     "user002",
			Gender:       domain.GenderFemale,
			BirthDate:    time.Date(1995, 3, 20, 0, 0, 0, 0, time.UTC),
			Prefecture:   domain.PrefectureOsaka,
			Verified:     true,
			EloRating:    1500,
			CreatedAt:    time.Now(),
			LastActiveAt: time.Now(),
		},
		{
			ID:           "user003",
			Nickname:     "user003",
			Gender:       domain.GenderMale,
			BirthDate:    time.Date(1992, 7, 10, 0, 0, 0, 0, time.UTC),
			Prefecture:   domain.PrefectureKyoto,
			Verified:     true,
			EloRating:    1500,
			CreatedAt:    time.Now(),
			LastActiveAt: time.Now(),
		},
	}
	for _, user := range users {
		if err := userRepo.Create(ctx, user); err != nil {
			t.Fatalf("ユーザー作成失敗: %v", err)
		}
	}

	// user002が2人からいいねを受け取る
	now := time.Now()
	likes := []*Like{
		{
			ID:         "like001",
			FromUserID: "user001",
			ToUserID:   "user002",
			CreatedAt:  now.Add(-1 * time.Hour),
		},
		{
			ID:         "like002",
			FromUserID: "user003",
			ToUserID:   "user002",
			CreatedAt:  now,
		},
	}
	for _, like := range likes {
		if err := repo.Save(ctx, like); err != nil {
			t.Fatalf("いいね保存失敗: %v", err)
		}
	}

	// user002が受け取ったいいねを取得
	found, err := repo.FindByTargetUserID(ctx, "user002")
	if err != nil {
		t.Fatalf("いいね取得失敗: %v", err)
	}

	if len(found) != 2 {
		t.Fatalf("いいね数 = %d, want 2", len(found))
	}

	// 新しい順にソートされているか確認
	if found[0].FromUserID != "user003" {
		t.Errorf("1番目のFromUserID = %s, want user003", found[0].FromUserID)
	}
	if found[1].FromUserID != "user001" {
		t.Errorf("2番目のFromUserID = %s, want user001", found[1].FromUserID)
	}
}

func TestLikeRepository_FindByTargetUserID_EmptyResult(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewLikeRepository(td.Pool)
	ctx := context.Background()

	// テストユーザー作成（いいねを受け取っていない）
	userRepo := NewUserRepository(td.Pool)
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
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("ユーザー作成失敗: %v", err)
	}

	// いいねを受け取っていないユーザーで検索
	likes, err := repo.FindByTargetUserID(ctx, "user001")
	if err != nil {
		t.Fatalf("いいね取得失敗: %v", err)
	}

	if len(likes) != 0 {
		t.Errorf("いいね数 = %d, want 0", len(likes))
	}
}

func TestLikeRepository_FindByTargetUserID_EmptyUserID(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewLikeRepository(td.Pool)
	ctx := context.Background()

	// 空のUserIDで検索
	likes, err := repo.FindByTargetUserID(ctx, "")
	if err != nil {
		t.Fatalf("いいね取得失敗: %v", err)
	}

	if len(likes) != 0 {
		t.Errorf("いいね数 = %d, want 0", len(likes))
	}
}

func TestLikeRepository_CheckMutual_MutualExists(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewLikeRepository(td.Pool)
	ctx := context.Background()

	// テストユーザー2人作成
	userRepo := NewUserRepository(td.Pool)
	user1 := &domain.User{
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
	user2 := &domain.User{
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
	if err := userRepo.Create(ctx, user1); err != nil {
		t.Fatalf("ユーザー1作成失敗: %v", err)
	}
	if err := userRepo.Create(ctx, user2); err != nil {
		t.Fatalf("ユーザー2作成失敗: %v", err)
	}

	// 相互いいね作成
	like1 := &Like{
		ID:         "like001",
		FromUserID: "user001",
		ToUserID:   "user002",
		CreatedAt:  time.Now(),
	}
	like2 := &Like{
		ID:         "like002",
		FromUserID: "user002",
		ToUserID:   "user001",
		CreatedAt:  time.Now(),
	}
	if err := repo.Save(ctx, like1); err != nil {
		t.Fatalf("いいね1保存失敗: %v", err)
	}
	if err := repo.Save(ctx, like2); err != nil {
		t.Fatalf("いいね2保存失敗: %v", err)
	}

	// 相互いいねチェック
	isMutual, err := repo.CheckMutual(ctx, "user001", "user002")
	if err != nil {
		t.Fatalf("相互いいねチェック失敗: %v", err)
	}

	if !isMutual {
		t.Error("相互いいねが存在すべき")
	}
}

func TestLikeRepository_CheckMutual_OnlyOneWay(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewLikeRepository(td.Pool)
	ctx := context.Background()

	// テストユーザー2人作成
	userRepo := NewUserRepository(td.Pool)
	user1 := &domain.User{
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
	user2 := &domain.User{
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
	if err := userRepo.Create(ctx, user1); err != nil {
		t.Fatalf("ユーザー1作成失敗: %v", err)
	}
	if err := userRepo.Create(ctx, user2); err != nil {
		t.Fatalf("ユーザー2作成失敗: %v", err)
	}

	// 片方向のいいねのみ
	like := &Like{
		ID:         "like001",
		FromUserID: "user001",
		ToUserID:   "user002",
		CreatedAt:  time.Now(),
	}
	if err := repo.Save(ctx, like); err != nil {
		t.Fatalf("いいね保存失敗: %v", err)
	}

	// 相互いいねチェック
	isMutual, err := repo.CheckMutual(ctx, "user001", "user002")
	if err != nil {
		t.Fatalf("相互いいねチェック失敗: %v", err)
	}

	if isMutual {
		t.Error("相互いいねが存在しないべき")
	}
}

func TestLikeRepository_CheckMutual_NoLikes(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewLikeRepository(td.Pool)
	ctx := context.Background()

	// テストユーザー2人作成（いいねなし）
	userRepo := NewUserRepository(td.Pool)
	user1 := &domain.User{
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
	user2 := &domain.User{
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
	if err := userRepo.Create(ctx, user1); err != nil {
		t.Fatalf("ユーザー1作成失敗: %v", err)
	}
	if err := userRepo.Create(ctx, user2); err != nil {
		t.Fatalf("ユーザー2作成失敗: %v", err)
	}

	// 相互いいねチェック（いいねなし）
	isMutual, err := repo.CheckMutual(ctx, "user001", "user002")
	if err != nil {
		t.Fatalf("相互いいねチェック失敗: %v", err)
	}

	if isMutual {
		t.Error("相互いいねが存在しないべき")
	}
}

func TestLikeRepository_FindByFromUserID_NormalCase(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewLikeRepository(td.Pool)
	ctx := context.Background()

	// テストユーザー3人作成（user001が2人にいいねを送る）
	userRepo := NewUserRepository(td.Pool)
	users := []*domain.User{
		{
			ID:           "user001",
			Nickname:     "user001",
			Gender:       domain.GenderMale,
			BirthDate:    time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
			Prefecture:   domain.PrefectureTokyo,
			Verified:     true,
			EloRating:    1500,
			CreatedAt:    time.Now(),
			LastActiveAt: time.Now(),
		},
		{
			ID:           "user002",
			Nickname:     "user002",
			Gender:       domain.GenderFemale,
			BirthDate:    time.Date(1995, 3, 20, 0, 0, 0, 0, time.UTC),
			Prefecture:   domain.PrefectureOsaka,
			Verified:     true,
			EloRating:    1500,
			CreatedAt:    time.Now(),
			LastActiveAt: time.Now(),
		},
		{
			ID:           "user003",
			Nickname:     "user003",
			Gender:       domain.GenderFemale,
			BirthDate:    time.Date(1992, 7, 10, 0, 0, 0, 0, time.UTC),
			Prefecture:   domain.PrefectureKyoto,
			Verified:     true,
			EloRating:    1500,
			CreatedAt:    time.Now(),
			LastActiveAt: time.Now(),
		},
	}
	for _, user := range users {
		if err := userRepo.Create(ctx, user); err != nil {
			t.Fatalf("ユーザー作成失敗: %v", err)
		}
	}

	// user001が2人にいいねを送る
	now := time.Now()
	likes := []*Like{
		{
			ID:         "like001",
			FromUserID: "user001",
			ToUserID:   "user002",
			CreatedAt:  now.Add(-1 * time.Hour),
		},
		{
			ID:         "like002",
			FromUserID: "user001",
			ToUserID:   "user003",
			CreatedAt:  now,
		},
	}
	for _, like := range likes {
		if err := repo.Save(ctx, like); err != nil {
			t.Fatalf("いいね保存失敗: %v", err)
		}
	}

	// user001が送ったいいねを取得
	found, err := repo.FindByFromUserID(ctx, "user001")
	if err != nil {
		t.Fatalf("いいね取得失敗: %v", err)
	}

	if len(found) != 2 {
		t.Fatalf("いいね数 = %d, want 2", len(found))
	}

	// 新しい順にソートされているか確認
	if found[0].ToUserID != "user003" {
		t.Errorf("1番目のToUserID = %s, want user003", found[0].ToUserID)
	}
	if found[1].ToUserID != "user002" {
		t.Errorf("2番目のToUserID = %s, want user002", found[1].ToUserID)
	}
}

func TestLikeRepository_FindByFromUserID_EmptyResult(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewLikeRepository(td.Pool)
	ctx := context.Background()

	// テストユーザー作成（いいねを送っていない）
	userRepo := NewUserRepository(td.Pool)
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
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("ユーザー作成失敗: %v", err)
	}

	// いいねを送っていないユーザーで検索
	likes, err := repo.FindByFromUserID(ctx, "user001")
	if err != nil {
		t.Fatalf("いいね取得失敗: %v", err)
	}

	if len(likes) != 0 {
		t.Errorf("いいね数 = %d, want 0", len(likes))
	}
}

func TestLikeRepository_FindByFromUserID_EmptyUserID(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewLikeRepository(td.Pool)
	ctx := context.Background()

	// 空のUserIDで検索
	likes, err := repo.FindByFromUserID(ctx, "")
	if err != nil {
		t.Fatalf("いいね取得失敗: %v", err)
	}

	if len(likes) != 0 {
		t.Errorf("いいね数 = %d, want 0", len(likes))
	}
}
