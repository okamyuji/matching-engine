package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"

	_ "github.com/go-sql-driver/mysql"
)

// Config データベース接続設定
type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	Debug           bool
}

// NewDB 新しいデータベース接続を作成する
func NewDB(cfg *Config) (*bun.DB, error) {
	// DSN構築
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=UTC",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	// SQL DB接続
	sqldb, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// コネクションプール設定
	sqldb.SetMaxOpenConns(cfg.MaxOpenConns)
	sqldb.SetMaxIdleConns(cfg.MaxIdleConns)
	sqldb.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// 接続テスト
	if err := sqldb.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// BUN DB作成
	db := bun.NewDB(sqldb, mysqldialect.New())

	// デバッグモードは将来的に実装可能
	// if cfg.Debug {
	//     db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	// }

	return db, nil
}

// Close データベース接続を閉じる
func Close(db *bun.DB) error {
	return db.Close()
}
