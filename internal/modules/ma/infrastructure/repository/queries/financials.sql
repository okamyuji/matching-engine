-- name: ListFinancialsByCompany :many
SELECT id, company_id, fiscal_year, revenue, ebitda, net_income, total_assets, total_liabilities, equity, roe, roa, debt_equity_ratio, current_ratio, created_at
FROM ma_financials
WHERE company_id = $1
ORDER BY fiscal_year DESC
LIMIT $2;

-- name: GetLatestFinancials :one
SELECT id, company_id, fiscal_year, revenue, ebitda, net_income, total_assets, total_liabilities, equity, roe, roa, debt_equity_ratio, current_ratio, created_at
FROM ma_financials
WHERE company_id = $1
ORDER BY fiscal_year DESC
LIMIT 1;

-- name: UpsertFinancials :one
INSERT INTO ma_financials (company_id, fiscal_year, revenue, ebitda, net_income, total_assets, total_liabilities, equity, roe, roa, debt_equity_ratio, current_ratio, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
ON CONFLICT (company_id, fiscal_year) DO UPDATE SET
  revenue = EXCLUDED.revenue,
  ebitda = EXCLUDED.ebitda,
  net_income = EXCLUDED.net_income,
  total_assets = EXCLUDED.total_assets,
  total_liabilities = EXCLUDED.total_liabilities,
  equity = EXCLUDED.equity,
  roe = EXCLUDED.roe,
  roa = EXCLUDED.roa,
  debt_equity_ratio = EXCLUDED.debt_equity_ratio,
  current_ratio = EXCLUDED.current_ratio
RETURNING id;
