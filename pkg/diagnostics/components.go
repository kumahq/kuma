package diagnostics

import (
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
)

func SetupServer(rt core_runtime.Runtime) error {
	return rt.Add(
		&diagnosticsServer{
			metrics:        rt.Metrics(),
			port:           rt.Config().Diagnostics.ServerPort,
			debugEndpoints: rt.Config().Diagnostics.DebugEndpoints,
		},
	)
}
