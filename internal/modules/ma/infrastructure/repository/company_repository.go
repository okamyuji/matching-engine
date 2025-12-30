package repository

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/yourorg/matching-engine/internal/modules/ma/domain"
)

// CompanyRepository 企業データアクセス用インターフェース
type CompanyRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Company, error)
	FindByPurpose(ctx context.Context, purpose domain.MatchingPurpose, criteria *domain.MAMatchingCriteria) ([]*CompanyWithFinancials, error)
	FindCriteria(ctx context.Context, companyID string) (*domain.MAMatchingCriteria, error)
	Create(ctx context.Context, company *domain.Company) error
	Update(ctx context.Context, company *domain.Company) error
}

// CompanyWithFinancials 企業と財務情報を結合する
type CompanyWithFinancials struct {
	Company    *domain.Company
	Financials []*domain.Financials
}

// companyRepository CompanyRepositoryのBUN実装
type companyRepository struct {
	db *bun.DB
}

// NewCompanyRepository 新しいCompanyRepositoryを作成する
func NewCompanyRepository(db *bun.DB) CompanyRepository {
	return &companyRepository{db: db}
}

// FindByID IDにより企業を取得する
func (r *companyRepository) FindByID(ctx context.Context, id string) (*domain.Company, error) {
	company := &domain.Company{}
	err := r.db.NewSelect().
		Model(company).
		Where("id = ?", id).
		Relation("Technologies").
		Relation("Markets").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return company, nil
}

// FindByPurpose 目的（buyer/seller）で候補企業を取得する
// 要件:
// - 業界、地域、売上、従業員数でフィルタ
// - 最大500件の候補
func (r *companyRepository) FindByPurpose(
	ctx context.Context,
	purpose domain.MatchingPurpose,
	criteria *domain.MAMatchingCriteria,
) ([]*CompanyWithFinancials, error) {
	var companies []*domain.Company

	query := r.db.NewSelect().
		Model(&companies).
		Where("matching_purpose = ?", purpose)

	// 業界フィルタ
	if criteria != nil && len(criteria.TargetIndustries) > 0 {
		industries := make([]string, len(criteria.TargetIndustries))
		for i, ind := range criteria.TargetIndustries {
			industries[i] = string(ind.Industry)
		}
		query = query.Where("industry IN (?)", bun.In(industries))
	}

	// 従業員数フィルタ
	if criteria != nil {
		if criteria.EmployeeMin > 0 {
			query = query.Where("employee_count >= ?", criteria.EmployeeMin)
		}
		if criteria.EmployeeMax > 0 {
			query = query.Where("employee_count <= ?", criteria.EmployeeMax)
		}
	}

	query = query.Limit(500)

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	// 各企業の財務情報とテクノロジー、マーケットをロード
	results := make([]*CompanyWithFinancials, 0, len(companies))
	for _, company := range companies {
		// 財務情報を取得（最新5年分）
		var financials []*domain.Financials
		err := r.db.NewSelect().
			Model(&financials).
			Where("company_id = ?", company.ID).
			Order("fiscal_year DESC").
			Limit(5).
			Scan(ctx)

		if err != nil {
			// 財務情報がない場合は空配列
			financials = []*domain.Financials{}
		}

		// テクノロジーとマーケットをロード
		err = r.db.NewSelect().
			Model(company).
			Where("id = ?", company.ID).
			Relation("Technologies").
			Relation("Markets").
			Scan(ctx)
		if err != nil {
			return nil, err
		}

		// 売上・EBITDAフィルタ（最新年度の財務情報で判定）
		if len(financials) > 0 {
			latest := financials[0]
			if criteria != nil {
				// 売上フィルタ
				if criteria.RevenueMin > 0 && latest.Revenue < criteria.RevenueMin {
					continue
				}
				if criteria.RevenueMax > 0 && latest.Revenue > criteria.RevenueMax {
					continue
				}
				// EBITDAフィルタ
				if criteria.EBITDAMin > 0 && latest.EBITDA < criteria.EBITDAMin {
					continue
				}
				// 負債比率フィルタ
				if criteria.MaxDebtEquityRatio > 0 && latest.DebtEquityRatio > criteria.MaxDebtEquityRatio {
					continue
				}
			}
		}

		results = append(results, &CompanyWithFinancials{
			Company:    company,
			Financials: financials,
		})
	}

	return results, nil
}

// FindCriteria 企業のマッチング基準を取得する
func (r *companyRepository) FindCriteria(ctx context.Context, companyID string) (*domain.MAMatchingCriteria, error) {
	criteria := &domain.MAMatchingCriteria{}
	err := r.db.NewSelect().
		Model(criteria).
		Where("company_id = ?", companyID).
		Relation("TargetIndustries").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return criteria, nil
}

// Create 新しい企業を挿入する
func (r *companyRepository) Create(ctx context.Context, company *domain.Company) error {
	_, err := r.db.NewInsert().
		Model(company).
		Exec(ctx)
	return err
}

// Update 既存の企業を更新する
func (r *companyRepository) Update(ctx context.Context, company *domain.Company) error {
	_, err := r.db.NewUpdate().
		Model(company).
		Where("id = ?", company.ID).
		Exec(ctx)
	return err
}
