package config

import (
	"log/slog"
	"os"
	"strings"
)

var Logger *slog.Logger

func InitLogger(levelStr string, verbose bool) {
	var level slog.Level
	if verbose {
		level = slog.LevelDebug
	} else {
		switch strings.ToUpper(levelStr) {
		case "DEBUG":
			level = slog.LevelDebug
		case "WARN":
			level = slog.LevelWarn
		case "ERROR":
			level = slog.LevelError
		default:
			level = slog.LevelInfo
		}
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})
	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}
