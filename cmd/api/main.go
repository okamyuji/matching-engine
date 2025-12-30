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

	"github.com/yourorg/matching-engine/internal/core/matching"
	datingAPI "github.com/yourorg/matching-engine/internal/modules/dating/api"
	datingApp "github.com/yourorg/matching-engine/internal/modules/dating/application"
	datingInfra "github.com/yourorg/matching-engine/internal/modules/dating/infrastructure/mapper"
	datingRepo "github.com/yourorg/matching-engine/internal/modules/dating/infrastructure/repository"
	maAPI "github.com/yourorg/matching-engine/internal/modules/ma/api"
	maApp "github.com/yourorg/matching-engine/internal/modules/ma/application"
	maInfra "github.com/yourorg/matching-engine/internal/modules/ma/infrastructure/mapper"
	maRepo "github.com/yourorg/matching-engine/internal/modules/ma/infrastructure/repository"
	"github.com/yourorg/matching-engine/internal/shared/config"
	"github.com/yourorg/matching-engine/internal/shared/database"
	"github.com/yourorg/matching-engine/internal/shared/health"
	"github.com/yourorg/matching-engine/internal/shared/logger"
)

func main() {
	// 設定読み込み
	cfg := config.Load()

	// ロガー初期化
	logger.Setup(cfg.Env)
	slog.Info("starting matching engine", slog.String("env", cfg.Env))

	// データベース接続
	db, err := database.NewDB(&database.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		Database:        cfg.Database.Database,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		Debug:           cfg.Database.Debug,
	})
	if err != nil {
		slog.Error("failed to connect database", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		if err := database.Close(db); err != nil {
			slog.Error("failed to close database", slog.Any("error", err))
		}
	}()
	slog.Info("database connected")

	// Datingモジュール初期化
	slog.Info("initializing dating module")
	datingConfig, err := matching.LoadConfig("configs/dating/matching.json")
	if err != nil {
		slog.Error("failed to load dating config", slog.Any("error", err))
		os.Exit(1)
	}

	datingEngine, err := matching.NewConfigurableEngine(datingConfig)
	if err != nil {
		slog.Error("failed to create dating engine", slog.Any("error", err))
		os.Exit(1)
	}

	// Dating repositories
	datingUserRepo := datingRepo.NewUserRepository(db)
	datingProfileRepo := datingRepo.NewProfileRepository(db)
	datingPreferenceRepo := datingRepo.NewPreferenceRepository(db)
	datingLikeRepo := datingRepo.NewLikeRepository(db)
	datingMatchRepo := datingRepo.NewMatchRepository(db)

	// Dating mapper
	datingMapper := datingInfra.NewDatingFeatureMapper()

	// Dating services
	datingMatchingService := datingApp.NewDatingMatchingService(
		datingEngine,
		datingUserRepo,
		datingProfileRepo,
		datingPreferenceRepo,
		datingMatchRepo,
		datingMapper,
	)

	datingLikeService := datingApp.NewLikeService(
		datingLikeRepo,
		datingMatchRepo,
	)

	// Dating API handler
	datingHandler := datingAPI.NewHandler(datingMatchingService, datingLikeService)
	slog.Info("dating module initialized")

	// M&Aモジュール初期化
	slog.Info("initializing ma module")
	maConfig, err := matching.LoadConfig("configs/ma/matching.json")
	if err != nil {
		slog.Error("failed to load ma config", slog.Any("error", err))
		os.Exit(1)
	}

	maEngine, err := matching.NewConfigurableEngine(maConfig)
	if err != nil {
		slog.Error("failed to create ma engine", slog.Any("error", err))
		os.Exit(1)
	}

	// M&A repositories
	maCompanyRepo := maRepo.NewCompanyRepository(db)
	maFinancialsRepo := maRepo.NewFinancialsRepository(db)
	maInterestRepo := maRepo.NewInterestRepository(db)
	maMatchRepo := maRepo.NewMAMatchRepository(db)

	// M&A mapper
	maMapper := maInfra.NewMAFeatureMapper()

	// M&A services
	maSynergyCalculator := maApp.NewSynergyCalculator()

	maMatchingService := maApp.NewMAMatchingService(
		maEngine,
		maCompanyRepo,
		maFinancialsRepo,
		maInterestRepo,
		maMatchRepo,
		maMapper,
		maSynergyCalculator,
	)

	maValuationService := maApp.NewValuationService(maFinancialsRepo)

	// M&A API handler
	maHandler := maAPI.NewHandler(maMatchingService, maValuationService)
	slog.Info("ma module initialized")

	// ルーター設定
	mux := http.NewServeMux()

	// ヘルスチェック
	healthHandler := health.NewHandler(db)
	mux.HandleFunc("GET /health/live", healthHandler.LivenessHandler)
	mux.HandleFunc("GET /health/ready", healthHandler.ReadinessHandler)

	// Dating APIルート
	datingAPI.SetupRoutes(mux, datingHandler)

	// M&A APIルート
	maAPI.SetupRoutes(mux, maHandler)

	// HTTPサーバー設定
	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
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
