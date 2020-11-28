package tracing

import (
	"context"

	"go.elastic.co/apm"
)

func WrapError(ctx context.Context, err error) error {
	apm.CaptureError(ctx, err).Send()
	return err
}
