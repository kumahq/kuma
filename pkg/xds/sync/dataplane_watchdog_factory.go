package sync

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/kumahq/kuma/pkg/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	util_watchdog "github.com/kumahq/kuma/pkg/util/watchdog"
)

type dataplaneWatchdogFactory struct {
	xdsGenerations       prometheus.Summary
	xdsGenerationsErrors prometheus.Counter
	refreshInterval      time.Duration

	deps DataplaneWatchdogDependencies
}

func NewDataplaneWatchdogFactory(
	metrics core_metrics.Metrics,
	refreshInterval time.Duration,
	deps DataplaneWatchdogDependencies,
) (DataplaneWatchdogFactory, error) {
	xdsGenerations := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "xds_generation",
		Help:       "Summary of XDS Snapshot generation",
		Objectives: core_metrics.DefaultObjectives,
	})
	if err := metrics.Register(xdsGenerations); err != nil {
		return nil, err
	}
	xdsGenerationsErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "xds_generation_errors",
		Help: "Counter of errors during XDS generation",
	})
	if err := metrics.Register(xdsGenerationsErrors); err != nil {
		return nil, err
	}

	return &dataplaneWatchdogFactory{
		deps:                 deps,
		refreshInterval:      refreshInterval,
		xdsGenerations:       xdsGenerations,
		xdsGenerationsErrors: xdsGenerationsErrors,
	}, nil
}

func (d *dataplaneWatchdogFactory) New(key core_model.ResourceKey, streamId int64) util_watchdog.Watchdog {
	log := xdsServerLog.WithName("dataplane-sync-watchdog").WithValues("dataplaneKey", key)
	dataplaneWatchdog := NewDataplaneWatchdog(d.deps, key, streamId)
	return &util_watchdog.SimpleWatchdog{
		NewTicker: func() *time.Ticker {
			return time.NewTicker(d.refreshInterval)
		},
		OnTick: func() error {
			start := core.Now()
			defer func() {
				d.xdsGenerations.Observe(float64(core.Now().Sub(start).Milliseconds()))
			}()
			return dataplaneWatchdog.Sync()
		},
		OnError: func(err error) {
			d.xdsGenerationsErrors.Inc()
			log.Error(err, "OnTick() failed")
		},
		OnStop: func() {
			if err := dataplaneWatchdog.Cleanup(); err != nil {
				log.Error(err, "OnTick() failed")
			}
		},
	}
}
