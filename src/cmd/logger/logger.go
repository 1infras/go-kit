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

//Normal Logging
//Info
func Info(message string) {
	zap.L().Info(message)
}

//Warn
func Warn(message string) {
	zap.L().Warn(message)
}

//Error
func Error(message string) {
	zap.L().Error(message)
}

//Panic
func Panic(message string) {
	zap.L().Panic(message)
}

//Debug
func Debug(message string) {
	zap.L().Debug(message)
}

//Logging with APM tracing
//Info with APM tracing
func Infot(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().With(apmTraceContextWrapper(ctx)...).Info(message, fields...)
}

//Warn with APM tracing
func Warnt(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().With(apmTraceContextWrapper(ctx)...).Warn(message, fields...)
}

//Error with APM tracing
func Errort(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().With(apmTraceContextWrapper(ctx)...).Error(message, fields...)
}

//Debug with APM tracing
func Debugt(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().With(apmTraceContextWrapper(ctx)...).Debug(message, fields...)
}

//Panic with APM tracing
func Panict(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().With(apmTraceContextWrapper(ctx)...).Panic(message, fields...)
}

//Logger with format
//Info with format
func Infof(message string, args ...interface{}) {
	zap.S().Infof(message, args...)
}

//Warn with format
func Warnf(message string, args ...interface{}) {
	zap.S().Warnf(message, args...)
}

//Error with format
func Errorf(message string, args ...interface{}) {
	zap.S().Errorf(message, args...)
}

//Debug with format
func Debugf(message string, args ...interface{}) {
	zap.S().Debugf(message, args...)
}

//Panic with format
func Panicf(message string, args ...interface{}) {
	zap.S().Panicf(message, args...)
}

//Logging with sugared
//Info with sugared
func Infow(message string, keyAndValues ...interface{}) {
	zap.S().Infow(message, keyAndValues...)
}

//Warn with sugared
func Warnw(message string, keyAndValues ...interface{}) {
	zap.S().Warnw(message, keyAndValues...)
}

//Error with sugared
func Errorw(message string, keyAndValues ...interface{}) {
	zap.S().Errorw(message, keyAndValues...)
}

//Debug with sugared
func Debugw(message string, keyAndValues ...interface{}) {
	zap.S().Debugw(message, keyAndValues...)
}

//Panic with sugared
func Panicw(message string, keyAndValues ...interface{}) {
	zap.S().Panicw(message, keyAndValues...)
}