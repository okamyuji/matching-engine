-- name: InsertMatch :exec
INSERT INTO dating_matches (id, user_id_a, user_id_b, score, breakdown, matched_at) VALUES ($1, $2, $3, $4, $5, $6);

-- name: ListMatchesByUser :many
SELECT id, user_id_a, user_id_b, score, breakdown, matched_at
FROM dating_matches
WHERE user_id_a = $1 OR user_id_b = $1
ORDER BY matched_at DESC, id;

-- name: ListMutualMatchesByUser :many
SELECT m.id, m.user_id_a, m.user_id_b, m.score, m.breakdown, m.matched_at
FROM dating_matches m
WHERE (m.user_id_a = $1 OR m.user_id_b = $1)
  AND EXISTS (SELECT 1 FROM dating_likes l WHERE l.from_user_id = m.user_id_a AND l.to_user_id = m.user_id_b)
  AND EXISTS (SELECT 1 FROM dating_likes l WHERE l.from_user_id = m.user_id_b AND l.to_user_id = m.user_id_a)
ORDER BY m.matched_at DESC, m.id;
