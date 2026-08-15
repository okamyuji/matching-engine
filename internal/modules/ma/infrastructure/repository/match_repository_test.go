package repository

import (
	"context"
	"testing"
	"time"

	"github.com/okamyuji/matching-engine/internal/modules/ma/domain"
	"github.com/okamyuji/matching-engine/internal/testutil"
)

func TestMAMatchRepository_Save_NormalCase(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	matchRepo := NewMAMatchRepository(td.Pool)
	compRepo := NewCompanyRepository(td.Pool)
	intRepo := NewInterestRepository(td.Pool)
	ctx := context.Background()

	// テスト企業2社作成
	companies := []*domain.Company{
		{
			ID:              "company_a",
			Name:            "買い手A",
			Industry:        domain.IndustryTechnology,
			Location:        "東京都",
			EmployeeCount:   200,
			Founded:         time.Date(2010, 5, 15, 0, 0, 0, 0, time.UTC),
			ListingStatus:   domain.ListingPublic,
			MatchingPurpose: domain.PurposeBuyer,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "company_b",
			Name:            "売り手B",
			Industry:        domain.IndustryTechnology,
			Location:        "東京都",
			EmployeeCount:   100,
			Founded:         time.Date(2015, 3, 20, 0, 0, 0, 0, time.UTC),
			ListingStatus:   domain.ListingPrivate,
			MatchingPurpose: domain.PurposeSeller,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	for _, company := range companies {
		if err := compRepo.Create(ctx, company); err != nil {
			t.Fatalf("企業作成失敗: %v", err)
		}
	}

	// 双方向の興味表明作成
	interests := []*domain.Interest{
		{
			ID:            "interest_a_to_b",
			FromCompanyID: "company_a",
			ToCompanyID:   "company_b",
			CreatedAt:     time.Now(),
		},
		{
			ID:            "interest_b_to_a",
			FromCompanyID: "company_b",
			ToCompanyID:   "company_a",
			CreatedAt:     time.Now(),
		},
	}

	for _, interest := range interests {
		if err := intRepo.Save(ctx, interest); err != nil {
			t.Fatalf("興味表明保存失敗: %v", err)
		}
	}

	// マッチ作成
	match := &domain.MAMatch{
		ID:         "match001",
		CompanyIDA: "company_a",
		CompanyIDB: "company_b",
		Score:      0.85,
		Breakdown: map[string]float64{
			"synergy":   0.9,
			"financial": 0.8,
		},
		SynergySummary: &domain.SynergySummary{
			Type:            domain.SynergyHorizontal,
			ExpectedSynergy: 1000000000,
			TechnologyFit:   0.9,
			CustomerFit:     0.85,
			OperationalFit:  0.8,
		},
		MatchedAt: time.Now(),
	}

	err := matchRepo.Save(ctx, match)
	if err != nil {
		t.Fatalf("マッチ保存失敗: %v", err)
	}

	// 保存されたマッチを確認
	found, err := matchRepo.FindByID(ctx, "match001")
	if err != nil {
		t.Fatalf("マッチ取得失敗: %v", err)
	}

	if found.CompanyIDA != "company_a" {
		t.Errorf("CompanyIDA = %s, want company_a", found.CompanyIDA)
	}
	if found.CompanyIDB != "company_b" {
		t.Errorf("CompanyIDB = %s, want company_b", found.CompanyIDB)
	}
	if found.Score != 0.85 {
		t.Errorf("Score = %f, want 0.85", found.Score)
	}
}

func TestMAMatchRepository_Save_DuplicateID(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	matchRepo := NewMAMatchRepository(td.Pool)
	compRepo := NewCompanyRepository(td.Pool)
	intRepo := NewInterestRepository(td.Pool)
	ctx := context.Background()

	// テスト企業2社作成
	companies := []*domain.Company{
		{
			ID:              "company_c",
			Name:            "買い手C",
			Industry:        domain.IndustryTechnology,
			Location:        "東京都",
			EmployeeCount:   200,
			Founded:         time.Date(2010, 5, 15, 0, 0, 0, 0, time.UTC),
			ListingStatus:   domain.ListingPublic,
			MatchingPurpose: domain.PurposeBuyer,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "company_d",
			Name:            "売り手D",
			Industry:        domain.IndustryTechnology,
			Location:        "東京都",
			EmployeeCount:   100,
			Founded:         time.Date(2015, 3, 20, 0, 0, 0, 0, time.UTC),
			ListingStatus:   domain.ListingPrivate,
			MatchingPurpose: domain.PurposeSeller,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	for _, company := range companies {
		if err := compRepo.Create(ctx, company); err != nil {
			t.Fatalf("企業作成失敗: %v", err)
		}
	}

	// 双方向の興味表明作成
	interests := []*domain.Interest{
		{
			ID:            "interest_c_to_d",
			FromCompanyID: "company_c",
			ToCompanyID:   "company_d",
			CreatedAt:     time.Now(),
		},
		{
			ID:            "interest_d_to_c",
			FromCompanyID: "company_d",
			ToCompanyID:   "company_c",
			CreatedAt:     time.Now(),
		},
	}

	for _, interest := range interests {
		if err := intRepo.Save(ctx, interest); err != nil {
			t.Fatalf("興味表明保存失敗: %v", err)
		}
	}

	// 1回目のマッチ保存
	match1 := &domain.MAMatch{
		ID:         "match002",
		CompanyIDA: "company_c",
		CompanyIDB: "company_d",
		Score:      0.85,
		MatchedAt:  time.Now(),
	}
	err := matchRepo.Save(ctx, match1)
	if err != nil {
		t.Fatalf("1回目のマッチ保存失敗: %v", err)
	}

	// 2回目のマッチ保存（同じID）
	match2 := &domain.MAMatch{
		ID:         "match002",
		CompanyIDA: "company_c",
		CompanyIDB: "company_d",
		Score:      0.90,
		MatchedAt:  time.Now(),
	}
	err = matchRepo.Save(ctx, match2)
	if err == nil {
		t.Error("重複IDでエラーが発生すべき")
	}
}

func TestMAMatchRepository_FindByID_NormalCase(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	matchRepo := NewMAMatchRepository(td.Pool)
	compRepo := NewCompanyRepository(td.Pool)
	intRepo := NewInterestRepository(td.Pool)
	ctx := context.Background()

	// テスト企業2社作成
	companies := []*domain.Company{
		{
			ID:              "company_e",
			Name:            "買い手E",
			Industry:        domain.IndustryTechnology,
			Location:        "東京都",
			EmployeeCount:   200,
			Founded:         time.Date(2010, 5, 15, 0, 0, 0, 0, time.UTC),
			ListingStatus:   domain.ListingPublic,
			MatchingPurpose: domain.PurposeBuyer,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "company_f",
			Name:            "売り手F",
			Industry:        domain.IndustryTechnology,
			Location:        "東京都",
			EmployeeCount:   100,
			Founded:         time.Date(2015, 3, 20, 0, 0, 0, 0, time.UTC),
			ListingStatus:   domain.ListingPrivate,
			MatchingPurpose: domain.PurposeSeller,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	for _, company := range companies {
		if err := compRepo.Create(ctx, company); err != nil {
			t.Fatalf("企業作成失敗: %v", err)
		}
	}

	// 双方向の興味表明作成
	interests := []*domain.Interest{
		{
			ID:            "interest_e_to_f",
			FromCompanyID: "company_e",
			ToCompanyID:   "company_f",
			CreatedAt:     time.Now(),
		},
		{
			ID:            "interest_f_to_e",
			FromCompanyID: "company_f",
			ToCompanyID:   "company_e",
			CreatedAt:     time.Now(),
		},
	}

	for _, interest := range interests {
		if err := intRepo.Save(ctx, interest); err != nil {
			t.Fatalf("興味表明保存失敗: %v", err)
		}
	}

	// マッチ作成
	match := &domain.MAMatch{
		ID:         "match003",
		CompanyIDA: "company_e",
		CompanyIDB: "company_f",
		Score:      0.92,
		Breakdown: map[string]float64{
			"synergy":   0.95,
			"financial": 0.89,
		},
		SynergySummary: &domain.SynergySummary{
			Type:            domain.SynergyVertical,
			ExpectedSynergy: 2000000000,
			TechnologyFit:   0.88,
			CustomerFit:     0.92,
			OperationalFit:  0.85,
		},
		MatchedAt: time.Now(),
	}
	err := matchRepo.Save(ctx, match)
	if err != nil {
		t.Fatalf("マッチ保存失敗: %v", err)
	}

	// IDで取得
	found, err := matchRepo.FindByID(ctx, "match003")
	if err != nil {
		t.Fatalf("マッチ取得失敗: %v", err)
	}

	if found.ID != "match003" {
		t.Errorf("ID = %s, want match003", found.ID)
	}
	if found.Score != 0.92 {
		t.Errorf("Score = %f, want 0.92", found.Score)
	}
	if found.SynergySummary.Type != domain.SynergyVertical {
		t.Errorf("SynergyType = %s, want %s", found.SynergySummary.Type, domain.SynergyVertical)
	}
}

func TestMAMatchRepository_FindByID_NotFound(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	matchRepo := NewMAMatchRepository(td.Pool)
	ctx := context.Background()

	// 存在しないIDで検索
	_, err := matchRepo.FindByID(ctx, "nonexistent")
	if err == nil {
		t.Error("存在しないマッチIDでエラーが発生すべき")
	}
}

func TestMAMatchRepository_FindByID_EmptyID(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	matchRepo := NewMAMatchRepository(td.Pool)
	ctx := context.Background()

	// 空のIDで検索
	_, err := matchRepo.FindByID(ctx, "")
	if err == nil {
		t.Error("空のIDでエラーが発生すべき")
	}
}

func TestMAMatchRepository_FindMutual_NormalCase(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	matchRepo := NewMAMatchRepository(td.Pool)
	compRepo := NewCompanyRepository(td.Pool)
	intRepo := NewInterestRepository(td.Pool)
	ctx := context.Background()

	// テスト企業3社作成
	companies := []*domain.Company{
		{
			ID:              "company_g",
			Name:            "買い手G",
			Industry:        domain.IndustryTechnology,
			Location:        "東京都",
			EmployeeCount:   200,
			Founded:         time.Date(2010, 5, 15, 0, 0, 0, 0, time.UTC),
			ListingStatus:   domain.ListingPublic,
			MatchingPurpose: domain.PurposeBuyer,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "company_h",
			Name:            "売り手H",
			Industry:        domain.IndustryTechnology,
			Location:        "東京都",
			EmployeeCount:   100,
			Founded:         time.Date(2015, 3, 20, 0, 0, 0, 0, time.UTC),
			ListingStatus:   domain.ListingPrivate,
			MatchingPurpose: domain.PurposeSeller,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "company_i",
			Name:            "売り手I",
			Industry:        domain.IndustryFinance,
			Location:        "大阪府",
			EmployeeCount:   150,
			Founded:         time.Date(2012, 7, 10, 0, 0, 0, 0, time.UTC),
			ListingStatus:   domain.ListingPrivate,
			MatchingPurpose: domain.PurposeSeller,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	for _, company := range companies {
		if err := compRepo.Create(ctx, company); err != nil {
			t.Fatalf("企業作成失敗: %v", err)
		}
	}

	// 双方向の興味表明作成（G-H, G-I）
	interests := []*domain.Interest{
		{
			ID:            "interest_g_to_h",
			FromCompanyID: "company_g",
			ToCompanyID:   "company_h",
			CreatedAt:     time.Now(),
		},
		{
			ID:            "interest_h_to_g",
			FromCompanyID: "company_h",
			ToCompanyID:   "company_g",
			CreatedAt:     time.Now(),
		},
		{
			ID:            "interest_g_to_i",
			FromCompanyID: "company_g",
			ToCompanyID:   "company_i",
			CreatedAt:     time.Now(),
		},
		{
			ID:            "interest_i_to_g",
			FromCompanyID: "company_i",
			ToCompanyID:   "company_g",
			CreatedAt:     time.Now(),
		},
	}

	for _, interest := range interests {
		if err := intRepo.Save(ctx, interest); err != nil {
			t.Fatalf("興味表明保存失敗: %v", err)
		}
	}

	// マッチ作成（G-H, G-I）
	now := time.Now()
	matches := []*domain.MAMatch{
		{
			ID:         "match_g_h",
			CompanyIDA: "company_g",
			CompanyIDB: "company_h",
			Score:      0.85,
			MatchedAt:  now.Add(-1 * time.Hour),
		},
		{
			ID:         "match_g_i",
			CompanyIDA: "company_g",
			CompanyIDB: "company_i",
			Score:      0.78,
			MatchedAt:  now,
		},
	}

	for _, match := range matches {
		if err := matchRepo.Save(ctx, match); err != nil {
			t.Fatalf("マッチ保存失敗: %v", err)
		}
	}

	// company_gの相互マッチを取得
	found, err := matchRepo.FindMutual(ctx, "company_g")
	if err != nil {
		t.Fatalf("相互マッチ取得失敗: %v", err)
	}

	if len(found) != 2 {
		t.Fatalf("相互マッチ数 = %d, want 2", len(found))
	}

	// 新しい順にソートされているか確認
	if found[0].ID != "match_g_i" {
		t.Errorf("1番目のマッチID = %s, want match_g_i", found[0].ID)
	}
	if found[1].ID != "match_g_h" {
		t.Errorf("2番目のマッチID = %s, want match_g_h", found[1].ID)
	}
}

func TestMAMatchRepository_FindMutual_OneWayInterest(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	matchRepo := NewMAMatchRepository(td.Pool)
	compRepo := NewCompanyRepository(td.Pool)
	intRepo := NewInterestRepository(td.Pool)
	ctx := context.Background()

	// テスト企業2社作成
	companies := []*domain.Company{
		{
			ID:              "company_j",
			Name:            "買い手J",
			Industry:        domain.IndustryTechnology,
			Location:        "東京都",
			EmployeeCount:   200,
			Founded:         time.Date(2010, 5, 15, 0, 0, 0, 0, time.UTC),
			ListingStatus:   domain.ListingPublic,
			MatchingPurpose: domain.PurposeBuyer,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "company_k",
			Name:            "売り手K",
			Industry:        domain.IndustryTechnology,
			Location:        "東京都",
			EmployeeCount:   100,
			Founded:         time.Date(2015, 3, 20, 0, 0, 0, 0, time.UTC),
			ListingStatus:   domain.ListingPrivate,
			MatchingPurpose: domain.PurposeSeller,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	for _, company := range companies {
		if err := compRepo.Create(ctx, company); err != nil {
			t.Fatalf("企業作成失敗: %v", err)
		}
	}

	// 片方向の興味表明のみ作成（J→K）
	interest := &domain.Interest{
		ID:            "interest_j_to_k",
		FromCompanyID: "company_j",
		ToCompanyID:   "company_k",
		CreatedAt:     time.Now(),
	}
	if err := intRepo.Save(ctx, interest); err != nil {
		t.Fatalf("興味表明保存失敗: %v", err)
	}

	// マッチ作成（双方向興味がないのでFindMutualでは取得されない）
	match := &domain.MAMatch{
		ID:         "match_j_k",
		CompanyIDA: "company_j",
		CompanyIDB: "company_k",
		Score:      0.80,
		MatchedAt:  time.Now(),
	}
	err := matchRepo.Save(ctx, match)
	if err != nil {
		t.Fatalf("マッチ保存失敗: %v", err)
	}

	// company_jの相互マッチを取得（片方向なので0件）
	found, err := matchRepo.FindMutual(ctx, "company_j")
	if err != nil {
		t.Fatalf("相互マッチ取得失敗: %v", err)
	}

	if len(found) != 0 {
		t.Errorf("相互マッチ数 = %d, want 0 (片方向興味のみ)", len(found))
	}
}

func TestMAMatchRepository_FindMutual_EmptyResult(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	matchRepo := NewMAMatchRepository(td.Pool)
	compRepo := NewCompanyRepository(td.Pool)
	ctx := context.Background()

	// テスト企業作成（マッチなし）
	company := &domain.Company{
		ID:              "company_l",
		Name:            "買い手L",
		Industry:        domain.IndustryTechnology,
		Location:        "東京都",
		EmployeeCount:   200,
		Founded:         time.Date(2010, 5, 15, 0, 0, 0, 0, time.UTC),
		ListingStatus:   domain.ListingPublic,
		MatchingPurpose: domain.PurposeBuyer,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := compRepo.Create(ctx, company); err != nil {
		t.Fatalf("企業作成失敗: %v", err)
	}

	// マッチなしで検索
	found, err := matchRepo.FindMutual(ctx, "company_l")
	if err != nil {
		t.Fatalf("相互マッチ取得失敗: %v", err)
	}

	if len(found) != 0 {
		t.Errorf("相互マッチ数 = %d, want 0", len(found))
	}
}

func TestMAMatchRepository_FindMutual_EmptyCompanyID(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	matchRepo := NewMAMatchRepository(td.Pool)
	ctx := context.Background()

	// 空のCompanyIDで検索
	found, err := matchRepo.FindMutual(ctx, "")
	if err != nil {
		t.Fatalf("相互マッチ取得失敗: %v", err)
	}

	if len(found) != 0 {
		t.Errorf("相互マッチ数 = %d, want 0", len(found))
	}
}
