package server

import (
	"context"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/maps"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/kds/v2/reconcile"
	"github.com/kumahq/kuma/pkg/multitenant"
	util_watchdog "github.com/kumahq/kuma/pkg/util/watchdog"
)

type EventBasedWatchdog struct {
	Ctx                  context.Context
	Node                 *envoy_core.Node
	Listener             events.Listener
	Reconciler           reconcile.Reconciler
	ProvidedTypes        map[model.ResourceType]struct{}
	KdsGenerations       prometheus.Summary
	KdsGenerationsErrors prometheus.Counter
	Log                  logr.Logger
	FlushInterval        time.Duration
	FullResyncInterval   time.Duration
}

var _ util_watchdog.Watchdog = &EventBasedWatchdog{}

func (e *EventBasedWatchdog) Start(stop <-chan struct{}) {
	tenantID, _ := multitenant.TenantFromCtx(e.Ctx)
	flushTicker := time.NewTicker(e.FlushInterval)
	defer flushTicker.Stop()
	fullResyncTicker := time.NewTicker(e.FullResyncInterval)
	defer fullResyncTicker.Stop()

	// for the first reconcile assign all types
	changedTypes := maps.Clone(e.ProvidedTypes)

	for {
		select {
		case <-stop:
			if err := e.Reconciler.Clear(e.Ctx, e.Node); err != nil {
				e.Log.Error(err, "reconcile clear failed")
			}
			e.Listener.Close()
			return
		case <-flushTicker.C:
			if len(changedTypes) == 0 {
				continue
			}
			e.Log.V(1).Info("reconcile", "changedTypes", changedTypes)
			start := core.Now()
			if err := e.Reconciler.Reconcile(e.Ctx, e.Node, changedTypes); err != nil {
				e.Log.Error(err, "reconcile failed", "changedTypes", changedTypes)
				e.KdsGenerationsErrors.Inc()
			} else {
				e.KdsGenerations.Observe(float64(core.Now().Sub(start).Milliseconds()))
				changedTypes = map[model.ResourceType]struct{}{}
			}
		case <-fullResyncTicker.C:
			e.Log.V(1).Info("schedule full resync")
			changedTypes = maps.Clone(e.ProvidedTypes)
		case event := <-e.Listener.Recv():
			resChange, ok := event.(events.ResourceChangedEvent)
			if !ok {
				continue
			}
			if resChange.TenantID != tenantID {
				continue
			}
			if _, ok := e.ProvidedTypes[resChange.Type]; !ok {
				continue
			}
			e.Log.V(1).Info("schedule sync for type", "typ", resChange.Type)
			changedTypes[resChange.Type] = struct{}{}
		}
	}
}
