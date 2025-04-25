package log

import (
	"errors"
	"io"
	"log"
	"os"
	"strings"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	*zap.Logger
}

func AddLoggerWith(level, output string) (*Logger, error) {
	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, err
	}
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var out io.Writer
	switch output {
	case "stdout", "":
		out = os.Stdout
	case "stderr":
		out = os.Stderr
	default:
		out = logRotator(output, 100, 3, true)
	}

	return &Logger{
		zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(cfg), zapcore.AddSync(out), lvl)),
	}, nil
}

func logRotator(path string, maxSize int, maxBackups int, compress bool) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   path,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     28,
		Compress:   compress,
	}
}

func (l *Logger) Sync() {
	if err := l.Logger.Sync(); err != nil && !errors.Is(err, syscall.ENOTTY) {
		log.Printf("cannot sync logger: %v", err)
	}
}

func (l *Logger) Fatal(msg string, add ...string) {
	l.Logger.Fatal(buildMsg(msg, add...))
}

func (l *Logger) Error(msg string, add ...string) {
	l.Logger.Error(buildMsg(msg, add...))
}

func (l *Logger) Warn(msg string, add ...string) {
	l.Logger.Warn(buildMsg(msg, add...))
}

func (l *Logger) Debug(msg string, add ...string) {
	l.Logger.Debug(buildMsg(msg, add...))
}

func (l *Logger) Info(msg string, add ...string) {
	l.Logger.Info(buildMsg(msg, add...))
}

func buildMsg(msg string, add ...string) (string, zap.Field) {
	return msg, zap.String("", strings.Join(add, " "))
}
