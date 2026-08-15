-- name: InsertMAMatch :exec
INSERT INTO ma_matches (id, company_id_a, company_id_b, score, breakdown, synergy_summary, matched_at) VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetMAMatch :one
SELECT id, company_id_a, company_id_b, score, breakdown, synergy_summary, matched_at FROM ma_matches WHERE id = $1;

-- name: ListMutualMAMatchesByCompany :many
SELECT m.id, m.company_id_a, m.company_id_b, m.score, m.breakdown, m.synergy_summary, m.matched_at
FROM ma_matches m
WHERE (m.company_id_a = $1 OR m.company_id_b = $1)
  AND EXISTS (SELECT 1 FROM ma_interests i WHERE i.from_company_id = m.company_id_a AND i.to_company_id = m.company_id_b)
  AND EXISTS (SELECT 1 FROM ma_interests i WHERE i.from_company_id = m.company_id_b AND i.to_company_id = m.company_id_a)
ORDER BY m.matched_at DESC, m.id;
