package logging

import (
	"io"
	"log/slog"
	"os"
)

var base *slog.Logger

type Logger struct {
	*slog.Logger
}

func Init() {
	if err := os.MkdirAll("logs", 0755); err != nil && !os.IsExist(err) {
		panic("can't create log dir")
	}

	allFile, err := os.OpenFile("logs/all.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		panic(err)
	}

	writer := io.MultiWriter(os.Stdout, allFile)
	handler := slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})

	base = slog.New(handler)
	slog.SetDefault(base)
	base.Info("logger initialized", "log_file", "logs/all.log")
}

func GetLogger() Logger {
	if base == nil {
		Init()
	}

	return Logger{base}
}

func (l Logger) With(args ...any) Logger {
	return Logger{l.Logger.With(args...)}
}

func (l Logger) Debug(msg string, args ...any) {
	l.Logger.Debug(msg, args...)
}

func (l Logger) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

func (l Logger) Warn(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

func (l Logger) Error(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}

func (l Logger) Fatal(msg string, args ...any) {
	l.Logger.Error(msg, args...)
	os.Exit(1)
}
