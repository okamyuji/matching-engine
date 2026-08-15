-- name: GetCompany :one
SELECT id, name, industry, location, employee_count, founded, listing_status, matching_purpose, created_at, updated_at
FROM ma_companies
WHERE id = $1;

-- name: CreateCompany :exec
INSERT INTO ma_companies (id, name, industry, location, employee_count, founded, listing_status, matching_purpose, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: UpdateCompany :exec
UPDATE ma_companies
SET name = $2, industry = $3, location = $4, employee_count = $5, founded = $6, listing_status = $7, matching_purpose = $8, updated_at = now()
WHERE id = $1;

-- name: ListCompaniesByPurpose :many
-- 目的が一致する企業を最大500件返す。業種配列が空、従業員条件が NULL のときは無視する。
SELECT id, name, industry, location, employee_count, founded, listing_status, matching_purpose, created_at, updated_at
FROM ma_companies
WHERE matching_purpose = @purpose
  AND (cardinality(@industries::text[]) = 0 OR industry = ANY(@industries::text[]))
  AND (sqlc.narg('employee_min')::int IS NULL OR employee_count >= sqlc.narg('employee_min')::int)
  AND (sqlc.narg('employee_max')::int IS NULL OR employee_count <= sqlc.narg('employee_max')::int)
ORDER BY id
LIMIT 500;

-- name: ListCompanyTechnologies :many
SELECT id, company_id, technology FROM ma_company_technologies WHERE company_id = $1 ORDER BY id;

-- name: InsertCompanyTechnology :one
INSERT INTO ma_company_technologies (company_id, technology) VALUES ($1, $2) RETURNING id;

-- name: ListCompanyMarkets :many
SELECT id, company_id, market FROM ma_company_markets WHERE company_id = $1 ORDER BY id;

-- name: InsertCompanyMarket :one
INSERT INTO ma_company_markets (company_id, market) VALUES ($1, $2) RETURNING id;

-- name: GetCriteria :one
SELECT company_id, purpose, revenue_min, revenue_max, ebitda_min, employee_min, employee_max, max_debt_equity_ratio, updated_at
FROM ma_matching_criteria
WHERE company_id = $1;

-- name: UpsertCriteria :exec
INSERT INTO ma_matching_criteria (company_id, purpose, revenue_min, revenue_max, ebitda_min, employee_min, employee_max, max_debt_equity_ratio, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now())
ON CONFLICT (company_id) DO UPDATE SET
  purpose = EXCLUDED.purpose,
  revenue_min = EXCLUDED.revenue_min,
  revenue_max = EXCLUDED.revenue_max,
  ebitda_min = EXCLUDED.ebitda_min,
  employee_min = EXCLUDED.employee_min,
  employee_max = EXCLUDED.employee_max,
  max_debt_equity_ratio = EXCLUDED.max_debt_equity_ratio,
  updated_at = now();

-- name: ListCriteriaIndustries :many
SELECT id, company_id, industry FROM ma_criteria_industries WHERE company_id = $1 ORDER BY id;

-- name: DeleteCriteriaIndustries :exec
DELETE FROM ma_criteria_industries WHERE company_id = $1;

-- name: InsertCriteriaIndustry :one
INSERT INTO ma_criteria_industries (company_id, industry) VALUES ($1, $2) RETURNING id;
