package tracing

import (
	"context"

	"go.elastic.co/apm/module/apmzap"
	"go.uber.org/zap"
)

func WrapZapLogger(z *zap.Logger) {
	z.WithOptions(zap.WrapCore((&apmzap.Core{}).WrapCore))
}

func WrapZapInfo(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().With(apmzap.TraceContext(ctx)...).Info(message, fields...)
}

func WrapZapWarn(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().With(apmzap.TraceContext(ctx)...).Warn(message, fields...)
}

func WrapZapDebug(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().With(apmzap.TraceContext(ctx)...).Debug(message, fields...)
}

func WrapZapError(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().With(apmzap.TraceContext(ctx)...).Error(message, fields...)
}

func WrapZapPanic(ctx context.Context, message string, fields ...zap.Field) {
	zap.L().With(apmzap.TraceContext(ctx)...).Panic(message, fields...)
}
