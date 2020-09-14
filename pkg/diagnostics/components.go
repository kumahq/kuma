package diagnostics

import (
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
)

func SetupServer(rt core_runtime.Runtime) error {
	return rt.Add(
		// diagnostics server
		&diagnosticsServer{
			port:    rt.Config().XdsServer.DiagnosticsPort,
			metrics: rt.Metrics(),
		},
	)
}
