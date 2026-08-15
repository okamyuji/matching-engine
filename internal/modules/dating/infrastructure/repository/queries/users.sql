-- name: GetUser :one
SELECT id, nickname, gender, birth_date, prefecture, verified, elo_rating, created_at, last_active_at
FROM dating_users
WHERE id = $1;

-- name: CreateUser :exec
INSERT INTO dating_users (id, nickname, gender, birth_date, prefecture, verified, elo_rating, created_at, last_active_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: UpdateUser :exec
UPDATE dating_users
SET nickname = $2, gender = $3, birth_date = $4, prefecture = $5, verified = $6, elo_rating = $7, last_active_at = $8
WHERE id = $1;

-- name: FindCandidateUsers :many
-- 認証済みユーザーのうち、プロフィールを持ち条件に合う候補を最大1000件返す。
-- 各条件は NULL または空配列のとき無視する。
SELECT u.id, u.nickname, u.gender, u.birth_date, u.prefecture, u.verified, u.elo_rating, u.created_at, u.last_active_at
FROM dating_users u
INNER JOIN dating_profiles p ON p.user_id = u.id
WHERE u.id <> @user_id
  AND u.verified = TRUE
  AND (sqlc.narg('birth_min')::date IS NULL OR u.birth_date >= sqlc.narg('birth_min')::date)
  AND (sqlc.narg('birth_max')::date IS NULL OR u.birth_date <= sqlc.narg('birth_max')::date)
  AND (cardinality(@prefectures::text[]) = 0 OR u.prefecture = ANY(@prefectures::text[]))
  AND (sqlc.narg('height_min')::int IS NULL OR p.height >= sqlc.narg('height_min')::int)
  AND (sqlc.narg('height_max')::int IS NULL OR p.height <= sqlc.narg('height_max')::int)
  AND (sqlc.narg('income_min')::int IS NULL OR p.income_level >= sqlc.narg('income_min')::int)
  AND (cardinality(@educations::text[]) = 0 OR p.education = ANY(@educations::text[]))
  AND (cardinality(@marriage_desires::text[]) = 0 OR p.marriage_desire = ANY(@marriage_desires::text[]))
  AND (cardinality(@smoking_statuses::text[]) = 0 OR p.smoking = ANY(@smoking_statuses::text[]))
  AND (cardinality(@drinking_statuses::text[]) = 0 OR p.drinking = ANY(@drinking_statuses::text[]))
ORDER BY u.id
LIMIT 1000;
