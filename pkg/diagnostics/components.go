package diagnostics

import (
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
)

func SetupServer(rt core_runtime.Runtime) error {
	return rt.Add(
		&diagnosticsServer{
			isReady: rt.Ready,
			config:  rt.Config().Diagnostics,
			metrics: rt.Metrics(),
		},
	)
}
