package cmd

import (
	"context"

	"github.com/kumahq/kuma/pkg/core"
)

type RunCmdOpts struct {
	// Stop signals are SIGINT and SIGTERM
	// We can start graceful shutdown when first context is closed and forcefully stop when the second one is closed.
	// Note that the handler closes usr2Received as soon as SIGTERM has been
	// received, exactly one SIGUSR2 is buffered and notifications are
	// non-blocking, so the guarantee is that at least one SIGUSR2 is delivered.
	SetupSignalHandler func() (firstStopSignalReceived, secondStopSignalReceived context.Context, usr2Received <-chan struct{})
}

var DefaultRunCmdOpts = RunCmdOpts{
	SetupSignalHandler: core.SetupSignalHandler,
}
