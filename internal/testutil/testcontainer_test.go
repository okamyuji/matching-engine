package testutil

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestGetSharedTestDB_SeedAndClean(t *testing.T) {
	td := GetSharedTestDB(t)
	ctx := context.Background()

	td.SeedTestData(t)

	var users int
	if err := td.Pool.QueryRow(ctx, "SELECT count(*) FROM dating_users").Scan(&users); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if users != 5 {
		t.Errorf("seeded users = %d, want 5", users)
	}

	td.CleanTables(t)
	if err := td.Pool.QueryRow(ctx, "SELECT count(*) FROM dating_users").Scan(&users); err != nil {
		t.Fatalf("count users after clean: %v", err)
	}
	if users != 0 {
		t.Errorf("users after clean = %d, want 0", users)
	}
}

func TestGetSharedTestDB_ReturnsSameInstance(t *testing.T) {
	a := GetSharedTestDB(t)
	b := GetSharedTestDB(t)
	if a != b {
		t.Fatal("GetSharedTestDB は同じインスタンスを返すべき")
	}
	if a.DBName == "" || a.DSN == "" {
		t.Fatalf("DBName/DSN が空: %+v", a)
	}
}

func TestReplaceDatabase(t *testing.T) {
	cases := map[string]string{
		"postgres://u:p@h:1/postgres?sslmode=disable": "postgres://u:p@h:1/x?sslmode=disable",
		"postgres://u:p@h:1/postgres":                 "postgres://u:p@h:1/x",
	}
	for in, want := range cases {
		if got := replaceDatabase(in, "x"); got != want {
			t.Errorf("replaceDatabase(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEnsureTemplate_RecreatesWhenHashDiffers(t *testing.T) {
	td := GetSharedTestDB(t)
	ctx := context.Background()
	adminDSN := replaceDatabase(td.DSN, adminDBName)
	admin, err := connectWithRetry(ctx, adminDSN, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)

	if _, err := admin.Exec(ctx, "SELECT pg_advisory_lock($1)", advisoryLockKey); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = admin.Exec(ctx, "SELECT pg_advisory_unlock($1)", advisoryLockKey) }()

	// 現在のマイグレーションでは何もしない（ハッシュ一致）
	migrationPath, err := repoFile("db", "migrations", "001_create_tables.sql")
	if err != nil {
		t.Fatal(err)
	}
	migrationSQL, err := os.ReadFile(migrationPath) //nolint:gosec // テスト用の固定パス
	if err != nil {
		t.Fatal(err)
	}
	if err := ensureTemplate(ctx, admin, adminDSN, migrationSQL); err != nil {
		t.Fatalf("同一ハッシュ: %v", err)
	}

	// 別のマイグレーション内容なら作り直す（テンプレートに接続が無いことが前提）
	altered := append([]byte("-- altered\n"), migrationSQL...)
	if err := ensureTemplate(ctx, admin, adminDSN, altered); err != nil {
		t.Fatalf("ハッシュ不一致で再作成: %v", err)
	}
	got, err := readTemplateHash(ctx, replaceDatabase(adminDSN, templateDBName))
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(altered)
	if got != hex.EncodeToString(sum[:]) {
		t.Errorf("再作成後のハッシュが違う: %s", got)
	}

	// 元に戻しておく（後続プロセスのため）
	if err := ensureTemplate(ctx, admin, adminDSN, migrationSQL); err != nil {
		t.Fatalf("復元: %v", err)
	}
	// 存在確認のヘルパー
	exists, err := databaseExists(ctx, admin, templateDBName)
	if err != nil || !exists {
		t.Errorf("テンプレートが存在するべき: %v %v", exists, err)
	}
	exists, err = databaseExists(ctx, admin, "no_such_db_xyz")
	if err != nil || exists {
		t.Errorf("存在しないDB: %v %v", exists, err)
	}
}

func TestDropStaleTestDBs(t *testing.T) {
	td := GetSharedTestDB(t)
	ctx := context.Background()
	adminDSN := replaceDatabase(td.DSN, adminDBName)
	admin, err := connectWithRetry(ctx, adminDSN, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)

	old := fmt.Sprintf("test_%d_deadbeef", time.Now().Add(-2*staleTestDBAge).Unix())
	fresh := fmt.Sprintf("test_%d_cafebabe", time.Now().Unix())
	for _, n := range []string{old, fresh} {
		if _, err := admin.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{n}.Sanitize()); err != nil {
			t.Fatalf("create %s: %v", n, err)
		}
	}
	dropStaleTestDBs(ctx, admin)
	oldExists, _ := databaseExists(ctx, admin, old)
	freshExists, _ := databaseExists(ctx, admin, fresh)
	if oldExists {
		t.Error("古いテストDBは削除されるべき")
	}
	if !freshExists {
		t.Error("新しいテストDBは残るべき")
	}
	_, _ = admin.Exec(ctx, "DROP DATABASE IF EXISTS "+pgx.Identifier{fresh}.Sanitize()+" WITH (FORCE)")
}

func TestConnectWithRetry_Timeout(t *testing.T) {
	ctx := context.Background()
	if _, err := connectWithRetry(ctx, "postgres://u:p@127.0.0.1:1/x?sslmode=disable", 1200*time.Millisecond); err == nil {
		t.Error("到達不能ならタイムアウトでエラーになるべき")
	}
}
