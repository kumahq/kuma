package hds

import (
	"context"
	"sync"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_service_health "github.com/envoyproxy/go-control-plane/envoy/service/health/v3"
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

type tracker struct {
	resourceManager manager.ResourceManager
	config          *dp_server.HdsConfig
	reconciler      *reconciler

	proxyByStreamIDMux sync.RWMutex
	proxyByStreamID    map[int64]*envoy_core.Node
}

func NewTracker(
	resourceManager manager.ResourceManager,
	readOnlyResourceManager manager.ReadOnlyResourceManager,
	cache util_xds_v3.SnapshotCache,
	config *dp_server.HdsConfig,
) Callbacks {
	return &tracker{
		resourceManager: resourceManager,
		proxyByStreamID: map[int64]*envoy_core.Node{},
		config:          config,
		reconciler: &reconciler{
			cache:     cache,
			hasher:    &hasher{},
			versioner: envoy_cache.SnapshotAutoVersioner{UUID: core.NewUUID},
			generator: NewSnapshotGenerator(readOnlyResourceManager, config),
		},
	}
}

func (t *tracker) OnHealthCheckRequest(streamID int64, req *envoy_service_health.HealthCheckRequest) error {
	t.addProxy(streamID, req.Node)
	return t.reconciler.Reconcile(req.Node)
}

func (t *tracker) OnEndpointHealthResponse(streamID int64, resp *envoy_service_health.EndpointHealthResponse) error {
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

	node, exist := t.getProxy(streamID)
	if !exist {
		return errors.Errorf("no proxy for streamID = %d", streamID)
	}

	// we are using the fact that this method OnEndpointHealthResponse will be called
	// periodically with the interval we provided in HealthCheckSpecifier
	return t.reconciler.Reconcile(node)
}

func (t *tracker) OnStreamClosed(streamID int64) {
	t.deleteProxy(streamID)
}

func (t *tracker) getProxy(streamID int64) (*envoy_core.Node, bool) {
	t.proxyByStreamIDMux.RLock()
	defer t.proxyByStreamIDMux.RUnlock()

	node, exist := t.proxyByStreamID[streamID]
	return node, exist
}

func (t *tracker) addProxy(streamID int64, node *envoy_core.Node) {
	t.proxyByStreamIDMux.Lock()
	defer t.proxyByStreamIDMux.Unlock()

	t.proxyByStreamID[streamID] = node
}

func (t *tracker) deleteProxy(streamID int64) {
	t.proxyByStreamIDMux.Lock()
	defer t.proxyByStreamIDMux.Unlock()

	delete(t.proxyByStreamID, streamID)
}

func (t *tracker) updateDataplane(streamID int64, port uint32, ready bool) error {
	node, exist := t.getProxy(streamID)
	if !exist {
		return errors.Errorf("no proxy for streamID = %d", streamID)
	}
	proxyId, err := xds.ParseProxyIdFromString(node.Id)
	if err != nil {
		return err
	}
	dp := mesh.NewDataplaneResource()
	if err := t.resourceManager.Get(context.Background(), dp, store.GetByKey(proxyId.Name, proxyId.Mesh)); err != nil {
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
