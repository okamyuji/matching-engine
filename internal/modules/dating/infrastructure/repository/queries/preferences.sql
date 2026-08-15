-- name: GetPreference :one
SELECT user_id, age_min, age_max, height_min, height_max, income_min, updated_at
FROM dating_preferences
WHERE user_id = $1;

-- name: UpsertPreference :exec
INSERT INTO dating_preferences (user_id, age_min, age_max, height_min, height_max, income_min, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, now())
ON CONFLICT (user_id) DO UPDATE SET
  age_min = EXCLUDED.age_min,
  age_max = EXCLUDED.age_max,
  height_min = EXCLUDED.height_min,
  height_max = EXCLUDED.height_max,
  income_min = EXCLUDED.income_min,
  updated_at = now();

-- name: ListPreferencePrefectures :many
SELECT id, user_id, prefecture FROM dating_preference_prefectures WHERE user_id = $1 ORDER BY id;

-- name: DeletePreferencePrefectures :exec
DELETE FROM dating_preference_prefectures WHERE user_id = $1;

-- name: InsertPreferencePrefecture :one
INSERT INTO dating_preference_prefectures (user_id, prefecture) VALUES ($1, $2) RETURNING id;

-- name: ListPreferenceEducations :many
SELECT id, user_id, education FROM dating_preference_educations WHERE user_id = $1 ORDER BY id;

-- name: DeletePreferenceEducations :exec
DELETE FROM dating_preference_educations WHERE user_id = $1;

-- name: InsertPreferenceEducation :one
INSERT INTO dating_preference_educations (user_id, education) VALUES ($1, $2) RETURNING id;

-- name: ListPreferenceMarriageDesires :many
SELECT id, user_id, marriage_desire FROM dating_preference_marriage_desires WHERE user_id = $1 ORDER BY id;

-- name: DeletePreferenceMarriageDesires :exec
DELETE FROM dating_preference_marriage_desires WHERE user_id = $1;

-- name: InsertPreferenceMarriageDesire :one
INSERT INTO dating_preference_marriage_desires (user_id, marriage_desire) VALUES ($1, $2) RETURNING id;

-- name: ListPreferenceSmokingStatuses :many
SELECT id, user_id, smoking_status FROM dating_preference_smoking_statuses WHERE user_id = $1 ORDER BY id;

-- name: DeletePreferenceSmokingStatuses :exec
DELETE FROM dating_preference_smoking_statuses WHERE user_id = $1;

-- name: InsertPreferenceSmokingStatus :one
INSERT INTO dating_preference_smoking_statuses (user_id, smoking_status) VALUES ($1, $2) RETURNING id;

-- name: ListPreferenceDrinkingStatuses :many
SELECT id, user_id, drinking_status FROM dating_preference_drinking_statuses WHERE user_id = $1 ORDER BY id;

-- name: DeletePreferenceDrinkingStatuses :exec
DELETE FROM dating_preference_drinking_statuses WHERE user_id = $1;

-- name: InsertPreferenceDrinkingStatus :one
INSERT INTO dating_preference_drinking_statuses (user_id, drinking_status) VALUES ($1, $2) RETURNING id;
