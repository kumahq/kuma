package sync

import (
	"time"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	util_watchdog "github.com/kumahq/kuma/pkg/util/watchdog"
	xds_metrics "github.com/kumahq/kuma/pkg/xds/metrics"
)

type dataplaneWatchdogFactory struct {
	xdsMetrics      *xds_metrics.Metrics
	refreshInterval time.Duration

	deps DataplaneWatchdogDependencies
}

func NewDataplaneWatchdogFactory(
	xdsSyncMetrics *xds_metrics.Metrics,
	refreshInterval time.Duration,
	deps DataplaneWatchdogDependencies,
) (DataplaneWatchdogFactory, error) {
	return &dataplaneWatchdogFactory{
		deps:            deps,
		refreshInterval: refreshInterval,
		xdsMetrics:      xdsSyncMetrics,
	}, nil
}

func (d *dataplaneWatchdogFactory) New(key core_model.ResourceKey, streamId int64, proxyType mesh_proto.DpType) util_watchdog.Watchdog {
	log := xdsServerLog.WithName("dataplane-sync-watchdog").WithValues("dataplaneKey", key)
	dataplaneWatchdog := NewDataplaneWatchdog(d.deps, key, streamId, proxyType)
	return &util_watchdog.SimpleWatchdog{
		NewTicker: func() *time.Ticker {
			return time.NewTicker(d.refreshInterval)
		},
		OnTick: func() error {
			start := core.Now()
			defer func() {
				d.xdsMetrics.XdsGenerations.Observe(float64(core.Now().Sub(start).Milliseconds()))
			}()
			return dataplaneWatchdog.Sync()
		},
		OnError: func(err error) {
			d.xdsMetrics.XdsGenerationsErrors.Inc()
			log.Error(err, "OnTick() failed")
		},
		OnStop: func() {
			if err := dataplaneWatchdog.Cleanup(); err != nil {
				log.Error(err, "OnTick() failed")
			}
		},
	}
}
