-- name: InsertInterest :exec
INSERT INTO ma_interests (id, from_company_id, to_company_id, created_at) VALUES ($1, $2, $3, $4);

-- name: ListInterestsByToCompany :many
SELECT id, from_company_id, to_company_id, created_at FROM ma_interests WHERE to_company_id = $1 ORDER BY created_at DESC, id;

-- name: InterestExists :one
SELECT EXISTS (SELECT 1 FROM ma_interests WHERE from_company_id = $1 AND to_company_id = $2) AS exists;
