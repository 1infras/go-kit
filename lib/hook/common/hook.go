package common

import "context"

type HookProcess interface {
	BeforeProcess(ctx context.Context, name string) context.Context
	AfterProcess(ctx context.Context, name string)
}

// Hook
type Hook struct {
	hooks []HookProcess
}

func (_this *Hook) AddHook(hook HookProcess) {
	_this.hooks = append(_this.hooks, hook)
}

func (_this *Hook) Process(ctx context.Context, fn func(), name string) {
	for _, h := range _this.hooks {
		ctx = h.BeforeProcess(ctx, name)
	}

	fn()

	for _, h := range _this.hooks {
		h.AfterProcess(ctx, name)
	}
}
