package reconcile

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache"

	mads_cache "github.com/Kong/kuma/pkg/mads/cache"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
)

func NewSnapshotter() Snapshotter {
	return &snapshotter{}
}

type snapshotter struct {
}

func (_ snapshotter) Snapshot(*envoy_core.Node) (util_xds.Snapshot, error) {
	snapshot := mads_cache.NewSnapshot("", []envoy_cache.Resource{})
	return &snapshot, nil
}
