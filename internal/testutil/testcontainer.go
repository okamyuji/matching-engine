// Package testutil テスト用の PostgreSQL 環境を提供する。
//
// コンテナはテストごとに起動・終了せず、名前付きで再利用する（WithReuseByName）。
// 複数のテストプロセス（go test ./... の各パッケージ）が同時に同じコンテナを共有できるよう、
// マイグレーション済みのテンプレートDBを一度だけ作り、プロセスごとに
// CREATE DATABASE ... TEMPLATE で独立したDBを払い出す。テーブルの初期化やシード投入は
// 払い出したDBの中だけで行うため、パッケージ間で干渉しない。
package testutil

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	// PostgresImage テストで使う PostgreSQL のイメージ
	PostgresImage = "postgres:18-alpine"
	// containerName 再利用するコンテナ名。同名のコンテナがあればそれを使う
	containerName = "matching-engine-pg18-test"
	// templateDBName マイグレーション済みテンプレートDB
	templateDBName = "matching_template"
	// adminDBName 管理操作に使う既定DB
	adminDBName = "postgres"
	// advisoryLockKey テンプレート作成とテストDB払い出しを直列化するロック
	advisoryLockKey = 727272
	// staleTestDBAge この時間より古いテストDBは起動時に破棄する
	staleTestDBAge = 30 * time.Minute

	dbUser     = "test"
	dbPassword = "test"
)

// TestDatabase テストデータベース情報
type TestDatabase struct {
	// Pool このプロセス専用のテストDBへのプール
	Pool *pgxpool.Pool
	// DSN このプロセス専用のテストDBへの接続文字列
	DSN string
	// DBName 払い出したDB名
	DBName string
}

var (
	sharedMu   sync.Mutex
	sharedOnce sync.Once
	sharedDB   *TestDatabase
	sharedErr  error
)

// GetSharedTestDB プロセス内で共有するテストDBを返す。初回だけコンテナの起動（または再利用）、
// テンプレートDBの準備、プロセス専用DBの払い出しを行う。
func GetSharedTestDB(t *testing.T) *TestDatabase {
	t.Helper()
	sharedMu.Lock()
	defer sharedMu.Unlock()

	sharedOnce.Do(func() {
		sharedDB, sharedErr = setup(context.Background())
	})
	if sharedErr != nil {
		t.Fatalf("テストDB準備失敗: %v", sharedErr)
	}
	return sharedDB
}

// CleanTables 全テーブルのデータを削除する（スキーマは残す）
func (td *TestDatabase) CleanTables(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	if err := TruncateAll(ctx, td.Pool); err != nil {
		t.Fatalf("テーブル初期化失敗: %v", err)
	}
}

// SeedTestData テストデータを投入する（先に全テーブルを空にする）
func (td *TestDatabase) SeedTestData(t *testing.T) {
	t.Helper()
	td.CleanTables(t)

	seedPath, err := repoFile("db", "testdata", "seed.sql")
	if err != nil {
		t.Fatalf("シードデータのパス解決失敗: %v", err)
	}
	seedSQL, err := os.ReadFile(seedPath) //nolint:gosec // テスト内で固定パスを読む
	if err != nil {
		t.Fatalf("シードデータ読み込み失敗: %v", err)
	}
	if _, err := td.Pool.Exec(context.Background(), string(seedSQL)); err != nil {
		t.Fatalf("シードデータ投入失敗: %v", err)
	}
}

// TruncateAll public スキーマの全テーブルを TRUNCATE する
func TruncateAll(ctx context.Context, pool *pgxpool.Pool) error {
	rows, err := pool.Query(ctx, `SELECT tablename FROM pg_tables WHERE schemaname = 'public'`)
	if err != nil {
		return fmt.Errorf("list tables: %w", err)
	}
	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return fmt.Errorf("scan table name: %w", err)
		}
		tables = append(tables, pgx.Identifier{name}.Sanitize())
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate tables: %w", err)
	}
	if len(tables) == 0 {
		return nil
	}
	_, err = pool.Exec(ctx, "TRUNCATE TABLE "+strings.Join(tables, ", ")+" RESTART IDENTITY CASCADE")
	if err != nil {
		return fmt.Errorf("truncate: %w", err)
	}
	return nil
}

// setup コンテナとDBを準備する
func setup(ctx context.Context) (*TestDatabase, error) {
	adminDSN, err := startOrReuseContainer(ctx)
	if err != nil {
		return nil, err
	}

	migrationPath, err := repoFile("db", "migrations", "001_create_tables.sql")
	if err != nil {
		return nil, err
	}
	migrationSQL, err := os.ReadFile(migrationPath) //nolint:gosec // テスト内で固定パスを読む
	if err != nil {
		return nil, fmt.Errorf("マイグレーション読み込み失敗: %w", err)
	}

	admin, err := connectWithRetry(ctx, adminDSN, 60*time.Second)
	if err != nil {
		return nil, err
	}
	defer func() { _ = admin.Close(ctx) }()

	// テンプレート作成とDB払い出しは advisory lock で直列化する。
	// CREATE DATABASE ... TEMPLATE はテンプレートへの接続が無いことを要求するため、
	// テンプレートへの接続はロック内で閉じてから払い出す。
	if _, err := admin.Exec(ctx, "SELECT pg_advisory_lock($1)", advisoryLockKey); err != nil {
		return nil, fmt.Errorf("advisory lock: %w", err)
	}
	defer func() { _, _ = admin.Exec(ctx, "SELECT pg_advisory_unlock($1)", advisoryLockKey) }()

	if err := ensureTemplate(ctx, admin, adminDSN, migrationSQL); err != nil {
		return nil, err
	}
	dropStaleTestDBs(ctx, admin)

	dbName := newTestDBName()
	if _, err := admin.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s TEMPLATE %s",
		pgx.Identifier{dbName}.Sanitize(), pgx.Identifier{templateDBName}.Sanitize())); err != nil {
		return nil, fmt.Errorf("create test database: %w", err)
	}

	dsn := replaceDatabase(adminDSN, dbName)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping test database: %w", err)
	}
	return &TestDatabase{Pool: pool, DSN: dsn, DBName: dbName}, nil
}

// startOrReuseContainer コンテナを起動または再利用し、管理DBへのDSNを返す
func startOrReuseContainer(ctx context.Context) (string, error) {
	if dsn := os.Getenv("TEST_DATABASE_ADMIN_DSN"); dsn != "" {
		// 既存の PostgreSQL（CI のサービスコンテナなど）を使う
		return dsn, nil
	}

	reuse := os.Getenv("TESTCONTAINERS_REUSE") != "false"
	if reuse {
		// 再利用コンテナは Ryuk に回収させない（プロセス終了後も残して次回に使う）
		_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	}

	opts := []testcontainers.ContainerCustomizer{
		postgres.WithDatabase(adminDBName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.BasicWaitStrategies(),
	}
	if reuse {
		opts = append(opts, testcontainers.WithReuseByName(containerName))
	}
	ctr, err := postgres.Run(ctx, PostgresImage, opts...)
	if err != nil {
		return "", fmt.Errorf("PostgreSQLコンテナ起動失敗: %w", err)
	}
	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return "", fmt.Errorf("接続文字列取得失敗: %w", err)
	}
	return dsn, nil
}

// ensureTemplate マイグレーション済みテンプレートDBを用意する。マイグレーションのハッシュが
// 変わっていれば作り直す。
func ensureTemplate(ctx context.Context, admin *pgx.Conn, adminDSN string, migrationSQL []byte) error {
	sum := sha256.Sum256(migrationSQL)
	wantHash := hex.EncodeToString(sum[:])

	exists, err := databaseExists(ctx, admin, templateDBName)
	if err != nil {
		return err
	}
	if exists {
		gotHash, err := readTemplateHash(ctx, replaceDatabase(adminDSN, templateDBName))
		if err == nil && gotHash == wantHash {
			return nil
		}
		if err := dropTemplate(ctx, admin); err != nil {
			return err
		}
	}
	return createTemplate(ctx, admin, adminDSN, migrationSQL, wantHash)
}

func databaseExists(ctx context.Context, admin *pgx.Conn, name string) (bool, error) {
	var exists bool
	if err := admin.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", name).Scan(&exists); err != nil {
		return false, fmt.Errorf("check database %s: %w", name, err)
	}
	return exists, nil
}

// dropTemplate 古いテンプレートを削除する（テンプレート属性を外してから強制削除する）
func dropTemplate(ctx context.Context, admin *pgx.Conn) error {
	name := pgx.Identifier{templateDBName}.Sanitize()
	if _, err := admin.Exec(ctx, fmt.Sprintf("ALTER DATABASE %s IS_TEMPLATE false", name)); err != nil {
		return fmt.Errorf("unmark template: %w", err)
	}
	if _, err := admin.Exec(ctx, fmt.Sprintf("DROP DATABASE %s WITH (FORCE)", name)); err != nil {
		return fmt.Errorf("drop stale template: %w", err)
	}
	return nil
}

// createTemplate テンプレートDBを作り、マイグレーションとハッシュを書き込んでテンプレート属性を付ける
func createTemplate(ctx context.Context, admin *pgx.Conn, adminDSN string, migrationSQL []byte, hash string) error {
	name := pgx.Identifier{templateDBName}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+name); err != nil {
		return fmt.Errorf("create template: %w", err)
	}
	if err := populateTemplate(ctx, replaceDatabase(adminDSN, templateDBName), migrationSQL, hash); err != nil {
		return err
	}
	if _, err := admin.Exec(ctx, fmt.Sprintf("ALTER DATABASE %s IS_TEMPLATE true", name)); err != nil {
		return fmt.Errorf("mark template: %w", err)
	}
	return nil
}

// populateTemplate テンプレートDBに接続してマイグレーションとハッシュを適用し、接続を閉じる
func populateTemplate(ctx context.Context, dsn string, migrationSQL []byte, hash string) error {
	tconn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect template: %w", err)
	}
	defer func() { _ = tconn.Close(ctx) }()

	stmts := []string{
		string(migrationSQL),
		"CREATE TABLE schema_meta (migration_hash text NOT NULL)",
	}
	for _, stmt := range stmts {
		if _, err := tconn.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("populate template: %w", err)
		}
	}
	if _, err := tconn.Exec(ctx, "INSERT INTO schema_meta (migration_hash) VALUES ($1)", hash); err != nil {
		return fmt.Errorf("record migration hash: %w", err)
	}
	return tconn.Close(ctx)
}

func readTemplateHash(ctx context.Context, dsn string) (string, error) {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return "", err
	}
	defer func() { _ = conn.Close(ctx) }()
	var hash string
	err = conn.QueryRow(ctx, "SELECT migration_hash FROM schema_meta LIMIT 1").Scan(&hash)
	return hash, err
}

// dropStaleTestDBs 名前に埋め込んだ作成時刻が古く、接続の無いテストDBを削除する
func dropStaleTestDBs(ctx context.Context, admin *pgx.Conn) {
	rows, err := admin.Query(ctx, `
		SELECT d.datname FROM pg_database d
		WHERE d.datname LIKE 'test\_%'
		  AND NOT EXISTS (SELECT 1 FROM pg_stat_activity a WHERE a.datname = d.datname)`)
	if err != nil {
		return
	}
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err == nil {
			names = append(names, n)
		}
	}
	rows.Close()
	cutoff := time.Now().Add(-staleTestDBAge).Unix()
	for _, n := range names {
		parts := strings.Split(n, "_")
		if len(parts) < 3 {
			continue
		}
		created, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || created > cutoff {
			continue
		}
		_, _ = admin.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", pgx.Identifier{n}.Sanitize()))
	}
}

func newTestDBName() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("test_%d_%s", time.Now().Unix(), hex.EncodeToString(b))
}

// replaceDatabase DSN のデータベース名部分を置き換える
func replaceDatabase(dsn, dbName string) string {
	// 形式: postgres://user:pass@host:port/dbname?params
	q := ""
	base := dsn
	if i := strings.Index(dsn, "?"); i >= 0 {
		base, q = dsn[:i], dsn[i:]
	}
	slash := strings.LastIndex(base, "/")
	if slash < 0 {
		return dsn
	}
	return base[:slash+1] + dbName + q
}

func connectWithRetry(ctx context.Context, dsn string, timeout time.Duration) (*pgx.Conn, error) {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		conn, err := pgx.Connect(ctx, dsn)
		if err == nil {
			if pingErr := conn.Ping(ctx); pingErr == nil {
				return conn, nil
			}
			_ = conn.Close(ctx)
		}
		lastErr = err
		time.Sleep(500 * time.Millisecond)
	}
	if lastErr == nil {
		lastErr = errors.New("timeout")
	}
	return nil, fmt.Errorf("データベース接続待機失敗: %w", lastErr)
}

// repoFile リポジトリルート（go.mod のある場所）からの相対パスを解決する
func repoFile(parts ...string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(append([]string{dir}, parts...)...), nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("go.mod が見つかりません")
		}
		dir = parent
	}
}
