package server

import (
	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"

	"github.com/kumahq/kuma/v3/pkg/kds/util"
	util_xds_v3 "github.com/kumahq/kuma/v3/pkg/util/xds/v3"
)

// streamScopeCallbacks stamps the stream ID on the node of every request, so the
// snapshot cache key (see Hasher) is unique per stream. Two streams of the same
// zone (leader handover) must not share a snapshot: the one that closes clears it.
type streamScopeCallbacks struct {
	util_xds_v3.NoopCallbacks
}

func (*streamScopeCallbacks) OnStreamDeltaRequest(streamID int64, request *envoy_sd.DeltaDiscoveryRequest) error {
	// A request without a node gets the first request's node from the delta
	// server after the callbacks run, and that one is already stamped.
	if request.Node != nil {
		util.FillStreamMetadata(streamID, request.Node)
	}
	return nil
}
