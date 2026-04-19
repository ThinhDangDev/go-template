package logger

import (
	"go.uber.org/zap"
)

// Logger interface for structured logging
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Fatal(msg string, keysAndValues ...interface{})
	Sync() error
}

type zapLogger struct {
	logger *zap.SugaredLogger
}

// NewZapLogger creates a new Zap logger
func NewZapLogger() Logger {
	logger, _ := zap.NewProduction()
	return &zapLogger{logger: logger.Sugar()}
}

func (l *zapLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Infow(msg, keysAndValues...)
}

func (l *zapLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Errorw(msg, keysAndValues...)
}

func (l *zapLogger) Fatal(msg string, keysAndValues ...interface{}) {
	l.logger.Fatalw(msg, keysAndValues...)
}

func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}
