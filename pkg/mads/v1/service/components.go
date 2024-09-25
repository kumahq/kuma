package service

import (
	"context"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	meshmetrics_generator "github.com/kumahq/kuma/pkg/mads/v1/generator"
	mads_reconcile "github.com/kumahq/kuma/pkg/mads/v1/reconcile"
	util_watchdog "github.com/kumahq/kuma/pkg/util/watchdog"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

type restReconcilerCallbacks struct {
	reconciler mads_reconcile.Reconciler
	log        logr.Logger
}

func (r *restReconcilerCallbacks) OnFetchRequest(ctx context.Context, request util_xds.DiscoveryRequest) error {
	nodei := request.Node()

	node, ok := nodei.(*envoy_core.Node)
	if !ok {
		return errors.Errorf("expecting a v3 Node, got: %v", nodei)
	}

	knownClients := r.reconciler.KnownClientIds()
	if !knownClients[node.Id] {
		node.Id = meshmetrics_generator.DefaultKumaClientId
	}

	if r.reconciler.NeedsReconciliation(node) {
		return r.reconciler.Reconcile(ctx, r.log)
	}
	return nil
}

func (r *restReconcilerCallbacks) OnFetchResponse(request util_xds.DiscoveryRequest, response util_xds.DiscoveryResponse) {
}

func NewReconcilerRestCallbacks(reconciler mads_reconcile.Reconciler, log logr.Logger) util_xds.RestCallbacks {
	return &restReconcilerCallbacks{reconciler: reconciler, log: log}
}

func NewSyncTracker(reconciler mads_reconcile.Reconciler, refresh time.Duration, log logr.Logger) envoy_xds.Callbacks {
	return util_xds_v3.NewWatchdogCallbacks(func(_ context.Context, node *envoy_core.Node, streamID int64) (util_xds_v3.Watchdog, error) {
		log := log.WithValues("streamID", streamID, "node", node)
		return &util_watchdog.SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				return time.NewTicker(refresh)
			},
			OnTick: func(ctx context.Context) error {
				log.V(1).Info("on tick")
				return reconciler.Reconcile(ctx, log)
			},
			OnError: func(err error) {
				log.Error(err, "OnTick() failed")
			},
		}, nil
	})
}
