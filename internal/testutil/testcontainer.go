package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
)

var (
	sharedContainer *mysql.MySQLContainer
	sharedDB        *bun.DB
	containerMutex  sync.Mutex
	containerOnce   sync.Once
)

// TestDatabase テストデータベース情報
type TestDatabase struct {
	Container *mysql.MySQLContainer
	DB        *bun.DB
	DSN       string
}

// GetSharedTestDB 共有テストデータベースを取得する
// すべてのテストで同じコンテナを使用するため、起動コストを削減
func GetSharedTestDB(t *testing.T) *TestDatabase {
	t.Helper()
	containerMutex.Lock()
	defer containerMutex.Unlock()

	// 既に初期化済みならそのまま返す
	if sharedDB != nil && sharedContainer != nil {
		return &TestDatabase{
			Container: sharedContainer,
			DB:        sharedDB,
			DSN:       buildDSN(t, sharedContainer),
		}
	}

	// 初回のみコンテナを起動
	containerOnce.Do(func() {
		ctx := context.Background()

		// MySQLコンテナ起動
		container, err := mysql.Run(ctx,
			"mysql:8.0",
			mysql.WithDatabase("testdb"),
			mysql.WithUsername("testuser"),
			mysql.WithPassword("testpass"),
		)
		if err != nil {
			t.Fatalf("MySQLコンテナ起動失敗: %v", err)
		}

		// DSN取得
		dsn := buildDSN(t, container)

		// データベース接続
		sqldb, err := sql.Open("mysql", dsn)
		if err != nil {
			if termErr := container.Terminate(ctx); termErr != nil {
				t.Logf("コンテナ終了警告: %v", termErr)
			}
			t.Fatalf("データベース接続失敗: %v", err)
		}

		// 接続プール設定
		sqldb.SetMaxOpenConns(10)
		sqldb.SetMaxIdleConns(5)
		sqldb.SetConnMaxLifetime(5 * time.Minute)

		// BUN初期化
		db := bun.NewDB(sqldb, mysqldialect.New())

		// 接続テスト（リトライあり）
		if err := waitForDB(db, 30*time.Second); err != nil {
			if closeErr := sqldb.Close(); closeErr != nil {
				t.Logf("データベースクローズ警告: %v", closeErr)
			}
			if termErr := container.Terminate(ctx); termErr != nil {
				t.Logf("コンテナ終了警告: %v", termErr)
			}
			t.Fatalf("データベース接続待機失敗: %v", err)
		}

		// マイグレーション実行
		if err := runMigrations(db); err != nil {
			if closeErr := db.Close(); closeErr != nil {
				t.Logf("データベースクローズ警告: %v", closeErr)
			}
			if termErr := container.Terminate(ctx); termErr != nil {
				t.Logf("コンテナ終了警告: %v", termErr)
			}
			t.Fatalf("マイグレーション実行失敗: %v", err)
		}

		sharedContainer = container
		sharedDB = db

		// Ryukが自動的にコンテナをクリーンアップするため、明示的なCleanupは不要
	})

	return &TestDatabase{
		Container: sharedContainer,
		DB:        sharedDB,
		DSN:       buildDSN(t, sharedContainer),
	}
}

// CleanTables テーブルデータをクリーンアップする
func (td *TestDatabase) CleanTables(t *testing.T) {
	t.Helper()

	ctx := context.Background()
	// 外部キー制約を考慮した順序で削除
	tables := []string{
		// Dating tables
		"dating_matches",
		"dating_likes",
		"dating_profile_tags",
		"dating_profile_photos",
		"dating_preference_prefectures",
		"dating_preference_educations",
		"dating_preference_marriage_desires",
		"dating_preference_smoking_statuses",
		"dating_preference_drinking_statuses",
		"dating_preferences",
		"dating_profiles",
		"dating_users",
		// M&A tables
		"ma_matches",
		"ma_interests",
		"ma_company_markets",
		"ma_company_technologies",
		"ma_criteria_industries",
		"ma_matching_criteria",
		"ma_financials",
		"ma_companies",
	}

	for _, table := range tables {
		_, err := td.DB.NewRaw(fmt.Sprintf("DELETE FROM %s", table)).Exec(ctx)
		if err != nil {
			t.Logf("テーブル %s のクリーンアップ警告: %v", table, err)
		}
	}
}

// SeedTestData テストデータを投入する
func (td *TestDatabase) SeedTestData(t *testing.T) {
	t.Helper()

	// まずクリーンアップ
	td.CleanTables(t)

	// 複数の可能なパスを試す
	possiblePaths := []string{
		"db/testdata/seed.sql",
		"../../db/testdata/seed.sql",
		"../../../db/testdata/seed.sql",
		"../../../../db/testdata/seed.sql",
		"../../../../../db/testdata/seed.sql",
		"../../../../../../db/testdata/seed.sql",
	}

	var seedSQL []byte
	var err error
	for _, path := range possiblePaths {
		seedSQL, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}
	if err != nil {
		t.Fatalf("シードデータ読み込み失敗: %v", err)
	}

	ctx := context.Background()
	_, err = td.DB.ExecContext(ctx, string(seedSQL))
	if err != nil {
		t.Fatalf("シードデータ投入失敗: %v", err)
	}
}

// buildDSN DSN文字列を構築する
func buildDSN(t *testing.T, container *mysql.MySQLContainer) string {
	t.Helper()

	ctx := context.Background()
	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("コンテナホスト取得失敗: %v", err)
	}

	port, err := container.MappedPort(ctx, "3306/tcp")
	if err != nil {
		t.Fatalf("コンテナポート取得失敗: %v", err)
	}

	return fmt.Sprintf("testuser:testpass@tcp(%s:%s)/testdb?parseTime=true&multiStatements=true", host, port.Port())
}

// waitForDB データベース接続を待機する
func waitForDB(db *bun.DB, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("データベース接続タイムアウト")
		case <-ticker.C:
			if err := db.PingContext(ctx); err == nil {
				return nil
			}
		}
	}
}

// runMigrations マイグレーションを実行する
func runMigrations(db *bun.DB) error {
	// 複数の可能なパスを試す
	possiblePaths := []string{
		"db/migrations/001_create_tables.sql",
		"../../db/migrations/001_create_tables.sql",
		"../../../db/migrations/001_create_tables.sql",
		"../../../../db/migrations/001_create_tables.sql",
		"../../../../../db/migrations/001_create_tables.sql",
		"../../../../../../db/migrations/001_create_tables.sql",
	}

	var migrationSQL []byte
	var err error
	for _, path := range possiblePaths {
		migrationSQL, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("マイグレーションファイル読み込み失敗: %w", err)
	}

	ctx := context.Background()
	_, err = db.ExecContext(ctx, string(migrationSQL))
	if err != nil {
		return fmt.Errorf("マイグレーション実行失敗: %w", err)
	}

	return nil
}
