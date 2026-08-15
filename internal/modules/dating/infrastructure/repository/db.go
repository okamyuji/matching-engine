package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB リポジトリが必要とするデータベース操作。*pgxpool.Pool が満たす
type DB interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

var _ DB = (*pgxpool.Pool)(nil)

// ErrNotFound 対象の行が存在しない
var ErrNotFound = pgx.ErrNoRows

func int32PtrFromInt(v int) *int32 {
	i := int32(v) //nolint:gosec // 身長・収入等の小さな値のみを扱う
	return &i
}

func intFromInt32Ptr(v *int32) int {
	if v == nil {
		return 0
	}
	return int(*v)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func strFromPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
