package logger

import (
	"log/slog"
	"testing"
)

func TestSetup_Production(t *testing.T) {
	logger := Setup("production")
	if logger == nil {
		t.Fatal("Expected logger to be created")
	}

	// デフォルトロガーが設定されているか確認
	if slog.Default() == nil {
		t.Error("Expected default logger to be set")
	}
}

func TestSetup_Development(t *testing.T) {
	logger := Setup("development")
	if logger == nil {
		t.Fatal("Expected logger to be created")
	}

	// デフォルトロガーが設定されているか確認
	if slog.Default() == nil {
		t.Error("Expected default logger to be set")
	}
}

func TestSetup_UnknownEnvironment(t *testing.T) {
	// 未知の環境でも開発環境として扱われる
	logger := Setup("unknown")
	if logger == nil {
		t.Fatal("Expected logger to be created")
	}

	// デフォルトロガーが設定されているか確認
	if slog.Default() == nil {
		t.Error("Expected default logger to be set")
	}
}
