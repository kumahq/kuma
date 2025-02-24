package sync

import (
	"context"
	"errors"
	"time"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/user"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_watchdog "github.com/kumahq/kuma/pkg/util/watchdog"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
	xds_metrics "github.com/kumahq/kuma/pkg/xds/metrics"
)

func NewDataplaneWatchdogFactory(deps DataplaneWatchdogDependencies, refreshInterval time.Duration, xdsMetrics *xds_metrics.Metrics) DataplaneWatchdogFactory {
	return DataplaneWatchdogFactoryFunc(func(dpKey model.ResourceKey, fetchMeta func() *core_xds.DataplaneMetadata) util_xds_v3.Watchdog {
		l := xdsServerLog.WithName("dataplane-sync-watchdog").WithValues("key", dpKey)
		dataplaneWatchdog := NewDataplaneWatchdog(l, deps, dpKey)
		return &util_watchdog.SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				return time.NewTicker(refreshInterval)
			},
			OnTick: func(ctx context.Context) error {
				ctx = user.Ctx(ctx, user.ControlPlane)
				start := core.Now()
				meta := fetchMeta()
				if meta == nil {
					return errors.New("metadata cannot be nil")
				}
				result, err := dataplaneWatchdog.Sync(ctx, meta)
				if err != nil {
					return err
				}
				xdsMetrics.XdsGenerations.
					WithLabelValues(string(result.ProxyType), string(result.Status)).
					Observe(float64(core.Now().Sub(start).Milliseconds()))
				return nil
			},
			OnError: func(err error) {
				xdsMetrics.XdsGenerationsErrors.Inc()
				l.Error(err, "OnTick() failed")
			},
			OnStop: func() {
				if err := dataplaneWatchdog.Cleanup(); err != nil {
					l.Error(err, "OnTick() failed")
				}
			},
		}
	})
}
