package logger

import (
	"io"
	"log/slog"
)

func NewJSON(out io.Writer, level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{Level: level}))
}
