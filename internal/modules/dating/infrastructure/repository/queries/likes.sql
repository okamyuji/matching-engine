-- name: InsertLike :exec
INSERT INTO dating_likes (id, from_user_id, to_user_id, created_at) VALUES ($1, $2, $3, $4);

-- name: ListLikesByToUser :many
SELECT id, from_user_id, to_user_id, created_at FROM dating_likes WHERE to_user_id = $1 ORDER BY created_at DESC, id;

-- name: ListLikesByFromUser :many
SELECT id, from_user_id, to_user_id, created_at FROM dating_likes WHERE from_user_id = $1 ORDER BY created_at DESC, id;

-- name: LikeExists :one
SELECT EXISTS (SELECT 1 FROM dating_likes WHERE from_user_id = $1 AND to_user_id = $2) AS exists;
