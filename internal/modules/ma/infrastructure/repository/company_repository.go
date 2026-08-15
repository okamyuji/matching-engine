package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/okamyuji/matching-engine/internal/modules/ma/domain"
	"github.com/okamyuji/matching-engine/internal/modules/ma/infrastructure/repository/sqlcgen"
)

// CompanyRepository 企業データアクセス用インターフェース
type CompanyRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Company, error)
	FindByPurpose(ctx context.Context, purpose domain.MatchingPurpose, criteria *domain.MAMatchingCriteria) ([]*CompanyWithFinancials, error)
	FindCriteria(ctx context.Context, companyID string) (*domain.MAMatchingCriteria, error)
	Create(ctx context.Context, company *domain.Company) error
	Update(ctx context.Context, company *domain.Company) error
}

// CompanyWithFinancials 企業と財務データを結合する
type CompanyWithFinancials struct {
	Company    *domain.Company
	Financials []*domain.Financials
}

// companyRepository CompanyRepository の sqlc 実装
type companyRepository struct {
	q *sqlcgen.Queries
}

// NewCompanyRepository 新しい CompanyRepository を作成する
func NewCompanyRepository(db DB) CompanyRepository {
	return &companyRepository{q: sqlcgen.New(db)}
}

// FindByID IDにより企業（技術・市場を含む）を取得する
func (r *companyRepository) FindByID(ctx context.Context, id string) (*domain.Company, error) {
	row, err := r.q.GetCompany(ctx, id)
	if err != nil {
		return nil, err
	}
	company := companyFromRow(row)
	if err := r.loadRelations(ctx, company); err != nil {
		return nil, err
	}
	return company, nil
}

// FindByPurpose 目的と条件に合う企業を最大500件、直近5年分の財務データ付きで取得する
// 財務条件（売上・EBITDA・負債比率）は最新年度の値で判定する
func (r *companyRepository) FindByPurpose(
	ctx context.Context,
	purpose domain.MatchingPurpose,
	criteria *domain.MAMatchingCriteria,
) ([]*CompanyWithFinancials, error) {
	params := sqlcgen.ListCompaniesByPurposeParams{
		Purpose:    string(purpose),
		Industries: []string{},
	}
	if criteria != nil {
		params.Industries = criteria.GetIndustryStrings()
		if criteria.EmployeeMin > 0 {
			params.EmployeeMin = int32PtrFromInt(criteria.EmployeeMin)
		}
		if criteria.EmployeeMax > 0 {
			params.EmployeeMax = int32PtrFromInt(criteria.EmployeeMax)
		}
	}
	rows, err := r.q.ListCompaniesByPurpose(ctx, params)
	if err != nil {
		return nil, err
	}

	results := make([]*CompanyWithFinancials, 0, len(rows))
	for _, row := range rows {
		company := companyFromRow(row)
		if err := r.loadRelations(ctx, company); err != nil {
			return nil, err
		}
		finRows, err := r.q.ListFinancialsByCompany(ctx, sqlcgen.ListFinancialsByCompanyParams{CompanyID: company.ID, Limit: 5})
		if err != nil {
			return nil, fmt.Errorf("list financials %s: %w", company.ID, err)
		}
		financials := financialsFromRows(finRows)
		if len(financials) > 0 && criteria != nil && !matchesFinancialCriteria(financials[0], criteria) {
			continue
		}
		results = append(results, &CompanyWithFinancials{Company: company, Financials: financials})
	}
	return results, nil
}

func matchesFinancialCriteria(latest *domain.Financials, c *domain.MAMatchingCriteria) bool {
	if c.RevenueMin > 0 && latest.Revenue < c.RevenueMin {
		return false
	}
	if c.RevenueMax > 0 && latest.Revenue > c.RevenueMax {
		return false
	}
	if c.EBITDAMin > 0 && latest.EBITDA < c.EBITDAMin {
		return false
	}
	if c.MaxDebtEquityRatio > 0 && latest.DebtEquityRatio > c.MaxDebtEquityRatio {
		return false
	}
	return true
}

// FindCriteria 企業のマッチング条件（対象業種を含む）を取得する
func (r *companyRepository) FindCriteria(ctx context.Context, companyID string) (*domain.MAMatchingCriteria, error) {
	row, err := r.q.GetCriteria(ctx, companyID)
	if err != nil {
		return nil, err
	}
	criteria := &domain.MAMatchingCriteria{
		CompanyID:          row.CompanyID,
		Purpose:            domain.MatchingPurpose(row.Purpose),
		RevenueMin:         int64FromPtr(row.RevenueMin),
		RevenueMax:         int64FromPtr(row.RevenueMax),
		EBITDAMin:          int64FromPtr(row.EbitdaMin),
		EmployeeMin:        intFromInt32Ptr(row.EmployeeMin),
		EmployeeMax:        intFromInt32Ptr(row.EmployeeMax),
		MaxDebtEquityRatio: float64FromPtr(row.MaxDebtEquityRatio),
		UpdatedAt:          row.UpdatedAt,
	}
	inds, err := r.q.ListCriteriaIndustries(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("list criteria industries: %w", err)
	}
	for _, i := range inds {
		criteria.TargetIndustries = append(criteria.TargetIndustries, domain.CriteriaIndustry{ID: i.ID, CompanyID: i.CompanyID, Industry: domain.Industry(i.Industry)})
	}
	return criteria, nil
}

// Create 新しい企業を挿入する
func (r *companyRepository) Create(ctx context.Context, company *domain.Company) error {
	now := time.Now()
	createdAt := company.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}
	updatedAt := company.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = now
	}
	listing := company.ListingStatus
	if listing == "" {
		listing = domain.ListingPrivate
	}
	return r.q.CreateCompany(ctx, sqlcgen.CreateCompanyParams{
		ID:              company.ID,
		Name:            company.Name,
		Industry:        string(company.Industry),
		Location:        company.Location,
		EmployeeCount:   int32(company.EmployeeCount), //nolint:gosec // 従業員数
		Founded:         company.Founded,
		ListingStatus:   string(listing),
		MatchingPurpose: string(company.MatchingPurpose),
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	})
}

// Update 既存の企業を更新する
func (r *companyRepository) Update(ctx context.Context, company *domain.Company) error {
	listing := company.ListingStatus
	if listing == "" {
		listing = domain.ListingPrivate
	}
	return r.q.UpdateCompany(ctx, sqlcgen.UpdateCompanyParams{
		ID:              company.ID,
		Name:            company.Name,
		Industry:        string(company.Industry),
		Location:        company.Location,
		EmployeeCount:   int32(company.EmployeeCount), //nolint:gosec // 従業員数
		Founded:         company.Founded,
		ListingStatus:   string(listing),
		MatchingPurpose: string(company.MatchingPurpose),
	})
}

func (r *companyRepository) loadRelations(ctx context.Context, company *domain.Company) error {
	techs, err := r.q.ListCompanyTechnologies(ctx, company.ID)
	if err != nil {
		return fmt.Errorf("list technologies %s: %w", company.ID, err)
	}
	company.Technologies = make([]*domain.CompanyTechnology, 0, len(techs))
	for _, t := range techs {
		company.Technologies = append(company.Technologies, &domain.CompanyTechnology{ID: t.ID, CompanyID: t.CompanyID, Technology: t.Technology})
	}
	markets, err := r.q.ListCompanyMarkets(ctx, company.ID)
	if err != nil {
		return fmt.Errorf("list markets %s: %w", company.ID, err)
	}
	company.Markets = make([]*domain.CompanyMarket, 0, len(markets))
	for _, m := range markets {
		company.Markets = append(company.Markets, &domain.CompanyMarket{ID: m.ID, CompanyID: m.CompanyID, Market: m.Market})
	}
	return nil
}

func companyFromRow(row sqlcgen.MaCompany) *domain.Company {
	return &domain.Company{
		ID:              row.ID,
		Name:            row.Name,
		Industry:        domain.Industry(row.Industry),
		Location:        row.Location,
		EmployeeCount:   int(row.EmployeeCount),
		Founded:         row.Founded,
		ListingStatus:   domain.ListingStatus(row.ListingStatus),
		MatchingPurpose: domain.MatchingPurpose(row.MatchingPurpose),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

// InsertTechnology 企業の技術を追加する（テストデータ投入や管理操作向け）
func InsertTechnology(ctx context.Context, db DB, tech *domain.CompanyTechnology) error {
	id, err := sqlcgen.New(db).InsertCompanyTechnology(ctx, sqlcgen.InsertCompanyTechnologyParams{CompanyID: tech.CompanyID, Technology: tech.Technology})
	if err != nil {
		return err
	}
	tech.ID = id
	return nil
}

// InsertMarket 企業の市場を追加する（テストデータ投入や管理操作向け）
func InsertMarket(ctx context.Context, db DB, market *domain.CompanyMarket) error {
	id, err := sqlcgen.New(db).InsertCompanyMarket(ctx, sqlcgen.InsertCompanyMarketParams{CompanyID: market.CompanyID, Market: market.Market})
	if err != nil {
		return err
	}
	market.ID = id
	return nil
}

// UpsertCriteria マッチング条件を挿入または更新する。対象業種は criteria.TargetIndustries で置き換える
func UpsertCriteria(ctx context.Context, db DB, criteria *domain.MAMatchingCriteria) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := sqlcgen.New(tx)
	if err := q.UpsertCriteria(ctx, criteriaParams(criteria)); err != nil {
		return err
	}
	if len(criteria.TargetIndustries) > 0 {
		if err := q.DeleteCriteriaIndustries(ctx, criteria.CompanyID); err != nil {
			return err
		}
		for i := range criteria.TargetIndustries {
			ind := &criteria.TargetIndustries[i]
			id, err := q.InsertCriteriaIndustry(ctx, sqlcgen.InsertCriteriaIndustryParams{CompanyID: criteria.CompanyID, Industry: string(ind.Industry)})
			if err != nil {
				return err
			}
			ind.ID = id
		}
	}
	return tx.Commit(ctx)
}

// criteriaParams ドメインの条件を sqlc のパラメータに変換する。0 は「条件なし」として NULL にする
func criteriaParams(c *domain.MAMatchingCriteria) sqlcgen.UpsertCriteriaParams {
	return sqlcgen.UpsertCriteriaParams{
		CompanyID:          c.CompanyID,
		Purpose:            string(c.Purpose),
		RevenueMin:         positiveInt64Ptr(c.RevenueMin),
		RevenueMax:         positiveInt64Ptr(c.RevenueMax),
		EbitdaMin:          positiveInt64Ptr(c.EBITDAMin),
		EmployeeMin:        positiveInt32Ptr(c.EmployeeMin),
		EmployeeMax:        positiveInt32Ptr(c.EmployeeMax),
		MaxDebtEquityRatio: positiveFloat64Ptr(c.MaxDebtEquityRatio),
	}
}

func positiveInt64Ptr(v int64) *int64 {
	if v <= 0 {
		return nil
	}
	return int64Ptr(v)
}

func positiveInt32Ptr(v int) *int32 {
	if v <= 0 {
		return nil
	}
	return int32PtrFromInt(v)
}

func positiveFloat64Ptr(v float64) *float64 {
	if v <= 0 {
		return nil
	}
	return float64Ptr(v)
}

// InsertCriteriaIndustry マッチング条件の対象業種を1件追加する
func InsertCriteriaIndustry(ctx context.Context, db DB, ind *domain.CriteriaIndustry) error {
	id, err := sqlcgen.New(db).InsertCriteriaIndustry(ctx, sqlcgen.InsertCriteriaIndustryParams{CompanyID: ind.CompanyID, Industry: string(ind.Industry)})
	if err != nil {
		return err
	}
	ind.ID = id
	return nil
}

// IsNotFound 行が見つからないエラーかを返す
func IsNotFound(err error) bool { return errors.Is(err, pgx.ErrNoRows) }
