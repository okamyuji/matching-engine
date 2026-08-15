package repository

import (
	"context"
	"testing"
	"time"

	"github.com/okamyuji/matching-engine/internal/modules/ma/domain"
	"github.com/okamyuji/matching-engine/internal/testutil"
)

func TestCompanyRepository_Create_NormalCase(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewCompanyRepository(td.Pool)
	ctx := context.Background()

	// テスト企業作成
	company := &domain.Company{
		ID:              "company001",
		Name:            "テスト株式会社",
		Industry:        domain.IndustryTechnology,
		Location:        "東京都",
		EmployeeCount:   100,
		Founded:         time.Date(2010, 5, 15, 0, 0, 0, 0, time.UTC),
		ListingStatus:   domain.ListingPrivate,
		MatchingPurpose: domain.PurposeBuyer,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err := repo.Create(ctx, company)
	if err != nil {
		t.Fatalf("企業作成失敗: %v", err)
	}

	// 作成された企業を取得して確認
	found, err := repo.FindByID(ctx, "company001")
	if err != nil {
		t.Fatalf("企業取得失敗: %v", err)
	}

	if found.ID != company.ID {
		t.Errorf("ID = %v, want %v", found.ID, company.ID)
	}
	if found.Name != company.Name {
		t.Errorf("Name = %v, want %v", found.Name, company.Name)
	}
	if found.Industry != company.Industry {
		t.Errorf("Industry = %v, want %v", found.Industry, company.Industry)
	}
	if found.EmployeeCount != company.EmployeeCount {
		t.Errorf("EmployeeCount = %v, want %v", found.EmployeeCount, company.EmployeeCount)
	}
}

func TestCompanyRepository_Create_DuplicateID(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewCompanyRepository(td.Pool)
	ctx := context.Background()

	// 1回目の企業作成
	company1 := &domain.Company{
		ID:              "company001",
		Name:            "テスト株式会社",
		Industry:        domain.IndustryTechnology,
		Location:        "東京都",
		EmployeeCount:   100,
		Founded:         time.Date(2010, 5, 15, 0, 0, 0, 0, time.UTC),
		ListingStatus:   domain.ListingPrivate,
		MatchingPurpose: domain.PurposeBuyer,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err := repo.Create(ctx, company1)
	if err != nil {
		t.Fatalf("1回目の企業作成失敗: %v", err)
	}

	// 2回目の企業作成（同じID）
	company2 := &domain.Company{
		ID:              "company001",
		Name:            "別の会社",
		Industry:        domain.IndustryFinance,
		Location:        "大阪府",
		EmployeeCount:   200,
		Founded:         time.Date(2015, 3, 20, 0, 0, 0, 0, time.UTC),
		ListingStatus:   domain.ListingPublic,
		MatchingPurpose: domain.PurposeSeller,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err = repo.Create(ctx, company2)
	if err == nil {
		t.Error("重複IDでエラーが発生すべき")
	}
}

func TestCompanyRepository_FindByID_NormalCase(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewCompanyRepository(td.Pool)
	ctx := context.Background()

	// テスト企業作成
	company := &domain.Company{
		ID:              "company002",
		Name:            "検索テスト株式会社",
		Industry:        domain.IndustryFinance,
		Location:        "大阪府",
		EmployeeCount:   200,
		Founded:         time.Date(2015, 3, 20, 0, 0, 0, 0, time.UTC),
		ListingStatus:   domain.ListingPublic,
		MatchingPurpose: domain.PurposeSeller,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err := repo.Create(ctx, company)
	if err != nil {
		t.Fatalf("企業作成失敗: %v", err)
	}

	// Technologies追加
	tech := &domain.CompanyTechnology{
		CompanyID:  "company002",
		Technology: "AI",
	}
	err = InsertTechnology(ctx, td.Pool, tech)
	if err != nil {
		t.Fatalf("Technology追加失敗: %v", err)
	}

	// Markets追加
	market := &domain.CompanyMarket{
		CompanyID: "company002",
		Market:    "国内",
	}
	err = InsertMarket(ctx, td.Pool, market)
	if err != nil {
		t.Fatalf("Market追加失敗: %v", err)
	}

	// FindByIDテスト
	found, err := repo.FindByID(ctx, "company002")
	if err != nil {
		t.Fatalf("企業取得失敗: %v", err)
	}

	if found.ID != company.ID {
		t.Errorf("ID = %v, want %v", found.ID, company.ID)
	}
	if len(found.Technologies) != 1 {
		t.Errorf("Technologies count = %d, want 1", len(found.Technologies))
	}
	if len(found.Markets) != 1 {
		t.Errorf("Markets count = %d, want 1", len(found.Markets))
	}
}

func TestCompanyRepository_FindByID_NotFound(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewCompanyRepository(td.Pool)
	ctx := context.Background()

	// 存在しないIDで検索
	_, err := repo.FindByID(ctx, "nonexistent")
	if err == nil {
		t.Error("存在しない企業でエラーが発生すべき")
	}
}

func TestCompanyRepository_FindByID_EmptyID(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewCompanyRepository(td.Pool)
	ctx := context.Background()

	// 空のIDで検索
	_, err := repo.FindByID(ctx, "")
	if err == nil {
		t.Error("空のIDでエラーが発生すべき")
	}
}

func TestCompanyRepository_Update_NormalCase(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewCompanyRepository(td.Pool)
	ctx := context.Background()

	// テスト企業作成
	company := &domain.Company{
		ID:              "company003",
		Name:            "更新前の会社名",
		Industry:        domain.IndustryTechnology,
		Location:        "東京都",
		EmployeeCount:   100,
		Founded:         time.Date(2010, 5, 15, 0, 0, 0, 0, time.UTC),
		ListingStatus:   domain.ListingPrivate,
		MatchingPurpose: domain.PurposeBuyer,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err := repo.Create(ctx, company)
	if err != nil {
		t.Fatalf("企業作成失敗: %v", err)
	}

	// 企業情報更新
	company.Name = "更新後の会社名"
	company.EmployeeCount = 150
	company.Location = "神奈川県"
	err = repo.Update(ctx, company)
	if err != nil {
		t.Fatalf("企業更新失敗: %v", err)
	}

	// 更新を確認
	found, err := repo.FindByID(ctx, "company003")
	if err != nil {
		t.Fatalf("企業取得失敗: %v", err)
	}

	if found.Name != "更新後の会社名" {
		t.Errorf("Name = %v, want 更新後の会社名", found.Name)
	}
	if found.EmployeeCount != 150 {
		t.Errorf("EmployeeCount = %v, want 150", found.EmployeeCount)
	}
	if found.Location != "神奈川県" {
		t.Errorf("Location = %v, want 神奈川県", found.Location)
	}
}

func TestCompanyRepository_Update_NonExistentCompany(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewCompanyRepository(td.Pool)
	ctx := context.Background()

	// 存在しない企業を更新
	company := &domain.Company{
		ID:              "nonexistent",
		Name:            "存在しない会社",
		Industry:        domain.IndustryTechnology,
		Location:        "東京都",
		EmployeeCount:   100,
		Founded:         time.Date(2010, 5, 15, 0, 0, 0, 0, time.UTC),
		ListingStatus:   domain.ListingPrivate,
		MatchingPurpose: domain.PurposeBuyer,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err := repo.Update(ctx, company)
	// Updateは存在しない場合でもエラーにならない（影響行数0）
	if err != nil {
		t.Errorf("Update should not error for non-existent company: %v", err)
	}
}

func TestCompanyRepository_FindByPurpose_NormalCase(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewCompanyRepository(td.Pool)
	ctx := context.Background()

	// 買い手企業を2社作成
	buyers := []*domain.Company{
		{
			ID:              "buyer001",
			Name:            "買い手1",
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
			ID:              "buyer002",
			Name:            "買い手2",
			Industry:        domain.IndustryFinance,
			Location:        "大阪府",
			EmployeeCount:   300,
			Founded:         time.Date(2015, 3, 20, 0, 0, 0, 0, time.UTC),
			ListingStatus:   domain.ListingPrivate,
			MatchingPurpose: domain.PurposeBuyer,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	// 売り手企業を1社作成
	seller := &domain.Company{
		ID:              "seller001",
		Name:            "売り手1",
		Industry:        domain.IndustryTechnology,
		Location:        "東京都",
		EmployeeCount:   100,
		Founded:         time.Date(2018, 7, 10, 0, 0, 0, 0, time.UTC),
		ListingStatus:   domain.ListingPrivate,
		MatchingPurpose: domain.PurposeSeller,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	for _, buyer := range buyers {
		if err := repo.Create(ctx, buyer); err != nil {
			t.Fatalf("買い手企業作成失敗: %v", err)
		}
	}
	if err := repo.Create(ctx, seller); err != nil {
		t.Fatalf("売り手企業作成失敗: %v", err)
	}

	// 買い手企業を検索（criteriaなし）
	results, err := repo.FindByPurpose(ctx, domain.PurposeBuyer, nil)
	if err != nil {
		t.Fatalf("候補検索失敗: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("候補数 = %d, want 2", len(results))
	}
}

func TestCompanyRepository_FindByPurpose_WithIndustryFilter(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewCompanyRepository(td.Pool)
	ctx := context.Background()

	// Technology業界の買い手を2社、Finance業界の買い手を1社作成
	companies := []*domain.Company{
		{
			ID:              "tech_buyer001",
			Name:            "Tech買い手1",
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
			ID:              "tech_buyer002",
			Name:            "Tech買い手2",
			Industry:        domain.IndustryTechnology,
			Location:        "大阪府",
			EmployeeCount:   300,
			Founded:         time.Date(2015, 3, 20, 0, 0, 0, 0, time.UTC),
			ListingStatus:   domain.ListingPrivate,
			MatchingPurpose: domain.PurposeBuyer,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "finance_buyer001",
			Name:            "Finance買い手1",
			Industry:        domain.IndustryFinance,
			Location:        "東京都",
			EmployeeCount:   250,
			Founded:         time.Date(2012, 8, 1, 0, 0, 0, 0, time.UTC),
			ListingStatus:   domain.ListingPublic,
			MatchingPurpose: domain.PurposeBuyer,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	for _, company := range companies {
		if err := repo.Create(ctx, company); err != nil {
			t.Fatalf("企業作成失敗: %v", err)
		}
	}

	// Technology業界のみの候補を検索
	criteria := &domain.MAMatchingCriteria{
		CompanyID: "test_company",
		Purpose:   domain.PurposeSeller,
		TargetIndustries: []domain.CriteriaIndustry{
			{Industry: domain.IndustryTechnology},
		},
		UpdatedAt: time.Now(),
	}

	results, err := repo.FindByPurpose(ctx, domain.PurposeBuyer, criteria)
	if err != nil {
		t.Fatalf("候補検索失敗: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("候補数 = %d, want 2 (Technology業界のみ)", len(results))
	}

	// 全てTechnology業界であることを確認
	for _, result := range results {
		if result.Company.Industry != domain.IndustryTechnology {
			t.Errorf("Industry = %v, want technology", result.Company.Industry)
		}
	}
}

func TestCompanyRepository_FindByPurpose_WithEmployeeFilter(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewCompanyRepository(td.Pool)
	ctx := context.Background()

	// 従業員数が異なる買い手を3社作成
	companies := []*domain.Company{
		{
			ID:              "small_buyer",
			Name:            "小規模買い手",
			Industry:        domain.IndustryTechnology,
			Location:        "東京都",
			EmployeeCount:   50,
			Founded:         time.Date(2010, 5, 15, 0, 0, 0, 0, time.UTC),
			ListingStatus:   domain.ListingPrivate,
			MatchingPurpose: domain.PurposeBuyer,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "medium_buyer",
			Name:            "中規模買い手",
			Industry:        domain.IndustryTechnology,
			Location:        "東京都",
			EmployeeCount:   200,
			Founded:         time.Date(2015, 3, 20, 0, 0, 0, 0, time.UTC),
			ListingStatus:   domain.ListingPrivate,
			MatchingPurpose: domain.PurposeBuyer,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "large_buyer",
			Name:            "大規模買い手",
			Industry:        domain.IndustryTechnology,
			Location:        "東京都",
			EmployeeCount:   500,
			Founded:         time.Date(2018, 7, 10, 0, 0, 0, 0, time.UTC),
			ListingStatus:   domain.ListingPublic,
			MatchingPurpose: domain.PurposeBuyer,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	for _, company := range companies {
		if err := repo.Create(ctx, company); err != nil {
			t.Fatalf("企業作成失敗: %v", err)
		}
	}

	// 従業員数100〜300の候補を検索
	criteria := &domain.MAMatchingCriteria{
		CompanyID:   "test_company",
		Purpose:     domain.PurposeSeller,
		EmployeeMin: 100,
		EmployeeMax: 300,
		UpdatedAt:   time.Now(),
	}

	results, err := repo.FindByPurpose(ctx, domain.PurposeBuyer, criteria)
	if err != nil {
		t.Fatalf("候補検索失敗: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("候補数 = %d, want 1", len(results))
	}

	if results[0].Company.EmployeeCount != 200 {
		t.Errorf("EmployeeCount = %d, want 200", results[0].Company.EmployeeCount)
	}
}

func TestCompanyRepository_FindByPurpose_EmptyResult(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewCompanyRepository(td.Pool)
	ctx := context.Background()

	// 売り手企業のみ作成
	seller := &domain.Company{
		ID:              "seller001",
		Name:            "売り手1",
		Industry:        domain.IndustryTechnology,
		Location:        "東京都",
		EmployeeCount:   100,
		Founded:         time.Date(2018, 7, 10, 0, 0, 0, 0, time.UTC),
		ListingStatus:   domain.ListingPrivate,
		MatchingPurpose: domain.PurposeSeller,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := repo.Create(ctx, seller); err != nil {
		t.Fatalf("売り手企業作成失敗: %v", err)
	}

	// 買い手企業を検索（存在しない）
	results, err := repo.FindByPurpose(ctx, domain.PurposeBuyer, nil)
	if err != nil {
		t.Fatalf("候補検索失敗: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("候補数 = %d, want 0", len(results))
	}
}

func TestCompanyRepository_FindCriteria_NormalCase(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewCompanyRepository(td.Pool)
	ctx := context.Background()

	// テスト企業作成
	company := &domain.Company{
		ID:              "company004",
		Name:            "基準テスト株式会社",
		Industry:        domain.IndustryTechnology,
		Location:        "東京都",
		EmployeeCount:   100,
		Founded:         time.Date(2010, 5, 15, 0, 0, 0, 0, time.UTC),
		ListingStatus:   domain.ListingPrivate,
		MatchingPurpose: domain.PurposeBuyer,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err := repo.Create(ctx, company)
	if err != nil {
		t.Fatalf("企業作成失敗: %v", err)
	}

	// マッチング基準作成
	criteria := &domain.MAMatchingCriteria{
		CompanyID:          "company004",
		Purpose:            domain.PurposeBuyer,
		RevenueMin:         1000000000,
		RevenueMax:         10000000000,
		EBITDAMin:          100000000,
		EmployeeMin:        50,
		EmployeeMax:        500,
		MaxDebtEquityRatio: 1.5,
		UpdatedAt:          time.Now(),
	}
	err = UpsertCriteria(ctx, td.Pool, criteria)
	if err != nil {
		t.Fatalf("基準作成失敗: %v", err)
	}

	// TargetIndustries追加
	industry := &domain.CriteriaIndustry{
		CompanyID: "company004",
		Industry:  domain.IndustryTechnology,
	}
	err = InsertCriteriaIndustry(ctx, td.Pool, industry)
	if err != nil {
		t.Fatalf("TargetIndustry追加失敗: %v", err)
	}

	// FindCriteriaテスト
	found, err := repo.FindCriteria(ctx, "company004")
	if err != nil {
		t.Fatalf("基準取得失敗: %v", err)
	}

	if found.CompanyID != "company004" {
		t.Errorf("CompanyID = %v, want company004", found.CompanyID)
	}
	if found.RevenueMin != 1000000000 {
		t.Errorf("RevenueMin = %v, want 1000000000", found.RevenueMin)
	}
	if len(found.TargetIndustries) != 1 {
		t.Errorf("TargetIndustries count = %d, want 1", len(found.TargetIndustries))
	}
}

func TestCompanyRepository_FindCriteria_NotFound(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewCompanyRepository(td.Pool)
	ctx := context.Background()

	// 存在しない企業IDで基準を検索
	_, err := repo.FindCriteria(ctx, "nonexistent")
	if err == nil {
		t.Error("存在しない基準でエラーが発生すべき")
	}
}

func TestCompanyRepository_FindCriteria_EmptyCompanyID(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	repo := NewCompanyRepository(td.Pool)
	ctx := context.Background()

	// 空のCompanyIDで基準を検索
	_, err := repo.FindCriteria(ctx, "")
	if err == nil {
		t.Error("空のCompanyIDでエラーが発生すべき")
	}
}
