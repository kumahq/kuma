package cni

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
)

func Inject(ctx context.Context, netns string, intermediateConfig *IntermediateConfig, logger logr.Logger) error {
	return errors.New("only implemented on linux")
}
