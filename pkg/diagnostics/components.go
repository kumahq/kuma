package diagnostics

import (
	core_runtime "github.com/kumahq/kuma/v2/pkg/core/runtime"
	kuma_log "github.com/kumahq/kuma/v2/pkg/log"
)

func SetupServer(rt core_runtime.Runtime) error {
	return rt.Add(
		&diagnosticsServer{
			isReady:     rt.Ready,
			config:      rt.Config().Diagnostics,
			metrics:     rt.Metrics(),
			logRegistry: kuma_log.GlobalComponentLevelRegistry(),
		},
	)
}
