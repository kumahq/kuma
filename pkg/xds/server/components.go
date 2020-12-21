package server

import (
	"time"

	xds_callbacks "github.com/kumahq/kuma/pkg/xds/server/callbacks"

	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
)

func DefaultDataplaneStatusTracker(rt core_runtime.Runtime) xds_callbacks.DataplaneStatusTracker {
	return xds_callbacks.NewDataplaneStatusTracker(rt, func(accessor xds_callbacks.SubscriptionStatusAccessor) xds_callbacks.DataplaneInsightSink {
		return xds_callbacks.NewDataplaneInsightSink(
			accessor,
			func() *time.Ticker {
				return time.NewTicker(rt.Config().XdsServer.DataplaneStatusFlushInterval)
			},
			rt.Config().XdsServer.DataplaneStatusFlushInterval/10,
			xds_callbacks.NewDataplaneInsightStore(rt.ResourceManager()),
		)
	})
}
