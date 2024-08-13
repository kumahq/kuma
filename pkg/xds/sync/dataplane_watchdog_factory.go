package sync

import (
	"context"
	"time"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/user"
	util_watchdog "github.com/kumahq/kuma/pkg/util/watchdog"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
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

func (d *dataplaneWatchdogFactory) New(dpKey model.ResourceKey) util_xds_v3.Watchdog {
	log := xdsServerLog.WithName("dataplane-sync-watchdog").WithValues("dataplaneKey", dpKey)
	dataplaneWatchdog := NewDataplaneWatchdog(d.deps, dpKey)
	return &util_watchdog.SimpleWatchdog{
		NewTicker: func() *time.Ticker {
			return time.NewTicker(d.refreshInterval)
		},
		OnTick: func(ctx context.Context) error {
			ctx = user.Ctx(ctx, user.ControlPlane)
			start := core.Now()
			result, err := dataplaneWatchdog.Sync(ctx)
			if err != nil {
				return err
			}
			d.xdsMetrics.XdsGenerations.
				WithLabelValues(string(result.ProxyType), string(result.Status)).
				Observe(float64(core.Now().Sub(start).Milliseconds()))
			return nil
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
