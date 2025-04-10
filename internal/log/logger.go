package log

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

func NewEmpty() *Logger {
	return &Logger{
		Logger: zap.NewNop(),
	}
}

func AddLoggerWith(level, path string) (*Logger, error) {
	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, err
	}
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder

	if path == "" {
		return &Logger{
			zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(cfg), zapcore.AddSync(os.Stdout), lvl)),
		}, nil
	}

	logFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}

	return &Logger{
		zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(cfg), zapcore.AddSync(logFile), lvl)),
	}, nil

}

func StringField(key string, value string) zap.Field {
	return zap.String(key, value)
}

func IntField(key string, value int) zap.Field {
	return zap.Int(key, value)
}

func DurationField(key string, value time.Duration) zap.Field {
	return zap.Duration(key, value)
}
