package server

import (
	"context"
	"errors"
	"strings"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/go-logr/logr"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/kds/v2/reconcile"
	"github.com/kumahq/kuma/pkg/multitenant"
	util_maps "github.com/kumahq/kuma/pkg/util/maps"
)

type EventBasedWatchdog struct {
	Node                *envoy_core.Node
	EventBus            events.EventBus
	Reconciler          reconcile.Reconciler
	ProvidedTypes       map[model.ResourceType]struct{}
	Metrics             *Metrics
	Log                 logr.Logger
	NewFlushTicker      func() *time.Ticker
	NewFullResyncTicker func() *time.Ticker
}

func (e *EventBasedWatchdog) Start(ctx context.Context) {
	tenantID, _ := multitenant.TenantFromCtx(ctx)
	listener := e.EventBus.Subscribe(func(event events.Event) bool {
		switch ev := event.(type) {
		case events.ResourceChangedEvent:
			_, ok := e.ProvidedTypes[ev.Type]
			return ok && ev.TenantID == tenantID
		case events.TriggerKDSResyncEvent:
			return ev.NodeID == e.Node.Id
		}
		return false
	})
	flushTicker := e.NewFlushTicker()
	defer flushTicker.Stop()
	fullResyncTicker := e.NewFullResyncTicker()
	defer fullResyncTicker.Stop()

	// for the first reconcile assign all types
	changedTypes := util_maps.Clone(e.ProvidedTypes)
	reasons := map[string]struct{}{
		ReasonResync: {},
	}

	for {
		select {
		case <-ctx.Done():
			if err := e.Reconciler.Clear(e.Node); err != nil {
				e.Log.Error(err, "reconcile clear failed")
			}
			listener.Close()
			return
		case <-flushTicker.C:
			if len(changedTypes) == 0 {
				continue
			}
			reason := strings.Join(util_maps.SortedKeys(reasons), "_and_")
			e.Log.V(1).Info("reconcile", "changedTypes", changedTypes, "reason", reason)
			start := core.Now()
			err, changed := e.Reconciler.Reconcile(ctx, e.Node, changedTypes, e.Log)
			if err != nil && !errors.Is(err, context.Canceled) {
				e.Log.Error(err, "reconcile failed", "changedTypes", changedTypes, "reason", reason)
				e.Metrics.KdsGenerationErrors.Inc()
			} else {
				result := ResultNoChanges
				if changed {
					result = ResultChanged
				}
				// we want to combine reason. One of the reasons we introduce this metric is to check if we need full resync
				// If we just keep a single reason, we might get into races where full resync ticker runs,
				// then listener, and we would lose information what triggered flush.
				e.Metrics.KdsGenerations.WithLabelValues(reason, result).Observe(float64(core.Now().Sub(start).Milliseconds()))
				changedTypes = map[model.ResourceType]struct{}{}
				reasons = map[string]struct{}{}
			}
		case <-fullResyncTicker.C:
			e.Log.V(1).Info("schedule full resync")
			changedTypes = util_maps.Clone(e.ProvidedTypes)
			reasons[ReasonResync] = struct{}{}
		case event := <-listener.Recv():
			switch ev := event.(type) {
			case events.ResourceChangedEvent:
				e.Log.V(1).Info("schedule sync for type", "typ", ev.Type, "event", "ResourceChanged")
				changedTypes[ev.Type] = struct{}{}
				reasons[ReasonEvent] = struct{}{}
			case events.TriggerKDSResyncEvent:
				e.Log.V(1).Info("schedule sync for type", "typ", ev.Type, "event", "TriggerKDSResync")
				changedTypes[ev.Type] = struct{}{}
				reasons[ReasonEvent] = struct{}{}
			}
		}
	}
}
