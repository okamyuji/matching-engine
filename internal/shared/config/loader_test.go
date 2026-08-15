package config

import (
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	for _, k := range []string{"ENV", "PORT", "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE", "DB_MAX_CONNS", "DB_MIN_CONNS", "DB_CONN_MAX_LIFETIME"} {
		t.Setenv(k, "")
	}
	cfg := Load()
	if cfg.Env != "development" || cfg.Port != 8080 {
		t.Errorf("既定値が違う: %+v", cfg)
	}
	if cfg.Database.Port != 5432 || cfg.Database.User != "postgres" || cfg.Database.SSLMode != "disable" {
		t.Errorf("DB 既定値が違う: %+v", cfg.Database)
	}
	if cfg.Database.ConnMaxLifetime != 300*time.Second {
		t.Errorf("ConnMaxLifetime = %v", cfg.Database.ConnMaxLifetime)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	t.Setenv("ENV", "production")
	t.Setenv("PORT", "9090")
	t.Setenv("DB_PORT", "6543")
	t.Setenv("DB_MAX_CONNS", "notanumber")
	t.Setenv("DB_CONN_MAX_LIFETIME", "10")
	cfg := Load()
	if cfg.Env != "production" || cfg.Port != 9090 || cfg.Database.Port != 6543 {
		t.Errorf("環境変数が反映されない: %+v", cfg)
	}
	if cfg.Database.MaxConns != 25 {
		t.Errorf("数値でない値は既定値に戻るべき: %d", cfg.Database.MaxConns)
	}
	if cfg.Database.ConnMaxLifetime != 10*time.Second {
		t.Errorf("ConnMaxLifetime = %v", cfg.Database.ConnMaxLifetime)
	}
}
