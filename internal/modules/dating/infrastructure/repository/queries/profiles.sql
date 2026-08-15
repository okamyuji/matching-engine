-- name: GetProfile :one
SELECT user_id, height, body_type, education, occupation, income_level, marriage_desire, children_desire, smoking, drinking, self_introduction, updated_at
FROM dating_profiles
WHERE user_id = $1;

-- name: UpsertProfile :exec
INSERT INTO dating_profiles (user_id, height, body_type, education, occupation, income_level, marriage_desire, children_desire, smoking, drinking, self_introduction, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, now())
ON CONFLICT (user_id) DO UPDATE SET
  height = EXCLUDED.height,
  body_type = EXCLUDED.body_type,
  education = EXCLUDED.education,
  occupation = EXCLUDED.occupation,
  income_level = EXCLUDED.income_level,
  marriage_desire = EXCLUDED.marriage_desire,
  children_desire = EXCLUDED.children_desire,
  smoking = EXCLUDED.smoking,
  drinking = EXCLUDED.drinking,
  self_introduction = EXCLUDED.self_introduction,
  updated_at = now();

-- name: ListProfileTags :many
SELECT id, user_id, tag FROM dating_profile_tags WHERE user_id = $1 ORDER BY id;

-- name: ListProfilePhotos :many
SELECT id, user_id, url, is_primary, display_order, created_at FROM dating_profile_photos WHERE user_id = $1 ORDER BY display_order, id;
