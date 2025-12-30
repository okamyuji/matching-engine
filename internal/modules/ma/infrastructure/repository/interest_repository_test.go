package repository

import (
	"context"
	"testing"
	"time"

	"github.com/yourorg/matching-engine/internal/modules/ma/domain"
	"github.com/yourorg/matching-engine/internal/testutil"
)

func TestInterestRepository_Save_NormalCase(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	intRepo := NewInterestRepository(td.DB)
	compRepo := NewCompanyRepository(td.DB)
	ctx := context.Background()

	// テスト企業2社作成
	companies := []*domain.Company{
		{
			ID:              "buyer001",
			Name:            "買い手企業",
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
			ID:              "seller001",
			Name:            "売り手企業",
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

	// 興味表明作成
	interest := &domain.Interest{
		ID:            "interest001",
		FromCompanyID: "buyer001",
		ToCompanyID:   "seller001",
		CreatedAt:     time.Now(),
	}

	err := intRepo.Save(ctx, interest)
	if err != nil {
		t.Fatalf("興味表明保存失敗: %v", err)
	}

	// 保存された興味表明を確認
	interests, err := intRepo.FindByToCompany(ctx, "seller001")
	if err != nil {
		t.Fatalf("興味表明取得失敗: %v", err)
	}

	if len(interests) != 1 {
		t.Errorf("興味表明数 = %d, want 1", len(interests))
	}
	if interests[0].FromCompanyID != "buyer001" {
		t.Errorf("FromCompanyID = %s, want buyer001", interests[0].FromCompanyID)
	}
}

func TestInterestRepository_Save_DuplicateInterest(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	intRepo := NewInterestRepository(td.DB)
	compRepo := NewCompanyRepository(td.DB)
	ctx := context.Background()

	// テスト企業2社作成
	companies := []*domain.Company{
		{
			ID:              "buyer002",
			Name:            "買い手企業2",
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
			ID:              "seller002",
			Name:            "売り手企業2",
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

	// 1回目の興味表明
	interest1 := &domain.Interest{
		ID:            "interest002",
		FromCompanyID: "buyer002",
		ToCompanyID:   "seller002",
		CreatedAt:     time.Now(),
	}
	err := intRepo.Save(ctx, interest1)
	if err != nil {
		t.Fatalf("1回目の興味表明保存失敗: %v", err)
	}

	// 2回目の興味表明（同じID）
	interest2 := &domain.Interest{
		ID:            "interest002",
		FromCompanyID: "buyer002",
		ToCompanyID:   "seller002",
		CreatedAt:     time.Now(),
	}
	err = intRepo.Save(ctx, interest2)
	if err == nil {
		t.Error("重複IDでエラーが発生すべき")
	}
}

func TestInterestRepository_FindByToCompany_MultipleInterests(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	intRepo := NewInterestRepository(td.DB)
	compRepo := NewCompanyRepository(td.DB)
	ctx := context.Background()

	// テスト企業3社作成（1社の売り手に対して2社の買い手が興味表明）
	companies := []*domain.Company{
		{
			ID:              "buyer003",
			Name:            "買い手企業3",
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
			ID:              "buyer004",
			Name:            "買い手企業4",
			Industry:        domain.IndustryFinance,
			Location:        "大阪府",
			EmployeeCount:   300,
			Founded:         time.Date(2012, 7, 10, 0, 0, 0, 0, time.UTC),
			ListingStatus:   domain.ListingPublic,
			MatchingPurpose: domain.PurposeBuyer,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "seller003",
			Name:            "売り手企業3",
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

	// 2社から興味表明
	now := time.Now()
	interests := []*domain.Interest{
		{
			ID:            "interest003",
			FromCompanyID: "buyer003",
			ToCompanyID:   "seller003",
			CreatedAt:     now.Add(-1 * time.Hour),
		},
		{
			ID:            "interest004",
			FromCompanyID: "buyer004",
			ToCompanyID:   "seller003",
			CreatedAt:     now,
		},
	}

	for _, interest := range interests {
		if err := intRepo.Save(ctx, interest); err != nil {
			t.Fatalf("興味表明保存失敗: %v", err)
		}
	}

	// seller003が受け取った興味表明を取得
	found, err := intRepo.FindByToCompany(ctx, "seller003")
	if err != nil {
		t.Fatalf("興味表明取得失敗: %v", err)
	}

	if len(found) != 2 {
		t.Fatalf("興味表明数 = %d, want 2", len(found))
	}

	// 新しい順にソートされているか確認
	if found[0].FromCompanyID != "buyer004" {
		t.Errorf("1番目のFromCompanyID = %s, want buyer004", found[0].FromCompanyID)
	}
	if found[1].FromCompanyID != "buyer003" {
		t.Errorf("2番目のFromCompanyID = %s, want buyer003", found[1].FromCompanyID)
	}
}

func TestInterestRepository_FindByToCompany_EmptyResult(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	intRepo := NewInterestRepository(td.DB)
	compRepo := NewCompanyRepository(td.DB)
	ctx := context.Background()

	// テスト企業作成（興味表明を受けていない）
	company := &domain.Company{
		ID:              "seller004",
		Name:            "売り手企業4",
		Industry:        domain.IndustryTechnology,
		Location:        "東京都",
		EmployeeCount:   100,
		Founded:         time.Date(2015, 3, 20, 0, 0, 0, 0, time.UTC),
		ListingStatus:   domain.ListingPrivate,
		MatchingPurpose: domain.PurposeSeller,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := compRepo.Create(ctx, company); err != nil {
		t.Fatalf("企業作成失敗: %v", err)
	}

	// 興味表明を受けていない企業で検索
	interests, err := intRepo.FindByToCompany(ctx, "seller004")
	if err != nil {
		t.Fatalf("興味表明取得失敗: %v", err)
	}

	if len(interests) != 0 {
		t.Errorf("興味表明数 = %d, want 0", len(interests))
	}
}

func TestInterestRepository_FindByToCompany_EmptyCompanyID(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	intRepo := NewInterestRepository(td.DB)
	ctx := context.Background()

	// 空のCompanyIDで検索
	interests, err := intRepo.FindByToCompany(ctx, "")
	if err != nil {
		t.Fatalf("興味表明取得失敗: %v", err)
	}

	if len(interests) != 0 {
		t.Errorf("興味表明数 = %d, want 0", len(interests))
	}
}

func TestInterestRepository_Exists_True(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	intRepo := NewInterestRepository(td.DB)
	compRepo := NewCompanyRepository(td.DB)
	ctx := context.Background()

	// テスト企業2社作成
	companies := []*domain.Company{
		{
			ID:              "buyer005",
			Name:            "買い手企業5",
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
			ID:              "seller005",
			Name:            "売り手企業5",
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

	// 興味表明作成
	interest := &domain.Interest{
		ID:            "interest005",
		FromCompanyID: "buyer005",
		ToCompanyID:   "seller005",
		CreatedAt:     time.Now(),
	}
	err := intRepo.Save(ctx, interest)
	if err != nil {
		t.Fatalf("興味表明保存失敗: %v", err)
	}

	// 存在チェック
	exists, err := intRepo.Exists(ctx, "buyer005", "seller005")
	if err != nil {
		t.Fatalf("存在チェック失敗: %v", err)
	}

	if !exists {
		t.Error("興味表明が存在すべき")
	}
}

func TestInterestRepository_Exists_False(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	intRepo := NewInterestRepository(td.DB)
	compRepo := NewCompanyRepository(td.DB)
	ctx := context.Background()

	// テスト企業2社作成
	companies := []*domain.Company{
		{
			ID:              "buyer006",
			Name:            "買い手企業6",
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
			ID:              "seller006",
			Name:            "売り手企業6",
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

	// 興味表明なしで存在チェック
	exists, err := intRepo.Exists(ctx, "buyer006", "seller006")
	if err != nil {
		t.Fatalf("存在チェック失敗: %v", err)
	}

	if exists {
		t.Error("興味表明が存在しないべき")
	}
}

func TestInterestRepository_Exists_EmptyCompanyID(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	intRepo := NewInterestRepository(td.DB)
	ctx := context.Background()

	// 空のCompanyIDで存在チェック
	exists, err := intRepo.Exists(ctx, "", "")
	if err != nil {
		t.Fatalf("存在チェック失敗: %v", err)
	}

	if exists {
		t.Error("空のIDで興味表明が存在しないべき")
	}
}
