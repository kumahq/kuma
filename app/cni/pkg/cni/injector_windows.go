package cni

import (
	"errors"

	"github.com/go-logr/logr"
)

func Inject(netns string, logger logr.Logger, intermediateConfig *IntermediateConfig) error {
	return errors.New("only implemented on linux")
}
