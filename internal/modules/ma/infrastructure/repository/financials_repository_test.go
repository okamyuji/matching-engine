package repository

import (
	"context"
	"testing"
	"time"

	"github.com/okamyuji/matching-engine/internal/modules/ma/domain"
	"github.com/okamyuji/matching-engine/internal/testutil"
)

func TestFinancialsRepository_Save_Insert(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	finRepo := NewFinancialsRepository(td.Pool)
	compRepo := NewCompanyRepository(td.Pool)
	ctx := context.Background()

	// テスト企業作成
	company := &domain.Company{
		ID:              "company001",
		Name:            "財務テスト株式会社",
		Industry:        domain.IndustryTechnology,
		Location:        "東京都",
		EmployeeCount:   100,
		Founded:         time.Date(2010, 5, 15, 0, 0, 0, 0, time.UTC),
		ListingStatus:   domain.ListingPrivate,
		MatchingPurpose: domain.PurposeBuyer,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err := compRepo.Create(ctx, company)
	if err != nil {
		t.Fatalf("企業作成失敗: %v", err)
	}

	// 財務情報作成
	financials := &domain.Financials{
		CompanyID:        "company001",
		FiscalYear:       2023,
		Revenue:          1000000000,
		EBITDA:           200000000,
		NetIncome:        100000000,
		TotalAssets:      2000000000,
		TotalLiabilities: 800000000,
		Equity:           1200000000,
		ROE:              0.083,
		ROA:              0.05,
		DebtEquityRatio:  0.67,
		CurrentRatio:     1.5,
	}

	err = finRepo.Save(ctx, financials)
	if err != nil {
		t.Fatalf("財務情報保存失敗: %v", err)
	}

	// 保存された財務情報を取得して確認
	found, err := finRepo.FindLatest(ctx, "company001")
	if err != nil {
		t.Fatalf("財務情報取得失敗: %v", err)
	}

	if found.CompanyID != financials.CompanyID {
		t.Errorf("CompanyID = %v, want %v", found.CompanyID, financials.CompanyID)
	}
	if found.FiscalYear != financials.FiscalYear {
		t.Errorf("FiscalYear = %v, want %v", found.FiscalYear, financials.FiscalYear)
	}
	if found.Revenue != financials.Revenue {
		t.Errorf("Revenue = %v, want %v", found.Revenue, financials.Revenue)
	}
	if found.EBITDA != financials.EBITDA {
		t.Errorf("EBITDA = %v, want %v", found.EBITDA, financials.EBITDA)
	}
}

func TestFinancialsRepository_Save_Update(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	finRepo := NewFinancialsRepository(td.Pool)
	compRepo := NewCompanyRepository(td.Pool)
	ctx := context.Background()

	// テスト企業作成
	company := &domain.Company{
		ID:              "company002",
		Name:            "更新テスト株式会社",
		Industry:        domain.IndustryTechnology,
		Location:        "東京都",
		EmployeeCount:   100,
		Founded:         time.Date(2010, 5, 15, 0, 0, 0, 0, time.UTC),
		ListingStatus:   domain.ListingPrivate,
		MatchingPurpose: domain.PurposeBuyer,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err := compRepo.Create(ctx, company)
	if err != nil {
		t.Fatalf("企業作成失敗: %v", err)
	}

	// 初回の財務情報保存
	financials := &domain.Financials{
		CompanyID:        "company002",
		FiscalYear:       2023,
		Revenue:          1000000000,
		EBITDA:           200000000,
		NetIncome:        100000000,
		TotalAssets:      2000000000,
		TotalLiabilities: 800000000,
		Equity:           1200000000,
		ROE:              0.083,
		ROA:              0.05,
		DebtEquityRatio:  0.67,
		CurrentRatio:     1.5,
	}
	err = finRepo.Save(ctx, financials)
	if err != nil {
		t.Fatalf("初回保存失敗: %v", err)
	}

	// 財務情報更新
	financials.Revenue = 1200000000
	financials.EBITDA = 250000000
	financials.NetIncome = 120000000
	err = finRepo.Save(ctx, financials)
	if err != nil {
		t.Fatalf("更新保存失敗: %v", err)
	}

	// 更新を確認
	found, err := finRepo.FindLatest(ctx, "company002")
	if err != nil {
		t.Fatalf("財務情報取得失敗: %v", err)
	}

	if found.Revenue != 1200000000 {
		t.Errorf("Revenue = %v, want 1200000000", found.Revenue)
	}
	if found.EBITDA != 250000000 {
		t.Errorf("EBITDA = %v, want 250000000", found.EBITDA)
	}
	if found.NetIncome != 120000000 {
		t.Errorf("NetIncome = %v, want 120000000", found.NetIncome)
	}
}

func TestFinancialsRepository_FindByCompanyID_MultipleYears(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	finRepo := NewFinancialsRepository(td.Pool)
	compRepo := NewCompanyRepository(td.Pool)
	ctx := context.Background()

	// テスト企業作成
	company := &domain.Company{
		ID:              "company003",
		Name:            "複数年テスト株式会社",
		Industry:        domain.IndustryTechnology,
		Location:        "東京都",
		EmployeeCount:   100,
		Founded:         time.Date(2010, 5, 15, 0, 0, 0, 0, time.UTC),
		ListingStatus:   domain.ListingPrivate,
		MatchingPurpose: domain.PurposeBuyer,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err := compRepo.Create(ctx, company)
	if err != nil {
		t.Fatalf("企業作成失敗: %v", err)
	}

	// 5年分の財務情報作成
	for year := 2019; year <= 2023; year++ {
		financials := &domain.Financials{
			CompanyID:        "company003",
			FiscalYear:       year,
			Revenue:          int64(year * 100000000),
			EBITDA:           int64(year * 20000000),
			NetIncome:        int64(year * 10000000),
			TotalAssets:      int64(year * 200000000),
			TotalLiabilities: int64(year * 80000000),
			Equity:           int64(year * 120000000),
			ROE:              0.083,
			ROA:              0.05,
			DebtEquityRatio:  0.67,
			CurrentRatio:     1.5,
		}
		err = finRepo.Save(ctx, financials)
		if err != nil {
			t.Fatalf("財務情報保存失敗 (%d年度): %v", year, err)
		}
	}

	// 直近3年分を取得
	found, err := finRepo.FindByCompanyID(ctx, "company003", 3)
	if err != nil {
		t.Fatalf("財務情報取得失敗: %v", err)
	}

	if len(found) != 3 {
		t.Fatalf("取得件数 = %d, want 3", len(found))
	}

	// 新しい順にソートされているか確認
	if found[0].FiscalYear != 2023 {
		t.Errorf("1番目の年度 = %d, want 2023", found[0].FiscalYear)
	}
	if found[1].FiscalYear != 2022 {
		t.Errorf("2番目の年度 = %d, want 2022", found[1].FiscalYear)
	}
	if found[2].FiscalYear != 2021 {
		t.Errorf("3番目の年度 = %d, want 2021", found[2].FiscalYear)
	}
}

func TestFinancialsRepository_FindByCompanyID_EmptyResult(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	finRepo := NewFinancialsRepository(td.Pool)
	compRepo := NewCompanyRepository(td.Pool)
	ctx := context.Background()

	// テスト企業作成（財務情報なし）
	company := &domain.Company{
		ID:              "company004",
		Name:            "財務なし株式会社",
		Industry:        domain.IndustryTechnology,
		Location:        "東京都",
		EmployeeCount:   100,
		Founded:         time.Date(2010, 5, 15, 0, 0, 0, 0, time.UTC),
		ListingStatus:   domain.ListingPrivate,
		MatchingPurpose: domain.PurposeBuyer,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err := compRepo.Create(ctx, company)
	if err != nil {
		t.Fatalf("企業作成失敗: %v", err)
	}

	// 財務情報なしの企業で検索
	found, err := finRepo.FindByCompanyID(ctx, "company004", 5)
	if err != nil {
		t.Fatalf("財務情報取得失敗: %v", err)
	}

	if len(found) != 0 {
		t.Errorf("取得件数 = %d, want 0", len(found))
	}
}

func TestFinancialsRepository_FindLatest_NormalCase(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	finRepo := NewFinancialsRepository(td.Pool)
	compRepo := NewCompanyRepository(td.Pool)
	ctx := context.Background()

	// テスト企業作成
	company := &domain.Company{
		ID:              "company005",
		Name:            "最新年度テスト株式会社",
		Industry:        domain.IndustryTechnology,
		Location:        "東京都",
		EmployeeCount:   100,
		Founded:         time.Date(2010, 5, 15, 0, 0, 0, 0, time.UTC),
		ListingStatus:   domain.ListingPrivate,
		MatchingPurpose: domain.PurposeBuyer,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err := compRepo.Create(ctx, company)
	if err != nil {
		t.Fatalf("企業作成失敗: %v", err)
	}

	// 複数年の財務情報作成
	years := []int{2021, 2022, 2023}
	for _, year := range years {
		financials := &domain.Financials{
			CompanyID:        "company005",
			FiscalYear:       year,
			Revenue:          int64(year * 100000000),
			EBITDA:           int64(year * 20000000),
			NetIncome:        int64(year * 10000000),
			TotalAssets:      int64(year * 200000000),
			TotalLiabilities: int64(year * 80000000),
			Equity:           int64(year * 120000000),
			ROE:              0.083,
			ROA:              0.05,
			DebtEquityRatio:  0.67,
			CurrentRatio:     1.5,
		}
		err = finRepo.Save(ctx, financials)
		if err != nil {
			t.Fatalf("財務情報保存失敗 (%d年度): %v", year, err)
		}
	}

	// 最新年度を取得
	found, err := finRepo.FindLatest(ctx, "company005")
	if err != nil {
		t.Fatalf("最新財務情報取得失敗: %v", err)
	}

	if found.FiscalYear != 2023 {
		t.Errorf("FiscalYear = %d, want 2023", found.FiscalYear)
	}
	if found.Revenue != 2023*100000000 {
		t.Errorf("Revenue = %d, want %d", found.Revenue, 2023*100000000)
	}
}

func TestFinancialsRepository_FindLatest_NotFound(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	finRepo := NewFinancialsRepository(td.Pool)
	ctx := context.Background()

	// 存在しない企業IDで検索
	_, err := finRepo.FindLatest(ctx, "nonexistent")
	if err == nil {
		t.Error("存在しない財務情報でエラーが発生すべき")
	}
}

func TestFinancialsRepository_FindLatest_EmptyCompanyID(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	finRepo := NewFinancialsRepository(td.Pool)
	ctx := context.Background()

	// 空のCompanyIDで検索
	_, err := finRepo.FindLatest(ctx, "")
	if err == nil {
		t.Error("空のCompanyIDでエラーが発生すべき")
	}
}

func TestFinancialsRepository_Save_Idempotent(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)

	finRepo := NewFinancialsRepository(td.Pool)
	compRepo := NewCompanyRepository(td.Pool)
	ctx := context.Background()

	// テスト企業作成
	company := &domain.Company{
		ID:              "company006",
		Name:            "冪等性テスト株式会社",
		Industry:        domain.IndustryTechnology,
		Location:        "東京都",
		EmployeeCount:   100,
		Founded:         time.Date(2010, 5, 15, 0, 0, 0, 0, time.UTC),
		ListingStatus:   domain.ListingPrivate,
		MatchingPurpose: domain.PurposeBuyer,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	err := compRepo.Create(ctx, company)
	if err != nil {
		t.Fatalf("企業作成失敗: %v", err)
	}

	// 同じ財務情報を3回保存
	financials := &domain.Financials{
		CompanyID:        "company006",
		FiscalYear:       2023,
		Revenue:          1000000000,
		EBITDA:           200000000,
		NetIncome:        100000000,
		TotalAssets:      2000000000,
		TotalLiabilities: 800000000,
		Equity:           1200000000,
		ROE:              0.083,
		ROA:              0.05,
		DebtEquityRatio:  0.67,
		CurrentRatio:     1.5,
	}

	for i := 0; i < 3; i++ {
		err = finRepo.Save(ctx, financials)
		if err != nil {
			t.Fatalf("Save %d回目失敗: %v", i+1, err)
		}
	}

	// 1件のみ存在することを確認
	found, err := finRepo.FindByCompanyID(ctx, "company006", 10)
	if err != nil {
		t.Fatalf("財務情報取得失敗: %v", err)
	}

	if len(found) != 1 {
		t.Errorf("取得件数 = %d, want 1", len(found))
	}
}
