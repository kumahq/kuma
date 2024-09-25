package v3

import (
	"context"
	"sync"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/xds"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

var warmingForcerLog = xdsServerLog.WithName("warming-forcer")

// The problem
//
//	When you send Cluster of type EDS to Envoy, it updates the config, but Cluster is marked as warming and is not used until you send EDS request.
//	https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol#xds-protocol-resource-warming
//	The main problem is when you update the Cluster itself (for example enable mTLS or change a cluster property via ProxyTemplate)
//	If you don't send EDS request after that (no endpoint has changed) then the cluster is stuck in warming state indefinitely.
//
// The solution
//
//	The easiest solution would be to just set a new version of endpoints also when cluster changes (in pkg/xds/server/SnapshotReconciler) . The problem is that go-control-plane does not support resource ordering
//	* https://github.com/envoyproxy/go-control-plane/issues/59
//	* https://github.com/envoyproxy/go-control-plane/issues/218
//	* https://github.com/envoyproxy/go-control-plane/issues/235
//	Therefore even if we were to set a new version of EDS + CDS on one snapshot, there is no guarantee that EDS will be delivered after CDS.
//
//	The alternative solution is this based on a callbacks.
//	Nonce is a sequence indicator of a sent DiscoveryResponse on a stream. We use ADS therefore every single DiscoveryResponse regardless of a type is sent with incremented nonce.
//	Typical sequence of CDS + EDS looks like this:
//	1) Envoy sends DiscoveryRequest [type=CDS, version="", nonce=""] // ask for the clusters
//	2) Kuma sends DiscoveryResponse [type=CDS, version="UUID-1", nonce="1"] // response with clusters
//	3) Envoy sends DiscoveryRequest [type=EDS, version="", nonce=""] // ask for the endpoints
//	4) Envoy sends DiscoveryRequest [type=CDS, version="UUID-1", nonce="1"] // confirmations that it received clusters (ACK)
//	5) Kuma sends DiscoveryResponse [type=EDS, version="UUID-2", nonce="2"] // response with endpoints
//	6) Envoy sends DiscoveryRequest [type=EDS, version="UUID-2", nonce="2"] // confirmations that it received endpoints (ACK)
//
//	Then if we send a Cluster update (continuing the flow above)
//	7) Kuma sends DiscoveryResponse [type=CDS, version="UUID-2", nonce="3"] // response with cluster update
//	8) Envoy sends DiscoveryRequest [type=CDS, version="UUID-2", nonce="3"] // ACK
//	9) Envoy sends DiscoveryRequest [type=EDS, version="UUID-2", nonce="2"] // Envoy sends a request which looks like the second ACK for the previous endpoints
//
//	Updated Cluster is now in warming state until we send DiscoveryResponse with EDS.
//	We could sent the same DiscoveryResponse again: DiscoveryRequest [type=EDS, version="UUID-2", nonce="2"], but there is no API in go-control-plane to do it.
//	Instead we set a new version of the Endpoints to force a new EDS exchange:
//	10) Kuma sends DiscoveryResponse [type=EDS, version="UUID-3", nonce="3"] // triggered because we set snapshot with a new version
//	11) Envoy sends DiscoveryRequest [type=EDS, version="UUID-3", nonce="3"] // ACK
//	After this exchange, cluster is now out of the warming state.
//
// The same problem is with Listeners and Routes (change of the Listener that uses RDS requires RDS DiscoveryResponse), but since we don't use RDS now, the implementation is for EDS only.
// More reading of how Envoy is trying to solve it https://github.com/envoyproxy/envoy/issues/13009
type resourceWarmingForcer struct {
	util_xds_v3.NoopCallbacks
	cache  envoy_cache.SnapshotCache
	hasher envoy_cache.NodeHash

	sync.Mutex
	lastEndpointNonces map[xds.StreamID]string
	nodeIDs            map[xds.StreamID]string
	snapshotCacheMux   *sync.Mutex
}

func newResourceWarmingForcer(cache envoy_cache.SnapshotCache, hasher envoy_cache.NodeHash, snapshotCacheMux *sync.Mutex) *resourceWarmingForcer {
	return &resourceWarmingForcer{
		cache:              cache,
		hasher:             hasher,
		lastEndpointNonces: map[xds.StreamID]string{},
		nodeIDs:            map[xds.StreamID]string{},
		snapshotCacheMux:   snapshotCacheMux,
	}
}

func (r *resourceWarmingForcer) OnStreamClosed(streamID int64, _ *envoy_core.Node) {
	r.Lock()
	defer r.Unlock()
	delete(r.lastEndpointNonces, streamID)
	delete(r.nodeIDs, streamID)
}

func (r *resourceWarmingForcer) OnStreamRequest(streamID xds.StreamID, request *envoy_sd.DiscoveryRequest) error {
	if request.TypeUrl != envoy_resource.EndpointType {
		return nil // we force Cluster warming only on receiving the same EDS Discovery Request
	}
	if request.ResponseNonce == "" {
		return nil // initial request, no need to force warming
	}
	if request.ErrorDetail != nil {
		return nil // we only care about ACKs, otherwise we can get 2 Nonces with multiple NACKs
	}

	r.Lock()
	lastEndpointNonce := r.lastEndpointNonces[streamID]
	r.lastEndpointNonces[streamID] = request.ResponseNonce
	nodeID := r.nodeIDs[streamID]
	if nodeID == "" {
		nodeID = r.hasher.ID(request.Node) // request.Node can be set only on first request therefore we need to save it
		r.nodeIDs[streamID] = nodeID
	}
	r.Unlock()

	if lastEndpointNonce == request.ResponseNonce {
		warmingForcerLog.V(1).Info("received second Endpoint DiscoveryRequest with same Nonce. Forcing new version of Endpoints to warm the Cluster")
		if err := r.forceNewEndpointsVersion(nodeID); err != nil {
			warmingForcerLog.Error(err, "could not force cluster warming")
		}
	}
	return nil
}

func (r *resourceWarmingForcer) forceNewEndpointsVersion(nodeID string) error {
	r.snapshotCacheMux.Lock()
	defer r.snapshotCacheMux.Unlock()
	snapshot, err := r.cache.GetSnapshot(nodeID)
	if err != nil {
		return nil // GetSnapshot returns an error if there is no snapshot. We don't need to force on a new snapshot
	}
	cacheSnapshot, ok := snapshot.(*envoy_cache.Snapshot)
	if !ok {
		return errors.New("couldn't convert snapshot from cache to envoy Snapshot")
	}
	endpoints := cacheSnapshot.Resources[types.Endpoint]
	endpoints.Version = core.NewUUID()
	cacheSnapshot.Resources[types.Endpoint] = endpoints
	if err := r.cache.SetSnapshot(context.TODO(), nodeID, snapshot); err != nil {
		return errors.Wrap(err, "could not set snapshot")
	}
	return nil
}

var _ envoy_xds.Callbacks = &resourceWarmingForcer{}
