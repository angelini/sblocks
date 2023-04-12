package log

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

var logKey = contextKey("log")

func shortTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	type appendTimeEncoder interface {
		AppendTimeLayout(time.Time, string)
	}

	layout := "15:04:05.000"

	if enc, ok := enc.(appendTimeEncoder); ok {
		enc.AppendTimeLayout(t, layout)
		return
	}

	enc.AppendString(t.Format(layout))
}

func Init(ctx context.Context) (context.Context, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeTime = shortTimeEncoder
	log, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, fmt.Errorf("cannot build zap logger: %w", err)
	}

	return context.WithValue(ctx, logKey, log), nil
}

func GetLogger(ctx context.Context) *zap.Logger {
	return ctx.Value(logKey).(*zap.Logger)
}

func Info(ctx context.Context, message string, fields ...zap.Field) {
	log := ctx.Value(logKey).(*zap.Logger)
	log.Info(message, fields...)
}

func Warn(ctx context.Context, message string, fields ...zap.Field) {
	log := ctx.Value(logKey).(*zap.Logger)
	log.Warn(message, fields...)
}

func Error(ctx context.Context, message string, fields ...zap.Field) {
	log := ctx.Value(logKey).(*zap.Logger)
	log.Error(message, fields...)
}
