package context

import (
	"context"

	"github.com/go-logr/logr"

	"github.com/kumahq/kuma/pkg/core"
)

func NewContext(ctx context.Context, logger logr.Logger) context.Context {
	return logr.NewContext(ctx, logger)
}

func FromContext(ctx context.Context) logr.Logger {
	logger, err := logr.FromContext(ctx)
	if err != nil {
		core.Log.V(1).Error(err, "could not extract logger from the context - using core.Log")
		return core.Log
	}

	return logger
}
