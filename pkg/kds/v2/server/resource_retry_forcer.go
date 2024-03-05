package server

import (
	"sync"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/xds"
	cache_v2 "github.com/kumahq/kuma/pkg/kds/v2/cache"
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
//
// In case of NACK, we invalidate the version in VersionMap for all the resources of a specific
// type to respond as soon as possible. Setting a different version of the snapshot is not possible
// in this case because the delta server calculates the hash of the object and stores them in VersionMap.
// The change of version in the map triggers update to be sent.
type kdsRetryForcer struct {
	util_xds_v3.NoopCallbacks
	hasher  envoy_cache.NodeHash
	cache   envoy_cache.SnapshotCache
	log     logr.Logger
	nodeIDs map[xds.StreamID]string

	sync.Mutex
}

func newKdsRetryForcer(log logr.Logger, cache envoy_cache.SnapshotCache, hasher envoy_cache.NodeHash) *kdsRetryForcer {
	return &kdsRetryForcer{
		cache:   cache,
		hasher:  hasher,
		log:     log,
		nodeIDs: map[xds.StreamID]string{},
	}
}

var _ envoy_xds.Callbacks = &kdsRetryForcer{}

func (r *kdsRetryForcer) OnDeltaStreamClosed(streamID int64, _ *envoy_core.Node) {
	r.Lock()
	defer r.Unlock()
	delete(r.nodeIDs, streamID)
}

func (r *kdsRetryForcer) OnStreamDeltaRequest(streamID xds.StreamID, request *envoy_sd.DeltaDiscoveryRequest) error {
	if request.ResponseNonce == "" {
		return nil // initial request, no need to force warming
	}

	if request.ErrorDetail == nil {
		return nil // not NACK, no need to retry
	}

	r.Lock()
	nodeID := r.nodeIDs[streamID]
	if nodeID == "" {
		nodeID = r.hasher.ID(request.Node) // request.Node can be set only on first request therefore we need to save it
		r.nodeIDs[streamID] = nodeID
	}
	r.Unlock()
	r.log.Info("received NACK", "nodeID", nodeID, "type", request.TypeUrl, "err", request.GetErrorDetail().GetMessage())
	snapshot, err := r.cache.GetSnapshot(nodeID)
	if err != nil {
		return nil // GetSnapshot returns an error if there is no snapshot. We don't need to force on a new snapshot
	}
	cacheSnapshot, ok := snapshot.(*cache_v2.Snapshot)
	if !ok {
		return errors.New("couldn't convert snapshot from cache to envoy Snapshot")
	}
	for resourceName := range cacheSnapshot.VersionMap[model.ResourceType(request.TypeUrl)] {
		cacheSnapshot.VersionMap[model.ResourceType(request.TypeUrl)][resourceName] = ""
	}

	r.log.V(1).Info("forced the new verion of resources", "nodeID", nodeID, "type", request.TypeUrl)
	return nil
}
