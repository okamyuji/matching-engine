package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/okamyuji/matching-engine/internal/app"
	"github.com/okamyuji/matching-engine/internal/shared/config"
	"github.com/okamyuji/matching-engine/internal/shared/database"
	"github.com/okamyuji/matching-engine/internal/shared/logger"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", slog.Any("error", err))
		os.Exit(1)
	}
}

// run サーバーの起動から停止までを行う。エラーは呼び出し元で終了コードに変換する
func run() error {
	// 設定読み込み
	cfg := config.Load()

	// ロガー初期化
	logger.Setup(cfg.Env)
	slog.Info("starting matching engine", slog.String("env", cfg.Env))

	// データベース接続
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 30*time.Second)
	db, err := database.NewPool(dbCtx, &database.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		Database:        cfg.Database.Database,
		SSLMode:         cfg.Database.SSLMode,
		MaxConns:        clampConns(cfg.Database.MaxConns),
		MinConns:        clampConns(cfg.Database.MinConns),
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	})
	dbCancel()
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()
	slog.Info("database connected")

	// ルーター構築（モジュールの配線は internal/app に集約）
	handler, err := app.NewRouter(db, app.Options{
		DatingConfigPath: "configs/dating/matching.json",
		MAConfigPath:     "configs/ma/matching.json",
	})
	if err != nil {
		return fmt.Errorf("build router: %w", err)
	}
	slog.Info("modules initialized")

	// HTTPサーバー設定
	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// サーバー起動。エラーはチャネルで受け取り、シグナルと合わせて待つ
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("server listening", slog.String("addr", addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	case <-quit:
	}

	slog.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}
	slog.Info("server stopped gracefully")
	return nil
}

// clampConns 環境変数由来の接続数を pgxpool が受け付ける範囲（1〜1000）に収めて int32 にする
func clampConns(n int) int32 {
	const maxConns = 1000
	if n < 1 {
		return 1
	}
	if n > maxConns {
		return maxConns
	}
	return int32(n)
}
