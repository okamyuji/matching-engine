package repository

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/matching-engine/internal/modules/dating/domain"
	"github.com/yourorg/matching-engine/internal/testutil"
)

func TestMatchRepository_Save_NormalCase(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewMatchRepository(td.DB)
	ctx := context.Background()

	// テストユーザー2人作成
	userRepo := NewUserRepository(td.DB)
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

	// マッチ作成
	match := &domain.Match{
		ID:      "match001",
		UserIDA: "user001",
		UserIDB: "user002",
		Score:   0.855,
		Breakdown: map[string]float64{
			"age":        0.9,
			"location":   0.8,
			"preference": 0.865,
		},
		MatchedAt: time.Now(),
	}

	err := repo.Save(ctx, match)
	if err != nil {
		t.Fatalf("マッチ保存失敗: %v", err)
	}

	// 保存されたことを確認
	matches, err := repo.FindByUserID(ctx, "user001")
	if err != nil {
		t.Fatalf("マッチ取得失敗: %v", err)
	}
	if len(matches) != 1 {
		t.Errorf("マッチ数 = %d, want 1", len(matches))
	}
	if matches[0].UserIDA != "user001" {
		t.Errorf("UserIDA = %s, want user001", matches[0].UserIDA)
	}
	if matches[0].UserIDB != "user002" {
		t.Errorf("UserIDB = %s, want user002", matches[0].UserIDB)
	}
	if matches[0].Score != 0.855 {
		t.Errorf("Score = %f, want 0.855", matches[0].Score)
	}
}

func TestMatchRepository_Save_DuplicateMatch(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewMatchRepository(td.DB)
	ctx := context.Background()

	// テストユーザー2人作成
	userRepo := NewUserRepository(td.DB)
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

	// 1回目のマッチ
	match1 := &domain.Match{
		ID:        "match001",
		UserIDA:   "user001",
		UserIDB:   "user002",
		Score:     0.855,
		Breakdown: map[string]float64{"age": 0.9},
		MatchedAt: time.Now(),
	}
	err := repo.Save(ctx, match1)
	if err != nil {
		t.Fatalf("1回目のマッチ保存失敗: %v", err)
	}

	// 2回目のマッチ（同じID）
	match2 := &domain.Match{
		ID:        "match001",
		UserIDA:   "user001",
		UserIDB:   "user002",
		Score:     0.9,
		Breakdown: map[string]float64{"age": 0.95},
		MatchedAt: time.Now(),
	}
	err = repo.Save(ctx, match2)
	if err == nil {
		t.Error("重複IDでエラーが発生すべき")
	}
}

func TestMatchRepository_FindByUserID_AsUserIDA(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewMatchRepository(td.DB)
	ctx := context.Background()

	// テストユーザー3人作成
	userRepo := NewUserRepository(td.DB)
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

	// user001がUserIDAとしてマッチ
	now := time.Now()
	matches := []*domain.Match{
		{
			ID:        "match001",
			UserIDA:   "user001",
			UserIDB:   "user002",
			Score:     0.855,
			Breakdown: map[string]float64{"age": 0.9},
			MatchedAt: now.Add(-1 * time.Hour),
		},
		{
			ID:        "match002",
			UserIDA:   "user001",
			UserIDB:   "user003",
			Score:     0.78,
			Breakdown: map[string]float64{"age": 0.8},
			MatchedAt: now,
		},
	}
	for _, match := range matches {
		if err := repo.Save(ctx, match); err != nil {
			t.Fatalf("マッチ保存失敗: %v", err)
		}
	}

	// user001のマッチを取得
	found, err := repo.FindByUserID(ctx, "user001")
	if err != nil {
		t.Fatalf("マッチ取得失敗: %v", err)
	}

	if len(found) != 2 {
		t.Fatalf("マッチ数 = %d, want 2", len(found))
	}

	// 新しい順にソートされているか確認
	if found[0].UserIDB != "user003" {
		t.Errorf("1番目のUserIDB = %s, want user003", found[0].UserIDB)
	}
	if found[1].UserIDB != "user002" {
		t.Errorf("2番目のUserIDB = %s, want user002", found[1].UserIDB)
	}
}

func TestMatchRepository_FindByUserID_AsUserIDB(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewMatchRepository(td.DB)
	ctx := context.Background()

	// テストユーザー3人作成
	userRepo := NewUserRepository(td.DB)
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

	// user002がUserIDBとしてマッチ
	now := time.Now()
	matches := []*domain.Match{
		{
			ID:        "match001",
			UserIDA:   "user001",
			UserIDB:   "user002",
			Score:     0.855,
			Breakdown: map[string]float64{"age": 0.9},
			MatchedAt: now.Add(-1 * time.Hour),
		},
		{
			ID:        "match002",
			UserIDA:   "user003",
			UserIDB:   "user002",
			Score:     0.78,
			Breakdown: map[string]float64{"age": 0.8},
			MatchedAt: now,
		},
	}
	for _, match := range matches {
		if err := repo.Save(ctx, match); err != nil {
			t.Fatalf("マッチ保存失敗: %v", err)
		}
	}

	// user002のマッチを取得
	found, err := repo.FindByUserID(ctx, "user002")
	if err != nil {
		t.Fatalf("マッチ取得失敗: %v", err)
	}

	if len(found) != 2 {
		t.Fatalf("マッチ数 = %d, want 2", len(found))
	}

	// 新しい順にソートされているか確認
	if found[0].UserIDA != "user003" {
		t.Errorf("1番目のUserIDA = %s, want user003", found[0].UserIDA)
	}
	if found[1].UserIDA != "user001" {
		t.Errorf("2番目のUserIDA = %s, want user001", found[1].UserIDA)
	}
}

func TestMatchRepository_FindByUserID_Mixed(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewMatchRepository(td.DB)
	ctx := context.Background()

	// テストユーザー3人作成
	userRepo := NewUserRepository(td.DB)
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

	// user001がUserIDAとUserIDBの両方として登場
	now := time.Now()
	matches := []*domain.Match{
		{
			ID:        "match001",
			UserIDA:   "user001",
			UserIDB:   "user002",
			Score:     0.855,
			Breakdown: map[string]float64{"age": 0.9},
			MatchedAt: now.Add(-2 * time.Hour),
		},
		{
			ID:        "match002",
			UserIDA:   "user003",
			UserIDB:   "user001",
			Score:     0.78,
			Breakdown: map[string]float64{"age": 0.8},
			MatchedAt: now,
		},
	}
	for _, match := range matches {
		if err := repo.Save(ctx, match); err != nil {
			t.Fatalf("マッチ保存失敗: %v", err)
		}
	}

	// user001のマッチを取得（両方取得できるはず）
	found, err := repo.FindByUserID(ctx, "user001")
	if err != nil {
		t.Fatalf("マッチ取得失敗: %v", err)
	}

	if len(found) != 2 {
		t.Fatalf("マッチ数 = %d, want 2", len(found))
	}

	// 新しい順にソートされているか確認
	if found[0].UserIDA != "user003" {
		t.Errorf("1番目のUserIDA = %s, want user003", found[0].UserIDA)
	}
	if found[1].UserIDB != "user002" {
		t.Errorf("2番目のUserIDB = %s, want user002", found[1].UserIDB)
	}
}

func TestMatchRepository_FindByUserID_EmptyResult(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewMatchRepository(td.DB)
	ctx := context.Background()

	// テストユーザー作成（マッチなし）
	userRepo := NewUserRepository(td.DB)
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

	// マッチがないユーザーで検索
	matches, err := repo.FindByUserID(ctx, "user001")
	if err != nil {
		t.Fatalf("マッチ取得失敗: %v", err)
	}

	if len(matches) != 0 {
		t.Errorf("マッチ数 = %d, want 0", len(matches))
	}
}

func TestMatchRepository_FindMutual_MutualExists(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	matchRepo := NewMatchRepository(td.DB)
	likeRepo := NewLikeRepository(td.DB)
	ctx := context.Background()

	// テストユーザー2人作成
	userRepo := NewUserRepository(td.DB)
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
	if err := likeRepo.Save(ctx, like1); err != nil {
		t.Fatalf("いいね1保存失敗: %v", err)
	}
	if err := likeRepo.Save(ctx, like2); err != nil {
		t.Fatalf("いいね2保存失敗: %v", err)
	}

	// マッチ作成
	match := &domain.Match{
		ID:        "match001",
		UserIDA:   "user001",
		UserIDB:   "user002",
		Score:     0.855,
		Breakdown: map[string]float64{"age": 0.9},
		MatchedAt: time.Now(),
	}
	if err := matchRepo.Save(ctx, match); err != nil {
		t.Fatalf("マッチ保存失敗: %v", err)
	}

	// 相互マッチ取得
	mutuals, err := matchRepo.FindMutual(ctx, "user001")
	if err != nil {
		t.Fatalf("相互マッチ取得失敗: %v", err)
	}

	if len(mutuals) != 1 {
		t.Fatalf("相互マッチ数 = %d, want 1", len(mutuals))
	}

	if mutuals[0].UserIDA != "user001" {
		t.Errorf("UserIDA = %s, want user001", mutuals[0].UserIDA)
	}
	if mutuals[0].UserIDB != "user002" {
		t.Errorf("UserIDB = %s, want user002", mutuals[0].UserIDB)
	}
}

func TestMatchRepository_FindMutual_OnlyOneWayLike(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	matchRepo := NewMatchRepository(td.DB)
	likeRepo := NewLikeRepository(td.DB)
	ctx := context.Background()

	// テストユーザー2人作成
	userRepo := NewUserRepository(td.DB)
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
	if err := likeRepo.Save(ctx, like); err != nil {
		t.Fatalf("いいね保存失敗: %v", err)
	}

	// マッチ作成（Matchエントリーは存在するが相互いいねではない）
	match := &domain.Match{
		ID:        "match001",
		UserIDA:   "user001",
		UserIDB:   "user002",
		Score:     0.855,
		Breakdown: map[string]float64{"age": 0.9},
		MatchedAt: time.Now(),
	}
	if err := matchRepo.Save(ctx, match); err != nil {
		t.Fatalf("マッチ保存失敗: %v", err)
	}

	// 相互マッチ取得（片方向なので0件）
	mutuals, err := matchRepo.FindMutual(ctx, "user001")
	if err != nil {
		t.Fatalf("相互マッチ取得失敗: %v", err)
	}

	if len(mutuals) != 0 {
		t.Errorf("相互マッチ数 = %d, want 0", len(mutuals))
	}
}

func TestMatchRepository_FindMutual_NoLikes(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	matchRepo := NewMatchRepository(td.DB)
	ctx := context.Background()

	// テストユーザー2人作成
	userRepo := NewUserRepository(td.DB)
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

	// マッチ作成（いいねなし）
	match := &domain.Match{
		ID:        "match001",
		UserIDA:   "user001",
		UserIDB:   "user002",
		Score:     0.855,
		Breakdown: map[string]float64{"age": 0.9},
		MatchedAt: time.Now(),
	}
	if err := matchRepo.Save(ctx, match); err != nil {
		t.Fatalf("マッチ保存失敗: %v", err)
	}

	// 相互マッチ取得（いいねがないので0件）
	mutuals, err := matchRepo.FindMutual(ctx, "user001")
	if err != nil {
		t.Fatalf("相互マッチ取得失敗: %v", err)
	}

	if len(mutuals) != 0 {
		t.Errorf("相互マッチ数 = %d, want 0", len(mutuals))
	}
}

func TestMatchRepository_FindMutual_MultipleMatches(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	matchRepo := NewMatchRepository(td.DB)
	likeRepo := NewLikeRepository(td.DB)
	ctx := context.Background()

	// テストユーザー3人作成
	userRepo := NewUserRepository(td.DB)
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

	// user001とuser002の相互いいね
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
	if err := likeRepo.Save(ctx, like1); err != nil {
		t.Fatalf("いいね1保存失敗: %v", err)
	}
	if err := likeRepo.Save(ctx, like2); err != nil {
		t.Fatalf("いいね2保存失敗: %v", err)
	}

	// user001とuser003の片方向いいね
	like3 := &Like{
		ID:         "like003",
		FromUserID: "user001",
		ToUserID:   "user003",
		CreatedAt:  time.Now(),
	}
	if err := likeRepo.Save(ctx, like3); err != nil {
		t.Fatalf("いいね3保存失敗: %v", err)
	}

	// 2つのマッチ作成
	now := time.Now()
	matches := []*domain.Match{
		{
			ID:        "match001",
			UserIDA:   "user001",
			UserIDB:   "user002",
			Score:     0.855,
			Breakdown: map[string]float64{"age": 0.9},
			MatchedAt: now.Add(-1 * time.Hour),
		},
		{
			ID:        "match002",
			UserIDA:   "user001",
			UserIDB:   "user003",
			Score:     0.78,
			Breakdown: map[string]float64{"age": 0.8},
			MatchedAt: now,
		},
	}
	for _, match := range matches {
		if err := matchRepo.Save(ctx, match); err != nil {
			t.Fatalf("マッチ保存失敗: %v", err)
		}
	}

	// 相互マッチ取得（user002のみ）
	mutuals, err := matchRepo.FindMutual(ctx, "user001")
	if err != nil {
		t.Fatalf("相互マッチ取得失敗: %v", err)
	}

	if len(mutuals) != 1 {
		t.Fatalf("相互マッチ数 = %d, want 1", len(mutuals))
	}

	if mutuals[0].UserIDB != "user002" {
		t.Errorf("UserIDB = %s, want user002", mutuals[0].UserIDB)
	}
}

func TestMatchRepository_FindMutual_EmptyResult(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewMatchRepository(td.DB)
	ctx := context.Background()

	// テストユーザー作成（マッチなし）
	userRepo := NewUserRepository(td.DB)
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

	// 相互マッチがないユーザーで検索
	mutuals, err := repo.FindMutual(ctx, "user001")
	if err != nil {
		t.Fatalf("相互マッチ取得失敗: %v", err)
	}

	if len(mutuals) != 0 {
		t.Errorf("相互マッチ数 = %d, want 0", len(mutuals))
	}
}
