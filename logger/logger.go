package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/1infras/go-kit/tracing"
)

// LogLevel -
type LogLevel string

const (
	// Panic -
	PanicLevel LogLevel = "panic"
	// Fatal -
	FatalLevel = "fatal"
	// Error -
	ErrorLevel = "error"
	// Warn -
	WarnLevel = "warn"
	// Info -
	InfoLevel = "info"
	// Debug -
	DebugLevel = "debug"
)

var (
	// Ensure this logger only created at once time
	syncOne sync.Once
)

// GetLogLevel - Transform logger level to zapcore level
func GetLogLevel(lvl LogLevel) zapcore.Level {
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

// InitLogger - Init a logger with logger level
func InitLogger(lvl LogLevel) {
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

		if tracing.Enabled {
			tracing.WrapZapLogger(logger)
		}

		zap.ReplaceGlobals(logger)
		zap.RedirectStdLog(logger)
	})
}

// Info -
func Info(message string, fields ...zap.Field) {
	zap.L().Info(message, fields...)
}

// Warn -
func Warn(message string, fields ...zap.Field) {
	zap.L().Warn(message, fields...)
}

// Error -
func Error(message string, fields ...zap.Field) {
	zap.L().Error(message, fields...)
}

// Panic -
func Panic(message string, fields ...zap.Field) {
	zap.L().Panic(message, fields...)
}

// Debug -
func Debug(message string, fields ...zap.Field) {
	zap.L().Debug(message, fields...)
}

// Infof -
func Infof(message string, args ...interface{}) {
	zap.S().Infof(message, args...)
}

// Warnf -
func Warnf(message string, args ...interface{}) {
	zap.S().Warnf(message, args...)
}

// Errorf -
func Errorf(message string, args ...interface{}) {
	zap.S().Errorf(message, args...)
}

// Debugf -
func Debugf(message string, args ...interface{}) {
	zap.S().Debugf(message, args...)
}

// Panicf -
func Panicf(message string, args ...interface{}) {
	zap.S().Panicf(message, args...)
}
