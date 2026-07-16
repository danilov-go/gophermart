// Package logger предоставляет глобальный логгер для приложения.
package logger

import (
	"go.uber.org/zap"
)

// Log определяет глобальный логгер приложения.
var Log *zap.Logger = zap.NewNop()

// Initialize инициализирует глобальный логгер с заданным уровнем детализации.
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl
	return nil
}
