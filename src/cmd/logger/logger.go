package logger

import (
	"context"
	"go.elastic.co/apm/module/apmzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
)

const (
	PanicLevel int = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)

var (
	//Ensure this logger only created at once time
	syncOne sync.Once
)

func GetLogLevel(lvl int) zapcore.Level {
	zapLevel := zap.DebugLevel

	switch lvl {
	case PanicLevel:
		zapLevel = zap.PanicLevel
	case FatalLevel:
		zapLevel = zap.FatalLevel
	case ErrorLevel:
		zapLevel = zap.ErrorLevel
	case WarnLevel:
		zapLevel = zap.WarnLevel
	case InfoLevel:
		zapLevel = zap.InfoLevel
	case DebugLevel:
		zapLevel = zap.DebugLevel
	}

	return zapLevel
}

func InitLogger(lvl int) {
	syncOne.Do(func() {
		zapLevel := GetLogLevel(lvl)
		logger, err := zap.Config{
			Level:    zap.NewAtomicLevelAt(zapLevel),
			Encoding: "json",
			EncoderConfig: zapcore.EncoderConfig{
				MessageKey:    "message",
				LevelKey:      "level",
				EncodeLevel:   zapcore.CapitalLevelEncoder,
				TimeKey:       "time",
				EncodeTime:    zapcore.ISO8601TimeEncoder,
				EncodeCaller:  zapcore.ShortCallerEncoder,
				StacktraceKey: "stacktrace",
			},
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
			Sampling:         nil,
		}.Build()

		if err != nil {
			panic(err)
		}

		logger.WithOptions(zap.WrapCore((&apmzap.Core{}).WrapCore))

		zap.ReplaceGlobals(logger)
		zap.RedirectStdLog(logger)
	})
}

func apmTraceContextWrapper(ctx context.Context) []zapcore.Field {
	if ctx != nil {
		return apmzap.TraceContext(ctx)
	}
	return apmzap.TraceContext(context.Background())
}

func Info(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().With(apmTraceContextWrapper(ctx)...).Info(message, fields...)
}

func Infow(message string, keyAndValues ...interface{}) {
	zap.S().Infow(message, keyAndValues)
}

func Infof(message string, args ...interface{}) {
	zap.S().Infof(message, args)
}

func Warn(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().With(apmTraceContextWrapper(ctx)...).Warn(message, fields...)
}

func Warnw(message string, keyAndValues ...interface{}) {
	zap.S().Warnw(message, keyAndValues)
}

func Warnf(message string, args ...interface{}) {
	zap.S().Warnf(message, args)
}

func Debug(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().With(apmTraceContextWrapper(ctx)...).Debug(message, fields...)
}

func Debugw(message string, keyAndValues ...interface{}) {
	zap.S().Debugw(message, keyAndValues)
}

func Debugf(message string, args ...interface{}) {
	zap.S().Debugf(message, args)
}

func Error(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().With(apmTraceContextWrapper(ctx)...).Error(message, fields...)
}

func Errorw(message string, keyAndValues ...interface{}) {
	zap.S().Errorw(message, keyAndValues)
}

func Errorf(message string, args ...interface{}) {
	zap.S().Errorf(message, args)
}

func Panic(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().With(apmTraceContextWrapper(ctx)...).Panic(message, fields...)
}

func Panicw(message string, keyAndValues ...interface{}) {
	zap.S().Panicw(message, keyAndValues)
}

func Panicf(message string, args ...interface{}) {
	zap.S().Panicf(message, args)
}
