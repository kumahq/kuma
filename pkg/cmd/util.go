package cmd

import (
	"context"

	"github.com/kumahq/kuma/pkg/core"
)

type RunCmdOpts struct {
	SetupSignalHandler func() context.Context
}

var DefaultRunCmdOpts = RunCmdOpts{
	SetupSignalHandler: core.SetupSignalHandler,
}
