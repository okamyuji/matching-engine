package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/okamyuji/matching-engine/internal/modules/ma/domain"
	"github.com/okamyuji/matching-engine/internal/testutil"
)

func TestMatchesFinancialCriteria(t *testing.T) {
	latest := &domain.Financials{Revenue: 1000, EBITDA: 100, DebtEquityRatio: 1.0}
	cases := []struct {
		name string
		c    domain.MAMatchingCriteria
		want bool
	}{
		{"条件なし", domain.MAMatchingCriteria{}, true},
		{"売上下限を満たす", domain.MAMatchingCriteria{RevenueMin: 1000}, true},
		{"売上下限を下回る", domain.MAMatchingCriteria{RevenueMin: 1001}, false},
		{"売上上限を満たす", domain.MAMatchingCriteria{RevenueMax: 1000}, true},
		{"売上上限を超える", domain.MAMatchingCriteria{RevenueMax: 999}, false},
		{"EBITDA 下限を満たす", domain.MAMatchingCriteria{EBITDAMin: 100}, true},
		{"EBITDA 下限を下回る", domain.MAMatchingCriteria{EBITDAMin: 101}, false},
		{"負債比率上限を満たす", domain.MAMatchingCriteria{MaxDebtEquityRatio: 1.0}, true},
		{"負債比率上限を超える", domain.MAMatchingCriteria{MaxDebtEquityRatio: 0.9}, false},
		{"複合で最後の条件が落ちる", domain.MAMatchingCriteria{RevenueMin: 1, RevenueMax: 5000, EBITDAMin: 1, MaxDebtEquityRatio: 0.5}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := matchesFinancialCriteria(latest, &tc.c); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	if !IsNotFound(pgx.ErrNoRows) {
		t.Error("ErrNoRows は not found")
	}
	if IsNotFound(errors.New("other")) {
		t.Error("他のエラーは not found ではない")
	}
	if !errors.Is(ErrNotFound, pgx.ErrNoRows) {
		t.Error("ErrNotFound は pgx.ErrNoRows と同一であるべき")
	}
}

func TestUpsertCriteria_WithIndustriesAndErrors(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	td.CleanTables(t)
	ctx := context.Background()
	repo := NewCompanyRepository(td.Pool)
	c := &domain.Company{ID: "crit-co", Name: "C", Industry: domain.IndustryTechnology, Location: "Tokyo", EmployeeCount: 10, Founded: time.Now(), MatchingPurpose: domain.PurposeBuyer}
	if err := repo.Create(ctx, c); err != nil {
		t.Fatal(err)
	}

	criteria := &domain.MAMatchingCriteria{CompanyID: "crit-co", Purpose: domain.PurposeBuyer, RevenueMin: 100, MaxDebtEquityRatio: 2,
		TargetIndustries: []domain.CriteriaIndustry{{CompanyID: "crit-co", Industry: domain.IndustryFinance}, {CompanyID: "crit-co", Industry: domain.IndustryRetail}}}
	if err := UpsertCriteria(ctx, td.Pool, criteria); err != nil {
		t.Fatalf("UpsertCriteria: %v", err)
	}
	if criteria.TargetIndustries[0].ID == 0 || criteria.TargetIndustries[1].ID == 0 {
		t.Error("業種の ID が採番されるべき")
	}
	// 2回目は業種を置き換える
	criteria.TargetIndustries = []domain.CriteriaIndustry{{CompanyID: "crit-co", Industry: domain.IndustryEnergy}}
	if err := UpsertCriteria(ctx, td.Pool, criteria); err != nil {
		t.Fatalf("UpsertCriteria 2回目: %v", err)
	}
	got, err := repo.FindCriteria(ctx, "crit-co")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.TargetIndustries) != 1 || got.TargetIndustries[0].Industry != domain.IndustryEnergy {
		t.Errorf("業種が置き換わっていない: %+v", got.TargetIndustries)
	}
	if got.RevenueMin != 100 || got.RevenueMax != 0 || got.MaxDebtEquityRatio != 2 {
		t.Errorf("条件値が違う: %+v", got)
	}

	// 存在しない企業は外部キー違反でロールバックされる
	bad := &domain.MAMatchingCriteria{CompanyID: "no-such", Purpose: domain.PurposeBuyer, TargetIndustries: []domain.CriteriaIndustry{{CompanyID: "no-such", Industry: domain.IndustryEnergy}}}
	if err := UpsertCriteria(ctx, td.Pool, bad); err == nil {
		t.Error("存在しない企業でエラーになるべき")
	}
	// 重複業種は一意制約違反
	dup := &domain.MAMatchingCriteria{CompanyID: "crit-co", Purpose: domain.PurposeBuyer, TargetIndustries: []domain.CriteriaIndustry{{CompanyID: "crit-co", Industry: domain.IndustryEnergy}, {CompanyID: "crit-co", Industry: domain.IndustryEnergy}}}
	if err := UpsertCriteria(ctx, td.Pool, dup); err == nil {
		t.Error("重複業種でエラーになるべき")
	}
}
