package telemetry

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/ThinhDangDev/go-template/internal/boilerplate/config"

	"gopkg.in/natefinch/lumberjack.v2"
)

func NewLogger(cfg config.Config) (*slog.Logger, error) {
	level := parseLevel(cfg.LogLevel)
	writers := make([]io.Writer, 0, 2)

	if cfg.LogEnableConsole {
		writers = append(writers, os.Stdout)
	}

	if cfg.LogEnableFile {
		if err := os.MkdirAll(filepath.Dir(cfg.LogFile), 0o755); err != nil {
			return nil, err
		}

		writers = append(writers, &lumberjack.Logger{
			Filename:   cfg.LogFile,
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   true,
		})
	}

	if len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}

	handler := slog.NewJSONHandler(io.MultiWriter(writers...), &slog.HandlerOptions{
		Level:     level,
		AddSource: false,
	})

	logger := slog.New(handler).With(
		"service", cfg.ServiceName,
		"env", cfg.Environment,
		"version", cfg.Version,
	)
	slog.SetDefault(logger)

	return logger, nil
}

func parseLevel(level string) slog.Leveler {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
