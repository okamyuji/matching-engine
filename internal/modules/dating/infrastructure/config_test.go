package infrastructure

import (
	"testing"

	"github.com/yourorg/matching-engine/internal/core/matching"
)

func TestLoadDatingConfig(t *testing.T) {
	config, err := matching.LoadConfig("../../../../configs/dating/matching.json")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if config.Version != "1.0" {
		t.Errorf("Version = %v, want 1.0", config.Version)
	}

	if config.Domain != "dating" {
		t.Errorf("Domain = %v, want dating", config.Domain)
	}

	if len(config.Scoring.Components) != 7 {
		t.Errorf("Components length = %v, want 7", len(config.Scoring.Components))
	}

	// 合計ウェイトがおよそ1.0であることを検証
	totalWeight := 0.0
	for _, comp := range config.Scoring.Components {
		totalWeight += comp.Weight
	}

	if totalWeight < 0.99 || totalWeight > 1.01 {
		t.Errorf("Total weight = %v, want approximately 1.0", totalWeight)
	}

	// この設定からエンジンが作成できることを検証
	engine, err := matching.NewConfigurableEngine(config)
	if err != nil {
		t.Fatalf("NewConfigurableEngine() error = %v", err)
	}

	if engine == nil {
		t.Error("NewConfigurableEngine() returned nil")
	}
}
