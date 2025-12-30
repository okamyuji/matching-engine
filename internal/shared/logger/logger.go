package logger

import (
	"log/slog"
	"os"
)

// Setup ロガーを設定する
func Setup(env string) *slog.Logger {
	var handler slog.Handler

	if env == "production" {
		// 本番環境: JSON形式
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		// 開発環境: テキスト形式
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}
