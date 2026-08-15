// Package app サーバーの配線（依存関係の組み立てとルーティング）を提供する。
// cmd/api と E2E テストの両方から使う。
package app

import (
	"fmt"
	"net/http"

	"github.com/okamyuji/matching-engine/internal/core/matching"
	datingAPI "github.com/okamyuji/matching-engine/internal/modules/dating/api"
	datingApp "github.com/okamyuji/matching-engine/internal/modules/dating/application"
	datingMapper "github.com/okamyuji/matching-engine/internal/modules/dating/infrastructure/mapper"
	datingRepo "github.com/okamyuji/matching-engine/internal/modules/dating/infrastructure/repository"
	maAPI "github.com/okamyuji/matching-engine/internal/modules/ma/api"
	maApp "github.com/okamyuji/matching-engine/internal/modules/ma/application"
	maMapper "github.com/okamyuji/matching-engine/internal/modules/ma/infrastructure/mapper"
	maRepo "github.com/okamyuji/matching-engine/internal/modules/ma/infrastructure/repository"
	"github.com/okamyuji/matching-engine/internal/shared/health"
)

// DB dating と ma の両リポジトリが必要とするデータベース操作。*pgxpool.Pool が満たす
type DB interface {
	datingRepo.DB
	maRepo.DB
	health.Pinger
}

// Options ルーター構築の設定
type Options struct {
	// DatingConfigPath dating モジュールのマッチング設定ファイル
	DatingConfigPath string
	// MAConfigPath ma モジュールのマッチング設定ファイル
	MAConfigPath string
}

// NewRouter 依存関係を組み立て、全ルートを登録した http.Handler を返す
func NewRouter(db DB, opts Options) (http.Handler, error) {
	datingConfig, err := matching.LoadConfig(opts.DatingConfigPath)
	if err != nil {
		return nil, fmt.Errorf("load dating config: %w", err)
	}
	datingEngine, err := matching.NewConfigurableEngine(datingConfig)
	if err != nil {
		return nil, fmt.Errorf("create dating engine: %w", err)
	}
	datingHandler := datingAPI.NewHandler(
		datingApp.NewDatingMatchingService(
			datingEngine,
			datingRepo.NewUserRepository(db),
			datingRepo.NewProfileRepository(db),
			datingRepo.NewPreferenceRepository(db),
			datingRepo.NewMatchRepository(db),
			datingMapper.NewDatingFeatureMapper(),
		),
		datingApp.NewLikeService(datingRepo.NewLikeRepository(db), datingRepo.NewMatchRepository(db)),
	)

	maConfig, err := matching.LoadConfig(opts.MAConfigPath)
	if err != nil {
		return nil, fmt.Errorf("load ma config: %w", err)
	}
	maEngine, err := matching.NewConfigurableEngine(maConfig)
	if err != nil {
		return nil, fmt.Errorf("create ma engine: %w", err)
	}
	maFinancialsRepo := maRepo.NewFinancialsRepository(db)
	maHandler := maAPI.NewHandler(
		maApp.NewMAMatchingService(
			maEngine,
			maRepo.NewCompanyRepository(db),
			maFinancialsRepo,
			maRepo.NewInterestRepository(db),
			maRepo.NewMAMatchRepository(db),
			maMapper.NewMAFeatureMapper(),
			maApp.NewSynergyCalculator(),
		),
		maApp.NewValuationService(maFinancialsRepo),
	)

	mux := http.NewServeMux()
	healthHandler := health.NewHandler(db)
	mux.HandleFunc("GET /health/live", healthHandler.LivenessHandler)
	mux.HandleFunc("GET /health/ready", healthHandler.ReadinessHandler)
	datingAPI.SetupRoutes(mux, datingHandler)
	maAPI.SetupRoutes(mux, maHandler)
	return mux, nil
}
