package log

import (
	"io"
	"os"
	"strings"

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
		out = &lumberjack.Logger{
			Filename:   output,
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		}

	}

	return &Logger{
		zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(cfg), zapcore.AddSync(out), lvl)),
	}, nil

}

func (l *Logger) Fatal(msg ...string) {
	l.Logger.Fatal(buildMsg(msg...))
}

func (l *Logger) Error(msg ...string) {
	l.Logger.Error(buildMsg(msg...))
}

func (l *Logger) Warn(msg ...string) {
	l.Logger.Warn(buildMsg(msg...))
}

func (l *Logger) Debug(msg ...string) {
	l.Logger.Debug(buildMsg(msg...))
}

func (l *Logger) Info(msg ...string) {
	l.Logger.Info(buildMsg(msg...))
}

func buildMsg(msg ...string) string {
	if len(msg) == 0 {
		return ""
	}
	b := strings.Builder{}
	for i := range msg {
		b.WriteString(msg[i])
		b.WriteString(" ")
	}
	return b.String()
}
