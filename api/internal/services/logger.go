package services

import "log/slog"

type Severity string

type Logger interface {
	Log(message string, level slog.Level, args ...any) error
}
