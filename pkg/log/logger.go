// Package log includes zap based logger with various output and log level.
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

// Logger describes Logger instance.
type Logger struct {
	*zap.Logger
	level zapcore.Level
}

// AddLoggerWith creates new logger instance.
// level - log level value string.
// output - log output string. If it is empty or "stdout" log will be written to the stdout.
// If output sets to stderr, log will be written to the stderr.
// Set file path to the output if you need write logs to the file. 3 backup files will be written,
// with rotation and compression.
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
		level:  lvl,
		Logger: zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(cfg), zapcore.AddSync(out), lvl)),
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

// Sync syncs log. Call it before stop your app.
func (l *Logger) Sync() {
	if err := l.Logger.Sync(); err != nil && !errors.Is(err, syscall.ENOTTY) {
		log.Printf("cannot sync logger: %v", err)
	}
}

// Fatal writes log with fatal log level.
func (l *Logger) Fatal(msg string, add ...string) {
	if l.level > zapcore.FatalLevel {
		return
	}
	l.Logger.Fatal(buildMsg(msg, add...))
}

// Error writes log with error log level.
func (l *Logger) Error(msg string, add ...string) {
	if l.level > zapcore.ErrorLevel {
		return
	}
	l.Logger.Error(buildMsg(msg, add...))
}

// Warn writes log with warn log level.
func (l *Logger) Warn(msg string, add ...string) {
	if l.level > zapcore.WarnLevel {
		return
	}
	l.Logger.Warn(buildMsg(msg, add...))
}

// Debug writes log with debug log level.
func (l *Logger) Debug(msg string, add ...string) {
	l.Logger.Debug(buildMsg(msg, add...))
}

// Info writes log with info log level.
func (l *Logger) Info(msg string, add ...string) {
	if l.level > zapcore.InfoLevel {
		return
	}
	l.Logger.Info(buildMsg(msg, add...))
}

func buildMsg(msg string, add ...string) (string, zap.Field) {
	return msg, zap.String("", strings.Join(add, " "))
}
