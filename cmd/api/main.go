package main

import (
	"context"
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
		MaxConns:        int32(cfg.Database.MaxConns), //nolint:gosec // 設定値
		MinConns:        int32(cfg.Database.MinConns), //nolint:gosec // 設定値
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	})
	dbCancel()
	if err != nil {
		slog.Error("failed to connect database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("database connected")

	// ルーター構築（モジュールの配線は internal/app に集約）
	handler, err := app.NewRouter(db, app.Options{
		DatingConfigPath: "configs/dating/matching.json",
		MAConfigPath:     "configs/ma/matching.json",
	})
	if err != nil {
		slog.Error("failed to build router", slog.Any("error", err))
		os.Exit(1)
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

	// Graceful shutdown設定
	go func() {
		slog.Info("server listening", slog.String("addr", addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// シグナル待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")

	// Graceful shutdown実行
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server shutdown error", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("server stopped gracefully")
}
