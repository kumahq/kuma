package hds

import (
	"context"
	"sync"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_service_health "github.com/envoyproxy/go-control-plane/envoy/service/health/v3"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/util/watchdog"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	dp_server "github.com/kumahq/kuma/pkg/config/dp-server"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/xds"
	envoy_cache "github.com/kumahq/kuma/pkg/hds/cache"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
)

type streams struct {
	watchdogCancel context.CancelFunc
	activeStreams  map[xds.StreamID]bool
}

type tracker struct {
	resourceManager manager.ResourceManager
	config          *dp_server.HdsConfig
	reconciler      *reconciler

	sync.RWMutex       // protects access to the fields below
	streamsAssociation map[xds.StreamID]core_model.ResourceKey
	dpStreams          map[core_model.ResourceKey]streams
}

func NewTracker(
	resourceManager manager.ResourceManager,
	readOnlyResourceManager manager.ReadOnlyResourceManager,
	cache util_xds_v3.SnapshotCache,
	config *dp_server.HdsConfig,
) Callbacks {
	return &tracker{
		resourceManager:    resourceManager,
		streamsAssociation: map[xds.StreamID]core_model.ResourceKey{},
		dpStreams:          map[core_model.ResourceKey]streams{},
		config:             config,
		reconciler: &reconciler{
			cache:     cache,
			hasher:    &hasher{},
			versioner: envoy_cache.SnapshotAutoVersioner{UUID: core.NewUUID},
			generator: NewSnapshotGenerator(readOnlyResourceManager, config),
		},
	}
}

func (t *tracker) OnHealthCheckRequest(streamID xds.StreamID, req *envoy_service_health.HealthCheckRequest) error {
	id, err := xds.ParseProxyIdFromString(req.GetNode().GetId())
	if err != nil {
		hdsServerLog.Error(err, "failed to parse Dataplane Id out of HealthCheckRequest", "streamid", streamID, "req", req)
		return nil
	}

	dataplaneKey := core_model.ResourceKey{Mesh: id.Mesh, Name: id.Name}

	t.Lock()
	defer t.Unlock()

	streams := t.dpStreams[dataplaneKey]
	if streams.activeStreams == nil {
		streams.activeStreams = map[xds.StreamID]bool{}
	}
	streams.activeStreams[streamID] = true

	if streams.watchdogCancel == nil { // watchdog was not started yet
		stopCh := make(chan struct{})
		streams.watchdogCancel = func() {
			close(stopCh)
		}
		// kick off watchdog for that Dataplane
		go t.newWatchdog(req.Node).Start(stopCh)
		hdsServerLog.V(1).Info("started Watchdog for a Dataplane", "streamid", streamID, "proxyId", id, "dataplaneKey", dataplaneKey)
	}
	t.dpStreams[dataplaneKey] = streams
	t.streamsAssociation[streamID] = dataplaneKey
	return nil
}

func (t *tracker) newWatchdog(node *envoy_core.Node) watchdog.Watchdog {
	return &watchdog.SimpleWatchdog{
		NewTicker: func() *time.Ticker {
			return time.NewTicker(t.config.RefreshInterval)
		},
		OnTick: func() error {
			return t.reconciler.Reconcile(node)
		},
		OnError: func(err error) {
			hdsServerLog.Error(err, "OnTick() failed")
		},
		OnStop: func() {
			if err := t.reconciler.Clear(node); err != nil {
				hdsServerLog.Error(err, "OnTick() failed")
			}
		},
	}
}

func (t *tracker) OnEndpointHealthResponse(streamID xds.StreamID, resp *envoy_service_health.EndpointHealthResponse) error {
	hdsServerLog.Info("on-endpoint-health-response", "streamID", streamID, "resp", resp)
	for _, clusterHealth := range resp.GetClusterEndpointsHealth() {
		if len(clusterHealth.LocalityEndpointsHealth) == 0 {
			continue
		}
		if len(clusterHealth.LocalityEndpointsHealth[0].EndpointsHealth) == 0 {
			continue
		}
		status := clusterHealth.LocalityEndpointsHealth[0].EndpointsHealth[0].HealthStatus
		health := status == envoy_core.HealthStatus_HEALTHY || status == envoy_core.HealthStatus_UNKNOWN
		hdsServerLog.Info("on-endpoint-health-response", "health", health)
		port, err := names.GetPortForLocalClusterName(clusterHealth.ClusterName)
		if err != nil {
			return err
		}
		if err := t.updateDataplane(streamID, port, health); err != nil {
			return err
		}
	}
	return nil
}

func (t *tracker) OnStreamClosed(streamID xds.StreamID) {
	t.Lock()
	defer t.Unlock()

	dp, hasAssociation := t.streamsAssociation[streamID]
	if hasAssociation {
		delete(t.streamsAssociation, streamID)

		streams := t.dpStreams[dp]
		delete(streams.activeStreams, streamID)
		if len(streams.activeStreams) == 0 { // no stream is active, cancel watchdog
			if streams.watchdogCancel != nil {
				streams.watchdogCancel()
			}
			delete(t.dpStreams, dp)
		}
	}
}

func (t *tracker) updateDataplane(streamID xds.StreamID, port uint32, ready bool) error {
	t.RLock()
	defer t.RUnlock()
	dataplaneKey, hasAssociation := t.streamsAssociation[streamID]
	if !hasAssociation {
		return errors.Errorf("no proxy for streamID = %d", streamID)
	}

	dp := mesh.NewDataplaneResource()
	if err := t.resourceManager.Get(context.Background(), dp, store.GetBy(dataplaneKey)); err != nil {
		return err
	}

	changed := false
	for _, inbound := range dp.Spec.Networking.Inbound {
		intf := dp.Spec.Networking.ToInboundInterface(inbound)
		if intf.WorkloadPort != port {
			continue
		}
		if inbound.Health == nil || inbound.Health.Ready != ready {
			inbound.Health = &v1alpha1.Dataplane_Networking_Inbound_Health{
				Ready: ready,
			}
			changed = true
		}
	}

	if changed {
		return t.resourceManager.Update(context.Background(), dp)
	}

	return nil
}
