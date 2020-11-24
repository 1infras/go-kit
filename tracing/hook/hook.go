package hook

import (
	"context"

	"go.elastic.co/apm"

	"github.com/1infras/go-kit/lib/hook/common"
)

// hook is an implementation of common.HookProcess that reports command name as spans to Elastic APM.
type hook struct {
	spanName string
}

// NewHook returns a common.HookProcess that reports cmdName as spans to Elastic APM.
func NewHook(spanName string) common.HookProcess {
	return &hook{
		spanName: spanName,
	}
}

func (_this *hook) BeforeProcess(ctx context.Context, name string) context.Context {
	_, ctx = apm.StartSpan(ctx, name, _this.spanName)
	return ctx
}

func (_this *hook) AfterProcess(ctx context.Context, name string) {
	if span := apm.SpanFromContext(ctx); span != nil {
		span.End()
	}
}
