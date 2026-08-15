package database

import (
	"context"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/okamyuji/matching-engine/internal/testutil"
)

func TestConfig_DSN(t *testing.T) {
	c := &Config{Host: "h", Port: 5432, User: "u", Password: "p@ss/w", Database: "d"}
	got := c.DSN()
	want := "postgres://u:p%40ss%2Fw@h:5432/d?sslmode=disable"
	if got != want {
		t.Errorf("DSN = %q, want %q", got, want)
	}
	c.SSLMode = "require"
	if c.DSN() != "postgres://u:p%40ss%2Fw@h:5432/d?sslmode=require" {
		t.Errorf("sslmode が反映されない: %s", c.DSN())
	}
}

func TestNewPool_And_FromDSN(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := NewPoolFromDSN(ctx, td.DSN)
	if err != nil {
		t.Fatalf("NewPoolFromDSN: %v", err)
	}
	pool.Close()

	u, err := url.Parse(td.DSN)
	if err != nil {
		t.Fatal(err)
	}
	port, _ := strconv.Atoi(u.Port())
	pw, _ := u.User.Password()
	cfg := &Config{Host: u.Hostname(), Port: port, User: u.User.Username(), Password: pw, Database: td.DBName, MaxConns: 3, MinConns: 1, ConnMaxLifetime: time.Minute}
	pool, err = NewPool(ctx, cfg)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		t.Errorf("ping: %v", err)
	}
	pool.Close()
}

func TestNewPool_Errors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := NewPoolFromDSN(ctx, "::not a dsn::"); err == nil {
		t.Error("不正 DSN でエラーになるべき")
	}
	// 到達不能なポート（接続拒否）
	if _, err := NewPool(ctx, &Config{Host: "127.0.0.1", Port: 1, User: "u", Password: "p", Database: "d"}); err == nil {
		t.Error("到達不能な DB でエラーになるべき")
	}
	if _, err := NewPoolFromDSN(ctx, "postgres://u:p@127.0.0.1:1/d?sslmode=disable"); err == nil {
		t.Error("到達不能な DSN でエラーになるべき")
	}
}
