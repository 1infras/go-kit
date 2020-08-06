package logger

import (
	"context"
	"sync"

	"go.elastic.co/apm/module/apmzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	//PanicLevel - Lowest level is panic
	PanicLevel int = iota
	//FatalLevel - Lowest level is fatal
	FatalLevel
	//ErrorLevel - Lowest level is error
	ErrorLevel
	//WarnLevel - Lowest level is warn
	WarnLevel
	//InfoLevel - Lowest level is info
	InfoLevel
	//DebugLevel - Lowest level is debug
	DebugLevel
)

var (
	//Ensure this logger only created at once time
	syncOne sync.Once
)

//GetLogLevel - Transform logger level to zapcore level
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

//InitLogger - Init a logger with logger level
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

//Info -
func Info(message string) {
	zap.L().Info(message)
}

//Warn -
func Warn(message string) {
	zap.L().Warn(message)
}

//Error -
func Error(message string) {
	zap.L().Error(message)
}

//Panic -
func Panic(message string) {
	zap.L().Panic(message)
}

//Debug -
func Debug(message string) {
	zap.L().Debug(message)
}

//Infot - Info with APM tracing
func Infot(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().With(apmTraceContextWrapper(ctx)...).Info(message, fields...)
}

//Warnt - Warn with APM tracing
func Warnt(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().With(apmTraceContextWrapper(ctx)...).Warn(message, fields...)
}

//Errort - Error with APM tracing
func Errort(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().With(apmTraceContextWrapper(ctx)...).Error(message, fields...)
}

//Debugt - Debug with APM tracing
func Debugt(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().With(apmTraceContextWrapper(ctx)...).Debug(message, fields...)
}

//Panict - Panic with APM tracing
func Panict(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().With(apmTraceContextWrapper(ctx)...).Panic(message, fields...)
}

//Infof - Info with format
func Infof(message string, args ...interface{}) {
	zap.S().Infof(message, args...)
}

//Warnf - Warn with format
func Warnf(message string, args ...interface{}) {
	zap.S().Warnf(message, args...)
}

//Errorf - Error with format
func Errorf(message string, args ...interface{}) {
	zap.S().Errorf(message, args...)
}

//Debugf - Debug with format
func Debugf(message string, args ...interface{}) {
	zap.S().Debugf(message, args...)
}

//Panicf - Panic with format
func Panicf(message string, args ...interface{}) {
	zap.S().Panicf(message, args...)
}

//Infow - Info with sugared
func Infow(message string, keyAndValues ...interface{}) {
	zap.S().Infow(message, keyAndValues...)
}

//Warnw - Warn with sugared
func Warnw(message string, keyAndValues ...interface{}) {
	zap.S().Warnw(message, keyAndValues...)
}

//Errorw - Error with sugared
func Errorw(message string, keyAndValues ...interface{}) {
	zap.S().Errorw(message, keyAndValues...)
}

//Debugw - Debug with sugared
func Debugw(message string, keyAndValues ...interface{}) {
	zap.S().Debugw(message, keyAndValues...)
}

//Panicw - Panic with sugared
func Panicw(message string, keyAndValues ...interface{}) {
	zap.S().Panicw(message, keyAndValues...)
}
