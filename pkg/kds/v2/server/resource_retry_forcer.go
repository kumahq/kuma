package server

import (
	"sync"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/go-logr/logr"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/kds/v2/util"
	"github.com/kumahq/kuma/pkg/log"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

// The Problem:
//
// KDS is utilized to synchronize resources between the zone and the global and vice versa.
// When either of them runs on Kubernetes, webhooks are employed to validate the resource's
// correctness or the owner reference. To do so, the control-plane exposes APIs through the
// Kubernetes service 'kuma-control-plane.kuma-system.svc:443.' However, the service may be
// available before the control-plane is ready to handle requests, which results in failed
// requests and errors.
//
// The SOTW server, in the case of NACK, was sending the response again, which was effective
// for the above-described behavior. However, Delta xDS does not have that behavior, and in
// the case of NACK, it does not respond with the same message, which is the correct behavior
// based on the xDS specification.
//
// The Solution:
// In case of NACK, we notify Reconciler to force set a new version for all resources of a NACK type
// Because KDS do not send NACK for specific resource (it's always ResourceNamesSubscribe=*) we have to do this for all resources.
// We cannot invalidate existing VersionMap of a Snapshot, because versions are also stored in StreamState.
// The exchange looks like this:
// 1) We send a resource with version="xyz". "xyz" is computed by Snapshot#ConstructVersionMap() and it's SHA of a resource.
// 2) KDS client responds with NACK for "xyz"
// 3) We send the same resource with version="" (set manually after ConstructVersionMap call)
// 4) KDS client responds with NACK for ""
// We loop those steps with a backoff until the client recovers. In this case we either see
// 1) ACK for "xyz"
// 2) ACK for "". In which case we will also send "xyz" and receive ACK for "xyz"
// There is no easy way to force resource sending without changing its version.
// We cannot simply invalidate existing snapshot because versions are also set in StreamState
type kdsRetryForcer struct {
	util_xds_v3.NoopCallbacks
	forceFn       func(*envoy_core.Node, model.ResourceType)
	log           logr.Logger
	nodes         map[xds.StreamID]*envoy_core.Node
	backoff       time.Duration
	emitter       events.Emitter
	hasher        envoy_cache.NodeHash
	streamToDelay map[xds.StreamID]bool

	sync.Mutex
}

func newKdsRetryForcer(
	log logr.Logger,
	forceFn func(*envoy_core.Node, model.ResourceType),
	backoff time.Duration,
	emitter events.Emitter,
	hasher envoy_cache.NodeHash,
) *kdsRetryForcer {
	return &kdsRetryForcer{
		forceFn:       forceFn,
		log:           log,
		nodes:         map[xds.StreamID]*envoy_core.Node{},
		backoff:       backoff,
		emitter:       emitter,
		hasher:        hasher,
		streamToDelay: map[int64]bool{},
	}
}

var _ envoy_xds.Callbacks = &kdsRetryForcer{}

func (r *kdsRetryForcer) OnDeltaStreamClosed(streamID int64, _ *envoy_core.Node) {
	r.Lock()
	defer r.Unlock()
	delete(r.nodes, streamID)
}

func (r *kdsRetryForcer) OnStreamDeltaRequest(streamID xds.StreamID, request *envoy_sd.DeltaDiscoveryRequest) error {
	if request.ResponseNonce == "" {
		return nil // initial request, no need to force warming
	}

	if request.ErrorDetail == nil {
		return nil // not NACK, no need to retry
	}

	r.Lock()
	node, ok := r.nodes[streamID]
	if !ok {
		node = request.Node
		r.nodes[streamID] = node
		// store information about NACK resources, to delete force retries once ACK
	}
	if _, found := r.streamToDelay[streamID]; !found {
		r.streamToDelay[streamID] = true
	}
	r.Unlock()
	logger := r.log.WithValues(
		"nodeID", r.hasher.ID(request.Node),
		"type", request.TypeUrl,
	)
	if tenantID, found := util.TenantFromMetadata(node); found {
		logger = logger.WithValues(log.TenantLoggerKey, tenantID)
	}
	logger.Info("received NACK, will retry", "err", request.GetErrorDetail().GetMessage(), "backoff", r.backoff)
	r.forceFn(node, model.ResourceType(request.TypeUrl))
	r.emitter.Send(events.TriggerKDSResyncEvent{
		Type:   model.ResourceType(request.TypeUrl),
		NodeID: node.Id,
	})
	logger.V(1).Info("forced the new version of resources")
	return nil
}

func (r *kdsRetryForcer) OnStreamDeltaResponse(streamID int64, req *envoy_sd.DeltaDiscoveryRequest, resp *envoy_sd.DeltaDiscoveryResponse) {
	r.Lock()
	_, found := r.streamToDelay[streamID]
	r.Unlock()
	if found {
		time.Sleep(r.backoff)
		r.Lock()
		delete(r.streamToDelay, streamID)
		r.Unlock()
	}
}
