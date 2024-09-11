package tracker

import (
	"context"
	"sync"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_service_health "github.com/envoyproxy/go-control-plane/envoy/service/health/v3"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	dp_server "github.com/kumahq/kuma/pkg/config/dp-server"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/core/xds"
	hds_callbacks "github.com/kumahq/kuma/pkg/hds/callbacks"
	hds_metrics "github.com/kumahq/kuma/pkg/hds/metrics"
	"github.com/kumahq/kuma/pkg/util/watchdog"
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
	log             logr.Logger
	metrics         *hds_metrics.Metrics

	sync.RWMutex       // protects access to the fields below
	streamsAssociation map[xds.StreamID]core_model.ResourceKey
	dpStreams          map[core_model.ResourceKey]streams
}

func NewCallbacks(
	log logr.Logger,
	resourceManager manager.ResourceManager,
	readOnlyResourceManager manager.ReadOnlyResourceManager,
	cache util_xds_v3.SnapshotCache,
	config *dp_server.HdsConfig,
	hasher util_xds_v3.NodeHash,
	metrics *hds_metrics.Metrics,
	defaultAdminPort uint32,
) hds_callbacks.Callbacks {
	return &tracker{
		resourceManager:    resourceManager,
		streamsAssociation: map[xds.StreamID]core_model.ResourceKey{},
		dpStreams:          map[core_model.ResourceKey]streams{},
		config:             config,
		log:                log,
		metrics:            metrics,
		reconciler: &reconciler{
			cache:     cache,
			hasher:    hasher,
			versioner: util_xds_v3.SnapshotAutoVersioner{UUID: core.NewUUID},
			generator: NewSnapshotGenerator(readOnlyResourceManager, config, defaultAdminPort),
		},
	}
}

func (t *tracker) OnStreamOpen(ctx context.Context, streamID int64) error {
	t.metrics.StreamsActiveInc()
	return nil
}

func (t *tracker) OnStreamClosed(streamID xds.StreamID) {
	t.metrics.StreamsActiveDec()

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

func (t *tracker) OnHealthCheckRequest(streamID xds.StreamID, req *envoy_service_health.HealthCheckRequest) error {
	t.metrics.RequestsReceivedMetric.Inc()

	proxyId, err := xds.ParseProxyIdFromString(req.GetNode().GetId())
	if err != nil {
		t.log.Error(err, "failed to parse Dataplane Id out of HealthCheckRequest", "streamid", streamID, "req", req)
		return nil
	}

	dataplaneKey := proxyId.ToResourceKey()

	t.Lock()
	defer t.Unlock()

	streams := t.dpStreams[dataplaneKey]
	if streams.activeStreams == nil {
		streams.activeStreams = map[xds.StreamID]bool{}
	}
	streams.activeStreams[streamID] = true

	if streams.watchdogCancel == nil { // watchdog was not started yet
		ctx, cancel := context.WithCancel(context.Background())
		streams.watchdogCancel = cancel
		// kick off watchdog for that Dataplane
		go t.newWatchdog(req.Node).Start(ctx)
		t.log.V(1).Info("started Watchdog for a Dataplane", "streamid", streamID, "proxyId", proxyId, "dataplaneKey", dataplaneKey)
	}
	t.dpStreams[dataplaneKey] = streams
	t.streamsAssociation[streamID] = dataplaneKey
	return nil
}

func (t *tracker) newWatchdog(node *envoy_core.Node) util_xds_v3.Watchdog {
	return &watchdog.SimpleWatchdog{
		NewTicker: func() *time.Ticker {
			return time.NewTicker(t.config.RefreshInterval.Duration)
		},
		OnTick: func(ctx context.Context) error {
			start := core.Now()
			defer func() {
				t.metrics.HdsGenerations.Observe(float64(core.Now().Sub(start).Milliseconds()))
			}()
			return t.reconciler.Reconcile(ctx, node)
		},
		OnError: func(err error) {
			t.metrics.HdsGenerationsErrors.Inc()
			t.log.Error(err, "OnTick() failed")
		},
		OnStop: func() {
			if err := t.reconciler.Clear(node); err != nil {
				t.log.Error(err, "OnTick() failed")
			}
		},
	}
}

func (t *tracker) OnEndpointHealthResponse(streamID xds.StreamID, resp *envoy_service_health.EndpointHealthResponse) error {
	t.metrics.ResponsesReceivedMetric.Inc()

	healthMap := map[uint32]bool{}
	envoyHealth := true // if there is no Envoy HC, assume it's healthy

	for _, clusterHealth := range resp.GetClusterEndpointsHealth() {
		if len(clusterHealth.LocalityEndpointsHealth) == 0 {
			continue
		}
		if len(clusterHealth.LocalityEndpointsHealth[0].EndpointsHealth) == 0 {
			continue
		}
		status := clusterHealth.LocalityEndpointsHealth[0].EndpointsHealth[0].HealthStatus
		health := status == envoy_core.HealthStatus_HEALTHY || status == envoy_core.HealthStatus_UNKNOWN

		if clusterHealth.ClusterName == names.GetEnvoyAdminClusterName() {
			envoyHealth = health
		} else {
			port, err := names.GetPortForLocalClusterName(clusterHealth.ClusterName)
			if err != nil {
				return err
			}
			healthMap[port] = health
		}
	}
	if err := t.updateDataplane(streamID, healthMap, envoyHealth); err != nil {
		return err
	}
	return nil
}

func (t *tracker) updateDataplane(streamID xds.StreamID, healthMap map[uint32]bool, envoyHealth bool) error {
	ctx := user.Ctx(context.Background(), user.ControlPlane)
	t.RLock()
	defer t.RUnlock()
	dataplaneKey, hasAssociation := t.streamsAssociation[streamID]
	if !hasAssociation {
		return errors.Errorf("no proxy for streamID = %d", streamID)
	}

	dp := mesh.NewDataplaneResource()
	if err := t.resourceManager.Get(ctx, dp, store.GetBy(dataplaneKey)); err != nil {
		return err
	}

	changed := false
	for _, inbound := range dp.Spec.Networking.Inbound {
		intf := dp.Spec.Networking.ToInboundInterface(inbound)
		workloadHealth, exist := healthMap[intf.WorkloadPort]
		if exist {
			workloadHealth = workloadHealth && envoyHealth
		} else {
			workloadHealth = envoyHealth
		}
		if workloadHealth && inbound.State == mesh_proto.Dataplane_Networking_Inbound_NotReady {
			inbound.State = mesh_proto.Dataplane_Networking_Inbound_Ready
			// write health for backwards compatibility with Kuma 2.5 and older
			inbound.Health = &mesh_proto.Dataplane_Networking_Inbound_Health{
				Ready: true,
			}
			changed = true
		} else if !workloadHealth && inbound.State == mesh_proto.Dataplane_Networking_Inbound_Ready {
			inbound.State = mesh_proto.Dataplane_Networking_Inbound_NotReady
			// write health for backwards compatibility with Kuma 2.5 and older
			inbound.Health = &mesh_proto.Dataplane_Networking_Inbound_Health{
				Ready: false,
			}
			changed = true
		}
	}

	if changed {
		t.log.V(1).Info("status updated", "dataplaneKey", dataplaneKey)
		return t.resourceManager.Update(ctx, dp)
	}

	return nil
}
