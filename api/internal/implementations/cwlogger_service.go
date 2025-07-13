package implementations

import (
	"context"
	"log/slog"
	"os"
)

type CwLogger struct {
	loggerAPI LoggerAPI
}

type LoggerAPI interface {
	LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr)
}

func (cwl *CwLogger) Log(message string, level slog.Level, args ...any) error {
	var attrs []slog.Attr

	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			key = "!BAD_KEY"
		}

		if i+1 < len(args) {
			attrs = append(attrs, slog.Any(key, args[i+1]))
		}
	}

	cwl.loggerAPI.LogAttrs(context.Background(), level, message, attrs...)

	return nil
}

func NewSlogLogger() *CwLogger {
	loggerAPI := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return &CwLogger{
		loggerAPI,
	}
}
