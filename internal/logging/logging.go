// Package logging provides logging setup for the CLI.
package logging

import (
	"context"
	"log/slog"
	"os"
)

// Level represents the logging level.
type Level int

const (
	LevelDebug Level = -4
	LevelInfo  Level = 0
	LevelWarn  Level = 4
	LevelError Level = 8
)

// Setup configures the global logger with the specified level.
// It uses slog (Go 1.21+) for structured logging.
func Setup(verbose bool) *slog.Logger {
	var level Level
	if verbose {
		level = LevelDebug
	} else {
		level = LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.Level(level),
	}))

	slog.SetDefault(logger)
	return logger
}

// SetupWithContext returns a context with the logger attached.
func SetupWithContext(ctx context.Context, verbose bool) context.Context {
	var level Level
	if verbose {
		level = LevelDebug
	} else {
		level = LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.Level(level),
	}))

	return context.WithValue(ctx, "logger", logger)
}
